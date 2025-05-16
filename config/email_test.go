package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmailConfig(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name        string
		envVars     map[string]string
		wantConfig  *EmailConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "Success with all fields",
			envVars: map[string]string{
				"SMTP_HOST":          "smtp.example.com",
				"SMTP_PORT":          "587",
				"SMTP_SENDER_NAME":   "Test Sender",
				"SMTP_AUTH_EMAIL":    "user@example.com",
				"SMTP_AUTH_PASSWORD": "password123",
			},
			wantConfig: &EmailConfig{
				Host:         "smtp.example.com",
				Port:         587,
				SenderName:   "Test Sender",
				AuthEmail:    "user@example.com",
				AuthPassword: "password123",
			},
			wantErr: false,
		},
		{
			name: "Success with minimum required fields",
			envVars: map[string]string{
				"SMTP_HOST":          "smtp.example.com",
				"SMTP_AUTH_EMAIL":    "user@example.com",
				"SMTP_AUTH_PASSWORD": "password123",
			},
			wantConfig: &EmailConfig{
				Host:         "smtp.example.com",
				Port:         587, // default
				SenderName:   "",  // optional
				AuthEmail:    "user@example.com",
				AuthPassword: "password123",
			},
			wantErr: false,
		},
		{
			name: "Invalid port number",
			envVars: map[string]string{
				"SMTP_HOST":          "smtp.example.com",
				"SMTP_PORT":          "not-a-number",
				"SMTP_AUTH_EMAIL":    "user@example.com",
				"SMTP_AUTH_PASSWORD": "password123",
			},
			wantErr:     true,
			errContains: "invalid syntax",
		},
		{
			name: "Missing required host",
			envVars: map[string]string{
				"SMTP_AUTH_EMAIL":    "user@example.com",
				"SMTP_AUTH_PASSWORD": "password123",
			},
			wantErr:     true,
			errContains: "email configuration is incomplete",
		},
		{
			name: "Missing required auth email",
			envVars: map[string]string{
				"SMTP_HOST":          "smtp.example.com",
				"SMTP_AUTH_PASSWORD": "password123",
			},
			wantErr:     true,
			errContains: "email configuration is incomplete",
		},
		{
			name: "Missing required auth password",
			envVars: map[string]string{
				"SMTP_HOST":       "smtp.example.com",
				"SMTP_AUTH_EMAIL": "user@example.com",
			},
			wantErr:     true,
			errContains: "email configuration is incomplete",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Set up environment
				for k, v := range tt.envVars {
					err := os.Setenv(k, v)
					if err != nil {
						panic(err)
					}
				}
				defer func() {
					// Clean up environment
					for k := range tt.envVars {
						err := os.Unsetenv(k)
						if err != nil {
							panic(err)
						}
					}
				}()

				// Execute
				got, err := NewEmailConfig()

				// Verify
				if tt.wantErr {
					require.Error(t, err)
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.wantConfig, got)
				}
			},
		)
	}
}
