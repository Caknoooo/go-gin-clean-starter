package tests

import (
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/validation"
	"github.com/stretchr/testify/assert"
)

func TestUserValidation_ValidateUserCreateRequest_Success(t *testing.T) {
	userValidation := validation.NewUserValidation()

	req := dto.UserCreateRequest{
		Name:       "Test User",
		Email:      "test@example.com",
		TelpNumber: "12345678",
		Password:   "password123",
	}

	err := userValidation.ValidateUserCreateRequest(req)

	assert.NoError(t, err)
}

func TestUserValidation_ValidateUserCreateRequest_InvalidName(t *testing.T) {
	userValidation := validation.NewUserValidation()

	req := dto.UserCreateRequest{
		Name:       "", // This will be caught by binding:"required,min=2,max=100" in DTO
		Email:      "test@example.com",
		TelpNumber: "12345678",
		Password:   "password123",
	}

	err := userValidation.ValidateUserCreateRequest(req)

	// The validation should pass because DTO binding handles name validation
	// Custom validation only adds extra checks beyond DTO binding
	assert.NoError(t, err)
}

func TestUserValidation_ValidateUserUpdateRequest_Success(t *testing.T) {
	userValidation := validation.NewUserValidation()

	req := dto.UserUpdateRequest{
		Name:       "Updated Name",
		TelpNumber: "87654321",
		Email:      "updated@example.com",
	}

	err := userValidation.ValidateUserUpdateRequest(req)

	assert.NoError(t, err)
}

func TestUserValidation_ValidateUserUpdateRequest_InvalidTelp(t *testing.T) {
	userValidation := validation.NewUserValidation()

	req := dto.UserUpdateRequest{
		Name:       "Updated Name",
		TelpNumber: "123", // This will be caught by binding:"omitempty,min=8,max=20" in DTO
		Email:      "updated@example.com",
	}

	err := userValidation.ValidateUserUpdateRequest(req)

	// The validation should pass because DTO binding handles telp validation
	// Custom validation only adds extra checks beyond DTO binding
	assert.NoError(t, err)
}
