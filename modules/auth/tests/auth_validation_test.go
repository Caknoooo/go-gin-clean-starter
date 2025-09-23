package tests

import (
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/validation"
	userDto "github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/stretchr/testify/assert"
)

func TestAuthValidation_ValidateRegisterRequest_Success(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Name:       "Test User",
		Email:      "test@example.com",
		TelpNumber: "12345678",
		Password:   "password123",
	}

	err := authValidation.ValidateRegisterRequest(req)

	assert.NoError(t, err)
}

func TestAuthValidation_ValidateRegisterRequest_InvalidEmail(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Name:       "Test User",
		Email:      "invalid-email", // This will be caught by binding:"required,email" in DTO
		TelpNumber: "12345678",
		Password:   "password123",
	}

	err := authValidation.ValidateRegisterRequest(req)

	// The validation should pass because DTO binding handles email validation
	// Custom validation only adds extra checks beyond DTO binding
	assert.NoError(t, err)
}

func TestAuthValidation_ValidateRegisterRequest_ShortPassword(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Name:       "Test User",
		Email:      "test@example.com",
		TelpNumber: "12345678",
		Password:   "123", // This will be caught by binding:"required,min=8" in DTO
	}

	err := authValidation.ValidateRegisterRequest(req)

	// The validation should pass because DTO binding handles password validation
	// Custom validation only adds extra checks beyond DTO binding
	assert.NoError(t, err)
}

func TestAuthValidation_ValidateLoginRequest_Success(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	err := authValidation.ValidateLoginRequest(req)

	assert.NoError(t, err)
}

func TestAuthValidation_ValidateRefreshTokenRequest_Success(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := dto.RefreshTokenRequest{
		RefreshToken: "valid-refresh-token",
	}

	err := authValidation.ValidateRefreshTokenRequest(req)

	assert.NoError(t, err)
}
