package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

var (
	LogDir = "./config/logs/query_log"
)

func SetupLogger() logger.Interface {
	if err := os.MkdirAll(LogDir, 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	currentMonth := strings.ToLower(time.Now().Format("January"))
	logFileName := fmt.Sprintf("%s_query.log", currentMonth)
	logPath := filepath.Join(LogDir, logFileName)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	return logger.New(
		log.New(logFile, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
		},
	)
}
