package models

import "time"

type WritingTask struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Slug        string
	Pack        string
	SourceRef   string
	TestNo      int16
	TaskNo      int16
	Module      string
	TaskType    string
	Topic       string
	Title       string
	Prompt      string
	Figures     []byte `gorm:"type:jsonb"`
	Facts       []byte `gorm:"type:jsonb"`
	Positions   []byte `gorm:"type:jsonb"`
	MinWords    int16
	Minutes     int16
	Status      string
	ContentHash string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (WritingTask) TableName() string { return "writing_tasks" }

type WritingLesson struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TaskID       string `gorm:"type:uuid"`
	Position     string
	TaskHash     string
	Content      []byte `gorm:"type:jsonb"`
	Model        string
	InputTokens  int
	OutputTokens int
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (WritingLesson) TableName() string { return "writing_lessons" }

// WritingLessonVocab is a lesson's useful vocabulary (Epic 7).
type WritingLessonVocab struct {
	LessonID     string `gorm:"primaryKey;type:uuid"`
	Content      []byte `gorm:"type:jsonb"`
	Model        string
	InputTokens  int
	OutputTokens int
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (WritingLessonVocab) TableName() string { return "writing_lesson_vocab" }

type WritingAttempt struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string `gorm:"type:uuid"`
	TaskID      string `gorm:"type:uuid"`
	LessonID    string `gorm:"type:uuid"`
	Level       int16
	Position    string
	PlanAnswers []byte `gorm:"type:jsonb"`
	Status      string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	SubmittedAt *time.Time
}

func (WritingAttempt) TableName() string { return "writing_attempts" }

type WritingAttemptParagraph struct {
	AttemptID    string `gorm:"primaryKey;type:uuid"`
	ParagraphKey string `gorm:"primaryKey"`
	Answers      []byte `gorm:"type:jsonb"`
	Feedback     []byte `gorm:"type:jsonb"`
	CheckedCount int
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (WritingAttemptParagraph) TableName() string { return "writing_attempt_paragraphs" }

type WritingSubmission struct {
	AttemptID    string `gorm:"primaryKey;type:uuid"`
	Text         string
	WordCount    int
	SecondsSpent int
	Feedback     []byte `gorm:"type:jsonb"`
	BandLow      float64
	BandHigh     float64
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (WritingSubmission) TableName() string { return "writing_submissions" }

type WritingAIRequest struct {
	ID           string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       string  `gorm:"type:uuid"`
	AttemptID    *string `gorm:"type:uuid"`
	Kind         string
	Model        string
	InputTokens  int
	OutputTokens int
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (WritingAIRequest) TableName() string { return "writing_ai_requests" }
