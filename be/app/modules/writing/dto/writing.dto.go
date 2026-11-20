package dto

import (
	"time"

	"ielts-practice/app/modules/writing/domain/entity"
)

// ---- requests ----------------------------------------------------------------

type TaskListQueryDto struct {
	TaskNo int    `form:"task_no" binding:"omitempty,oneof=1 2"`
	Type   string `form:"type" binding:"max=32"`
	Topic  string `form:"topic" binding:"max=32"`
}

type StartAttemptRequestDto struct {
	Level    int    `json:"level" binding:"required,min=1,max=4"`
	Position string `json:"position" binding:"max=24"`
}

type PlanRequestDto struct {
	Answers []int `json:"answers" binding:"required,max=10"`
}

type CheckRequestDto struct {
	Answers []string `json:"answers" binding:"required,min=1,max=6,dive,max=400"`
}

type SubmitRequestDto struct {
	Text         string `json:"text" binding:"required,max=6000"`
	SecondsSpent int    `json:"seconds_spent" binding:"min=0,max=86400"`
}

// ---- tasks -------------------------------------------------------------------

type FigureDto struct {
	URL    string `json:"url"`
	Alt    string `json:"alt"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type TaskSummaryDto struct {
	ID       string     `json:"id"`
	TaskNo   int        `json:"task_no"`
	TaskType string     `json:"task_type"`
	Topic    string     `json:"topic"`
	Title    string     `json:"title"`
	Figure   *FigureDto `json:"figure"`
	Attempts int        `json:"attempts"`
	BestBand *float64   `json:"best_band"`
}

type TaskDto struct {
	TaskSummaryDto
	Prompt    string      `json:"prompt"`
	Figures   []FigureDto `json:"figures"`
	Positions []string    `json:"positions"`
	MinWords  int         `json:"min_words"`
	Minutes   int         `json:"minutes"`
}

type TaskDetailDto struct {
	TaskDto
	History []AttemptSummaryDto `json:"history"`
}

// ---- attempts ----------------------------------------------------------------

type PlanQuestionDto struct {
	QuestionVI string   `json:"question_vi"`
	OptionsVI  []string `json:"options_vi"`
	// Answer and ExplainVI are sent once the learner has answered.
	Answer    *int   `json:"answer"`
	ExplainVI string `json:"explain_vi,omitempty"`
}

type SentenceDto struct {
	// Prompt is what the learner works from at their level: Vietnamese sentence, idea note or English outline.
	Prompt string        `json:"prompt"`
	Hints  []entity.Hint `json:"hints"`
	// ModelEN is sent once the paragraph has been checked.
	ModelEN string `json:"model_en,omitempty"`
}

type ParagraphDto struct {
	Key          string                    `json:"key"`
	GoalVI       string                    `json:"goal_vi"`
	Sentences    []SentenceDto             `json:"sentences"`
	Answers      []string                  `json:"answers"`
	Feedback     []entity.SentenceFeedback `json:"feedback"`
	CheckedCount int                       `json:"checked_count"`
}

type SubmissionDto struct {
	Text         string               `json:"text"`
	WordCount    int                  `json:"word_count"`
	SecondsSpent int                  `json:"seconds_spent"`
	BandLow      float64              `json:"band_low"`
	BandHigh     float64              `json:"band_high"`
	Feedback     entity.EssayFeedback `json:"feedback"`
	CreatedAt    time.Time            `json:"created_at"`
}

type AttemptDto struct {
	ID             string            `json:"id"`
	Task           TaskDto           `json:"task"`
	Level          int               `json:"level"`
	Position       string            `json:"position"`
	Status         string            `json:"status"`
	IdeasVI        []string          `json:"ideas_vi"`
	Questions      []PlanQuestionDto `json:"questions"`
	PlanAnswers    []int             `json:"plan_answers"`
	Paragraphs     []ParagraphDto    `json:"paragraphs"`
	VocabAvailable bool              `json:"vocab_available"` // GET /attempts/:id/vocab may be called
	ModelAnswer    string            `json:"model_answer,omitempty"`
	Submission     *SubmissionDto    `json:"submission"`
	CreatedAt      time.Time         `json:"created_at"`
	SubmittedAt    *time.Time        `json:"submitted_at"`
}

type AttemptSummaryDto struct {
	ID          string     `json:"id"`
	TaskID      string     `json:"task_id"`
	TaskTitle   string     `json:"task_title"`
	TaskNo      int        `json:"task_no"`
	TaskType    string     `json:"task_type"`
	Level       int        `json:"level"`
	Position    string     `json:"position"`
	Status      string     `json:"status"`
	BandLow     *float64   `json:"band_low"`
	BandHigh    *float64   `json:"band_high"`
	CreatedAt   time.Time  `json:"created_at"`
	SubmittedAt *time.Time `json:"submitted_at"`
}

type QuotaDto struct {
	Limit     int `json:"limit"`
	Used      int `json:"used"`
	Remaining int `json:"remaining"`
}

type StatusDto struct {
	Configured  bool     `json:"configured"`
	Checks      QuotaDto `json:"checks"`
	Submissions QuotaDto `json:"submissions"`
}

// VocabDto is a lesson's useful vocabulary, in group order (topic, paraphrase, task).
type VocabDto struct {
	Groups []entity.VocabGroup `json:"groups"`
}
