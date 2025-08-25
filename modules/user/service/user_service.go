package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	authDto "github.com/Caknoooo/go-gin-clean-starter/modules/auth/dto"
	authRepo "github.com/Caknoooo/go-gin-clean-starter/modules/auth/repository"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/repository"
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
	// Check if email already exists
	_, exists, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return dto.UserResponse{}, err
	}
	if exists {
		return dto.UserResponse{}, dto.ErrEmailAlreadyExists
	}

	// Create user entity
	user := entities.User{
		ID:         uuid.New(),
		Name:       req.Name,
		Email:      req.Email,
		TelpNumber: req.TelpNumber,
		Password:   req.Password, // Will be hashed in BeforeCreate hook
		Role:       "user",
		IsVerified: false,
	}

	// Handle image upload if provided
	if req.Image != nil {
		// Handle image upload logic here
		user.ImageUrl = "" // Set image URL after upload
	}

	// Save user
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

	// Check password
	isValid, err := helpers.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !isValid {
		return authDto.TokenResponse{}, dto.ErrUserNotFound
	}

	// Generate tokens
	accessToken := s.jwtService.GenerateAccessToken(user.ID.String(), user.Role)
	refreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	// Save refresh token
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

	// Generate verification token
	verificationToken := s.jwtService.GenerateAccessToken(user.ID.String(), "verification")

	// Send email (implement email sending logic)
	subject := "Email Verification"
	body := "Please verify your email using this token: " + verificationToken

	return utils.SendMail(user.Email, subject, body)
}

func (s *userService) VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (dto.VerifyEmailResponse, error) {
	// Validate token
	token, err := s.jwtService.ValidateToken(req.Token)
	if err != nil || !token.Valid {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	// Get user ID from token
	userId, err := s.jwtService.GetUserIDByToken(req.Token)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	// Get user
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrUserNotFound
	}

	// Update verification status
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
	// Get existing user
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.UserUpdateResponse{}, dto.ErrUserNotFound
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.TelpNumber != "" {
		user.TelpNumber = req.TelpNumber
	}

	// Save updates
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
	// Find refresh token
	refreshToken, err := s.refreshTokenRepository.FindByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return authDto.TokenResponse{}, err
	}

	// Generate new tokens
	accessToken := s.jwtService.GenerateAccessToken(refreshToken.UserID.String(), refreshToken.User.Role)
	newRefreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	// Delete old refresh token
	err = s.refreshTokenRepository.DeleteByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return authDto.TokenResponse{}, err
	}

	// Create new refresh token
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
