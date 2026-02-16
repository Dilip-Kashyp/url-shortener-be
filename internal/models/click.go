package models

import (
	"time"

	"gorm.io/gorm"
)

type Click struct {
	gorm.Model
	URLID     uint      `json:"url_id" gorm:"not null;index"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP"`
	URL       URL       `gorm:"foreignKey:URLID"`
}
