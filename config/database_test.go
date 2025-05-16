package config

import (
	"os"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Unit tests
func TestGetEnv(t *testing.T) {
	t.Run(
		"Should return value when env exists", func(t *testing.T) {
			key := "TEST_KEY"
			value := "test_value"
			os.Setenv(key, value)
			defer os.Unsetenv(key)

			result := getEnv(key, "default")
			assert.Equal(t, value, result)
		},
	)

	t.Run(
		"Should return default when env doesn't exist", func(t *testing.T) {
			result := getEnv("NON_EXISTENT_KEY", "default_value")
			assert.Equal(t, "default_value", result)
		},
	)
}

func TestLoadEnv(t *testing.T) {
	t.Run(
		"Should load .env.test file in testing mode", func(t *testing.T) {
			// Setup
			originalEnv := os.Getenv("APP_ENV")
			os.Setenv("APP_ENV", constants.ENUM_RUN_TESTING)
			defer func() {
				os.Setenv("APP_ENV", originalEnv)
			}()

			// Create a temporary .env.test file
			content := "DB_USER=testuser\nDB_PASS=testpass\nDB_NAME=testdb\n"
			err := os.WriteFile(".env.test", []byte(content), 0644)
			require.NoError(t, err)
			defer os.Remove(".env.test")

			// Test
			loadEnv()

			// Verify
			assert.Equal(t, "testuser", os.Getenv("DB_USER"))
			assert.Equal(t, "testpass", os.Getenv("DB_PASS"))
			assert.Equal(t, "testdb", os.Getenv("DB_NAME"))
		},
	)

	t.Run(
		"Should panic when .env.test file is missing in testing mode", func(t *testing.T) {
			// Setup
			originalEnv := os.Getenv("APP_ENV")
			os.Setenv("APP_ENV", constants.ENUM_RUN_TESTING)
			defer func() {
				os.Setenv("APP_ENV", originalEnv)
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()

			// Ensure .env.test doesn't exist
			os.Remove(".env.test")

			// Test
			loadEnv()
		},
	)
}
