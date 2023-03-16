package dto

import (
	"github.com/google/uuid"
)

type UserCreateDTO struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Nama     string    `gorm:"type:varchar(100)" form:"nama" json:"nama" binding:"required"`
	NoTelp   string    `gorm:"type:varchar(20)" form:"no_telp" json:"no_telp" binding:"required"`
	Email    string    `gorm:"type:varchar(100)" form:"email" json:"email" binding:"required"`
	Password string    `gorm:"type:varchar(100)" form:"password" json:"password" binding:"required"`
}

type UserUpdateDTO struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Nama     string    `gorm:"type:varchar(100)" form:"nama" json:"nama" binding:"required"`
	NoTelp   string    `gorm:"type:varchar(20)" form:"no_telp" json:"no_telp" binding:"required"`
	Email    string    `gorm:"type:varchar(100)" form:"email" json:"email" binding:"required"`
	Password string    `gorm:"type:varchar(100)" form:"password" json:"password" binding:"required"`
}

type UserLoginDTO struct {
	Email    string `json:"email" binding:"email" form:"email"`
	Password string `json:"password" binding:"required" form:"password"`
}
