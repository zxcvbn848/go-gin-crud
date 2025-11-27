package models

import "time"

type RefreshToken struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Token     string `gorm:"type:varchar(500);uniqueIndex"`
	ExpiresAt time.Time
}
