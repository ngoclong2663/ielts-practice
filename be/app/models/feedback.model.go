package models

import "time"

type Feedback struct {
	ID         string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID     string `gorm:"type:uuid"`
	Category   string
	Message    string
	Page       string
	UILocale   string `gorm:"column:ui_locale"`
	UserAgent  string
	Status     string `gorm:"default:new"`
	AdminNote  string
	NotifiedAt *time.Time
	ResolvedAt *time.Time
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

func (Feedback) TableName() string { return "feedback" }
