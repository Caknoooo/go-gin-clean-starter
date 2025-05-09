package entity

import (
	"github.com/Caknoooo/go-gin-clean-starter/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name       string    `gorm:"type:varchar(100);not null" json:"name" validate:"required,min=2,max=100"`
	Email      string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email" validate:"required,email"`
	TelpNumber string    `gorm:"type:varchar(20);index" json:"telp_number" validate:"omitempty,required,min=8,max=20"`
	Password   string    `gorm:"type:varchar(255);not null" json:"-" validate:"required,min=8"`
	Role       string    `gorm:"type:varchar(50);not null;default:'user'" json:"role" validate:"required,oneof=user admin"`
	ImageUrl   string    `gorm:"type:varchar(255)" json:"image_url" validate:"omitempty,url"`
	IsVerified bool      `gorm:"default:false" json:"is_verified"`

	Timestamp
}

// BeforeCreate hook to hash password and set defaults
func (u *User) BeforeCreate(_ *gorm.DB) (err error) {
	// Hash password
	if u.Password != "" {
		u.Password, err = helpers.HashPassword(u.Password)
		if err != nil {
			return err
		}
	}

	// Ensure UUID is set
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	// Set default role if not specified
	if u.Role == "" {
		u.Role = "user"
	}

	return nil
}

// BeforeUpdate hook to handle password updates
func (u *User) BeforeUpdate(_ *gorm.DB) (err error) {
	// Only hash password if it has been changed
	if u.Password != "" {
		u.Password, err = helpers.HashPassword(u.Password)
		if err != nil {
			return err
		}
	}
	return nil
}
