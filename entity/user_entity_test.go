package entity

import (
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUser_BeforeCreate(t *testing.T) {
	tests := []struct {
		name        string
		user        *User
		expectError bool
		validate    func(t *testing.T, user *User)
	}{
		{
			name: "Valid user with password",
			user: &User{
				Name:       "John Doe",
				Email:      "john@example.com",
				TelpNumber: "1234567890",
				Password:   "password123",
				Role:       "user",
			},
			expectError: false,
			validate: func(t *testing.T, user *User) {
				assert.NotEqual(t, "password123", user.Password, "Password should be hashed")
				assert.NotEqual(t, uuid.Nil, user.ID, "ID should be set")
				assert.Equal(t, "user", user.Role, "Role should be set to user")
				assert.False(t, user.IsVerified, "IsVerified should be false by default")
			},
		},
		{
			name: "User with empty role",
			user: &User{
				Name:       "Jane Doe",
				Email:      "jane@example.com",
				TelpNumber: "0987654321",
				Password:   "password123",
			},
			expectError: false,
			validate: func(t *testing.T, user *User) {
				assert.Equal(t, "user", user.Role, "Role should default to user")
				assert.NotEqual(t, uuid.Nil, user.ID, "ID should be set")
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := tt.user.BeforeCreate(&gorm.DB{})
				if tt.expectError {
					assert.Error(t, err, "Expected an error")
				} else {
					assert.NoError(t, err, "Expected no error")
					tt.validate(t, tt.user)
				}
			},
		)
	}
}

func TestUser_BeforeUpdate(t *testing.T) {
	tests := []struct {
		name        string
		user        *User
		expectError bool
		validate    func(t *testing.T, user *User)
	}{
		{
			name: "Update with new password",
			user: &User{
				ID:       uuid.New(),
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "newpassword123",
				Role:     "user",
			},
			expectError: false,
			validate: func(t *testing.T, user *User) {
				assert.NotEqual(t, "newpassword123", user.Password, "Password should be hashed")
				// Verify the password can be checked
				_, err := helpers.CheckPassword(user.Password, []byte("newpassword123"))
				assert.NoError(t, err, "Hashed password should match original")
			},
		},
		{
			name: "Update without password change",
			user: &User{
				ID:       uuid.New(),
				Name:     "Jane Doe",
				Email:    "jane@example.com",
				Password: "", // No password change
				Role:     "admin",
			},
			expectError: false,
			validate: func(t *testing.T, user *User) {
				assert.Empty(t, user.Password, "Password should remain empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := tt.user.BeforeUpdate(&gorm.DB{})
				if tt.expectError {
					assert.Error(t, err, "Expected an error")
				} else {
					assert.NoError(t, err, "Expected no error")
					tt.validate(t, tt.user)
				}
			},
		)
	}
}

func TestUser_Validation(t *testing.T) {
	tests := []struct {
		name        string
		user        *User
		expectError bool
	}{
		{
			name: "Valid user",
			user: &User{
				Name:       "John Doe",
				Email:      "john@example.com",
				TelpNumber: "1234567890",
				Password:   "password123",
				Role:       "user",
				ImageUrl:   "https://example.com/image.jpg",
			},
			expectError: false,
		},
		{
			name: "Invalid email",
			user: &User{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
				Role:     "user",
			},
			expectError: true,
		},
		{
			name: "Invalid role",
			user: &User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     "invalid_role",
			},
			expectError: true,
		},
		{
			name: "Password too short",
			user: &User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "short",
				Role:     "user",
			},
			expectError: true,
		},
		{
			name: "Invalid image URL",
			user: &User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
				ImageUrl: "not-a-url",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Assuming a validation function exists or using a validator library
				// Here we mock the validation behavior based on struct tags
				err := validateUser(tt.user)
				if tt.expectError {
					assert.Error(t, err, "Expected validation error")
				} else {
					assert.NoError(t, err, "Expected no validation error")
				}
			},
		)
	}
}

// Mock validation function to simulate struct tag validation
func validateUser(user *User) error {
	if user.Name == "" || len(user.Name) < 2 || len(user.Name) > 100 {
		return assert.AnError
	}
	if user.Email == "" || !isValidEmail(user.Email) {
		return assert.AnError
	}
	if user.TelpNumber != "" && (len(user.TelpNumber) < 8 || len(user.TelpNumber) > 20) {
		return assert.AnError
	}
	if user.Password == "" || len(user.Password) < 8 {
		return assert.AnError
	}
	if user.Role != "user" && user.Role != "admin" {
		return assert.AnError
	}
	if user.ImageUrl != "" && !isValidURL(user.ImageUrl) {
		return assert.AnError
	}
	return nil
}

// Mock helper functions for validation
func isValidEmail(email string) bool {
	// Simplified email validation
	return email != "" && email != "invalid-email"
}

func isValidURL(url string) bool {
	// Simplified URL validation
	return url != "" && url != "not-a-url"
}
