package migrations

import (
	"errors"

	"github.com/Caknoooo/go-gin-clean-template/constants"
	"github.com/Caknoooo/go-gin-clean-template/entity"
	"gorm.io/gorm"
)

func Seeder(db *gorm.DB) error {
	if err := ListUserSeeder(db); err != nil {
		return err
	}

	return nil
}

func ListUserSeeder(db *gorm.DB) error {
	var listUser = []entity.User{
		{
			Name:       "Admin",
			TelpNumber: "081234567890",
			Email:      "admin@gmail.com",
			Password:   "admin123",
			Role:       constants.ENUM_ROLE_ADMIN,
			IsVerified: true,
		},
		{
			Name:       "User",
			TelpNumber: "081234567891",
			Email:      "user@gmail.com",
			Password:   "user123",
			Role:       constants.ENUM_ROLE_USER,
			IsVerified: true,
		},
	}

	hasTable := db.Migrator().HasTable(&entity.User{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entity.User{}); err != nil {
			return err
		}
	}

	for _, data := range listUser {
		var user entity.User
		err := db.Where(&entity.User{Email: data.Email}).First(&user).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		isData := db.Find(&user, "email = ?", data.Email).RowsAffected
		if isData == 0 {
			if err := db.Create(&data).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
