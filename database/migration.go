package database

import (
	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entities.User{},
		&entities.RefreshToken{},
	); err != nil {
		return err
	}

	return nil
}
