package models

import (
	"time"

	"gorm.io/gorm"
)

type URL struct {
	gorm.Model
	OriginalURL  string       `json:"original_url" validate:"required,url"`
	ShortCode    string       `json:"short_code" gorm:"unique;not null"`
	UserID       *uint        `json:"user_id" gorm:"index"`
	SessionID    *uint        `json:"session_id" gorm:"index"`
	Clicks       int          `json:"clicks" gorm:"default:0;check:clicks >= 0"`
	ExpiresAt    *time.Time   `json:"expires_at"`
	User         User         `gorm:"foreignKey:UserID"`
	GuestSession GuestSession `gorm:"foreignKey:SessionID"`
	ClicksData   []Click      `gorm:"foreignKey:URLID"`
}
