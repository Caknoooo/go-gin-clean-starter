package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-template/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, user entity.User) (entity.User, error)
	GetAllUser(ctx context.Context) ([]entity.User, error)
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

func (r *userRepository) GetAllUser(ctx context.Context) ([]entity.User, error) {
	var user []entity.User
	if err := r.db.Find(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
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
