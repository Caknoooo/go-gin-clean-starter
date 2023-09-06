package services

import (
	"context"

	"github.com/Caknoooo/golang-clean_template/constants"
	"github.com/Caknoooo/golang-clean_template/dto"
	"github.com/Caknoooo/golang-clean_template/entities"
	"github.com/Caknoooo/golang-clean_template/helpers"
	"github.com/Caknoooo/golang-clean_template/repository"
)

type UserService interface {
	RegisterUser(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error)
	GetAllUser(ctx context.Context) ([]dto.UserResponse, error)
	GetUserById(ctx context.Context, userId string) (dto.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (dto.UserResponse, error)
	UpdateStatusIsVerified(ctx context.Context, req dto.UpdateStatusIsVerifiedRequest, adminId string) (dto.UserResponse, error)
	CheckUser(ctx context.Context, email string) (bool, error)
	UpdateUser(ctx context.Context, req dto.UserUpdateRequest, userId string) error
	DeleteUser(ctx context.Context, userId string) error
	Verify(ctx context.Context, email string, password string) (bool, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(ur repository.UserRepository) UserService {
	return &userService{
		userRepo: ur,
	}
}

func (s *userService) RegisterUser(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error) {
	email, _ := s.userRepo.CheckEmail(ctx, req.Email)
	if email {
		return dto.UserResponse{}, dto.ErrEmailAlreadyExists
	}

	user := entities.User{
		Name:       req.Name,
		TelpNumber: req.TelpNumber,
		Role:       constants.ENUM_ROLE_USER,
		Email:      req.Email,
		Password:   req.Password,
		IsVerified: false,
	}

	userResponse, err := s.userRepo.RegisterUser(ctx, user)
	if err != nil {
		return dto.UserResponse{}, dto.ErrCreateUser
	}

	return dto.UserResponse{
		ID:         userResponse.ID.String(),
		Name:       userResponse.Name,
		TelpNumber: userResponse.TelpNumber,
		Role:       userResponse.Role,
		Email:      userResponse.Email,
		IsVerified: userResponse.IsVerified,
	}, nil
}

func (s *userService) GetAllUser(ctx context.Context) ([]dto.UserResponse, error) {
	users, err := s.userRepo.GetAllUser(ctx)
	if err != nil {
		return nil, dto.ErrGetAllUser
	}

	var userResponse []dto.UserResponse
	for _, user := range users {
		userResponse = append(userResponse, dto.UserResponse{
			ID:         user.ID.String(),
			Name:       user.Name,
			TelpNumber: user.TelpNumber,
			Role:       user.Role,
			Email:      user.Email,
			IsVerified: user.IsVerified,
		})
	}

	return userResponse, nil
}

func (s *userService) UpdateStatusIsVerified(ctx context.Context, req dto.UpdateStatusIsVerifiedRequest, adminId string) (dto.UserResponse, error) {
	admin, err := s.userRepo.GetUserById(ctx, adminId)
	if err != nil {
		return dto.UserResponse{}, dto.ErrUserNotFound
	}

	if admin.Role != constants.ENUM_ROLE_ADMIN {
		return dto.UserResponse{}, dto.ErrUserNotAdmin
	}

	user, err := s.userRepo.GetUserById(ctx, req.UserId)
	if err != nil {
		return dto.UserResponse{}, dto.ErrUserNotFound
	}

	userUpdate := entities.User{
		ID:         user.ID,
		IsVerified: req.IsVerified,
	}

	err = s.userRepo.UpdateUser(ctx, userUpdate)
	if err != nil {
		return dto.UserResponse{}, dto.ErrUpdateUser
	}

	return dto.UserResponse{
		ID:         user.ID.String(),
		Name:       user.Name,
		TelpNumber: user.TelpNumber,
		Role:       user.Role,
		Email:      user.Email,
		IsVerified: userUpdate.IsVerified,
	}, nil
}

func (s *userService) GetUserById(ctx context.Context, userId string) (dto.UserResponse, error) {
	user, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return dto.UserResponse{}, dto.ErrGetUserById
	}

	return dto.UserResponse{
		ID:         user.ID.String(),
		Name:       user.Name,
		TelpNumber: user.TelpNumber,
		Role:       user.Role,
		Email:      user.Email,
		IsVerified: user.IsVerified,
	}, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (dto.UserResponse, error) {
	emails, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return dto.UserResponse{}, dto.ErrGetUserByEmail
	}

	return dto.UserResponse{
		ID:         emails.ID.String(),
		Name:       emails.Name,
		TelpNumber: emails.TelpNumber,
		Role:       emails.Role,
		Email:      emails.Email,
		IsVerified: emails.IsVerified,
	}, nil
}

func (s *userService) CheckUser(ctx context.Context, email string) (bool, error) {
	res, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	if res.Email == "" {
		return false, err
	}
	return true, nil
}

func (s *userService) UpdateUser(ctx context.Context, req dto.UserUpdateRequest, userId string) error {
	user, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return dto.ErrUserNotFound
	}

	userUpdate := entities.User{
		ID:         user.ID,
		Name:       req.Name,
		TelpNumber: req.TelpNumber,
		Role:       user.Role,
		Email:      req.Email,
		Password:   req.Password,
		IsVerified: req.IsVerified,
	}

	err = s.userRepo.UpdateUser(ctx, userUpdate)
	if err != nil {
		return dto.ErrUpdateUser
	}

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, userId string) error {
	user, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return dto.ErrUserNotFound
	}

	err = s.userRepo.DeleteUser(ctx, user.ID.String())
	if err != nil {
		return dto.ErrDeleteUser
	}

	return nil
}

func (s *userService) Verify(ctx context.Context, email string, password string) (bool, error) {
	res, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return false, dto.ErrUserNotFound
	}

	if !res.IsVerified {
		return false, dto.ErrAccountNotVerified
	}

	checkPassword, err := helpers.CheckPassword(res.Password, []byte(password))
	if err != nil {
		return false, dto.ErrPasswordNotMatch
	}

	if res.Email == email && checkPassword {
		return true, nil
	}

	return false, dto.ErrEmailOrPassword
}
