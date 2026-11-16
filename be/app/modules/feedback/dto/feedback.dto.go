package dto

import "time"

type CreateFeedbackRequestDto struct {
	Category string `json:"category" binding:"required,oneof=bug idea content other"`
	// Trimmed, then 10–2000 characters (checked in the use case; max here only bounds the body).
	Message  string `json:"message" binding:"required,max=4000"`
	Page     string `json:"page" binding:"max=300"`
	UILocale string `json:"ui_locale" binding:"omitempty,oneof=vi en"`
}

type CreatedFeedbackDto struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type ListFeedbackQueryDto struct {
	// Default new; "all" lists every status.
	Status string `form:"status" binding:"omitempty,oneof=new seen resolved all"`
	// Empty or "all" lists every category.
	Category string `form:"category" binding:"omitempty,oneof=bug idea content other all"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// UpdateFeedbackRequestDto: nil fields are unchanged.
type UpdateFeedbackRequestDto struct {
	Status    *string `json:"status" binding:"omitempty,oneof=new seen resolved"`
	AdminNote *string `json:"admin_note" binding:"omitempty,max=1000"`
}

type FeedbackUserDto struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type FeedbackItemDto struct {
	ID         string          `json:"id"`
	Category   string          `json:"category"`
	Message    string          `json:"message"`
	Page       string          `json:"page"`
	UILocale   string          `json:"ui_locale"`
	UserAgent  string          `json:"user_agent"`
	Status     string          `json:"status"`
	AdminNote  string          `json:"admin_note"`
	Notified   bool            `json:"notified"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	ResolvedAt *time.Time      `json:"resolved_at"`
	User       FeedbackUserDto `json:"user"`
}

type StatusCountsDto struct {
	New      int `json:"new"`
	Seen     int `json:"seen"`
	Resolved int `json:"resolved"`
}

type FeedbackListDto struct {
	Items    []FeedbackItemDto `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Counts   StatusCountsDto   `json:"counts"`
}
