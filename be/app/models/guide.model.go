package models

import "time"

// GuideVideo is one guide video (Epic 7 §5), loaded by `api seed`.
type GuideVideo struct {
	ID              string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Slug            string
	Topic           string
	Sort            int16
	Lang            string
	Title           []byte `gorm:"type:jsonb"`
	Description     []byte `gorm:"type:jsonb"`
	VideoKey        string
	PosterKey       string
	DurationSeconds float64
	Status          string
	ContentHash     string
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (GuideVideo) TableName() string { return "guide_videos" }
