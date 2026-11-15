package rules

import (
	"strings"
	"testing"

	"ielts-practice/app/modules/study/domain/entity"
)

func cards(prefix string, n int) []entity.StudyCard {
	out := make([]entity.StudyCard, n)
	for i := range out {
		out[i] = entity.StudyCard{ID: prefix + string(rune('a'+i))}
	}
	return out
}

func ids(cs []entity.StudyCard) string {
	parts := make([]string, len(cs))
	for i, c := range cs {
		parts[i] = c.ID
	}
	return strings.Join(parts, ",")
}

func TestLearningFirst(t *testing.T) {
	q := BuildQueue(QueueInput{
		LearningNow: cards("L", 2), Reviews: cards("R", 5), News: cards("N", 5),
		ReviewRemaining: 5, NewRemaining: 5, Size: 4,
	})
	if !strings.HasPrefix(ids(q), "La,Lb,") || len(q) != 4 {
		t.Fatalf("learning cards must come first: %s", ids(q))
	}
}

func TestNewSpreadEvenlyAmongReviews(t *testing.T) {
	// 4 reviews and 2 new: new cards should appear roughly every third card.
	q := BuildQueue(QueueInput{
		Reviews: cards("R", 4), News: cards("N", 2),
		ReviewRemaining: 4, NewRemaining: 2, Size: 6,
	})
	if got := ids(q); got != "Na,Ra,Rb,Nb,Rc,Rd" {
		t.Fatalf("unexpected order %s", got)
	}
}

func TestSpreadContinuesAfterStudyingSomeToday(t *testing.T) {
	// Started the day with 4 reviews + 2 new; 1 new and 2 reviews done, so 1 new remains early.
	q := BuildQueue(QueueInput{
		Reviews: cards("R", 2), News: cards("N", 1),
		ReviewRemaining: 2, NewRemaining: 1, ReviewDoneToday: 2, NewDoneToday: 1, Size: 3,
	})
	if got := ids(q); got != "Na,Ra,Rb" {
		t.Fatalf("unexpected order %s", got)
	}
}

func TestLearnAheadOnlyWhenNothingElse(t *testing.T) {
	q := BuildQueue(QueueInput{Reviews: cards("R", 1), ReviewRemaining: 1, LearningAhead: cards("A", 1), Size: 5})
	if ids(q) != "Ra" {
		t.Fatalf("learn-ahead must wait while other cards exist: %s", ids(q))
	}
	q = BuildQueue(QueueInput{LearningAhead: cards("A", 2), Size: 5})
	if ids(q) != "Aa,Ab" {
		t.Fatalf("learn-ahead should be shown when nothing else is left: %s", ids(q))
	}
}

func TestLimitsRespected(t *testing.T) {
	q := BuildQueue(QueueInput{Reviews: cards("R", 5), News: cards("N", 5), ReviewRemaining: 1, NewRemaining: 0, Size: 5})
	if ids(q) != "Ra" {
		t.Fatalf("limits ignored: %s", ids(q))
	}
}

func TestApplyLimits(t *testing.T) {
	ten := 10
	stat := entity.DeckStat{NewPerDay: &ten, NewAvailable: 50, ReviewDue: 300}
	n, r := ApplyLimits(stat, entity.TodayStat{Introduced: 4, Reviewed: 150}, 20, 200)
	if n != 6 || r != 50 {
		t.Fatalf("deck override and today's progress: got new %d review %d", n, r)
	}
	n, _ = ApplyLimits(entity.DeckStat{NewAvailable: 3}, entity.TodayStat{}, 20, 200)
	if n != 3 {
		t.Fatalf("cannot exceed available cards: %d", n)
	}
	n, _ = ApplyLimits(entity.DeckStat{NewAvailable: 30}, entity.TodayStat{Introduced: 25}, 20, 200)
	if n != 0 {
		t.Fatalf("over the limit must not go negative: %d", n)
	}
}
