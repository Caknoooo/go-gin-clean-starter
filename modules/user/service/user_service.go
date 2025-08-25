package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	authDto "github.com/Caknoooo/go-gin-clean-starter/modules/auth/dto"
	authRepo "github.com/Caknoooo/go-gin-clean-starter/modules/auth/repository"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/repository"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	commonDto "github.com/Caknoooo/go-gin-clean-starter/pkg/dto"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/helpers"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	Register(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error)
	GetAllUserWithPagination(ctx context.Context, req commonDto.PaginationRequest) (dto.UserPaginationResponse, error)
	GetUserById(ctx context.Context, userId string) (dto.UserResponse, error)
	Verify(ctx context.Context, req dto.UserLoginRequest) (authDto.TokenResponse, error)
	SendVerificationEmail(ctx context.Context, req dto.SendVerificationEmailRequest) error
	VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (dto.VerifyEmailResponse, error)
	Update(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error)
	Delete(ctx context.Context, userId string) error
	RefreshToken(ctx context.Context, req authDto.RefreshTokenRequest) (authDto.TokenResponse, error)
}

type userService struct {
	userRepository         repository.UserRepository
	refreshTokenRepository authRepo.RefreshTokenRepository
	jwtService             authService.JWTService
	db                     *gorm.DB
}

func NewUserService(
	userRepo repository.UserRepository,
	refreshTokenRepo authRepo.RefreshTokenRepository,
	jwtService authService.JWTService,
	db *gorm.DB,
) UserService {
	return &userService{
		userRepository:         userRepo,
		refreshTokenRepository: refreshTokenRepo,
		jwtService:             jwtService,
		db:                     db,
	}
}

func (s *userService) Register(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error) {
	_, exists, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return dto.UserResponse{}, err
	}
	if exists {
		return dto.UserResponse{}, dto.ErrEmailAlreadyExists
	}

	user := entities.User{
		ID:         uuid.New(),
		Name:       req.Name,
		Email:      req.Email,
		TelpNumber: req.TelpNumber,
		Password:   req.Password,
		Role:       constants.ENUM_ROLE_USER,
		IsVerified: false,
	}

	createdUser, err := s.userRepository.Register(ctx, s.db, user)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:         createdUser.ID.String(),
		Name:       createdUser.Name,
		Email:      createdUser.Email,
		TelpNumber: createdUser.TelpNumber,
		Role:       createdUser.Role,
		ImageUrl:   createdUser.ImageUrl,
		IsVerified: createdUser.IsVerified,
	}, nil
}

func (s *userService) GetAllUserWithPagination(ctx context.Context, req commonDto.PaginationRequest) (dto.UserPaginationResponse, error) {
	result, err := s.userRepository.GetAllUserWithPagination(ctx, s.db, req)
	if err != nil {
		return dto.UserPaginationResponse{}, err
	}

	var userResponses []dto.UserResponse
	for _, user := range result.Users {
		userResponses = append(userResponses, dto.UserResponse{
			ID:         user.ID.String(),
			Name:       user.Name,
			Email:      user.Email,
			TelpNumber: user.TelpNumber,
			Role:       user.Role,
			ImageUrl:   user.ImageUrl,
			IsVerified: user.IsVerified,
		})
	}

	return dto.UserPaginationResponse{
		Data:               userResponses,
		PaginationResponse: result.PaginationResponse,
	}, nil
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

func (s *userService) Verify(ctx context.Context, req dto.UserLoginRequest) (authDto.TokenResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return authDto.TokenResponse{}, dto.ErrEmailNotFound
	}

	isValid, err := helpers.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !isValid {
		return authDto.TokenResponse{}, dto.ErrUserNotFound
	}

	accessToken := s.jwtService.GenerateAccessToken(user.ID.String(), user.Role)
	refreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	refreshToken := entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: expiresAt,
	}

	_, err = s.refreshTokenRepository.Create(ctx, s.db, refreshToken)
	if err != nil {
		return authDto.TokenResponse{}, err
	}

	return authDto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		Role:         user.Role,
	}, nil
}

func (s *userService) SendVerificationEmail(ctx context.Context, req dto.SendVerificationEmailRequest) error {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return dto.ErrEmailNotFound
	}

	if user.IsVerified {
		return dto.ErrAccountAlreadyVerified
	}

	verificationToken := s.jwtService.GenerateAccessToken(user.ID.String(), "verification")

	subject := "Email Verification"
	body := "Please verify your email using this token: " + verificationToken

	return utils.SendMail(user.Email, subject, body)
}

func (s *userService) VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (dto.VerifyEmailResponse, error) {
	token, err := s.jwtService.ValidateToken(req.Token)
	if err != nil || !token.Valid {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	userId, err := s.jwtService.GetUserIDByToken(req.Token)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrUserNotFound
	}

	user.IsVerified = true
	updatedUser, err := s.userRepository.Update(ctx, s.db, user)
	if err != nil {
		return dto.VerifyEmailResponse{}, err
	}

	return dto.VerifyEmailResponse{
		Email:      updatedUser.Email,
		IsVerified: updatedUser.IsVerified,
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

func (s *userService) RefreshToken(ctx context.Context, req authDto.RefreshTokenRequest) (authDto.TokenResponse, error) {
	refreshToken, err := s.refreshTokenRepository.FindByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return authDto.TokenResponse{}, err
	}

	accessToken := s.jwtService.GenerateAccessToken(refreshToken.UserID.String(), refreshToken.User.Role)
	newRefreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	err = s.refreshTokenRepository.DeleteByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return authDto.TokenResponse{}, err
	}

	newRefreshToken := entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    refreshToken.UserID,
		Token:     newRefreshTokenString,
		ExpiresAt: expiresAt,
	}

	_, err = s.refreshTokenRepository.Create(ctx, s.db, newRefreshToken)
	if err != nil {
		return authDto.TokenResponse{}, err
	}

	return authDto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenString,
		Role:         refreshToken.User.Role,
	}, nil
}
