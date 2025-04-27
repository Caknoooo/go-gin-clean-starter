package entity

import (
	"github.com/Caknoooo/go-gin-clean-starter/helpers"
	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name       string    `gorm:"type:varchar(100);not null" json:"name" validate:"required,min=2,max=100"`
	Email      string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email" validate:"required,email"`
	TelpNumber string    `gorm:"type:varchar(20);index" json:"telp_number" validate:"omitempty,required,min=8,max=20"`
	Password   string    `gorm:"type:varchar(255);not null" json:"-" validate:"required,min=8"`
	Role       string    `gorm:"type:varchar(50);not null;default:'user'" json:"role" validate:"required,oneof=user admin"`
	ImageURL   string    `gorm:"type:varchar(255)" json:"image_url" validate:"omitempty,url"`
	IsVerified bool      `gorm:"default:false" json:"is_verified"`

	Timestamp
}

func (u *User) BeforeCreate() error {
	var err error
	u.Password, err = helpers.HashPassword(u.Password)
	if err != nil {
		return err
	}
	return nil
}
