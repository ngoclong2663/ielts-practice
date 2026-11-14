package models

import "time"

type RefreshToken struct {
	ID         string `gorm:"primaryKey;type:uuid"`
	UserID     string `gorm:"type:uuid;index"`
	TokenHash  string `gorm:"uniqueIndex"`
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy *string `gorm:"type:uuid"`
	UserAgent  string
	IP         string    `gorm:"column:ip"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
