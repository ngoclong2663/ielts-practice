package models

import "time"

// ExamBook, ExamTest, ExamPart and ExamKey hold the Epic 10 test bank, loaded by `api seed`.
type ExamBook struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Slug      string
	SourceRef string
	Title     []byte `gorm:"type:jsonb"`
	Sort      int16
	Status    string
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (ExamBook) TableName() string { return "exam_books" }

type ExamTest struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	BookID    string `gorm:"type:uuid"`
	TestNo    int16
	Slug      string
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (ExamTest) TableName() string { return "exam_tests" }

type ExamPart struct {
	ID            string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TestID        string `gorm:"type:uuid"`
	Skill         string
	PartNo        int16
	Slug          string
	SourceRef     string
	Title         string
	Topic         string
	QFrom         int16
	QTo           int16
	QuestionTypes []byte `gorm:"type:jsonb"`
	Content       []byte `gorm:"type:jsonb"`
	Audio         []byte `gorm:"type:jsonb"`
	Status        string
	ContentHash   string
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (ExamPart) TableName() string { return "exam_parts" }

type ExamKey struct {
	PartID    string    `gorm:"primaryKey;type:uuid"`
	Key       []byte    `gorm:"type:jsonb"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (ExamKey) TableName() string { return "exam_keys" }

// ExamAttempt is one submitted run (Epic 10 §9).
type ExamAttempt struct {
	ID              string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID          string `gorm:"type:uuid"`
	Skill           string
	TestID          string `gorm:"type:uuid"`
	PartIDs         []byte `gorm:"column:part_ids;type:jsonb"`
	Kind            string
	ParentAttemptID *string `gorm:"type:uuid"`
	Mode            string
	TimeLimitSec    int
	DurationSec     int
	Questions       []byte `gorm:"type:jsonb"`
	Answers         []byte `gorm:"type:jsonb"`
	Results         []byte `gorm:"type:jsonb"`
	PartScores      []byte `gorm:"type:jsonb"`
	Correct         int16
	Total           int16
	Band            *float64
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

func (ExamAttempt) TableName() string { return "exam_attempts" }
