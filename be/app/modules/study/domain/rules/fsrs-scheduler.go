package rules

import (
	"math"
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v4"

	"ielts-practice/app/modules/study/domain/entity"
)

const LeechThreshold = 8

type SchedulerConfig struct {
	DesiredRetention float64
	LearningSteps    []float64 // minutes
	RelearningSteps  []float64 // minutes
	MaxIntervalDays  float64
	EnableFuzz       bool
}

// Scheduler wraps go-fsrs (FSRS-6). It is pure: no I/O, time is passed in.
type Scheduler struct {
	f *fsrs.FSRS
}

func NewScheduler(cfg SchedulerConfig) *Scheduler {
	p := fsrs.DefaultParam()
	if cfg.DesiredRetention > 0 && cfg.DesiredRetention < 1 {
		p.RequestRetention = cfg.DesiredRetention
	}
	if cfg.MaxIntervalDays >= 1 {
		p.MaximumInterval = math.Min(cfg.MaxIntervalDays, 36500)
	}
	if cfg.LearningSteps != nil {
		p.LearningSteps = cfg.LearningSteps
	}
	if cfg.RelearningSteps != nil {
		p.RelearningSteps = cfg.RelearningSteps
	}
	p.EnableFuzz = cfg.EnableFuzz
	return &Scheduler{f: fsrs.NewFSRS(p)}
}

func toFSRS(s entity.SchedulingState) fsrs.Card {
	c := fsrs.Card{
		Due:            s.Due,
		Stability:      s.Stability,
		Difficulty:     s.Difficulty,
		ScheduledDays:  uint64(max(s.ScheduledDays, 0)),
		Reps:           uint64(max(s.Reps, 0)),
		Lapses:         uint64(max(s.Lapses, 0)),
		State:          fsrs.State(s.State),
		RemainingSteps: s.RemainingSteps,
	}
	if s.LastReview != nil {
		c.LastReview = *s.LastReview
	}
	return c
}

func fromFSRS(c fsrs.Card) entity.SchedulingState {
	s := entity.SchedulingState{
		State:          int16(c.State),
		Due:            c.Due,
		Stability:      c.Stability,
		Difficulty:     c.Difficulty,
		ScheduledDays:  int(c.ScheduledDays),
		RemainingSteps: c.RemainingSteps,
		Reps:           int(c.Reps),
		Lapses:         int(c.Lapses),
	}
	if !c.LastReview.IsZero() {
		lr := c.LastReview
		s.LastReview = &lr
	}
	return s
}

// Preview returns, for each rating, how long until the card would be shown again.
func (s *Scheduler) Preview(state entity.SchedulingState, now time.Time) (map[entity.Rating]time.Duration, error) {
	log, err := s.f.Repeat(toFSRS(state), now)
	if err != nil {
		return nil, err
	}
	out := make(map[entity.Rating]time.Duration, 4)
	for _, r := range entity.Ratings {
		out[r] = log[fsrs.Rating(r)].Card.Due.Sub(now)
	}
	return out, nil
}

// Answer applies a rating and returns the new state plus whole days since the last review.
func (s *Scheduler) Answer(state entity.SchedulingState, rating entity.Rating, now time.Time) (entity.SchedulingState, int, error) {
	info, err := s.f.Next(toFSRS(state), now, fsrs.Rating(rating))
	if err != nil {
		return entity.SchedulingState{}, 0, err
	}
	elapsed := 0
	if state.LastReview != nil {
		elapsed = int(now.Sub(*state.LastReview).Hours() / 24)
	}
	return fromFSRS(info.Card), elapsed, nil
}

// BecomesLeech reports whether this answer made the card a leech (lapse count hit the threshold).
func BecomesLeech(before, after entity.SchedulingState) bool {
	return after.Lapses > before.Lapses && after.Lapses >= LeechThreshold
}
