package entity

import (
	"encoding/json"
	"time"
)

// Figure is one Task 1 image stored in S3 under a content-addressed key.
type Figure struct {
	Key         string `json:"key"`
	SHA256      string `json:"sha256"`
	ContentType string `json:"content_type"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Alt         string `json:"alt"`
}

// Task is one IELTS writing task from the task bank (Epic 5 §6).
type Task struct {
	ID        string
	Slug      string
	Pack      string
	SourceRef string // shown to learners as "Inspired by Cambridge IELTS 21 · Test 1"
	TestNo    int
	TaskNo    int
	Module    string
	TaskType  string
	Topic     string
	Title     string
	Prompt    string
	Figures   []Figure
	Facts     json.RawMessage
	// Positions a learner may take (Task 2); empty means the default for the task type.
	Positions []string
	MinWords  int
	Minutes   int
	Status    string
	// ContentHash lets the seed skip tasks whose pack entry and figures are unchanged.
	ContentHash string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
