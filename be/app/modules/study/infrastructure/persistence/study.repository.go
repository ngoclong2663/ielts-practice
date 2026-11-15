package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/study/domain"
	"ielts-practice/app/modules/study/domain/entity"
)

var _ domain.IStudyRepository = &StudyRepository{}

type StudyRepository struct {
	db *gorm.DB
}

func NewStudyRepository(db *gorm.DB) domain.IStudyRepository {
	return &StudyRepository{db: db}
}

func (r *StudyRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

// studyCardRow is a card joined with its deck name.
type studyCardRow struct {
	models.Card
	DeckName string
}

func toStudyCard(row studyCardRow) entity.StudyCard {
	return entity.StudyCard{
		ID:        row.ID,
		DeckID:    row.DeckID,
		DeckName:  row.DeckName,
		LexemeID:  row.LexemeID,
		Content:   row.Content,
		Suspended: row.Suspended,
		IsLeech:   row.IsLeech,
		SchedulingState: entity.SchedulingState{
			State:          row.State,
			Due:            row.Due,
			Stability:      row.Stability,
			Difficulty:     row.Difficulty,
			ScheduledDays:  row.ScheduledDays,
			RemainingSteps: int(row.RemainingSteps),
			Reps:           row.Reps,
			Lapses:         row.Lapses,
			LastReview:     row.LastReviewAt,
		},
	}
}

func toStudyCards(rows []studyCardRow) []entity.StudyCard {
	out := make([]entity.StudyCard, len(rows))
	for i := range rows {
		out[i] = toStudyCard(rows[i])
	}
	return out
}

const cardColumns = "c.*, d.name AS deck_name"

// activeCards selects the user's cards in non-archived decks, optionally in one deck.
func (r *StudyRepository) activeCards(ctx context.Context, userID, deckID string) *gorm.DB {
	q := r.conn(ctx).Table("cards c").
		Joins("JOIN decks d ON d.id = c.deck_id AND d.archived_at IS NULL").
		Where("c.user_id = ? AND NOT c.suspended", userID)
	if deckID != "" {
		q = q.Where("c.deck_id = ?", deckID)
	}
	return q
}

func (r *StudyRepository) DeckStats(ctx context.Context, userID, deckID string, tomorrowStart time.Time) ([]entity.DeckStat, error) {
	q := r.conn(ctx).Table("decks d").
		Select(`d.id AS deck_id, d.new_per_day, d.reviews_per_day,
			count(c.id) FILTER (WHERE c.state = 0 AND NOT c.suspended) AS new_available,
			count(c.id) FILTER (WHERE c.state IN (1, 3) AND NOT c.suspended AND c.due < ?) AS learning_due,
			count(c.id) FILTER (WHERE c.state = 2 AND NOT c.suspended AND c.due < ?) AS review_due`, tomorrowStart, tomorrowStart).
		Joins("LEFT JOIN cards c ON c.deck_id = d.id").
		Where("d.user_id = ? AND d.archived_at IS NULL", userID).
		Group("d.id").Order("min(d.created_at)")
	if deckID != "" {
		q = q.Where("d.id = ?", deckID)
	}
	var out []entity.DeckStat
	return out, q.Scan(&out).Error
}

func (r *StudyRepository) TodayStats(ctx context.Context, userID string, todayStart time.Time) (map[string]entity.TodayStat, error) {
	var rows []struct {
		DeckID     string
		Introduced int
		Reviewed   int
	}
	err := r.conn(ctx).Table("review_logs").
		Select(`deck_id,
			count(DISTINCT card_id) FILTER (WHERE state_before = 0) AS introduced,
			count(*) FILTER (WHERE state_before = 2) AS reviewed`).
		Where("user_id = ? AND reviewed_at >= ? AND undone_at IS NULL", userID, todayStart).
		Group("deck_id").Scan(&rows).Error
	out := make(map[string]entity.TodayStat, len(rows))
	for _, row := range rows {
		out[row.DeckID] = entity.TodayStat{Introduced: row.Introduced, Reviewed: row.Reviewed}
	}
	return out, err
}

func (r *StudyRepository) LearningCards(ctx context.Context, userID, deckID string, until time.Time, limit int) ([]entity.StudyCard, error) {
	var rows []studyCardRow
	err := r.activeCards(ctx, userID, deckID).Select(cardColumns).
		Where("c.state IN (1, 3) AND c.due <= ?", until).
		Order("c.due ASC").Limit(limit).Scan(&rows).Error
	return toStudyCards(rows), err
}

// cappedCards returns cards from several decks where each deck contributes at most caps[deck]
// rows (window function), ordered across decks by `order`.
func (r *StudyRepository) cappedCards(ctx context.Context, userID string, caps map[string]int, where string, whereArgs []any, partitionOrder, order string, limit int) ([]entity.StudyCard, error) {
	if len(caps) == 0 || limit <= 0 {
		return nil, nil
	}
	var values strings.Builder
	args := make([]any, 0, len(caps)*2+len(whereArgs)+2)
	i := 0
	for deckID, n := range caps {
		if i > 0 {
			values.WriteString(", ")
		}
		values.WriteString("(?::uuid, ?::int)")
		args = append(args, deckID, n)
		i++
	}
	args = append(args, userID)
	args = append(args, whereArgs...)
	args = append(args, limit)

	sql := fmt.Sprintf(`
		WITH caps(deck_id, n) AS (VALUES %s)
		SELECT * FROM (
			SELECT c.*, d.name AS deck_name, d.created_at AS deck_created_at, caps.n AS cap,
				row_number() OVER (PARTITION BY c.deck_id ORDER BY %s) AS rn
			FROM cards c
			JOIN caps ON caps.deck_id = c.deck_id
			JOIN decks d ON d.id = c.deck_id AND d.archived_at IS NULL
			WHERE c.user_id = ? AND NOT c.suspended AND %s
		) x
		WHERE x.rn <= x.cap
		ORDER BY %s
		LIMIT ?`, values.String(), partitionOrder, where, order)

	var rows []studyCardRow
	if err := r.conn(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return toStudyCards(rows), nil
}

func (r *StudyRepository) ReviewCards(ctx context.Context, userID string, caps map[string]int, before time.Time, limit int) ([]entity.StudyCard, error) {
	return r.cappedCards(ctx, userID, caps, "c.state = 2 AND c.due < ?", []any{before}, "c.due ASC", "x.due ASC", limit)
}

func (r *StudyRepository) NewCards(ctx context.Context, userID string, caps map[string]int, limit int) ([]entity.StudyCard, error) {
	return r.cappedCards(ctx, userID, caps, "c.state = 0", nil, "c.position ASC", "x.deck_created_at ASC, x.position ASC", limit)
}

func (r *StudyRepository) NextLearningDue(ctx context.Context, userID, deckID string, after, before time.Time) (*time.Time, error) {
	var row struct{ Next *time.Time }
	err := r.activeCards(ctx, userID, deckID).
		Where("c.state IN (1, 3) AND c.due > ? AND c.due < ?", after, before).
		Select("min(c.due) AS next").Scan(&row).Error
	return row.Next, err
}

func (r *StudyRepository) GetCard(ctx context.Context, userID, cardID string) (*entity.StudyCard, error) {
	var rows []studyCardRow
	err := r.conn(ctx).Table("cards c").Select(cardColumns).
		Joins("JOIN decks d ON d.id = c.deck_id AND d.archived_at IS NULL").
		Where("c.id = ? AND c.user_id = ?", cardID, userID).Limit(1).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	card := toStudyCard(rows[0])
	return &card, nil
}

func (r *StudyRepository) LockCard(ctx context.Context, userID, cardID string) (*entity.StudyCard, error) {
	var rows []studyCardRow
	err := r.conn(ctx).Table("cards c").Select(cardColumns).
		Joins("JOIN decks d ON d.id = c.deck_id AND d.archived_at IS NULL").
		Where("c.id = ? AND c.user_id = ?", cardID, userID).
		Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "c"}}).
		Limit(1).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	card := toStudyCard(rows[0])
	return &card, nil
}

