package entity

import (
	"time"

	cards "ielts-practice/app/modules/decks/domain/entity"
)

// Aliases so the vocab module speaks the same card schema as decks (Epic 3 §4).
type (
	CardContent = cards.CardContent
	CardImage   = cards.CardImage
	IPA         = cards.IPA
	WordForm    = cards.WordForm
	Example     = cards.Example
	Collocation = cards.Collocation
)

const (
	LexemePending     = "pending"
	LexemeReady       = "ready"
	LexemeInvalid     = "invalid"
	LexemeFailed      = "failed"
	LexemeNeedsReview = "needs_review"

	ItemQueued  = "queued"
	ItemContent = "content"
	ItemImage   = "image"
	ItemReady   = "ready"
	ItemFailed  = "failed"
	ItemInvalid = "invalid"

	JobQueued  = "queued"
	JobRunning = "running"
	JobDone    = "done"
	JobPartial = "partial"
	JobFailed  = "failed"

	MeaningLangVI = "vi"
)

// Entry is one normalised word or phrase from the paste box.
type Entry struct {
	Text      string `json:"text"`
	SenseHint string `json:"sense_hint"`
}

// Rejection reasons are stable codes translated by the frontend.
const (
	RejectTooLong      = "TOO_LONG"
	RejectTooManyWords = "TOO_MANY_WORDS"
	RejectInvalidChars = "INVALID_CHARS"
	RejectDuplicate    = "DUPLICATE"
)

type RejectedEntry struct {
	Text   string `json:"text"`
	Reason string `json:"reason"`
}

type Lexeme struct {
	ID              string
	Key             string
	Headword        string
	Lemma           string
	SenseHint       string
	MeaningLang     string
	Content         *CardContent
	Status          string
	Suggestion      string
	ImageQuery      string
	ImageCandidates []CardImage
	QualityFlags    []string
	ReportCount     int
}

type Job struct {
	ID              string
	UserID          string
	TargetDeckID    *string
	Status          string
	Total           int
	Done            int
	CommittedDeckID *string
	CreatedAt       time.Time
	FinishedAt      *time.Time
	Items           []JobItem
}

type JobItem struct {
	ID              string
	JobID           string
	UserID          string
	Position        int
	InputText       string
	SenseHint       string
	LexemeID        *string
	Status          string
	Regenerate      bool
	Charged         bool
	Cached          bool
	Content         *CardContent
	ImageCandidates []CardImage
	ErrorCode       string
	Suggestion      string
	Attempts        int
	NextAttemptAt   time.Time
	Removed         bool
	UpdatedAt       time.Time
}

// IsFinal: the worker will not touch the item again unless the user retries it.
func (i JobItem) IsFinal() bool {
	return i.Status == ItemReady || i.Status == ItemFailed || i.Status == ItemInvalid
}

// ImageLink tracks a provider link that expires (Pixabay) behind /api/v1/images/:provider/:id.
type ImageLink struct {
	Provider    string
	ProviderID  string
	URL         string
	ThumbURL    string
	ExpiresAt   time.Time
	Query       string
	Headword    string
	Replacement *CardImage
	Gone        bool
}

type Usage struct {
	InputTokens     int
	OutputTokens    int
	CacheReadTokens int
}

func (u *Usage) Add(o Usage) {
	u.InputTokens += o.InputTokens
	u.OutputTokens += o.OutputTokens
	u.CacheReadTokens += o.CacheReadTokens
}

// SuggestedWord is one candidate from a topic word list (Epic 4).
type SuggestedWord struct {
	Lemma     string `json:"lemma"`
	POS       string `json:"pos"`
	MeaningVI string `json:"meaning_vi"`
	CEFR      string `json:"cefr"`
}

type Suggestion struct {
	ID              string
	UserID          string
	DeckID          *string
	TopicID         string
	TopicText       string
	Requested       int
	Words           []SuggestedWord
	SkippedExisting int
	CreatedAt       time.Time
}
