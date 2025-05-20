package config_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/config"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EmailConfigTestSuite struct {
	suite.Suite
	emailContainer *container.TestDatabaseContainer
}

func (suite *EmailConfigTestSuite) SetupSuite() {

	container, err := container.StartTestContainer()
	require.NoError(suite.T(), err)
	suite.emailContainer = container

	err = os.Setenv("SMTP_HOST", container.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_PORT", container.Port)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_SENDER_NAME", "Test Sender")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_AUTH_EMAIL", "test@example.com")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_AUTH_PASSWORD", "password123")
	if err != nil {
		panic(err)
	}
}

func (suite *EmailConfigTestSuite) TearDownSuite() {

	err := os.Unsetenv("SMTP_HOST")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("SMTP_PORT")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("SMTP_SENDER_NAME")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("SMTP_AUTH_EMAIL")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("SMTP_AUTH_PASSWORD")
	if err != nil {
		panic(err)
	}

	if suite.emailContainer != nil {
		err := suite.emailContainer.Stop()
		require.NoError(suite.T(), err)
	}
}

func (suite *EmailConfigTestSuite) TestNewEmailConfig_Integration() {

	emailConfig, err := config.NewEmailConfig()
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), emailConfig)

	assert.Equal(suite.T(), os.Getenv("SMTP_HOST"), emailConfig.Host)
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	assert.Equal(suite.T(), port, emailConfig.Port)
	assert.Equal(suite.T(), os.Getenv("SMTP_SENDER_NAME"), emailConfig.SenderName)
	assert.Equal(suite.T(), os.Getenv("SMTP_AUTH_EMAIL"), emailConfig.AuthEmail)
	assert.Equal(suite.T(), os.Getenv("SMTP_AUTH_PASSWORD"), emailConfig.AuthPassword)
}

func TestEmailConfigTestSuite(t *testing.T) {
	suite.Run(t, new(EmailConfigTestSuite))
}
