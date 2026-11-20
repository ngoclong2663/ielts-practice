// Package entity holds the exam module's entities (Book, Test, Part, Attempt) and the part content format
// (Epic 10 §8.1).
package entity

import "time"

const (
	SkillReading   = "reading"
	SkillListening = "listening"

	KindPart = "part"
	KindFull = "full"
	KindRedo = "redo"

	ModePractice = "practice"
	ModeExam     = "exam"

	StatusPublished = "published"

	// MediaKeyPrefix is where listening audio lives in S3: exam/media/<sha256>.mp3 (content-addressed).
	MediaKeyPrefix = "exam/media/"
)

var Skills = []string{SkillReading, SkillListening}

// Localized is a text in the app's UI languages.
type Localized struct {
	VI string `json:"vi"`
	EN string `json:"en"`
}

type Book struct {
	ID        string
	Slug      string
	SourceRef string // internal only, never returned by the API
	Title     Localized
	Sort      int
	Status    string
}

type Test struct {
	ID     string
	BookID string
	TestNo int
	Slug   string
}

type Part struct {
	ID            string
	TestID        string
	Skill         string
	PartNo        int
	Slug          string
	SourceRef     string
	Title         string
	Topic         string
	QFrom, QTo    int
	QuestionTypes []string
	Content       Content
	Audio         *Audio
	Status        string
	ContentHash   string
}

// QuestionCount is the number of questions in the part.
func (p *Part) QuestionCount() int { return p.QTo - p.QFrom + 1 }

// Attempt is one submitted run.
type Attempt struct {
	ID              string
	UserID          string
	Skill           string
	TestID          string
	PartIDs         []string
	Kind            string
	ParentAttemptID string
	Mode            string
	TimeLimitSec    int
	DurationSec     int
	Questions       []int
	Answers         Answers
	Results         []Result
	PartScores      map[string]Score
	Correct, Total  int
	Band            *float64
	CreatedAt       time.Time
}

// Score is a correct/total pair.
type Score struct {
	Correct int `json:"correct"`
	Total   int `json:"total"`
}

// Result is the marking of one question.
type Result struct {
	N       int    `json:"n"`
	PartNo  int    `json:"part_no"`
	Type    string `json:"type"`
	Given   string `json:"given"`
	Correct bool   `json:"correct"`
	// OverLimit is set when the answer was marked wrong for going past the group's word limit.
	OverLimit bool `json:"over_limit,omitempty"`
}
