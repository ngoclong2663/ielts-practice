package dto

import (
	"time"

	"ielts-practice/app/modules/decks/domain/entity"
)

type DeckDto struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Source        string    `json:"source"`
	CollectionID  *string   `json:"collection_id"`
	NewPerDay     *int      `json:"new_per_day"`
	ReviewsPerDay *int      `json:"reviews_per_day"`
	CardCount     int       `json:"card_count"`
	CreatedAt     time.Time `json:"created_at"`
}

type DeckListDto struct {
	Items []DeckDto `json:"items"`
}

type CreateDeckRequestDto struct {
	Name        string `json:"name" binding:"required,min=1,max=80"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateDeckRequestDto: omitted fields stay unchanged; use_default_limits=true clears both limits.
type UpdateDeckRequestDto struct {
	Name             *string `json:"name" binding:"omitempty,min=1,max=80"`
	Description      *string `json:"description" binding:"omitempty,max=500"`
	NewPerDay        *int    `json:"new_per_day" binding:"omitempty,min=0,max=500"`
	ReviewsPerDay    *int    `json:"reviews_per_day" binding:"omitempty,min=0,max=9999"`
	UseDefaultLimits bool    `json:"use_default_limits"`
}

type CardSummaryDto struct {
	ID            string    `json:"id"`
	DeckID        string    `json:"deck_id"`
	Headword      string    `json:"headword"`
	MeaningVI     string    `json:"meaning_vi"`
	State         string    `json:"state"`
	Due           time.Time `json:"due"`
	ScheduledDays int       `json:"scheduled_days"`
	Reps          int       `json:"reps"`
	Lapses        int       `json:"lapses"`
	Suspended     bool      `json:"suspended"`
	IsLeech       bool      `json:"is_leech"`
	CreatedAt     time.Time `json:"created_at"`
}

type CardDto struct {
	CardSummaryDto
	Content entity.CardContent `json:"content"`
}

type CardListDto struct {
	Items    []CardSummaryDto `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
}

type ListCardsQueryDto struct {
	Q        string `form:"q" binding:"max=100"`
	State    string `form:"state" binding:"omitempty,oneof=new learning review suspended leech"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// ManualCardFields is the simple "add/edit card" form. Pointers: nil = unchanged.
type ManualCardFields struct {
	Headword     *string `json:"headword" binding:"omitempty,min=1,max=100"`
	MeaningVI    *string `json:"meaning_vi" binding:"omitempty,min=1,max=300"`
	IPA          *string `json:"ipa" binding:"omitempty,max=100"`
	PartOfSpeech *string `json:"part_of_speech" binding:"omitempty,max=30"`
	ExampleEN    *string `json:"example_en" binding:"omitempty,max=300"`
	ExampleVI    *string `json:"example_vi" binding:"omitempty,max=300"`
	// Empty string clears the image; otherwise it must be an http(s) URL (checked in the use case).
	ImageURL *string `json:"image_url" binding:"omitempty,max=1000"`
}

type CreateCardRequestDto struct {
	ManualCardFields
	AllowDuplicate bool `json:"allow_duplicate"`
}

type UpdateCardRequestDto struct {
	ManualCardFields
}

type BulkCardsRequestDto struct {
	CardIDs []string `json:"card_ids" binding:"required,min=1,max=500,dive,uuid"`
	Action  string   `json:"action" binding:"required,oneof=suspend unsuspend reset delete move"`
	DeckID  string   `json:"deck_id" binding:"omitempty,uuid"`
}

type BulkCardsResponseDto struct {
	Affected int64 `json:"affected"`
}
