package validation

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/dto"
	userDto "github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/go-playground/validator/v10"
)

type AuthValidation struct {
	validate *validator.Validate
}

func NewAuthValidation() *AuthValidation {
	validate := validator.New()

	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("email", validateEmail)

	return &AuthValidation{
		validate: validate,
	}
}

func (v *AuthValidation) ValidateRegisterRequest(req userDto.UserCreateRequest) error {
	return v.validate.Struct(req)
}

func (v *AuthValidation) ValidateLoginRequest(req userDto.UserLoginRequest) error {
	return v.validate.Struct(req)
}

func (v *AuthValidation) ValidateRefreshTokenRequest(req dto.RefreshTokenRequest) error {
	return v.validate.Struct(req)
}

func (v *AuthValidation) ValidateSendPasswordResetRequest(req dto.SendPasswordResetRequest) error {
	return v.validate.Struct(req)
}

func (v *AuthValidation) ValidateResetPasswordRequest(req dto.ResetPasswordRequest) error {
	return v.validate.Struct(req)
}

func (v *AuthValidation) ValidateSendVerificationEmailRequest(req userDto.SendVerificationEmailRequest) error {
	return v.validate.Struct(req)
}

func (v *AuthValidation) ValidateVerifyEmailRequest(req userDto.VerifyEmailRequest) error {
	return v.validate.Struct(req)
}

// Custom validators
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	// Add more password validation rules as needed
	return true
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	// Basic email validation - you can use regex for more complex validation
	return len(email) > 0 && len(email) < 255
}
