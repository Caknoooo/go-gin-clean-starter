package config

import (
	"fmt"
	"strconv"
)

type EmailConfig struct {
	Host         string
	Port         int
	SenderName   string
	AuthEmail    string
	AuthPassword string
}

var NewEmailConfig = func() (*EmailConfig, error) {
	port, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, err
	}

	config := &EmailConfig{
		Host:         getEnv("SMTP_HOST", ""),
		Port:         port,
		SenderName:   getEnv("SMTP_SENDER_NAME", ""),
		AuthEmail:    getEnv("SMTP_AUTH_EMAIL", ""),
		AuthPassword: getEnv("SMTP_AUTH_PASSWORD", ""),
	}

	// Validate required fields
	if config.Host == "" || config.AuthEmail == "" || config.AuthPassword == "" {
		return nil, fmt.Errorf("email configuration is incomplete")
	}

	return config, nil
}
