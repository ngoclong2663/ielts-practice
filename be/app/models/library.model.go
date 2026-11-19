package models

import "time"

// Vocabulary library (Epic 8 §3), loaded by `api seed`.

type LibraryCollection struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Slug        string
	Kind        string
	Title       []byte `gorm:"type:jsonb"`
	Description []byte `gorm:"type:jsonb"`
	Sort        int16
	Status      string
	ContentHash string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (LibraryCollection) TableName() string { return "library_collections" }

type LibraryTopic struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CollectionID string `gorm:"type:uuid"`
	TopicID      string
	Sort         int16
	ContentHash  string
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (LibraryTopic) TableName() string { return "library_topics" }

type LibraryWord struct {
	ID             string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	LibraryTopicID string `gorm:"type:uuid"`
	LexemeID       string `gorm:"type:uuid"`
	CEFR           string `gorm:"column:cefr"`
	Source         string
	Sort           int16
	SearchText     string
}

func (LibraryWord) TableName() string { return "library_words" }
