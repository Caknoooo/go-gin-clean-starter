package command_test

import (
	"github.com/Caknoooo/go-gin-clean-starter/command"
	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"os"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type CommandTestSuite struct {
	suite.Suite
	injector *do.Injector
	db       *gorm.DB
	oldArgs  []string
}

func (suite *CommandTestSuite) SetupSuite() {
	// Setup dependency injection
	suite.injector = do.New()

	// Setup test database (using your container package)
	testContainer, err := container.StartTestContainer()
	if err != nil {
		suite.T().Fatalf("Failed to start test container: %v", err)
	}

	// Set environment variables for database connection
	os.Setenv("DB_HOST", testContainer.Host)
	os.Setenv("DB_PORT", testContainer.Port)
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASS", "testpassword")
	os.Setenv("DB_NAME", "testdb")

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	suite.db = db

	// Provide the db instance to the injector
	do.ProvideNamed[*gorm.DB](
		suite.injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
			return db, nil
		},
	)

	// Backup original args
	suite.oldArgs = os.Args
}

func (suite *CommandTestSuite) TearDownSuite() {
	// Restore original args
	os.Args = suite.oldArgs

	// Clean up database connection
	if suite.db != nil {
		if err := container.CloseDatabaseConnection(suite.db); err != nil {
			suite.T().Logf("Failed to close database connection: %v", err)
		}
	}
}

func (suite *CommandTestSuite) TestCommands_Migrate() {
	// Set up test args
	os.Args = []string{"cmd", "--migrate"}

	// Execute command
	result := command.Commands(suite.injector)

	// Verify
	assert.False(suite.T(), result, "Expected run to be false when migrate flag is set")

	// Verify tables were created
	assert.True(suite.T(), suite.db.Migrator().HasTable("users"), "Users table should exist after migration")
	assert.True(
		suite.T(),
		suite.db.Migrator().HasTable("refresh_tokens"),
		"Refresh tokens table should exist after migration",
	)
}

func (suite *CommandTestSuite) TestCommands_Seed() {
	// First ensure tables exist
	suite.db.AutoMigrate(&entity.User{})

	// Set up test args
	os.Args = []string{"cmd", "--seed"}

	// Execute command
	result := command.Commands(suite.injector)

	// Verify
	assert.False(suite.T(), result, "Expected run to be false when seed flag is set")

	// Verify data was seeded
	var count int64
	suite.db.Model(&entity.User{}).Count(&count)
	assert.Greater(suite.T(), count, int64(0), "Expected users to be seeded")
}

func (suite *CommandTestSuite) TestCommands_Script() {
	// Set up test args
	os.Args = []string{"cmd", "--script:example_script"}

	// Execute command
	result := command.Commands(suite.injector)

	// Verify
	assert.False(suite.T(), result, "Expected run to be false when script flag is set")
}

func (suite *CommandTestSuite) TestCommands_Run() {
	// Set up test args
	os.Args = []string{"cmd", "--run"}

	// Execute command
	result := command.Commands(suite.injector)

	// Verify
	assert.True(suite.T(), result, "Expected run to be true when run flag is set")
}

func (suite *CommandTestSuite) TestCommands_NoFlags() {
	// Set up test args
	os.Args = []string{"cmd"}

	// Execute command
	result := command.Commands(suite.injector)

	// Verify
	assert.False(suite.T(), result, "Expected run to be false when no flags are set")
}

func TestCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommandTestSuite))
}
