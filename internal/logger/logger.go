package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ryzmae/hyperatomic/internal/config"
)

type Logger struct {
	logFile *os.File
	mutex   sync.Mutex
}

const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
)

func NewLogger(cfg *config.Config) (*Logger, error) {
	logPath := cfg.Logging.LogFile

	logDir := os.Getenv("HOME") + "/.config/hyperatomic/"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{logFile: logFile}, nil
}

func (l *Logger) Log(level string, format string, args ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	message := fmt.Sprintf("%s [%s] %s\n", time.Now().Format(time.RFC3339), level, fmt.Sprintf(format, args...))

	if _, err := l.logFile.WriteString(message); err != nil {
		fmt.Println("Error writing to log file:", err)
	}
}

func (l *Logger) Close() {
	l.logFile.Close()
}

func (l *Logger) Info(format string, args ...any) error {
	l.Log(INFO, format, args...)
	return nil
}

func (l *Logger) Debug(format string, args ...any) error {
	l.Log(DEBUG, format, args...)
	return nil
}

func (l *Logger) Warn(format string, args ...any) error {
	l.Log(WARN, format, args...)
	return nil
}

func (l *Logger) Error(format string, args ...any) error {
	l.Log(ERROR, format, args...)
	return nil
}
