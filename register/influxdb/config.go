package influxdb

type Config struct {
	URL             string
	Token           string
	Organization    string
	IntervalSeconds int
	ClearValue      bool
	RetryAttempts   int
	BackupDir       string
}
