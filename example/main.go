package main

import (
	"github.com/winey-dev/telemetry/register/influxdb"
)

func main() {

	register, err := influxdb.NewRegisterer(&influxdb.Config{
		URL:             "http://localhost:8086",
		Token:           "my-token",
		IntervalSeconds: 5,
		Organization:    "my-org",
	})
	if err != nil {
		panic(err)
	}

	register.Registers(MemoryUsage, CPUUsage, DiskUsage, NetworkTx, NetworkRx)

	register.Start()
	defer register.Stop()

	collect()
}
