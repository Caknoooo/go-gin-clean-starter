package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRouteRegistrar mocks the User route registration function
type MockUserRouteRegistrar struct {
	mock.Mock
}

func (m *MockUserRouteRegistrar) User(server *gin.Engine, injector *do.Injector) {
	m.Called(server, injector)
}

func TestRegisterRoutes(t *testing.T) {
	// Create mock server and injector
	mockEngine := gin.Default()
	mockInjector := do.New()

	t.Run(
		"Successfully registers routes", func(t *testing.T) {
			// Create mock for User function
			mockUserRegistrar := new(MockUserRouteRegistrar)

			// Set expectations
			mockUserRegistrar.On("User", mockEngine, mockInjector).Once()

			// Replace actual User function with mock
			originalUser := User
			User = mockUserRegistrar.User
			defer func() { User = originalUser }() // Restore original function

			// Call the function
			RegisterRoutes(mockEngine, mockInjector)

			// Verify
			mockUserRegistrar.AssertExpectations(t)
		},
	)

	t.Run(
		"Passes correct parameters", func(t *testing.T) {
			// Create mock for User function
			mockUserRegistrar := new(MockUserRouteRegistrar)

			// Set expectations with argument matchers
			mockUserRegistrar.On("User", mock.AnythingOfType("*gin.Engine"), mock.AnythingOfType("*do.Injector")).Once()

			// Replace actual User function with mock
			originalUser := User
			User = mockUserRegistrar.User
			defer func() { User = originalUser }()

			// Call the function
			RegisterRoutes(mockEngine, mockInjector)

			// Verify
			mockUserRegistrar.AssertExpectations(t)

			// Get the actual arguments passed
			args := mockUserRegistrar.Calls[0].Arguments
			serverArg := args.Get(0).(*gin.Engine)
			injectorArg := args.Get(1).(*do.Injector)

			assert.Equal(t, mockEngine, serverArg)
			assert.Equal(t, mockInjector, injectorArg)
		},
	)
}
