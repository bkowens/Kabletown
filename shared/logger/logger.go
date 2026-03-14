package logger

import (
	"log"
	"os"
	"sync"
)

// Logger provides structured logging for Kabletown services
type Logger struct {
	module string
	mu     sync.Mutex
}

// NewLogger creates a new logger instance
func NewLogger(module string) *Logger {
	return &Logger{module: module}
}

// Info logs an informational message
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	args := append([]interface{}{"[" + l.module + "] INFO", msg}, keysAndValues...)
	log.Println(args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	args := append([]interface{}{"[" + l.module + "] ERROR", msg}, keysAndValues...)
	log.Println(args...)
}

// Debug logs a debug message (conditionally based on env)
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		l.Info(msg, keysAndValues...)
	}
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	args := append([]interface{}{"[" + l.module + "] FATAL", msg}, keysAndValues...)
	log.Fatal(args...)
}
