package register

import (
	"sync"

	"github.com/winey-dev/telemetry/metric"
)

type Agent interface {
	Registerer
	Start() error
	Stop()
}

type Registry struct {
	mtx        sync.RWMutex
	collectors []metric.Collector
}

type Registerer interface {
	Register(metric.Collector) error
	Registers(...metric.Collector) error
}

func (r *Registry) Register(collector metric.Collector) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.collectors = append(r.collectors, collector)
	return nil
}

func (r *Registry) Registers(collectors ...metric.Collector) error {
	for _, collector := range collectors {
		if err := r.Register(collector); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) Gather() ([]metric.Metric, error) {
	var metrics []metric.Metric
	metricChan := make(chan metric.Metric)

	go func() {
		for _, c := range r.collectors {
			c.Collect(metricChan)

		}
		close(metricChan)
	}()

	for metric := range metricChan {
		metrics = append(metrics, metric)
	}
	return metrics, nil
}
