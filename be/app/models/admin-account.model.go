package models

import "time"

type AdminAccount struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string
	PasswordHash string
	LastLoginAt  *time.Time
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (AdminAccount) TableName() string { return "admin_accounts" }
