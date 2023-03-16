package repository

import (
	"context"

	"github.com/Caknoooo/golang-clean_template/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, user entities.User) (entities.User, error)
	GetAllUser(ctx context.Context) ([]entities.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (entities.User, error)
	UpdateUser(ctx context.Context, user entities.User) (error)
	DeleteUser(ctx context.Context, userID uuid.UUID) (error) 
}

type userRepository struct {
	connection *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository{
	return &userRepository{
		connection: db,
	}
}

func (ur *userRepository) RegisterUser(ctx context.Context, user entities.User) (entities.User, error){
	if err := ur.connection.Create(&user).Error; err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (ur *userRepository) GetAllUser(ctx context.Context) ([]entities.User, error){
	var user []entities.User
	if err := ur.connection.Find(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *userRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (entities.User, error){
	var user entities.User
	if err := ur.connection.Where("id = ?", userID).Take(&user).Error; err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (ur *userRepository) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	var user entities.User
	if err := ur.connection.Where("email = ?", email).Take(&user).Error; err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (ur *userRepository) UpdateUser(ctx context.Context, user entities.User) (error) {
	if err := ur.connection.Updates(&user).Error; err != nil {
		return err
	}
	return nil
}

func (ur *userRepository) DeleteUser(ctx context.Context, userID uuid.UUID) (error) {
	if err := ur.connection.Delete(&entities.User{}, &userID).Error; err != nil {
		return err
	}
	return nil
}