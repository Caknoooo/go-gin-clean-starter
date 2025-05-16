package config_test

import (
	"fmt"
	"github.com/Caknoooo/go-gin-clean-starter/config"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type LoggerIntegrationTestSuite struct {
	suite.Suite
	dbContainer *container.TestDatabaseContainer
	db          *gorm.DB
	testLogDir  string
}

func (suite *LoggerIntegrationTestSuite) SetupSuite() {
	// Setup test log directory
	suite.testLogDir = "./test_logs_integration"
	config.LogDir = suite.testLogDir
	err := os.MkdirAll(suite.testLogDir, 0755)
	if err != nil {
		return
	}

	// Start test database container
	container, err := container.StartTestContainer()
	require.NoError(suite.T(), err)
	suite.dbContainer = container

	// Set environment variables for the test
	err = os.Setenv("DB_HOST", container.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", container.Port)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}

	// Setup database connection with logger
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	suite.db, err = gorm.Open(
		postgres.Open(dsn), &gorm.Config{
			Logger: config.SetupLogger(),
		},
	)
	require.NoError(suite.T(), err)

	// Enable UUID extension
	err = suite.db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	require.NoError(suite.T(), err)
}

func (suite *LoggerIntegrationTestSuite) TearDownSuite() {
	// Close database connection
	if suite.db != nil {
		sqlDB, err := suite.db.DB()
		if err == nil {
			err := sqlDB.Close()
			if err != nil {
				panic(err)
			}
		}
	}

	// Stop container
	if suite.dbContainer != nil {
		err := suite.dbContainer.Stop()
		if err != nil {
			panic(err)
		}
	}

	// Clean up environment
	err := os.Unsetenv("DB_HOST")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("DB_PORT")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("DB_USER")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("DB_PASS")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("DB_NAME")
	if err != nil {
		panic(err)
	}

	// Clean up test log directory
	err = os.RemoveAll(suite.testLogDir)
	if err != nil {
		panic(err)
	}
}

func (suite *LoggerIntegrationTestSuite) TestLoggerWithDatabaseOperations() {
	// Create a simple table for testing
	type TestModel struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	err := suite.db.AutoMigrate(&TestModel{})
	require.NoError(suite.T(), err)

	// Perform operations that should be logged
	tests := []struct {
		name string
		op   func() error
	}{
		{
			name: "Create record",
			op: func() error {
				return suite.db.Create(&TestModel{Name: "test"}).Error
			},
		},
		{
			name: "Find existing record",
			op: func() error {
				var result TestModel
				return suite.db.First(&result, 1).Error
			},
		},
		{
			name: "Find non-existent record",
			op: func() error {
				var result TestModel
				return suite.db.First(&result, 999).Error
			},
		},
	}

	for _, tt := range tests {
		suite.Run(
			tt.name, func() {
				err := tt.op()
				if tt.name == "Find non-existent record" {
					assert.Error(suite.T(), err)
				} else {
					assert.NoError(suite.T(), err)
				}
			},
		)
	}

	// Verify log file was written to
	currentMonth := strings.ToLower(time.Now().Format("January"))
	logFileName := fmt.Sprintf("%s_query.log", currentMonth)
	logPath := filepath.Join(suite.testLogDir, logFileName)

	fileInfo, err := os.Stat(logPath)
	require.NoError(suite.T(), err)
	assert.Greater(suite.T(), fileInfo.Size(), int64(0))

	// Optionally: read the file and verify contents contain expected log entries
	// This would be more complex and might need regex matching
}

func TestLoggerIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerIntegrationTestSuite))
}
