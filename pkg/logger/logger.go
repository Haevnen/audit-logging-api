package logger

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

var (
	logger  *log.Logger
	logOnce sync.Once
	logMux  sync.Mutex
)

func GetLogger() *log.Logger {
	logMux.Lock()
	defer logMux.Unlock()
	return logger
}

// GetLogger returns the global logger
func InitLogger(logLevel string, logFileName string) *log.Logger {
	logMux.Lock()
	defer logMux.Unlock()

	logOnce.Do(func() {
		logger = log.New()
		logger.SetFormatter(&log.TextFormatter{})

		if logLevel == "debug" {
			logger.SetLevel(log.DebugLevel)
		} else {
			logger.SetLevel(log.InfoLevel)
		}

		file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			logger.SetOutput(file)
		} else {
			logger.Info("Failed to log to file, using default stderr", err)
		}
	})

	return logger
}

func WithFields(fields map[string]interface{}) *log.Entry {
	return GetLogger().WithFields(fields)
}
