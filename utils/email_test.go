package utils

import (
	"errors"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/gomail.v2"
)

// MockDialer is a mock implementation of gomail.Dialer
type MockDialer struct {
	mock.Mock
}

func (m *MockDialer) DialAndSend(messages ...*gomail.Message) error {
	args := m.Called(messages)
	return args.Error(0)
}

func TestSendMail(t *testing.T) {
	// Mock the NewEmailConfig function
	originalNewEmailConfig := config.NewEmailConfig
	defer func() { config.NewEmailConfig = originalNewEmailConfig }()

	tests := []struct {
		name           string
		emailConfig    *config.EmailConfig
		emailConfigErr error
		dialerErr      error
		wantErr        bool
	}{
		{
			name: "Successfully send email",
			emailConfig: &config.EmailConfig{
				Host:         "smtp.example.com",
				Port:         587,
				AuthEmail:    "test@example.com",
				AuthPassword: "password",
			},
			wantErr: false,
		},
		{
			name:           "Failed to get email config",
			emailConfig:    nil,
			emailConfigErr: errors.New("config error"),
			wantErr:        true,
		},
		{
			name: "Failed to send email",
			emailConfig: &config.EmailConfig{
				Host:         "smtp.example.com",
				Port:         587,
				AuthEmail:    "test@example.com",
				AuthPassword: "password",
			},
			dialerErr: errors.New("smtp error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Setup mock for NewEmailConfig
				config.NewEmailConfig = func() (*config.EmailConfig, error) {
					return tt.emailConfig, tt.emailConfigErr
				}

				// Create mock Dialer
				mockDialer := new(MockDialer)
				if tt.emailConfig != nil {
					mockDialer.On("DialAndSend", mock.Anything).Return(tt.dialerErr)
				}

				// Replace the real Dialer with our mock
				originalNewDialer := NewDialer
				defer func() { NewDialer = originalNewDialer }()
				NewDialer = func(host string, port int, username, password string) Dialer {
					return mockDialer
				}

				// Execute
				err := SendMail("recipient@example.com", "Test Subject", "<p>Test Body</p>")

				// Verify
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				// Verify mock expectations
				if tt.emailConfig != nil {
					mockDialer.AssertExpectations(t)
				}
			},
		)
	}
}
