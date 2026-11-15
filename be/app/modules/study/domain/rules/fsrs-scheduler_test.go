package rules

import (
	"testing"
	"time"

	"ielts-practice/app/modules/study/domain/entity"
)

var t0 = time.Date(2026, 11, 14, 7, 0, 0, 0, time.UTC)

func newScheduler() *Scheduler {
	return NewScheduler(SchedulerConfig{
		DesiredRetention: 0.9,
		LearningSteps:    []float64{1, 10},
		RelearningSteps:  []float64{10},
		MaxIntervalDays:  36500,
		EnableFuzz:       false, // deterministic tests
	})
}

func newCard() entity.SchedulingState {
	return entity.SchedulingState{State: entity.StateNew, Due: t0}
}

func answer(t *testing.T, s *Scheduler, st entity.SchedulingState, r entity.Rating, at time.Time) entity.SchedulingState {
	t.Helper()
	next, _, err := s.Answer(st, r, at)
	if err != nil {
		t.Fatalf("answer: %v", err)
	}
	return next
}

func TestNewCardLearningSteps(t *testing.T) {
	s := newScheduler()

	again := answer(t, s, newCard(), entity.Again, t0)
	if again.State != entity.StateLearning || again.Due.Sub(t0) != time.Minute {
		t.Fatalf("Again on new: want learning in 1m, got state %d in %v", again.State, again.Due.Sub(t0))
	}

	hard := answer(t, s, newCard(), entity.Hard, t0)
	if hard.State != entity.StateLearning || hard.Due.Sub(t0) != 6*time.Minute {
		t.Fatalf("Hard on first step: want ~6m (avg of 1m and 10m), got %v", hard.Due.Sub(t0))
	}

	good := answer(t, s, newCard(), entity.Good, t0)
	if good.State != entity.StateLearning || good.Due.Sub(t0) != 10*time.Minute {
		t.Fatalf("Good on new: want next step 10m, got state %d in %v", good.State, good.Due.Sub(t0))
	}

	// Good again after the last step graduates the card to review (days).
	t1 := t0.Add(10 * time.Minute)
	graduated := answer(t, s, good, entity.Good, t1)
	if graduated.State != entity.StateReview || graduated.ScheduledDays < 1 {
		t.Fatalf("Good on last step should graduate: state %d, %d days", graduated.State, graduated.ScheduledDays)
	}

	easy := answer(t, s, newCard(), entity.Easy, t0)
	if easy.State != entity.StateReview || easy.ScheduledDays <= graduated.ScheduledDays {
		t.Fatalf("Easy should graduate immediately with a longer interval: %d days vs %d", easy.ScheduledDays, graduated.ScheduledDays)
	}
}

func reviewCard(t *testing.T, s *Scheduler) (entity.SchedulingState, time.Time) {
	t.Helper()
	st := answer(t, s, newCard(), entity.Good, t0)
	st = answer(t, s, st, entity.Good, t0.Add(10*time.Minute))
	return st, st.Due
}

func TestReviewIntervalsAreOrdered(t *testing.T) {
	s := newScheduler()
	st, due := reviewCard(t, s)
	preview, err := s.Preview(st, due)
	if err != nil {
		t.Fatal(err)
	}
	if !(preview[entity.Again] < preview[entity.Hard] && preview[entity.Hard] < preview[entity.Good] && preview[entity.Good] < preview[entity.Easy]) {
		t.Fatalf("want Again < Hard < Good < Easy, got %v", preview)
	}
}

func TestLapseGoesToRelearning(t *testing.T) {
	s := newScheduler()
	st, due := reviewCard(t, s)
	lapsed := answer(t, s, st, entity.Again, due)
	if lapsed.State != entity.StateRelearning || lapsed.Lapses != st.Lapses+1 {
		t.Fatalf("Again on review: want relearning with lapses+1, got state %d lapses %d", lapsed.State, lapsed.Lapses)
	}
	if lapsed.Due.Sub(due) != 10*time.Minute {
		t.Fatalf("relearning step should be 10m, got %v", lapsed.Due.Sub(due))
	}
	if lapsed.Stability >= st.Stability {
		t.Fatalf("stability should drop after a lapse: %v -> %v", st.Stability, lapsed.Stability)
	}
	back := answer(t, s, lapsed, entity.Good, lapsed.Due)
	if back.State != entity.StateReview {
		t.Fatalf("Good after relearning should return to review, got %d", back.State)
	}
}

func TestLateReviewGetsLongerInterval(t *testing.T) {
	s := newScheduler()
	st, due := reviewCard(t, s)
	onTime := answer(t, s, st, entity.Good, due)
	late := answer(t, s, st, entity.Good, due.Add(10*24*time.Hour))
	if late.ScheduledDays <= onTime.ScheduledDays {
		t.Fatalf("a late successful review should grow the interval more: on time %d, late %d", onTime.ScheduledDays, late.ScheduledDays)
	}
}

func TestBecomesLeech(t *testing.T) {
	before := entity.SchedulingState{Lapses: 7}
	if !BecomesLeech(before, entity.SchedulingState{Lapses: 8}) {
		t.Fatal("8th lapse should make a leech")
	}
	if BecomesLeech(entity.SchedulingState{Lapses: 8}, entity.SchedulingState{Lapses: 8}) {
		t.Fatal("no new lapse, no change")
	}
}
