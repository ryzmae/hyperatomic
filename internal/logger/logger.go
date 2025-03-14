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
	logChan chan string
	wg      sync.WaitGroup
}

const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
)

func NewLogger(cfg *config.Config) (*Logger, error) {
	file, err := os.OpenFile(cfg.Logging.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open log file, using stdout instead")
		file = os.Stdout
	}

	logger := &Logger{
		logFile: file,
		logChan: make(chan string, 1000),
	}

	logger.wg.Add(1)
	go logger.logWorker()

	return logger, nil
}

func (l *Logger) logWorker() {
	defer l.wg.Done()
	for msg := range l.logChan {
		_, _ = l.logFile.WriteString(msg)
	}
}

func shouldLog(configLevel, messageLevel string) bool {
	levels := map[string]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}

	configLevelValue, configExists := levels[configLevel]
	messageLevelValue, messageExists := levels[messageLevel]

	if !configExists || !messageExists {
		return true
	}

	return messageLevelValue >= configLevelValue
}

func (l *Logger) Log(level, format string, v ...interface{}) {
	cfg := config.GetConfig() // Fetch live config

	if !shouldLog(cfg.Logging.LogLevel, level) {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, fmt.Sprintf(format, v...))

	select {
	case l.logChan <- message:
	default:
		fmt.Println("Log channel full, dropping log:", message)
	}
}
