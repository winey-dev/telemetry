package influxdb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	http2 "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/winey-dev/telemetry/pkg"
	"github.com/winey-dev/telemetry/register"
)

type agent struct {
	register.Registry
	client influxdb2.Client
	config *Config
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	logger pkg.Logger
}

func NewRegisterer(config *Config) (register.Agent, error) {
	client := influxdb2.NewClient(config.URL, config.Token)
	if client == nil {
		return nil, fmt.Errorf("failed to create InfluxDB client: %s", config.URL)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &agent{
		//client: client,
		config: config,
		ctx:    ctx,
		cancel: cancel,
		logger: pkg.DefaultLogger,
	}, nil
}

func (a *agent) Start() error {
	// Interval 시간에 맞춰서 Gather() 메서드를 호출
	// InfluxDB에 데이터를 쓰는 작업을 수행
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ticker := time.NewTicker(time.Duration(a.config.IntervalSeconds) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case now := <-ticker.C:
				fmt.Printf("Gathering metrics at %s\n", now.Format(time.RFC3339))
				if err := a.gather(now); err != nil {
					fmt.Printf("Error gathering metrics: %v\n", err)
					continue
				}
			case <-a.ctx.Done():
				return
			}
		}
	}()
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ticker := time.NewTicker(time.Duration(a.config.IntervalSeconds) * time.Second)
		defer ticker.Stop()
		// 주기적으로 batch backup Directory를 읽어서 write 작업을 수행
		// 파일명에 카테고리가 포함되어있음으로 해당 정보를 통해서 WriteAPI를 생성
		for {
			select {
			case <-ticker.C:
				a.record()
			case <-a.ctx.Done():
				return
			}
		}

	}()
	return nil
}

func (a *agent) Stop() {
	a.cancel()
	a.wg.Wait()
	a.client.Close()
}

func (a *agent) gather(now time.Time) error {
	bucket := NewBucket()

	metrics, err := a.Gather()

	if err != nil {
		return err
	}

	for _, metric := range metrics {
		bucket.Add(metric, now)
	}

	bucket.Summary(now)

	for bucketName, points := range bucket.items {
		writeAPI := a.client.WriteAPI(a.config.Organization, bucketName)
		failed := &failedCallback{
			category:      bucketName,
			now:           now,
			retryAttempts: uint(a.config.RetryAttempts),
			logger:        a.logger,
			backupDir:     a.config.BackupDir,
		}
		writeAPI.SetWriteFailedCallback(failed.callback)
		for _, point := range points {
			writeAPI.WritePoint(point)
		}
		writeAPI.Flush()
	}

	for _, metric := range metrics {
		if resetter, ok := metric.(interface{ Reset() }); ok {
			resetter.Reset()
		}
	}
	return nil
}

func (a *agent) record() {
	dirEntry, err := os.ReadDir(a.config.BackupDir)
	if err != nil {
		a.logger.Error("Failed to read backup directory(%s): %v", a.config.BackupDir, err)
		return
	}

	for _, entry := range dirEntry {
		if entry.IsDir() {
			continue // Skip directories
		}

		if entry.Name() == "." || entry.Name() == ".." {
			continue // Skip current and parent directory entries
		}

		fileName := entry.Name()
		category := strings.TrimSuffix(fileName, ".txt") // Assuming backup files are .txt
		filePath := a.config.BackupDir + "/" + fileName

		data, err := os.ReadFile(filePath)
		if err != nil {
			a.logger.Error("Failed to read backup file(%s): %v", filePath, err)
			continue
		}

		writeAPI := a.client.WriteAPI(a.config.Organization, category)
		failed := &failedCallback{
			category:      category,
			now:           time.Now(),
			retryAttempts: uint(a.config.RetryAttempts),
			logger:        a.logger,
			backupDir:     a.config.BackupDir,
		}
		writeAPI.SetWriteFailedCallback(failed.callback)

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if line == "" {
				continue // Skip empty lines
			}
			writeAPI.WriteRecord(line)
		}
		writeAPI.Flush()

		if err := os.Remove(filePath); err != nil {
			a.logger.Error("Failed to remove backup file(%s): %v", filePath, err)
		}
	}

}

type failedCallback struct {
	category      string
	now           time.Time
	retryAttempts uint
	backupDir     string
	logger        pkg.Logger
}

func (f *failedCallback) callback(batch string, err http2.Error, retryAttempts uint) bool {
	if retryAttempts < f.retryAttempts {
		return f.retryFailedWrites(batch, err, retryAttempts)
	}
	return f.fallBack(batch, err, retryAttempts)
}

func (f *failedCallback) retryFailedWrites(batch string, err http2.Error, retryAttempts uint) bool {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("InfluxDB write failed: %s:%v\n", f.category, f.now.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("- Error: %v\n", err))
	builder.WriteString(fmt.Sprintf("- Retry Attempts: (%d/%d)\n", retryAttempts, f.retryAttempts))
	builder.WriteString(fmt.Sprintf("- Batch: %s\n", batch))
	f.logger.Error(builder.String())
	return true // Return false to stop retrying
}

func (f *failedCallback) fallBack(batch string, err http2.Error, retryAttempts uint) bool {
	// batch 단위의 실패 처리 방식

	var _ = batch         // Use batch if needed for logging or processing
	var _ = err           // Use err if needed for logging or processing
	var _ = retryAttempts // Use retryAttempts if needed for logging or processing

	// 설정된 retryAttempts에 도달 했을 때 데이터를 내부 파일에 백업 저장
	// register에서는 주기적으로 데이터 위치를 읽어 batch 작업을 수행
	// batch 내용은 WriteAPI.WriteRecord 또는 WriteAPI.LineProtocol 메서드를 사용하여 기록하기 떄문에 별도로 Category Bucket위치를 알 수 없음
	// 따라서 파일명에 카테고리가 포함되어있어야함
	return false // Return false to stop retrying
}
