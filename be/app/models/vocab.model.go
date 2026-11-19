package models

import "time"

type Lexeme struct {
	ID              string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Key             string
	Headword        string
	Lemma           string
	SenseHint       string
	MeaningLang     string
	Content         []byte `gorm:"type:jsonb"`
	ContentVersion  int16
	Status          string
	Suggestion      string
	ImageQuery      string
	ImageCandidates []byte `gorm:"type:jsonb"`
	QualityFlags    []byte `gorm:"type:jsonb"`
	ReportCount     int
	Origin          string `gorm:"default:generated"`
	Topic           string
	ClaimedAt       *time.Time
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (Lexeme) TableName() string { return "lexemes" }

type GenerationJob struct {
	ID              string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID          string  `gorm:"type:uuid"`
	TargetDeckID    *string `gorm:"type:uuid"`
	Status          string
	Total           int
	Done            int
	CommittedDeckID *string `gorm:"type:uuid"`
	InputTokens     int
	OutputTokens    int
	CacheReadTokens int
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	FinishedAt      *time.Time
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (GenerationJob) TableName() string { return "generation_jobs" }

type GenerationJobItem struct {
	ID              string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	JobID           string `gorm:"type:uuid"`
	UserID          string `gorm:"type:uuid"`
	Position        int
	InputText       string
	SenseHint       string
	LexemeID        *string `gorm:"type:uuid"`
	Status          string
	Regenerate      bool
	Charged         bool
	Cached          bool
	Content         []byte `gorm:"type:jsonb"`
	ImageCandidates []byte `gorm:"type:jsonb"`
	ErrorCode       string
	Suggestion      string
	Attempts        int
	NextAttemptAt   time.Time
	Removed         bool
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (GenerationJobItem) TableName() string { return "generation_job_items" }

type ImageLink struct {
	Provider    string `gorm:"primaryKey"`
	ProviderID  string `gorm:"primaryKey"`
	URL         string
	ThumbURL    string
	ExpiresAt   time.Time
	Query       string
	Headword    string
	Replacement []byte `gorm:"type:jsonb"`
	Gone        bool
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (ImageLink) TableName() string { return "image_links" }

type TopicSuggestion struct {
	ID              string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID          string  `gorm:"type:uuid"`
	DeckID          *string `gorm:"type:uuid"`
	TopicID         string
	TopicText       string
	Requested       int
	Words           []byte `gorm:"type:jsonb"`
	SkippedExisting int
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (TopicSuggestion) TableName() string { return "topic_suggestions" }
