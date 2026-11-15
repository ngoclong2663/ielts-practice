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
	MinWords    int16
	Minutes     int16
	Status      string
	ContentHash string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (WritingTask) TableName() string { return "writing_tasks" }
