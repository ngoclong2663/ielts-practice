package entity

// Entities: Feedback. Value objects: FeedbackUser, FeedbackFilter, StatusCounts and the
// category and status values.

import "time"

const (
	CategoryBug     = "bug"
	CategoryIdea    = "idea"
	CategoryContent = "content"
	CategoryOther   = "other"

	StatusNew      = "new"
	StatusSeen     = "seen"
	StatusResolved = "resolved"
)

type Feedback struct {
	ID         string
	UserID     string
	Category   string
	Message    string
	Page       string
	UILocale   string
	UserAgent  string
	Status     string
	AdminNote  string
	NotifiedAt *time.Time
	ResolvedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	// User is filled by the admin queries (joined from users).
	User FeedbackUser
}

type FeedbackUser struct {
	ID          string
	Email       string
	DisplayName string
	AvatarURL   string
}

// FeedbackFilter: empty Status or Category means all.
type FeedbackFilter struct {
	Status   string
	Category string
	Page     int
	PageSize int
}

type StatusCounts struct {
	New      int
	Seen     int
	Resolved int
}
