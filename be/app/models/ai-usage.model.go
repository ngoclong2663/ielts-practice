package models

import "time"

// AIUsage is one call to a language model (Epic 7), written by the llm module.
type AIUsage struct {
	ID              string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Provider        string
	Model           string
	Feature         string
	OK              bool `gorm:"column:ok"`
	InputTokens     int
	OutputTokens    int
	CacheReadTokens int
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

func (AIUsage) TableName() string { return "ai_usage" }
