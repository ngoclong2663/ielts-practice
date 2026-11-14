package domain

import (
	"context"
	"time"

	"ielts-practice/app/modules/study/domain/entity"
)

type DayCount struct {
	Day   time.Time // study-day date (local)
	Count int
}

type TodaySummary struct {
	Reviews    int
	NewCards   int
	Again      int
	DurationMs int64
}

type CardTotals struct {
	New       int
	Learning  int
	Review    int
	Suspended int
}

// IStudyRepository: deckID == "" means all of the user's active decks.
type IStudyRepository interface {
	DeckStats(ctx context.Context, userID, deckID string, tomorrowStart time.Time) ([]entity.DeckStat, error)
	TodayStats(ctx context.Context, userID string, todayStart time.Time) (map[string]entity.TodayStat, error)

	LearningCards(ctx context.Context, userID, deckID string, until time.Time, limit int) ([]entity.StudyCard, error)
	ReviewCards(ctx context.Context, userID string, caps map[string]int, before time.Time, limit int) ([]entity.StudyCard, error)
	NewCards(ctx context.Context, userID string, caps map[string]int, limit int) ([]entity.StudyCard, error)
	NextLearningDue(ctx context.Context, userID, deckID string, after, before time.Time) (*time.Time, error)
	GetCard(ctx context.Context, userID, cardID string) (*entity.StudyCard, error)

	LockCard(ctx context.Context, userID, cardID string) (*entity.StudyCard, error)
	ReviewExists(ctx context.Context, userID, clientReviewID string) (bool, error)
	SaveSchedule(ctx context.Context, cardID string, s entity.SchedulingState, isLeech bool) error
	InsertLog(ctx context.Context, log *entity.ReviewLogEntity) error
	LastUndoableLog(ctx context.Context, userID string, since time.Time) (*entity.ReviewLogEntity, error)
	MarkUndone(ctx context.Context, logID string, at time.Time) error

	TodaySummary(ctx context.Context, userID string, todayStart time.Time) (TodaySummary, error)
	StudyDays(ctx context.Context, userID, timezone string, dayStartHour int, since time.Time) ([]time.Time, error)
	DueForecast(ctx context.Context, userID, timezone string, dayStartHour int, until time.Time) ([]DayCount, error)
	Retention(ctx context.Context, userID string, since time.Time) (passed, total int, err error)
	CardTotals(ctx context.Context, userID string) (CardTotals, error)
}
