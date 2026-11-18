package rules

import (
	"slices"
	"testing"

	"ielts-practice/app/modules/writing/domain/entity"
)

func TestBandRange(t *testing.T) {
	bands := func(b ...float64) []entity.CriterionFeedback {
		out := make([]entity.CriterionFeedback, len(b))
		for i, v := range b {
			out[i] = entity.CriterionFeedback{Band: v}
		}
		return out
	}
	cases := []struct {
		criteria  []entity.CriterionFeedback
		low, high float64
	}{
		{bands(6, 6, 6, 6), 5.5, 6}, // mean 6.0
		{bands(6, 6, 6, 5), 5.5, 6}, // mean 5.75 rounds up to 6.0
		{bands(6, 5, 5, 5), 5, 5.5}, // mean 5.25 rounds up to 5.5
		{bands(7, 6, 6, 6), 6, 6.5}, // mean 6.25 rounds up to 6.5
		{bands(0, 0, 0, 0), 0, 0},   // never below 0
		{nil, 0, 0},
	}
	for i, c := range cases {
		low, high := BandRange(c.criteria)
		if low != c.low || high != c.high {
			t.Errorf("case %d: got %.1f–%.1f, want %.1f–%.1f", i, low, high, c.low, c.high)
		}
	}
}

func TestParagraphKeys(t *testing.T) {
	if got := ParagraphKeys(1); !slices.Equal(got, []string{"introduction", "overview", "body1", "body2"}) {
		t.Errorf("task 1: %v", got)
	}
	if got := ParagraphKeys(2); !slices.Equal(got, []string{"introduction", "body1", "body2", "conclusion"}) {
		t.Errorf("task 2: %v", got)
	}
}

func TestPositions(t *testing.T) {
	if got := Positions(&entity.Task{TaskNo: 1}); !slices.Equal(got, []string{""}) {
		t.Errorf("task 1 has no position: %v", got)
	}
	if got := Positions(&entity.Task{TaskNo: 2, TaskType: "discussion"}); !slices.Equal(got, []string{"favour_first", "favour_second"}) {
		t.Errorf("discussion default: %v", got)
	}
	if got := Positions(&entity.Task{TaskNo: 2, TaskType: "two_part", Positions: []string{"agree", "disagree"}}); !slices.Equal(got, []string{"agree", "disagree"}) {
		t.Errorf("task's own positions win: %v", got)
	}
}

func TestVocabVisible(t *testing.T) {
	for level := 1; level <= MaxLevel; level++ {
		if got := VocabVisible(level, false); got != (level < MaxLevel) {
			t.Errorf("level %d while writing: %v", level, got)
		}
		if !VocabVisible(level, true) {
			t.Errorf("level %d submitted: hidden", level)
		}
	}
}
