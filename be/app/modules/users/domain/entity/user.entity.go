package entity

// Entities (have an identity): UserEntity, StudySettingsEntity (one per user).

import "time"

type UserEntity struct {
	ID           string
	GoogleSub    string
	Email        string
	DisplayName  string
	AvatarURL    string
	UILocale     string
	Timezone     string
	DayStartHour int
	LastLoginAt  *time.Time
	CreatedAt    time.Time
}

type StudySettingsEntity struct {
	UserID           string
	NewCardsPerDay   int
	MaxReviewsPerDay int
	DesiredRetention float64
	LearningSteps    string
	RelearningSteps  string
	AudioAutoplay    bool
	Accent           string
	ShowImageOnFront bool
	MaxIntervalDays  int
}
