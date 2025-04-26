package entity

import (
	"fmt"
	"log"

	"github.com/Caknoooo/go-gin-clean-starter/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name       string    `json:"name"`
	TelpNumber string    `json:"telp_number"`
	Email      string    `json:"email"`
	Password   string    `json:"password"`
	Role       string    `json:"role"`
	ImageUrl   string    `json:"image_url"`
	IsVerified bool      `json:"is_verified"`

	Timestamp
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic recovered in BeforeCreate: %v", r)
			err = fmt.Errorf("internal server error")
		}
	}()

	if u.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	hashedPassword, hashErr := helpers.HashPassword(u.Password)
	if hashErr != nil {
		return fmt.Errorf("failed to hash password: %w", hashErr)
	}
	u.Password = hashedPassword
	return nil
}
