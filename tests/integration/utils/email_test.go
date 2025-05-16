package utils_test

import (
	"context"
	"github.com/Caknoooo/go-gin-clean-starter/utils"
	"os"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

type EmailIntegrationTestSuite struct {
	suite.Suite
	smtpContainer testcontainers.Container
	dbContainer   *container.TestDatabaseContainer
	db            *gorm.DB
}

func (suite *EmailIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()

	// Start PostgreSQL container
	dbContainer, err := container.StartTestContainer()
	require.NoError(suite.T(), err)
	suite.dbContainer = dbContainer

	// Set database environment variables
	err = os.Setenv("DB_HOST", dbContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", dbContainer.Port)
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

	// Set up database connection
	suite.db = container.SetUpDatabaseConnection()

	// Start MailHog SMTP container
	smtpReq := testcontainers.ContainerRequest{
		Image:        "mailhog/mailhog",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForListeningPort("1025/tcp"),
	}
	smtpContainer, err := testcontainers.GenericContainer(
		ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: smtpReq,
			Started:          true,
		},
	)
	require.NoError(suite.T(), err)
	suite.smtpContainer = smtpContainer

	// Get SMTP container host and ports
	smtpHost, err := smtpContainer.Host(ctx)
	require.NoError(suite.T(), err)

	smtpPort, err := smtpContainer.MappedPort(ctx, "1025")
	require.NoError(suite.T(), err)

	// Set SMTP environment variables
	err = os.Setenv("SMTP_HOST", smtpHost)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_PORT", smtpPort.Port())
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_AUTH_EMAIL", "test@example.com")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_AUTH_PASSWORD", "password")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("SMTP_SENDER_NAME", "Test Sender")
	if err != nil {
		panic(err)
	}
}

func (suite *EmailIntegrationTestSuite) TearDownSuite() {
	ctx := context.Background()
	timeout := 10 * time.Second

	// Clean up environment variables
	for _, env := range []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASS", "DB_NAME",
		"SMTP_HOST", "SMTP_PORT", "SMTP_AUTH_EMAIL", "SMTP_AUTH_PASSWORD", "SMTP_SENDER_NAME",
	} {
		err := os.Unsetenv(env)
		if err != nil {
			panic(err)
		}
	}

	// Close database connection
	if suite.db != nil {
		err := container.CloseDatabaseConnection(suite.db)
		assert.NoError(suite.T(), err)
	}

	// Stop containers
	if suite.smtpContainer != nil {
		_ = suite.smtpContainer.Stop(ctx, &timeout)
	}
	if suite.dbContainer != nil {
		_ = suite.dbContainer.Stop()
	}
}

func (suite *EmailIntegrationTestSuite) TestSendMail_Integration() {
	tests := []struct {
		name      string
		toEmail   string
		subject   string
		body      string
		wantError bool
	}{
		{
			name:      "Successfully send email",
			toEmail:   "recipient@example.com",
			subject:   "Test Subject",
			body:      "<p>Test Body</p>",
			wantError: false,
		},
		{
			name:      "Invalid recipient email",
			toEmail:   "",
			subject:   "Test Subject",
			body:      "<p>Test Body</p>",
			wantError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(
			tt.name, func() {
				err := utils.SendMail(tt.toEmail, tt.subject, tt.body)
				if tt.wantError {
					assert.Error(suite.T(), err)
				} else {
					assert.NoError(suite.T(), err)
				}
			},
		)
	}
}
