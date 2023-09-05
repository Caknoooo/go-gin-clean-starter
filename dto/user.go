package dto

import "errors"

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY      = "gagal mendapatkan request dari body"
	MESSAGE_FAILED_REGISTER_USER           = "gagal menambahkan user"
	MESSAGE_FAILED_GET_LIST_USER           = "gagal mendapatkan list user"
	MESSAGE_FAILED_GET_USER_TOKEN          = "gagal mendapatkan token user"
	MESSAGE_FAILED_TOKEN_NOT_VALID         = "token tidak valid"
	MESSAGE_FAILED_GET_USER                = "gagal mendapatkan user"
	MESSAGE_FAILED_LOGIN                   = "gagal login"
	MESSAGE_FAILED_WRONG_EMAIL_OR_PASSWORD = "email atau password salah"
	MESSAGE_FAILED_UPDATE_USER             = "gagal update user"
	MESSAGE_FAILED_DELETE_USER             = "gagal delete user"

	// Success
	MESSAGE_SUCCESS_REGISTER_USER = "berhasil menambahkan user"
	MESSAGE_SUCCESS_GET_LIST_USER = "berhasil mendapatkan list user"
	MESSAGE_SUCCESS_GET_USER      = "berhasil mendapatkan user"
	MESSAGE_SUCCESS_LOGIN         = "berhasil login"
	MESSAGE_SUCCESS_UPDATE_USER   = "berhasil update user"
	MESSAGE_SUCCESS_DELETE_USER   = "berhasil delete user"
)

var (
	ErrCreateUser         = errors.New("failed to create user")
	ErrGetAllUser         = errors.New("failed to get all user")
	ErrGetUserById        = errors.New("failed to get user by id")
	ErrGetUserByEmail     = errors.New("failed to get user by email")
	ErrEmailAlreadyExists = errors.New("email already exist")
	ErrUpdateUser         = errors.New("failed to update user")
	ErrUserNotFound       = errors.New("user not found")
	ErrDeleteUser         = errors.New("failed to delete user")
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
