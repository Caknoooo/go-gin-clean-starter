package seeds

import (
	"encoding/json"
	"github.com/Caknoooo/go-gin-clean-starter/helpers"
	"io"
	"os"
	"path"

	"github.com/Caknoooo/go-gin-clean-starter/dto"
	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"gorm.io/gorm"
)

func ListUserSeeder(db *gorm.DB) error {
	projectDir, err := helpers.GetProjectRoot()
	if err != nil {
		return err
	}

	jsonFilePath := path.Join(projectDir, "migrations/json/users.json")
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return err
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			panic(err)
		}
	}(jsonFile)

	// Extend UserCreateRequest to include Role and IsVerified
	type SeedUserRequest struct {
		dto.UserCreateRequest
		Role       string `json:"role" binding:"required,oneof=user admin"`
		IsVerified bool   `json:"is_verified"`
	}

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var seedUsers []SeedUserRequest
	if err := json.Unmarshal(jsonData, &seedUsers); err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entity.User{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entity.User{}); err != nil {
			return err
		}
	}

	for _, seedUser := range seedUsers {
		// Convert SeedUserRequest to entity.User
		user := entity.User{
			Name:       seedUser.Name,
			TelpNumber: seedUser.TelpNumber,
			Email:      seedUser.Email,
			Password:   seedUser.Password,
			Role:       seedUser.Role,
			IsVerified: seedUser.IsVerified,
		}

		// Check if user already exists
		var existingUser entity.User
		isData := db.Where("email = ?", user.Email).Find(&existingUser).RowsAffected

		if isData == 0 {
			if err := db.Create(&user).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
