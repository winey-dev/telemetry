package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/winey-dev/telemetry/metric"
)

var (
	MemoryUsage = metric.NewItem(metric.ItemOpts{
		Category:    "system",
		SubCategory: "resource",
		ItemName:    "memory_usage",
		Description: "Tracks the memory usage of the system.",
		ConstraintTags: metric.ConstraintTags{
			TagNames:  []string{"env", "version"},
			TagValues: []string{"production", "v1.0"},
		},
	})
	CPUUsage = metric.NewItem(metric.ItemOpts{
		Category:    "system",
		SubCategory: "resource",
		ItemName:    "cpu_usage",
		Description: "Tracks the CPU usage of the system.",
		ConstraintTags: metric.ConstraintTags{
			TagNames:  []string{"env", "version"},
			TagValues: []string{"production", "v1.0"},
		},
	})
	DiskUsage = metric.NewItemVec(metric.ItemOpts{
		Category:    "system",
		SubCategory: "disk",
		ItemName:    "disk_usage",
		Description: "Tracks the disk usage of the system.",
		ConstraintTags: metric.ConstraintTags{
			TagNames:  []string{"env", "version"},
			TagValues: []string{"production", "v1.0"},
		},
	}, "disk_path")

	NetworkTx = metric.NewItemVec(metric.ItemOpts{
		Category:    "system",
		SubCategory: "network",
		ItemName:    "transmitted_traffic",
		Description: "Tracks the network traffic of the system.",
		ConstraintTags: metric.ConstraintTags{
			TagNames:  []string{"env", "version"},
			TagValues: []string{"production", "v1.0"},
		},
	}, "interface")
	NetworkRx = metric.NewItemVec(metric.ItemOpts{
		Category:    "system",
		SubCategory: "network",
		ItemName:    "received_traffic",
		Description: "Tracks the network traffic of the system.",
		ConstraintTags: metric.ConstraintTags{
			TagNames:  []string{"env", "version"},
			TagValues: []string{"production", "v1.0"},
		},
	}, "interface")
)

func getMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// m.Alloc is bytes allocated and still in use
	//fmt.Printf("Alloc = %v MiB\n", m.Alloc)
	MemoryUsage.Add(float64(m.Alloc))
}

func getCPUUsage() {
	percentages, _ := cpu.Percent(0, false)
	//fmt.Printf("CPU Usage: %v%%\n", percentages[0])
	CPUUsage.Add(percentages[0])

}

func getDiskUsage() {
	partitions, err := disk.Partitions(false)
	if err != nil {
		fmt.Println("Error getting disk partitions:", err)
		return
	}
	for _, p := range partitions {
		usageStat, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		//fmt.Printf("Disk Usage for %s: %v%%\n", p.Mountpoint, usageStat.Used)
		DiskUsage.WithTagValues(p.Mountpoint).Add(float64(usageStat.Used))
	}
}

func getNetworkUsage() {
	ioCounters, err := net.IOCounters(true)
	if err != nil {
		fmt.Println("Error getting network IO counters:", err)
		return
	}
	for _, io := range ioCounters {
		//fmt.Printf("Network Interface: %s, Bytes Sent: %d, Bytes Received: %d\n", io.Name, io.BytesSent, io.BytesRecv)
		NetworkTx.WithTagValues(io.Name).Add(float64(io.BytesSent))
		NetworkRx.WithTagValues(io.Name).Add(float64(io.BytesRecv))
	}
}
func collect() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		getMemoryUsage()
		getCPUUsage()
		getDiskUsage()
		getNetworkUsage()
	}
}
