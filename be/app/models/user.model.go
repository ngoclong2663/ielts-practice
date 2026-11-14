package models

import "time"

type User struct {
	ID           string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	GoogleSub    string     `gorm:"uniqueIndex" json:"-"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	AvatarURL    string     `json:"avatar_url"`
	UILocale     string     `gorm:"column:ui_locale" json:"ui_locale"`
	Timezone     string     `json:"timezone"`
	DayStartHour int16      `json:"day_start_hour"`
	LastLoginAt  *time.Time `json:"last_login_at"`

	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (User) TableName() string { return "users" }

type UserStudySettings struct {
	UserID           string  `gorm:"primaryKey;type:uuid"`
	NewCardsPerDay   int     `gorm:"column:new_cards_per_day"`
	MaxReviewsPerDay int     `gorm:"column:max_reviews_per_day"`
	DesiredRetention float64 `gorm:"column:desired_retention"`
	LearningSteps    string  `gorm:"column:learning_steps"`
	RelearningSteps  string  `gorm:"column:relearning_steps"`
	AudioAutoplay    bool    `gorm:"column:audio_autoplay"`
	Accent           string  `gorm:"column:accent"`
	ShowImageOnFront bool    `gorm:"column:show_image_on_front"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (UserStudySettings) TableName() string { return "user_study_settings" }
