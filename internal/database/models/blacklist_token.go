package models

import "time"

type BlacklistToken struct {
	ID        uint      `gorm:"primaryKey"`
	Token     string    `gorm:"type:varchar(500);uniqueIndex"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}
