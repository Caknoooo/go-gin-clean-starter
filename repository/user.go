package repository

import (
	"context"
	"math"

	"github.com/Caknoooo/go-gin-clean-template/dto"
	"github.com/Caknoooo/go-gin-clean-template/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, user entity.User) (entity.User, error)
	GetAllUserWithPagination(ctx context.Context, tx *gorm.DB, req dto.PaginationRequest) (dto.GetAllUserRepositoryResponse, error)
	GetUserById(ctx context.Context, userId string) (entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (entity.User, error)
	CheckEmail(ctx context.Context, email string) (bool, error)
	UpdateUser(ctx context.Context, user entity.User) (entity.User, error)
	DeleteUser(ctx context.Context, userId string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) RegisterUser(ctx context.Context, user entity.User) (entity.User, error) {
	if err := r.db.Create(&user).Error; err != nil {
		return entity.User{}, err
	}
	return user, nil
}

func (r *userRepository) GetAllUserWithPagination(ctx context.Context, tx *gorm.DB, req dto.PaginationRequest) (dto.GetAllUserRepositoryResponse, error) {
	if tx == nil {
		tx = r.db
	}

	var users []entity.User
	var err error
	var count int64

	if req.PerPage == 0 {
		req.PerPage = 10
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.Search != "" {
		err = tx.WithContext(ctx).Model(&entity.User{}).Where("name ILIKE ?", "%"+req.Search+"%").Count(&count).Error
		if err != nil {
			return dto.GetAllUserRepositoryResponse{}, err
		}
	} else {
		err = tx.WithContext(ctx).Model(&entity.User{}).Count(&count).Error
		if err != nil {
			return dto.GetAllUserRepositoryResponse{}, err
		}
	}

	stmt := tx.WithContext(ctx).Where("name ILIKE ?", "%"+req.Search+"%")
	maxPage := int64(math.Ceil(float64(count) / float64(req.PerPage)))

	offset := (req.Page - 1) * req.PerPage
	_ = stmt.Offset(offset).Limit(req.PerPage).Find(&users).Error

	return dto.GetAllUserRepositoryResponse{
		Users: users,
		PaginationResponse: dto.PaginationResponse{
			Page:    req.Page,
			PerPage: req.PerPage,
			MaxPage: maxPage,
			Count:   count,
		},
	}, nil
}

func (r *userRepository) GetUserById(ctx context.Context, userId string) (entity.User, error) {
	var user entity.User
	if err := r.db.Where("id = ?", userId).Take(&user).Error; err != nil {
		return entity.User{}, err
	}
	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (entity.User, error) {
	var user entity.User
	if err := r.db.Where("email = ?", email).Take(&user).Error; err != nil {
		return entity.User{}, err
	}
	return user, nil
}

func (r *userRepository) CheckEmail(ctx context.Context, email string) (bool, error) {
	var user entity.User
	if err := r.db.Where("email = ?", email).Take(&user).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user entity.User) (entity.User, error) {
	if err := r.db.Updates(&user).Error; err != nil {
		return entity.User{}, err
	}
	return user, nil
}

func (r *userRepository) DeleteUser(ctx context.Context, userId string) error {
	if err := r.db.Delete(&entity.User{}, &userId).Error; err != nil {
		return err
	}
	return nil
}