func (r *StudyRepository) ReviewExists(ctx context.Context, userID, clientReviewID string) (bool, error) {
	var n int64
	err := r.conn(ctx).Model(&models.ReviewLog{}).
		Where("user_id = ? AND client_review_id = ?", userID, clientReviewID).Count(&n).Error
	return n > 0, err
}

func scheduleColumns(s entity.SchedulingState) map[string]any {
	return map[string]any{
		"state":           s.State,
		"due":             s.Due,
		"stability":       s.Stability,
		"difficulty":      s.Difficulty,
		"scheduled_days":  s.ScheduledDays,
		"remaining_steps": s.RemainingSteps,
		"reps":            s.Reps,
		"lapses":          s.Lapses,
		"last_review_at":  s.LastReview,
	}
}

func (r *StudyRepository) SaveSchedule(ctx context.Context, cardID string, s entity.SchedulingState, isLeech bool) error {
	cols := scheduleColumns(s)
	cols["is_leech"] = isLeech
	cols["updated_at"] = time.Now()
	return r.conn(ctx).Model(&models.Card{}).Where("id = ?", cardID).Updates(cols).Error
}

func (r *StudyRepository) InsertLog(ctx context.Context, l *entity.ReviewLogEntity) error {
	b, a := l.Before, l.After
	return r.conn(ctx).Omit("ID").Create(&models.ReviewLog{
		UserID:               l.UserID,
		CardID:               l.CardID,
		DeckID:               l.DeckID,
		ClientReviewID:       l.ClientReviewID,
		Rating:               int16(l.Rating),
		StateBefore:          b.State,
		DueBefore:            b.Due,
		StabilityBefore:      b.Stability,
		DifficultyBefore:     b.Difficulty,
		ScheduledDaysBefore:  b.ScheduledDays,
		RemainingStepsBefore: int16(b.RemainingSteps),
		RepsBefore:           b.Reps,
		LapsesBefore:         b.Lapses,
		LastReviewBefore:     b.LastReview,
		IsLeechBefore:        l.IsLeechBefore,
		StateAfter:           a.State,
		DueAfter:             a.Due,
		StabilityAfter:       a.Stability,
		DifficultyAfter:      a.Difficulty,
		ScheduledDaysAfter:   a.ScheduledDays,
		ElapsedDays:          l.ElapsedDays,
		DurationMs:           l.DurationMs,
		ReviewedAt:           l.ReviewedAt,
	}).Error
}

