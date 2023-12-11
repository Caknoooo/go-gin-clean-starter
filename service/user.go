package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/Caknoooo/go-gin-clean-template/constants"
	"github.com/Caknoooo/go-gin-clean-template/dto"
	"github.com/Caknoooo/go-gin-clean-template/entity"
	"github.com/Caknoooo/go-gin-clean-template/helpers"
	"github.com/Caknoooo/go-gin-clean-template/repository"
	"github.com/Caknoooo/go-gin-clean-template/utils"
	"github.com/google/uuid"
)

type UserService interface {
	RegisterUser(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error)
	GetAllUserWithPagination(ctx context.Context, req dto.PaginationRequest) (dto.UserPaginationResponse, error)
	GetUserById(ctx context.Context, userId string) (dto.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (dto.UserResponse, error)
	UpdateStatusIsVerified(ctx context.Context, req dto.UpdateStatusIsVerifiedRequest, adminId string) (dto.UserResponse, error)
	SendVerificationEmail(ctx context.Context, req dto.SendVerificationEmailRequest) error
	VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (dto.VerifyEmailResponse, error)
	CheckUser(ctx context.Context, email string) (bool, error)
	UpdateUser(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error)
	DeleteUser(ctx context.Context, userId string) error
	Verify(ctx context.Context, email string, password string) (bool, error)
}

const (
	LOCAL_URL          = "http://localhost:3000"
	VERIFY_EMAIL_ROUTE = "register/verify_email"
)

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

	imageId := uuid.New()
	ext := utils.GetExtensions(req.Image.Filename)

	filename := fmt.Sprintf("profile/%s.%s", imageId, ext)
	if err := utils.UploadFile(req.Image, filename); err != nil {
		return dto.UserResponse{}, err
	}

	user := entity.User{
		Name:       req.Name,
		TelpNumber: req.TelpNumber,
		ImageUrl:   filename,
		Role:       constants.ENUM_ROLE_USER,
		Email:      req.Email,
		Password:   req.Password,
		IsVerified: false,
	}

	userReg, err := s.userRepo.RegisterUser(ctx, user)
	if err != nil {
		return dto.UserResponse{}, dto.ErrCreateUser
	}

	draftEmail, err := makeVerificationEmail(userReg.Email)
	if err != nil {
		return dto.UserResponse{}, err
	}

	err = utils.SendMail(userReg.Email, draftEmail["subject"], draftEmail["body"])
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:         userReg.ID.String(),
		Name:       userReg.Name,
		TelpNumber: userReg.TelpNumber,
		ImageUrl:   userReg.ImageUrl,
		Role:       userReg.Role,
		Email:      userReg.Email,
		IsVerified: userReg.IsVerified,
	}, nil
}

func makeVerificationEmail(receiverEmail string) (map[string]string, error) {
	expired := time.Now().Add(time.Hour * 24).Format("2006-01-02 15:04:05")
	plainText := receiverEmail + "_" + expired
	token, err := utils.AESEncrypt(plainText)
	if err != nil {
		return nil, err
	}

	verifyLink := LOCAL_URL + "/" + VERIFY_EMAIL_ROUTE + "?token=" + token

	readHtml, err := os.ReadFile("utils/email-template/base_mail.html")
	if err != nil {
		return nil, err
	}

	data := struct {
		Email  string
		Verify string
	}{
		Email:  receiverEmail,
		Verify: verifyLink,
	}

	tmpl, err := template.New("custom").Parse(string(readHtml))
	if err != nil {
		return nil, err
	}

	var strMail bytes.Buffer
	if err := tmpl.Execute(&strMail, data); err != nil {
		return nil, err
	}

	draftEmail := map[string]string{
		"subject": "Cakno - Go Gin Template",
		"body":    strMail.String(),
	}

	return draftEmail, nil
}

func (s *userService) SendVerificationEmail(ctx context.Context, req dto.SendVerificationEmailRequest) error {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return dto.ErrEmailNotFound
	}

	draftEmail, err := makeVerificationEmail(user.Email)
	if err != nil {
		return err
	}

	err = utils.SendMail(user.Email, draftEmail["subject"], draftEmail["body"])
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (dto.VerifyEmailResponse, error) {
	decryptedToken, err := utils.AESDecrypt(req.Token)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	if !strings.Contains(decryptedToken, "_") {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	decryptedTokenSplit := strings.Split(decryptedToken, "_")
	email := decryptedTokenSplit[0]
	expired := decryptedTokenSplit[1]

	now := time.Now()
	expiredTime, err := time.Parse("2006-01-02 15:04:05", expired)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	if expiredTime.Sub(now) < 0 {
		return dto.VerifyEmailResponse{
			Email:      email,
			IsVerified: false,
		}, dto.ErrTokenExpired
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrUserNotFound
	}

	if user.IsVerified {
		return dto.VerifyEmailResponse{}, dto.ErrAccountAlreadyVerified
	}

	updatedUser, err := s.userRepo.UpdateUser(ctx, entity.User{
		ID:         user.ID,
		IsVerified: true,
	})
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrUpdateUser
	}

	return dto.VerifyEmailResponse{
		Email:      email,
		IsVerified: updatedUser.IsVerified,
	}, nil
}

func (s *userService) GetAllUserWithPagination(ctx context.Context, req dto.PaginationRequest) (dto.UserPaginationResponse, error) {
	dataWithPaginate, err := s.userRepo.GetAllUserWithPagination(ctx, nil, req)
	if err != nil {
		return dto.UserPaginationResponse{}, err
	}

	var datas []dto.UserResponse
	for _, user := range dataWithPaginate.Users {
		data := dto.UserResponse{
			ID:         user.ID.String(),
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
			TelpNumber: user.TelpNumber,
			IsVerified: user.IsVerified,
		}

		datas = append(datas, data)
	}

	return dto.UserPaginationResponse{
		Data: datas,
		PaginationResponse: dto.PaginationResponse{
			Page:    dataWithPaginate.Page,
			PerPage: dataWithPaginate.PerPage,
			MaxPage: dataWithPaginate.MaxPage,
			Count:   dataWithPaginate.Count,
		},
	}, nil
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

	data := entity.User{
		ID:         user.ID,
		IsVerified: req.IsVerified,
	}

	userUpdate, err := s.userRepo.UpdateUser(ctx, data)
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

func (s *userService) UpdateUser(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error) {
	user, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return dto.UserUpdateResponse{}, dto.ErrUserNotFound
	}

	data := entity.User{
		ID:         user.ID,
		Name:       req.Name,
		TelpNumber: req.TelpNumber,
		Role:       user.Role,
		Email:      req.Email,
		Password:   req.Password,
	}

	userUpdate, err := s.userRepo.UpdateUser(ctx, data)
	if err != nil {
		return dto.UserUpdateResponse{}, dto.ErrUpdateUser
	}

	return dto.UserUpdateResponse{
		ID:         userUpdate.ID.String(),
		Name:       userUpdate.Name,
		TelpNumber: userUpdate.TelpNumber,
		Role:       userUpdate.Role,
		Email:      userUpdate.Email,
		IsVerified: userUpdate.IsVerified,
	}, nil
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
