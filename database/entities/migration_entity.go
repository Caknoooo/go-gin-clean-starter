package entities

import (
	"time"
)

type Migration struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	Batch     int       `gorm:"not null;index" json:"batch"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone" json:"created_at"`
}