func (r *StudyRepository) LastUndoableLog(ctx context.Context, userID string, since time.Time) (*entity.ReviewLogEntity, error) {
	var m models.ReviewLog
	err := r.conn(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND undone_at IS NULL AND reviewed_at >= ?", userID, since).
		Order("reviewed_at DESC").First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity.ReviewLogEntity{
		ID: m.ID, UserID: m.UserID, CardID: m.CardID, DeckID: m.DeckID, Rating: entity.Rating(m.Rating),
		IsLeechBefore: m.IsLeechBefore, ReviewedAt: m.ReviewedAt,
		Before: entity.SchedulingState{
			State: m.StateBefore, Due: m.DueBefore, Stability: m.StabilityBefore, Difficulty: m.DifficultyBefore,
			ScheduledDays: m.ScheduledDaysBefore, RemainingSteps: int(m.RemainingStepsBefore),
			Reps: m.RepsBefore, Lapses: m.LapsesBefore, LastReview: m.LastReviewBefore,
		},
	}, nil
}

func (r *StudyRepository) MarkUndone(ctx context.Context, logID string, at time.Time) error {
	return r.conn(ctx).Model(&models.ReviewLog{}).Where("id = ?", logID).Update("undone_at", at).Error
}

func (r *StudyRepository) TodaySummary(ctx context.Context, userID string, todayStart time.Time) (domain.TodaySummary, error) {
	var s domain.TodaySummary
	err := r.conn(ctx).Table("review_logs").
		Select(`count(*) AS reviews,
			count(DISTINCT card_id) FILTER (WHERE state_before = 0) AS new_cards,
			count(*) FILTER (WHERE rating = 1) AS again,
			coalesce(sum(duration_ms), 0) AS duration_ms`).
		Where("user_id = ? AND reviewed_at >= ? AND undone_at IS NULL", userID, todayStart).
		Scan(&s).Error
	return s, err
}

// studyDayExpr maps a timestamp to its study-day date in the user's zone (day starts at dayStartHour).
const studyDayExpr = "((%s AT TIME ZONE ?) - make_interval(hours => ?))::date"

func (r *StudyRepository) StudyDays(ctx context.Context, userID, timezone string, dayStartHour int, since time.Time) ([]time.Time, error) {
	var days []time.Time
	expr := fmt.Sprintf(studyDayExpr, "reviewed_at")
	err := r.conn(ctx).Table("review_logs").
		Select("DISTINCT "+expr+" AS day", timezone, dayStartHour).
		Where("user_id = ? AND reviewed_at >= ? AND undone_at IS NULL", userID, since).
		Order("day DESC").Pluck("day", &days).Error
	return days, err
}

func (r *StudyRepository) DueForecast(ctx context.Context, userID, timezone string, dayStartHour int, until time.Time) ([]domain.DayCount, error) {
	var rows []domain.DayCount
	expr := fmt.Sprintf(studyDayExpr, "c.due")
	err := r.activeCards(ctx, userID, "").
		Select(expr+" AS day, count(*) AS count", timezone, dayStartHour).
		Where("c.state IN (1, 2, 3) AND c.due < ?", until).
		Group("day").Order("day").Scan(&rows).Error
	return rows, err
}

func (r *StudyRepository) Retention(ctx context.Context, userID string, since time.Time) (int, int, error) {
	var row struct {
		Passed int
		Total  int
	}
	err := r.conn(ctx).Table("review_logs").
		Select("count(*) FILTER (WHERE rating > 1) AS passed, count(*) AS total").
		Where("user_id = ? AND state_before = 2 AND reviewed_at >= ? AND undone_at IS NULL", userID, since).
		Scan(&row).Error
	return row.Passed, row.Total, err
}

func (r *StudyRepository) CardTotals(ctx context.Context, userID string) (domain.CardTotals, error) {
	var t domain.CardTotals
	err := r.conn(ctx).Table("cards c").
		Joins("JOIN decks d ON d.id = c.deck_id AND d.archived_at IS NULL").
		Select(`count(*) FILTER (WHERE c.state = 0 AND NOT c.suspended) AS new,
			count(*) FILTER (WHERE c.state IN (1, 3) AND NOT c.suspended) AS learning,
			count(*) FILTER (WHERE c.state = 2 AND NOT c.suspended) AS review,
			count(*) FILTER (WHERE c.suspended) AS suspended`).
		Where("c.user_id = ?", userID).Scan(&t).Error
	return t, err
}
