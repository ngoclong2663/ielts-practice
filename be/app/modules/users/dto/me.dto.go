package dto

import "time"

type UserProfileDto struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	AvatarURL    string    `json:"avatar_url"`
	UILocale     string    `json:"ui_locale"`
	Timezone     string    `json:"timezone"`
	DayStartHour int       `json:"day_start_hour"`
	CreatedAt    time.Time `json:"created_at"`
}

type StudySettingsDto struct {
	NewCardsPerDay   int     `json:"new_cards_per_day"`
	MaxReviewsPerDay int     `json:"max_reviews_per_day"`
	DesiredRetention float64 `json:"desired_retention"`
	LearningSteps    string  `json:"learning_steps"`
	RelearningSteps  string  `json:"relearning_steps"`
	AudioAutoplay    bool    `json:"audio_autoplay"`
	Accent           string  `json:"accent"`
	ShowImageOnFront bool    `json:"show_image_on_front"`
	MaxIntervalDays  int     `json:"max_interval_days"`
}

type MeResponseDto struct {
	User          UserProfileDto   `json:"user"`
	StudySettings StudySettingsDto `json:"study_settings"`
}

// UpdateProfileRequestDto: every field is optional; only provided fields change.
type UpdateProfileRequestDto struct {
	DisplayName  *string `json:"display_name" binding:"omitempty,min=1,max=120"`
	UILocale     *string `json:"ui_locale" binding:"omitempty,oneof=vi en"`
	Timezone     *string `json:"timezone" binding:"omitempty,max=64"`
	DayStartHour *int    `json:"day_start_hour" binding:"omitempty,min=0,max=23"`
}

type UpdateStudySettingsRequestDto struct {
	NewCardsPerDay   *int     `json:"new_cards_per_day" binding:"omitempty,min=0,max=500"`
	MaxReviewsPerDay *int     `json:"max_reviews_per_day" binding:"omitempty,min=0,max=9999"`
	DesiredRetention *float64 `json:"desired_retention" binding:"omitempty,min=0.7,max=0.99"`
	AudioAutoplay    *bool    `json:"audio_autoplay"`
	Accent           *string  `json:"accent" binding:"omitempty,oneof=uk us"`
	ShowImageOnFront *bool    `json:"show_image_on_front"`
	// Advanced (Anki-style) options.
	LearningSteps   *string `json:"learning_steps" binding:"omitempty,max=64"`
	RelearningSteps *string `json:"relearning_steps" binding:"omitempty,max=64"`
	MaxIntervalDays *int    `json:"max_interval_days" binding:"omitempty,min=1,max=36500"`
}
