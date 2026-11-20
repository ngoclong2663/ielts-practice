package dto

import (
	"time"

	"ielts-practice/app/modules/exam/domain/entity"
)

type SkillQueryDto struct {
	Skill string `form:"skill" binding:"required,oneof=reading listening"`
}

type ListAttemptsQueryDto struct {
	Skill  string `form:"skill" binding:"required,oneof=reading listening"`
	Before string `form:"before" binding:"omitempty,uuid"`
}

// ---- test bank ---------------------------------------------------------------

type BookDto struct {
	Slug  string           `json:"slug"`
	Title entity.Localized `json:"title"`
	Tests []TestSummaryDto `json:"tests"`
}

type TestSummaryDto struct {
	Slug   string           `json:"slug"`
	TestNo int              `json:"test_no"`
	Parts  []PartSummaryDto `json:"parts"`
}

type PartSummaryDto struct {
	PartNo        int           `json:"part_no"`
	Title         string        `json:"title"`
	QFrom         int           `json:"q_from"`
	QTo           int           `json:"q_to"`
	QuestionTypes []string      `json:"question_types"`
	Best          *entity.Score `json:"best"` // the learner's best score on this part, null if never done
}

type PartDto struct {
	PartNo        int            `json:"part_no"`
	Title         string         `json:"title"`
	QFrom         int            `json:"q_from"`
	QTo           int            `json:"q_to"`
	QuestionTypes []string       `json:"question_types"`
	Content       entity.Content `json:"content"`
	Audio         *AudioDto      `json:"audio"`
	Key           entity.Key     `json:"key,omitempty"` // review only
}

type AudioDto struct {
	URL      string       `json:"url"`
	Duration float64      `json:"duration"`
	Cues     []entity.Cue `json:"cues"`
}

type TestDto struct {
	Slug      string           `json:"slug"`
	TestNo    int              `json:"test_no"`
	Skill     string           `json:"skill"`
	BookSlug  string           `json:"book_slug"`
	BookTitle entity.Localized `json:"book_title"`
	Parts     []PartDto        `json:"parts"`
}

// ---- attempts ----------------------------------------------------------------

type SubmitAttemptRequestDto struct {
	Test            string         `json:"test" binding:"required,max=32"`
	Skill           string         `json:"skill" binding:"required,oneof=reading listening"`
	Kind            string         `json:"kind" binding:"required,oneof=part full redo"`
	Parts           []int          `json:"parts" binding:"max=4,dive,min=1,max=4"`
	Questions       []int          `json:"questions" binding:"max=40,dive,min=1,max=40"` // redo only
	ParentAttemptID string         `json:"parent_attempt_id" binding:"omitempty,uuid"`
	Mode            string         `json:"mode" binding:"omitempty,oneof=practice exam"`
	TimeLimitSec    int            `json:"time_limit_sec" binding:"min=0,max=7200"`
	DurationSec     int            `json:"duration_sec" binding:"min=0,max=86400"`
	Answers         entity.Answers `json:"answers" binding:"max=40"`
}

type SubmitAttemptResponseDto struct {
	ID string `json:"id"`
}

type AttemptSummaryDto struct {
	ID           string           `json:"id"`
	Skill        string           `json:"skill"`
	TestSlug     string           `json:"test_slug"`
	TestNo       int              `json:"test_no"`
	BookTitle    entity.Localized `json:"book_title"`
	PartNos      []int            `json:"part_nos"`
	Kind         string           `json:"kind"`
	Mode         string           `json:"mode"`
	Correct      int              `json:"correct"`
	Total        int              `json:"total"`
	Band         *float64         `json:"band"`
	DurationSec  int              `json:"duration_sec"`
	TimeLimitSec int              `json:"time_limit_sec"`
	CreatedAt    time.Time        `json:"created_at"`
}

type AttemptDto struct {
	AttemptSummaryDto
	ParentAttemptID string          `json:"parent_attempt_id,omitempty"`
	Questions       []int           `json:"questions"`
	Answers         entity.Answers  `json:"answers"`
	Results         []entity.Result `json:"results"`
	Test            TestDto         `json:"test"` // the attempt's parts with their keys and explanations
}

type AttemptListDto struct {
	Items      []AttemptSummaryDto `json:"items"`
	NextBefore string              `json:"next_before,omitempty"`
}

type TypeStatDto struct {
	Type    string `json:"type"`
	Correct int    `json:"correct"`
	Total   int    `json:"total"`
}

type BandPointDto struct {
	AttemptID string    `json:"attempt_id"`
	Band      *float64  `json:"band"`
	Correct   int       `json:"correct"`
	At        time.Time `json:"at"`
}

type StatsDto struct {
	Attempts int            `json:"attempts"` // how many attempts the type figures cover
	ByType   []TypeStatDto  `json:"by_type"`  // weakest first
	Bands    []BandPointDto `json:"bands"`    // full tests, oldest first
}
