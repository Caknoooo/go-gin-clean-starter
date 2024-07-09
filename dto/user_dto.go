package dto

import (
	"errors"
	"mime/multipart"

	"github.com/Caknoooo/go-gin-clean-starter/entity"
)

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY      = "failed get data from body"
	MESSAGE_FAILED_REGISTER_USER           = "failed create user"
	MESSAGE_FAILED_GET_LIST_USER           = "failed get list user"
	MESSAGE_FAILED_GET_USER_TOKEN          = "failed get user token"
	MESSAGE_FAILED_TOKEN_NOT_VALID         = "token not valid"
	MESSAGE_FAILED_TOKEN_NOT_FOUND         = "token not found"
	MESSAGE_FAILED_GET_USER                = "failed get user"
	MESSAGE_FAILED_LOGIN                   = "failed login"
	MESSAGE_FAILED_WRONG_EMAIL_OR_PASSWORD = "wrong email or password"
	MESSAGE_FAILED_UPDATE_USER             = "failed update user"
	MESSAGE_FAILED_DELETE_USER             = "failed delete user"
	MESSAGE_FAILED_PROSES_REQUEST          = "failed proses request"
	MESSAGE_FAILED_DENIED_ACCESS           = "denied access"
	MESSAGE_FAILED_VERIFY_EMAIL            = "failed verify email"

	// Success
	MESSAGE_SUCCESS_REGISTER_USER           = "success create user"
	MESSAGE_SUCCESS_GET_LIST_USER           = "success get list user"
	MESSAGE_SUCCESS_GET_USER                = "success get user"
	MESSAGE_SUCCESS_LOGIN                   = "success login"
	MESSAGE_SUCCESS_UPDATE_USER             = "success update user"
	MESSAGE_SUCCESS_DELETE_USER             = "success delete user"
	MESSAGE_SEND_VERIFICATION_EMAIL_SUCCESS = "success send verification email"
	MESSAGE_SUCCESS_VERIFY_EMAIL            = "success verify email"
)

var (
	ErrCreateUser             = errors.New("failed to create user")
	ErrGetAllUser             = errors.New("failed to get all user")
	ErrGetUserById            = errors.New("failed to get user by id")
	ErrGetUserByEmail         = errors.New("failed to get user by email")
	ErrEmailAlreadyExists     = errors.New("email already exist")
	ErrUpdateUser             = errors.New("failed to update user")
	ErrUserNotAdmin           = errors.New("user not admin")
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailNotFound          = errors.New("email not found")
	ErrDeleteUser             = errors.New("failed to delete user")
	ErrPasswordNotMatch       = errors.New("password not match")
	ErrEmailOrPassword        = errors.New("wrong email or password")
	ErrAccountNotVerified     = errors.New("account not verified")
	ErrTokenInvalid           = errors.New("token invalid")
	ErrTokenExpired           = errors.New("token expired")
	ErrAccountAlreadyVerified = errors.New("account already verified")
)

type (
	UserCreateRequest struct {
		Name       string                `json:"name" form:"name"`
		TelpNumber string                `json:"telp_number" form:"telp_number"`
		Email      string                `json:"email" form:"email"`
		Image      *multipart.FileHeader `json:"image" form:"image"`
		Password   string                `json:"password" form:"password"`
	}

	UserResponse struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Email      string `json:"email"`
		TelpNumber string `json:"telp_number"`
		Role       string `json:"role"`
		ImageUrl   string `json:"image_url"`
		IsVerified bool   `json:"is_verified"`
	}

	UserPaginationResponse struct {
		Data []UserResponse `json:"data"`
		PaginationResponse
	}

	GetAllUserRepositoryResponse struct {
		Users []entity.User
		PaginationResponse
	}

	UserUpdateRequest struct {
		Name       string `json:"name" form:"name"`
		TelpNumber string `json:"telp_number" form:"telp_number"`
		Email      string `json:"email" form:"email"`
	}

	UserUpdateResponse struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		TelpNumber string `json:"telp_number"`
		Role       string `json:"role"`
		Email      string `json:"email"`
		IsVerified bool   `json:"is_verified"`
	}

	SendVerificationEmailRequest struct {
		Email string `json:"email" form:"email" binding:"required"`
	}

	VerifyEmailRequest struct {
		Token string `json:"token" form:"token" binding:"required"`
	}

	VerifyEmailResponse struct {
		Email      string `json:"email"`
		IsVerified bool   `json:"is_verified"`
	}

	UserLoginRequest struct {
		Email    string `json:"email" form:"email" binding:"required"`
		Password string `json:"password" form:"password" binding:"required"`
	}

	UserLoginResponse struct {
		Token string `json:"token"`
		Role  string `json:"role"`
	}

	UpdateStatusIsVerifiedRequest struct {
		UserId     string `json:"user_id" form:"user_id" binding:"required"`
		IsVerified bool   `json:"is_verified" form:"is_verified"`
	}
)
