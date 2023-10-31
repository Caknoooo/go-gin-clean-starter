package entity

import (
	"time"
	"gorm.io/gorm"
)

type Timestamp struct {
	CreatedAt time.Time `gorm:"type:timestamp with time zone" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone" json:"updated_at"`
	DeletedAt gorm.DeletedAt
}

type Authorization struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}