package command_test

import (
	"os"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/command"
	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"

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
	suite.injector = do.New()

	testContainer, err := container.StartTestContainer()
	if err != nil {
		suite.T().Fatalf("Failed to start test container: %v", err)
	}

	os.Setenv("DB_HOST", testContainer.Host)
	os.Setenv("DB_PORT", testContainer.Port)
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASS", "testpassword")
	os.Setenv("DB_NAME", "testdb")

	db := container.SetUpDatabaseConnection()
	suite.db = db

	do.ProvideNamed(
		suite.injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
			return db, nil
		},
	)

	suite.oldArgs = os.Args
}

func (suite *CommandTestSuite) TearDownSuite() {
	os.Args = suite.oldArgs

	if suite.db != nil {
		if err := container.CloseDatabaseConnection(suite.db); err != nil {
			suite.T().Logf("Failed to close database connection: %v", err)
		}
	}
}

func (suite *CommandTestSuite) TestCommands_Migrate() {
	os.Args = []string{"cmd", "--migrate"}

	result := command.Commands(suite.injector)

	assert.False(suite.T(), result, "Expected run to be false when migrate flag is set")

	assert.True(suite.T(), suite.db.Migrator().HasTable("users"), "Users table should exist after migration")
	assert.True(
		suite.T(),
		suite.db.Migrator().HasTable("refresh_tokens"),
		"Refresh tokens table should exist after migration",
	)
}

func (suite *CommandTestSuite) TestCommands_Seed() {
	suite.db.AutoMigrate(&entity.User{})

	os.Args = []string{"cmd", "--seed"}

	result := command.Commands(suite.injector)

	assert.False(suite.T(), result, "Expected run to be false when seed flag is set")

	var count int64
	suite.db.Model(&entity.User{}).Count(&count)
	assert.Greater(suite.T(), count, int64(0), "Expected users to be seeded")
}

func (suite *CommandTestSuite) TestCommands_Script() {
	os.Args = []string{"cmd", "--script:example_script"}

	result := command.Commands(suite.injector)

	assert.False(suite.T(), result, "Expected run to be false when script flag is set")
}

func (suite *CommandTestSuite) TestCommands_Run() {
	os.Args = []string{"cmd", "--run"}

	result := command.Commands(suite.injector)

	assert.True(suite.T(), result, "Expected run to be true when run flag is set")
}

func (suite *CommandTestSuite) TestCommands_NoFlags() {
	os.Args = []string{"cmd"}

	result := command.Commands(suite.injector)

	assert.False(suite.T(), result, "Expected run to be false when no flags are set")
}

func TestCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommandTestSuite))
}
