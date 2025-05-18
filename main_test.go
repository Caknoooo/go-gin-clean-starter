package main

import (
	"github.com/Caknoooo/go-gin-clean-starter/command"
	"github.com/Caknoooo/go-gin-clean-starter/provider"
	"github.com/Caknoooo/go-gin-clean-starter/routes"
	"github.com/gin-gonic/gin"
	"os"
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCommand is a mock implementation of the command package
type MockCommand struct {
	mock.Mock
}

func (m *MockCommand) Commands(injector *do.Injector) bool {
	args := m.Called(injector)
	return args.Bool(0)
}

// MockProvider is a mock implementation of the provider package
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) RegisterDependencies(injector *do.Injector) {
	m.Called(injector)
}

// MockRoutes is a mock implementation of the routes package
type MockRoutes struct {
	mock.Mock
}

func (m *MockRoutes) RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	m.Called(server, injector)
}

func TestArgs(t *testing.T) {
	t.Run(
		"with no arguments", func(t *testing.T) {
			// Save and restore original args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"test"}

			injector := do.New()
			result := args(injector)

			assert.True(t, result)
		},
	)

	t.Run(
		"with arguments", func(t *testing.T) {
			// Save and restore original args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"test", "some-command"}

			// Mock the command package
			mockCmd := new(MockCommand)
			mockCmd.On("Commands", mock.Anything).Return(false)
			command.Commands = mockCmd.Commands

			injector := do.New()
			result := args(injector)

			assert.False(t, result)
			mockCmd.AssertExpectations(t)
		},
	)
}

func TestRun(t *testing.T) {
	t.Run(
		"with custom port", func(t *testing.T) {
			// Setup
			oldPort := os.Getenv("PORT")
			defer func() { os.Setenv("PORT", oldPort) }()
			os.Setenv("PORT", "9999")

			server := gin.Default()
			called := false
			originalRun := run
			defer func() { run = originalRun }()
			run = func(s *gin.Engine) {
				called = true
				assert.Equal(t, server, s)
				assert.Equal(t, "9999", os.Getenv("PORT"))
			}

			// Execute
			run(server)

			// Verify
			assert.True(t, called)
		},
	)

	t.Run(
		"dev environment", func(t *testing.T) {
			// Setup
			oldEnv := os.Getenv("APP_ENV")
			defer func() { os.Setenv("APP_ENV", oldEnv) }()
			os.Setenv("APP_ENV", "dev")

			server := gin.Default()
			called := false
			originalRun := run
			defer func() { run = originalRun }()
			run = func(s *gin.Engine) {
				called = true
				assert.Equal(t, server, s)
			}

			// Execute
			run(server)

			// Verify
			assert.True(t, called)
		},
	)

	t.Run(
		"prod environment", func(t *testing.T) {
			// Setup
			oldEnv := os.Getenv("APP_ENV")
			defer func() { os.Setenv("APP_ENV", oldEnv) }()
			os.Setenv("APP_ENV", "prod")

			server := gin.Default()
			called := false
			originalRun := run
			defer func() { run = originalRun }()
			run = func(s *gin.Engine) {
				called = true
				assert.Equal(t, server, s)
			}

			// Execute
			run(server)

			// Verify
			assert.True(t, called)
		},
	)
}

func TestMainFunc(t *testing.T) {
	t.Run(
		"successful execution", func(t *testing.T) {
			// Mock dependencies
			mockProvider := new(MockProvider)
			mockProvider.On("RegisterDependencies", mock.Anything).Return()
			provider.RegisterDependencies = mockProvider.RegisterDependencies

			mockRoutes := new(MockRoutes)
			mockRoutes.On("RegisterRoutes", mock.Anything, mock.Anything).Return()
			routes.RegisterRoutes = mockRoutes.RegisterRoutes

			// Mock args to return true
			originalArgs := args
			defer func() { args = originalArgs }()
			args = func(injector *do.Injector) bool {
				return true
			}

			// Mock run function
			runCalled := false
			originalRun := run
			defer func() { run = originalRun }()
			run = func(server *gin.Engine) {
				runCalled = true
			}

			// Execute
			main()

			// Verify
			assert.True(t, runCalled)
			mockProvider.AssertExpectations(t)
			mockRoutes.AssertExpectations(t)
		},
	)

	t.Run(
		"early return from args", func(t *testing.T) {
			// Mock dependencies
			mockProvider := new(MockProvider)
			mockProvider.On("RegisterDependencies", mock.Anything).Return()
			provider.RegisterDependencies = mockProvider.RegisterDependencies

			// Mock args to return false
			originalArgs := args
			defer func() { args = originalArgs }()
			args = func(injector *do.Injector) bool {
				return false
			}

			// Mock run function (shouldn't be called)
			runCalled := false
			originalRun := run
			defer func() { run = originalRun }()
			run = func(server *gin.Engine) {
				runCalled = true
			}

			// Execute
			main()

			// Verify
			assert.False(t, runCalled)
			mockProvider.AssertExpectations(t)
		},
	)
}
