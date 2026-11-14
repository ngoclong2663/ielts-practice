package entity

import (
	"encoding/json"
	"time"
)

// Card states (same values as go-fsrs).
const (
	StateNew        int16 = 0
	StateLearning   int16 = 1
	StateReview     int16 = 2
	StateRelearning int16 = 3
)

type Rating int16

const (
	Again Rating = 1
	Hard  Rating = 2
	Good  Rating = 3
	Easy  Rating = 4
)

var Ratings = []Rating{Again, Hard, Good, Easy}

// SchedulingState is the FSRS memory and queue state stored on a card.
type SchedulingState struct {
	State          int16
	Due            time.Time
	Stability      float64
	Difficulty     float64
	ScheduledDays  int
	RemainingSteps int
	Reps           int
	Lapses         int
	LastReview     *time.Time
}

// StudyCard is a card as the study session needs it.
type StudyCard struct {
	ID        string
	DeckID    string
	DeckName  string
	Content   json.RawMessage
	Suspended bool
	IsLeech   bool
	SchedulingState
}

type ReviewLogEntity struct {
	ID             string
	UserID         string
	CardID         string
	DeckID         string
	ClientReviewID string
	Rating         Rating
	Before         SchedulingState
	IsLeechBefore  bool
	After          SchedulingState
	ElapsedDays    int
	DurationMs     int
	ReviewedAt     time.Time
}

// DeckStat holds raw per-deck counts before daily limits are applied.
type DeckStat struct {
	DeckID        string
	NewPerDay     *int
	ReviewsPerDay *int
	NewAvailable  int
	LearningDue   int
	ReviewDue     int
}

// TodayStat counts what was already studied today in a deck.
type TodayStat struct {
	Introduced int // new cards seen for the first time today
	Reviewed   int // answers on cards that were in the review state
}

// Counts are what the UI shows: new · learning · review.
type Counts struct {
	New      int `json:"new"`
	Learning int `json:"learning"`
	Review   int `json:"review"`
}
