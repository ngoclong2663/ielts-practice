package rules

import (
	"math"
	"strings"

	"ielts-practice/app/modules/writing/domain/entity"
)

const (
	MinLevel = 1
	MaxLevel = 4
	// MaxSentenceChars bounds one learner sentence; MaxEssayChars the full answer.
	MaxSentenceChars = 400
	MaxEssayChars    = 6000
	// MinSubmitWords: below this the answer is not worth an AI review.
	MinSubmitWords = 20
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
	"positive":      "the writer argues that this is mainly a positive development",
	"negative":      "the writer argues that this is mainly a negative development",
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

// BandRange turns the four criterion bands into the range shown to the learner. The overall band is the
// mean rounded the IELTS way (to the nearest half band, .25 and .75 rounding up); AI marking is only
// roughly as reliable as an examiner, so the range is that band and the half band below it.
func BandRange(criteria []entity.CriterionFeedback) (float64, float64) {
	if len(criteria) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, c := range criteria {
		sum += c.Band
	}
	overall := math.Floor(sum/float64(len(criteria))*2+0.5) / 2
	return max(0, overall-0.5), overall
}
