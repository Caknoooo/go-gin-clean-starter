package repository

import (
	"context"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, tx *gorm.DB, token entities.RefreshToken) (entities.RefreshToken, error)
	FindByToken(ctx context.Context, tx *gorm.DB, token string) (entities.RefreshToken, error)
	DeleteByUserID(ctx context.Context, tx *gorm.DB, userID string) error
	DeleteByToken(ctx context.Context, tx *gorm.DB, token string) error
	DeleteExpired(ctx context.Context, tx *gorm.DB) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{
		db: db,
	}
}

func (r *refreshTokenRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	token entities.RefreshToken,
) (entities.RefreshToken, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&token).Error; err != nil {
		return entities.RefreshToken{}, err
	}

	return token, nil
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, tx *gorm.DB, token string) (
	entities.RefreshToken,
	error,
) {
	if tx == nil {
		tx = r.db
	}

	var refreshToken entities.RefreshToken
	if err := tx.WithContext(ctx).Where("token = ?", token).Preload("User").Take(&refreshToken).Error; err != nil {
		return entities.RefreshToken{}, err
	}

	return refreshToken, nil
}

func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, tx *gorm.DB, userID string) error {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Where("user_id = ?", userID).Delete(&entities.RefreshToken{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *refreshTokenRepository) DeleteByToken(ctx context.Context, tx *gorm.DB, token string) error {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Where("token = ?", token).Delete(&entities.RefreshToken{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context, tx *gorm.DB) error {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&entities.RefreshToken{}).Error; err != nil {
		return err
	}

	return nil
}
