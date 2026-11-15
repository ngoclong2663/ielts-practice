package domain

import (
	"context"
	"errors"
	"strings"

	"ielts-practice/app/modules/writing/domain/entity"
)

const (
	StatusInProgress = "in_progress"
	StatusSubmitted  = "submitted"

	KindLesson = "lesson"
	KindCheck  = "check"
	KindSubmit = "submit"

	MinLevel = 1
	MaxLevel = 4
	// MaxSentenceChars bounds one learner sentence; MaxEssayChars the full answer.
	MaxSentenceChars = 400
	MaxEssayChars    = 6000
	// MinSubmitWords: below this the answer is not worth an AI review.
	MinSubmitWords = 20
)

var (
	ErrCoachNotConfigured = errors.New("writing: no AI model is configured")
	ErrCoachFailed        = errors.New("writing: the AI answer was not usable")
)

// ParagraphKeys is the paragraph order of each task (Epic 5 §4).
func ParagraphKeys(taskNo int) []string {
	if taskNo == 1 {
		return []string{"introduction", "overview", "body1", "body2"}
	}
	return []string{"introduction", "body1", "body2", "conclusion"}
}

// Task 2 positions. Task 1 has a single, empty position.
var knownPositions = map[string]string{
	"agree":         "the writer agrees with the statement",
	"disagree":      "the writer disagrees with the statement",
	"partly":        "the writer partly agrees: agrees with some points, disagrees with others",
	"favour_first":  "the writer discusses both views and supports the first view",
	"favour_second": "the writer discusses both views and supports the second view",
	"advantages":    "the writer argues that the advantages outweigh the disadvantages",
	"disadvantages": "the writer argues that the disadvantages outweigh the advantages",
	"default":       "the writer answers every part of the question directly",
}

var defaultPositions = map[string][]string{
	"opinion":                  {"agree", "disagree", "partly"},
	"discussion":               {"favour_first", "favour_second"},
	"advantages_disadvantages": {"advantages", "disadvantages"},
	"problem_solution":         {"default"},
	"two_part":                 {"default"},
}

func ValidPosition(p string) bool { _, ok := knownPositions[p]; return ok }

// PositionBrief describes a position for the lesson prompt.
func PositionBrief(p string) string { return knownPositions[p] }

// Positions a learner may take for a task: the task's own list, or the default for its type.
func Positions(t *entity.Task) []string {
	if t.TaskNo == 1 {
		return []string{""}
	}
	if len(t.Positions) > 0 {
		return t.Positions
	}
	return defaultPositions[t.TaskType]
}

// CountWords counts whitespace-separated words, as IELTS examiners do (numbers count as words).
func CountWords(s string) int { return len(strings.Fields(s)) }

// LessonRequest / CheckRequest / ReviewRequest are what the coach needs from the task.
type LessonRequest struct {
	Task     *entity.Task
	Position string
}

type CheckSentence struct {
	Prompt  string // what the learner saw: the Vietnamese sentence, idea note or English outline
	ModelEN string
	Answer  string
}

type CheckRequest struct {
	Task      *entity.Task
	Paragraph string
	GoalVI    string
	Level     int
	Sentences []CheckSentence
}

type ReviewRequest struct {
	Task      *entity.Task
	Position  string
	Text      string
	WordCount int
}

// ICoach writes lessons and gives feedback (Claude first, Gemini fallback; a fake offline).
type ICoach interface {
	Configured() bool
	WriteLesson(ctx context.Context, req LessonRequest) (entity.LessonContent, entity.AIUsage, error)
	CheckParagraph(ctx context.Context, req CheckRequest) ([]entity.SentenceFeedback, entity.AIUsage, error)
	ReviewEssay(ctx context.Context, req ReviewRequest) (entity.EssayFeedback, entity.AIUsage, error)
}
