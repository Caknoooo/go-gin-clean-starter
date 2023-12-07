package repository

import (
	"context"
	"math"

	"github.com/Caknoooo/go-gin-clean-template/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, user entity.User) (entity.User, error)
	GetAllUserWithPagination(ctx context.Context, tx *gorm.DB, search string, perPage int, page int) ([]entity.User, int64, int64, error)
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

func (r *userRepository) GetAllUserWithPagination(ctx context.Context, tx *gorm.DB, search string, perPage int, page int) ([]entity.User, int64, int64, error) {
	if tx == nil {
		tx = r.db
	}

	var users []entity.User
	var err error
	var count int64

	if search != "" {
		err = tx.WithContext(ctx).Model(&entity.User{}).Where("name ILIKE ?", "%"+search+"%").Count(&count).Error
		if err != nil {
			return nil, 0, 0, err
		}
	} else {
		err = tx.WithContext(ctx).Model(&entity.User{}).Count(&count).Error
		if err != nil {
			return nil, 0, 0, err
		}
	}

	stmt := tx.WithContext(ctx).Where("name ILIKE ?", "%"+search+"%")
	maxPage := int64(math.Ceil(float64(count) / float64(perPage)))

	if perPage <= 0 {
		stmt.Find(&users)
		return users, maxPage, count, nil
	}

	offset := (page - 1) * perPage
	_ = stmt.Offset(offset).Limit(perPage).Find(&users).Error

	return users, maxPage, count, nil
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
