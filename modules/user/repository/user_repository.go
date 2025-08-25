package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	commonDto "github.com/Caknoooo/go-gin-clean-starter/pkg/dto"
	"gorm.io/gorm"
)

type (
	UserRepository interface {
		Register(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error)
		GetAllUserWithPagination(
			ctx context.Context,
			tx *gorm.DB,
			req commonDto.PaginationRequest,
		) (dto.GetAllUserRepositoryResponse, error)
		GetUserById(ctx context.Context, tx *gorm.DB, userId string) (entities.User, error)
		GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, error)
		CheckEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, bool, error)
		Update(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error)
		Delete(ctx context.Context, tx *gorm.DB, userId string) error
	}

	userRepository struct {
		db *gorm.DB
	}
)

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Register(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetAllUserWithPagination(
	ctx context.Context,
	tx *gorm.DB,
	req commonDto.PaginationRequest,
) (dto.GetAllUserRepositoryResponse, error) {
	if tx == nil {
		tx = r.db
	}

	var users []entities.User
	var err error
	var count int64

	req.Default()

	query := tx.WithContext(ctx).Model(&entities.User{})
	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}

	if err := query.Count(&count).Error; err != nil {
		return dto.GetAllUserRepositoryResponse{}, err
	}

	if err := query.Scopes(Paginate(req)).Find(&users).Error; err != nil {
		return dto.GetAllUserRepositoryResponse{}, err
	}

	totalPage := TotalPage(count, int64(req.PerPage))
	return dto.GetAllUserRepositoryResponse{
		Users: users,
		PaginationResponse: commonDto.PaginationResponse{
			Page:    req.Page,
			PerPage: req.PerPage,
			Count:   count,
			MaxPage: totalPage,
		},
	}, err
}

func (r *userRepository) GetUserById(ctx context.Context, tx *gorm.DB, userId string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("id = ?", userId).Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) CheckEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, bool, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		return entities.User{}, false, err
	}

	return user, true, nil
}

func (r *userRepository) Update(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Updates(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) Delete(ctx context.Context, tx *gorm.DB, userId string) error {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Delete(&entities.User{}, "id = ?", userId).Error; err != nil {
		return err
	}

	return nil
}

// Helper functions
func Paginate(req commonDto.PaginationRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(req.GetOffset()).Limit(req.GetLimit())
	}
}

func TotalPage(totalData int64, limit int64) int64 {
	if totalData%limit == 0 {
		return totalData / limit
	}
	return (totalData / limit) + 1
}
