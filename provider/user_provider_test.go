package provider

import (
	"testing"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/Caknoooo/go-gin-clean-starter/controller"
	"github.com/Caknoooo/go-gin-clean-starter/service"
	"github.com/golang-jwt/jwt/v4"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockJWTService struct{}

func (m *mockJWTService) GenerateAccessToken(userID string, role string) string {
	return "mock-access-token"
}

func (m *mockJWTService) GenerateRefreshToken() (string, time.Time) {
	return "mock-refresh-token", time.Now().Add(24 * time.Hour)
}

func (m *mockJWTService) GetUserIDByToken(token string) (string, error) {
	return "mock-user-id", nil
}

func (m *mockJWTService) ValidateToken(token string) (*jwt.Token, error) {

	claims := jwt.MapClaims{
		"user_id": "mock-user-id",
		"role":    "user",
	}
	return &jwt.Token{
		Raw:    token,
		Claims: claims,
		Valid:  true,
	}, nil
}

func TestProvideUserDependencies(t *testing.T) {

	injector := do.New()

	mockDB := &gorm.DB{}
	do.ProvideNamedValue[*gorm.DB](injector, constants.DB, mockDB)

	mockJWT := &mockJWTService{}
	do.ProvideNamedValue[service.JWTService](injector, constants.JWTService, mockJWT)

	ProvideUserDependencies(injector)

	userController, err := do.Invoke[controller.UserController](injector)
	assert.NoError(t, err, "should provide UserController without error")
	assert.NotNil(t, userController, "UserController should not be nil")
}

func TestProvideUserDependencies_MissingDB(t *testing.T) {

	injector := do.New()

	mockJWT := &mockJWTService{}
	do.ProvideNamedValue[service.JWTService](injector, constants.JWTService, mockJWT)

	assert.Panics(
		t,
		func() {
			ProvideUserDependencies(injector)
		},
		"should panic when DB is missing",
	)
}

func TestProvideUserDependencies_MissingJWTService(t *testing.T) {

	injector := do.New()

	mockDB := &gorm.DB{}
	do.ProvideNamedValue[*gorm.DB](injector, constants.DB, mockDB)

	assert.Panics(
		t,
		func() {
			ProvideUserDependencies(injector)
		},
		"should panic when JWTService is missing",
	)
}
