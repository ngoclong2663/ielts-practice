package dto

import (
	"encoding/json"
	"time"

	"ielts-practice/app/modules/study/domain/entity"
)

type QueueQueryDto struct {
	DeckID string `form:"deck_id" binding:"omitempty,uuid"`
	Size   int    `form:"size" binding:"omitempty,min=1,max=10"`
}

// IntervalsDto: seconds until the card would be shown again for each answer button.
type IntervalsDto struct {
	Again int64 `json:"again"`
	Hard  int64 `json:"hard"`
	Good  int64 `json:"good"`
	Easy  int64 `json:"easy"`
}

type StudyCardDto struct {
	ID        string          `json:"id"`
	DeckID    string          `json:"deck_id"`
	DeckName  string          `json:"deck_name"`
	State     string          `json:"state"`
	Content   json.RawMessage `json:"content"`
	IsLeech   bool            `json:"is_leech"`
	Reps      int             `json:"reps"`
	Lapses    int             `json:"lapses"`
	Intervals IntervalsDto    `json:"intervals"`
}

type StudyQueueDto struct {
	Cards []StudyCardDto `json:"cards"`
	// Counts left for today: new · learning · review.
	Counts entity.Counts `json:"counts"`
	// Set when nothing is due now but learning cards come back later today.
	NextDueAt *time.Time `json:"next_due_at"`
	DayEndsAt time.Time  `json:"day_ends_at"`
}

type AnswerRequestDto struct {
	CardID         string `json:"card_id" binding:"required,uuid"`
	Rating         int    `json:"rating" binding:"required,min=1,max=4"`
	DurationMs     int    `json:"duration_ms" binding:"min=0"`
	ClientReviewID string `json:"client_review_id" binding:"required,uuid"`
	// Optional optimistic check: the card's reps when it was shown.
	ExpectedReps *int `json:"expected_reps" binding:"omitempty,min=0"`
	// Scope and size of the returned queue.
	DeckID string `json:"deck_id" binding:"omitempty,uuid"`
	Size   int    `json:"size" binding:"omitempty,min=1,max=10"`
}

type AnsweredDto struct {
	CardID      string    `json:"card_id"`
	Rating      int       `json:"rating"`
	State       string    `json:"state"`
	Due         time.Time `json:"due"`
	BecameLeech bool      `json:"became_leech"`
}

type AnswerResponseDto struct {
	StudyQueueDto
	Answered *AnsweredDto `json:"answered"`
}

type UndoRequestDto struct {
	DeckID string `json:"deck_id" binding:"omitempty,uuid"`
	Size   int    `json:"size" binding:"omitempty,min=1,max=10"`
}

type UndoResponseDto struct {
	StudyQueueDto
	UndoneCardID string `json:"undone_card_id"`
}

type DeckCountsDto struct {
	DeckID string `json:"deck_id"`
	entity.Counts
}

type CountsResponseDto struct {
	Decks     []DeckCountsDto `json:"decks"`
	Total     entity.Counts   `json:"total"`
	DayEndsAt time.Time       `json:"day_ends_at"`
}

type TodayDto struct {
	Reviews    int   `json:"reviews"`
	NewCards   int   `json:"new_cards"`
	Again      int   `json:"again"`
	DurationMs int64 `json:"duration_ms"`
}

type ForecastDayDto struct {
	Date  string `json:"date"` // YYYY-MM-DD (study day)
	Count int    `json:"count"`
}

type CardTotalsDto struct {
	New       int `json:"new"`
	Learning  int `json:"learning"`
	Review    int `json:"review"`
	Suspended int `json:"suspended"`
}

type StatsOverviewDto struct {
	Today           TodayDto         `json:"today"`
	StreakDays      int              `json:"streak_days"`
	StudiedToday    bool             `json:"studied_today"`
	Forecast        []ForecastDayDto `json:"forecast"`
	Retention30d    *float64         `json:"retention_30d"`
	RetentionSample int              `json:"retention_sample"`
	Totals          CardTotalsDto    `json:"totals"`
}
