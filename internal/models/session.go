package models

import (
	"time"

	"gorm.io/gorm"
)

type GuestSession struct {
	ID           uint           `gorm:"primaryKey"`
	Token        string         `gorm:"uniqueIndex;not null"`
	ExpiresAt    time.Time      `gorm:"column:expires_at;not null"`
	LastAccessed *time.Time     `gorm:"column:last_accessed"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}