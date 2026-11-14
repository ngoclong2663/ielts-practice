package models

import "time"

type Deck struct {
	ID            string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID        string `gorm:"type:uuid"`
	Name          string
	Description   string
	Source        string
	CollectionID  *string `gorm:"type:uuid"`
	NewPerDay     *int
	ReviewsPerDay *int
	ArchivedAt    *time.Time
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (Deck) TableName() string { return "decks" }

type Card struct {
	ID             string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID         string  `gorm:"type:uuid"`
	DeckID         string  `gorm:"type:uuid"`
	LexemeID       *string `gorm:"type:uuid"`
	Headword       string
	Content        []byte `gorm:"type:jsonb"`
	ContentVersion int16

	State          int16
	Due            time.Time
	Position       int
	Stability      float64
	Difficulty     float64
	ScheduledDays  int
	RemainingSteps int16
	Reps           int
	Lapses         int
	LastReviewAt   *time.Time
	Suspended      bool
	IsLeech        bool

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Card) TableName() string { return "cards" }

type ReviewLog struct {
	ID             string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID         string `gorm:"type:uuid"`
	CardID         string `gorm:"type:uuid"`
	DeckID         string `gorm:"type:uuid"`
	ClientReviewID string `gorm:"type:uuid"`
	Rating         int16

	StateBefore          int16
	DueBefore            time.Time
	StabilityBefore      float64
	DifficultyBefore     float64
	ScheduledDaysBefore  int
	RemainingStepsBefore int16
	RepsBefore           int
	LapsesBefore         int
	LastReviewBefore     *time.Time
	IsLeechBefore        bool

	StateAfter         int16
	DueAfter           time.Time
	StabilityAfter     float64
	DifficultyAfter    float64
	ScheduledDaysAfter int
	ElapsedDays        int

	DurationMs int
	ReviewedAt time.Time
	UndoneAt   *time.Time
}

func (ReviewLog) TableName() string { return "review_logs" }
