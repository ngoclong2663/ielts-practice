package rules

import "ielts-practice/app/modules/exam/domain/entity"

// bandStep is the lowest raw score (out of 40) that reaches a band.
type bandStep struct {
	min  int
	band float64
}

// The published Academic conversion tables (Epic 10 §7.3). ielts.org gives only anchor points and says the
// cut-offs vary slightly between test versions, so the result is shown as an estimate.
var (
	listeningBands = []bandStep{{39, 9}, {37, 8.5}, {35, 8}, {32, 7.5}, {30, 7}, {26, 6.5}, {23, 6}, {18, 5.5}, {16, 5}, {13, 4.5}, {10, 4}}
	readingBands   = []bandStep{{39, 9}, {37, 8.5}, {35, 8}, {33, 7.5}, {30, 7}, {27, 6.5}, {23, 6}, {19, 5.5}, {15, 5}, {13, 4.5}, {10, 4}}
)

// FullTestQuestions is the number of questions a band needs.
const FullTestQuestions = 40

// Band converts a full test's raw score to a band. Below 10 it returns nil ("below 4.0").
func Band(skill string, correct int) *float64 {
	table := readingBands
	if skill == entity.SkillListening {
		table = listeningBands
	}
	for _, s := range table {
		if correct >= s.min {
			b := s.band
			return &b
		}
	}
	return nil
}
