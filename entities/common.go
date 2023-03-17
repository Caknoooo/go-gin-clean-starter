package entities

import (
	"time"
)

type Timestamp struct {
	CreatedAt time.Time `gorm:"type:timestamp with time zone" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone" json:"updated_at"`
	// DeletedAt gorm.DeletedAt
}

type Authorization struct {
	Token string `gorm:"type:varchar(255)" json:"token"`
	Role  string `gorm:"type:varchar(30)" json:"role"`
}