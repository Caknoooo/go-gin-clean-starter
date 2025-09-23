package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/repository"
	"gorm.io/gorm"
)

type UserService interface {
	GetUserById(ctx context.Context, userId string) (dto.UserResponse, error)
	Update(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error)
	Delete(ctx context.Context, userId string) error
}

type userService struct {
	userRepository repository.UserRepository
	db             *gorm.DB
}

func NewUserService(
	userRepo repository.UserRepository,
	db *gorm.DB,
) UserService {
	return &userService{
		userRepository: userRepo,
		db:             db,
	}
}

func (s *userService) GetUserById(ctx context.Context, userId string) (dto.UserResponse, error) {
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:         user.ID.String(),
		Name:       user.Name,
		Email:      user.Email,
		TelpNumber: user.TelpNumber,
		Role:       user.Role,
		ImageUrl:   user.ImageUrl,
		IsVerified: user.IsVerified,
	}, nil
}

func (s *userService) Update(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error) {
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.UserUpdateResponse{}, dto.ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.TelpNumber != "" {
		user.TelpNumber = req.TelpNumber
	}

	updatedUser, err := s.userRepository.Update(ctx, s.db, user)
	if err != nil {
		return dto.UserUpdateResponse{}, err
	}

	return dto.UserUpdateResponse{
		ID:         updatedUser.ID.String(),
		Name:       updatedUser.Name,
		TelpNumber: updatedUser.TelpNumber,
		Role:       updatedUser.Role,
		Email:      updatedUser.Email,
		IsVerified: updatedUser.IsVerified,
	}, nil
}

func (s *userService) Delete(ctx context.Context, userId string) error {
	return s.userRepository.Delete(ctx, s.db, userId)
}
