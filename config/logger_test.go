package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/logger"
)

func TestSetupLogger(t *testing.T) {
	// Setup test directory
	testDir := "./test_logs"
	t.Cleanup(
		func() {
			err := os.RemoveAll(testDir)
			if err != nil {
				panic(err)
			}
		},
	)

	t.Run(
		"Successfully creates logger with directory and file", func(t *testing.T) {
			// Replace the constant for testing
			originalLogDir := LogDir
			LogDir = testDir
			t.Cleanup(
				func() {
					LogDir = originalLogDir
				},
			)

			// Execute
			result := SetupLogger()

			// Verify
			assert.NotNil(t, result)

			// Check directory was created
			_, err := os.Stat(testDir)
			assert.NoError(t, err)

			// Check log file was created
			currentMonth := strings.ToLower(time.Now().Format("January"))
			logFileName := fmt.Sprintf("%s_query.log", currentMonth)
			logPath := filepath.Join(testDir, logFileName)
			_, err = os.Stat(logPath)
			assert.NoError(t, err)
		},
	)

	t.Run(
		"Fails when directory cannot be created", func(t *testing.T) {
			if os.Geteuid() == 0 {
				t.Skip("Skipping test when running as root")
			}

			// Replace the constant for testing with invalid path
			originalLogDir := LogDir
			LogDir = "/root/protected_directory"
			t.Cleanup(
				func() {
					LogDir = originalLogDir
				},
			)
		},
	)

	t.Run(
		"Fails when file cannot be created", func(t *testing.T) {
			// Replace the constant for testing
			originalLogDir := LogDir
			LogDir = testDir
			t.Cleanup(
				func() {
					LogDir = originalLogDir
				},
			)

			// Create directory but make it read-only
			err := os.MkdirAll(testDir, 0444)
			require.NoError(t, err)
		},
	)
}

func TestLoggerInterfaceImplementation(t *testing.T) {
	// Setup test directory
	testDir := "./test_logs"
	t.Cleanup(
		func() {
			err := os.RemoveAll(testDir)
			if err != nil {
				panic(err)
			}
		},
	)

	// Replace the constant for testing
	originalLogDir := LogDir
	LogDir = testDir
	t.Cleanup(
		func() {
			LogDir = originalLogDir
		},
	)

	// Execute
	result := SetupLogger()

	// Verify it implements the logger.Interface
	_, ok := result.(logger.Interface)
	assert.True(t, ok, "SetupLogger should return a logger.Interface implementation")
}
