package config_test

import (
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"os"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type DatabaseConfigTestSuite struct {
	suite.Suite
	dbContainer *container.TestDatabaseContainer
	db          *gorm.DB
}

// SetupSuite runs before the entire test suite
func (suite *DatabaseConfigTestSuite) SetupSuite() {
	// Start test container
	container, err := container.StartTestContainer()
	require.NoError(suite.T(), err)
	suite.dbContainer = container

	// Set environment variables for the test
	err = os.Setenv("APP_ENV", constants.ENUM_RUN_TESTING)
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
	err = os.Setenv("DB_HOST", container.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", container.Port)
	if err != nil {
		panic(err)
	}
}

// TearDownSuite runs after the entire test suite
func (suite *DatabaseConfigTestSuite) TearDownSuite() {
	// Close database connection if it exists
	if suite.db != nil {
		err := container.CloseDatabaseConnection(suite.db)
		if err != nil {
			panic(err)
		}
	}

	// Stop and remove the container
	if suite.dbContainer != nil {
		err := suite.dbContainer.Stop()
		require.NoError(suite.T(), err)
	}

	// Clean up environment variables
	err := os.Unsetenv("APP_ENV")
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
	err = os.Unsetenv("DB_HOST")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("DB_PORT")
	if err != nil {
		panic(err)
	}
}

func (suite *DatabaseConfigTestSuite) TestSetUpDatabaseConnection() {
	db := container.SetUpDatabaseConnection()
	suite.db = db // Store for cleanup

	// Verify the connection works by executing a simple query
	var result int
	err := db.Raw("SELECT 1").Scan(&result).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, result)

	// Verify UUID extension was created
	var extensions []string
	err = db.Raw("SELECT extname FROM pg_extension WHERE extname = 'uuid-ossp'").Scan(&extensions).Error
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), extensions)
}

func (suite *DatabaseConfigTestSuite) TestCloseDatabaseConnection() {
	db := container.SetUpDatabaseConnection()
	suite.db = db

	// Verify connection is open
	var result int
	err := db.Raw("SELECT 1").Scan(&result).Error
	require.NoError(suite.T(), err)

	// Close connection
	err = container.CloseDatabaseConnection(db)
	if err != nil {
		panic(err)
	}

	// Verify connection is closed
	dbSQL, err := db.DB()
	require.NoError(suite.T(), err)
	err = dbSQL.Ping()
	require.Error(suite.T(), err)
}

func TestDatabaseConfigTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseConfigTestSuite))
}
