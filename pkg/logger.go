package pkg

import (
	"fmt"
	"time"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any) // zerolog는 지원하지 안ㅇ흠
	Error(msg string, args ...any)
}

var (
	DefaultLogger Logger = defLogger()
)

func defLogger() Logger {
	return &defaultLogger{}
}

type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, args ...any) {
	fmt.Printf("%s:[DEBUG] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(msg, args...))
	// Implement debug logging logic
}
func (l *defaultLogger) Error(msg string, args ...any) {
	fmt.Printf("%s:[ERROR] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(msg, args...))
	// Implement error logging logic
}
func (l *defaultLogger) Info(msg string, args ...any) {
	fmt.Printf("%s:[INFO] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(msg, args...))
	// Implement info logging logic
}
func (l *defaultLogger) Warn(msg string, args ...any) {
	fmt.Printf("%s:[WARN] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(msg, args...))
	// Implement warn logging logic
}
