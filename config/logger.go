package config

import (
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

const (
	LOG_DIR = "./config/logs/query_log"
)

func SetupLogger() logger.Interface {
	err := os.MkdirAll(LOG_DIR, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	currentMonth := time.Now().Format("January")
	currentMonth = strings.ToLower(currentMonth)
	logFileName := currentMonth + "_query.log"

	logFile, err := os.OpenFile(LOG_DIR+"/"+logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	newLogger := logger.New(
		log.New(logFile, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)

	return newLogger
}
