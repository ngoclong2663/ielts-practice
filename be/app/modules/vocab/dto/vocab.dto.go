package dto

import (
	"time"

	"ielts-practice/app/modules/vocab/domain/entity"
)

type NormalizeRequestDto struct {
	Text string `json:"text" binding:"max=20000"`
}

type NormalizeResponseDto struct {
	Entries  []entity.Entry         `json:"entries"`
	Rejected []entity.RejectedEntry `json:"rejected"`
	Max      int                    `json:"max"`
}

type EntryDto struct {
	Text      string `json:"text" binding:"required,max=100"`
	SenseHint string `json:"sense_hint" binding:"max=60"`
}

type CreateJobRequestDto struct {
	Entries      []EntryDto `json:"entries" binding:"required,min=1,max=200,dive"`
	TargetDeckID *string    `json:"target_deck_id" binding:"omitempty,uuid"`
}

type JobItemDto struct {
	ID              string              `json:"id"`
	Position        int                 `json:"position"`
	InputText       string              `json:"input_text"`
	SenseHint       string              `json:"sense_hint"`
	Status          string              `json:"status"`
	ErrorCode       string              `json:"error_code"`
	Suggestion      string              `json:"suggestion"`
	Cached          bool                `json:"cached"`
	LexemeID        *string             `json:"lexeme_id"`
	Content         *entity.CardContent `json:"content"`
	ImageCandidates []entity.CardImage  `json:"image_candidates"`
	Removed         bool                `json:"removed"`
}

type JobDto struct {
	ID              string       `json:"id"`
	Status          string       `json:"status"`
	Total           int          `json:"total"`
	Done            int          `json:"done"`
	TargetDeckID    *string      `json:"target_deck_id"`
	CommittedDeckID *string      `json:"committed_deck_id"`
	CreatedAt       time.Time    `json:"created_at"`
	FinishedAt      *time.Time   `json:"finished_at"`
	Items           []JobItemDto `json:"items"`
}

type JobSummaryDto struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	Total           int        `json:"total"`
	Done            int        `json:"done"`
	CommittedDeckID *string    `json:"committed_deck_id"`
	CreatedAt       time.Time  `json:"created_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	Words           []string   `json:"words"`
}

type JobListDto struct {
	Items []JobSummaryDto `json:"items"`
}

type RetryItemRequestDto struct {
	// InputText replaces the entry (e.g. accepting a typo or lemma suggestion).
	InputText *string `json:"input_text" binding:"omitempty,max=100"`
	SenseHint *string `json:"sense_hint" binding:"omitempty,max=60"`
	// Regenerate asks for fresh content for this item only (the shared cache is not changed).
	Regenerate bool `json:"regenerate"`
}

type ImageSearchQueryDto struct {
	Q        string `form:"q" binding:"max=100"`
	Provider string `form:"provider" binding:"omitempty,oneof=pixabay unsplash fake"`
}

type ImageSearchResponseDto struct {
	Images    []entity.CardImage `json:"images"`
	Provider  string             `json:"provider"`
	Providers []string           `json:"providers"`
}

type ImageRefDto struct {
	Provider   string `json:"provider" binding:"required,max=16"`
	ProviderID string `json:"provider_id" binding:"required,max=64"`
}

type UpdateItemRequestDto struct {
	Content    *entity.CardContent `json:"content"`
	Image      *ImageRefDto        `json:"image"`
	ClearImage bool                `json:"clear_image"`
	Removed    *bool               `json:"removed"`
}

type CommitRequestDto struct {
	DeckID      string   `json:"deck_id" binding:"omitempty,uuid"`
	NewDeckName string   `json:"new_deck_name" binding:"omitempty,max=80"`
	ItemIDs     []string `json:"item_ids" binding:"required,min=1,max=30,dive,uuid"`
}

type CommitResponseDto struct {
	DeckID   string   `json:"deck_id"`
	DeckName string   `json:"deck_name"`
	Created  int      `json:"created"`
	Skipped  []string `json:"skipped"`
}

type ReportRequestDto struct {
	Reason string `json:"reason" binding:"required,oneof=wrong_meaning bad_example wrong_image other"`
	Note   string `json:"note" binding:"max=500"`
}

type QuotaDto struct {
	Limit     int `json:"limit"`
	Used      int `json:"used"`
	Remaining int `json:"remaining"`
}

type StatusDto struct {
	Configured     bool     `json:"configured"`
	MaxEntries     int      `json:"max_entries"`
	Quota          QuotaDto `json:"quota"`
	ImageProviders []string `json:"image_providers"`
}
