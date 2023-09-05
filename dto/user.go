package dto

import "errors"

var (
	ErrCreateUser     = errors.New("Failed to create user")
	ErrGetAllUser     = errors.New("Failed to get all user")
	ErrGetUserById    = errors.New("Failed to get user by id")
	ErrGetUserByEmail = errors.New("Failed to get user by email")
	ErrUpdateUser     = errors.New("Failed to update user")
	ErrUserNotFound   = errors.New("User not found")
	ErrDeleteUser     = errors.New("Failed to delete user")
)

type (
	UserCreateRequest struct {
		Nama     string `json:"nama" form:"nama"`
		NoTelp   string `json:"no_telp" form:"no_telp"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}

	UserResponse struct {
		ID       string `json:"id"`
		Nama     string `json:"nama"`
		NoTelp   string `json:"no_telp"`
		Role     string `json:"role"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	UserUpdateRequest struct {
		Nama     string `json:"nama" form:"nama"`
		NoTelp   string `json:"no_telp" form:"no_telp"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}

	UserLoginRequest struct {
		Email    string `json:"email" form:"email" binding:"required"`
		Password string `json:"password" form:"password" binding:"required"`
	}
)
