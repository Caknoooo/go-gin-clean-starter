package validation

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/go-playground/validator/v10"
)

type UserValidation struct {
	validate *validator.Validate
}

func NewUserValidation() *UserValidation {
	validate := validator.New()

	validate.RegisterValidation("telp_number", validateTelpNumber)
	validate.RegisterValidation("name", validateName)

	return &UserValidation{
		validate: validate,
	}
}

func (v *UserValidation) ValidateUserCreateRequest(req dto.UserCreateRequest) error {
	return v.validate.Struct(req)
}

func (v *UserValidation) ValidateUserUpdateRequest(req dto.UserUpdateRequest) error {
	return v.validate.Struct(req)
}

func validateTelpNumber(fl validator.FieldLevel) bool {
	telp := fl.Field().String()
	// Basic phone number validation - should be numeric and have reasonable length
	if len(telp) < 8 || len(telp) > 15 {
		return false
	}
	// Add more phone validation rules as needed
	return true
}

func validateName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	// Name should not be empty and not too long
	return len(name) > 0 && len(name) <= 100
}
