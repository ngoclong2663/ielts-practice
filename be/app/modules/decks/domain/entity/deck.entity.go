package entity

import "time"

type DeckEntity struct {
	ID            string
	UserID        string
	Name          string
	Description   string
	Source        string
	CollectionID  *string
	NewPerDay     *int
	ReviewsPerDay *int
	CardCount     int
	CreatedAt     time.Time
}

// Card states, identical to go-fsrs values.
const (
	StateNew        int16 = 0
	StateLearning   int16 = 1
	StateReview     int16 = 2
	StateRelearning int16 = 3
)

func StateName(s int16) string {
	switch s {
	case StateLearning:
		return "learning"
	case StateReview:
		return "review"
	case StateRelearning:
		return "relearning"
	default:
		return "new"
	}
}

type CardEntity struct {
	ID            string
	UserID        string
	DeckID        string
	LexemeID      *string
	Headword      string
	Content       CardContent
	State         int16
	Due           time.Time
	Position      int
	ScheduledDays int
	Reps          int
	Lapses        int
	LastReviewAt  *time.Time
	Suspended     bool
	IsLeech       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CardFilter struct {
	DeckID   string
	Query    string
	State    string // "", new, learning, review, suspended, leech
	Page     int
	PageSize int
}
