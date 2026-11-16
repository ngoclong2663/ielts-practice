package persistence

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"ielts-practice/app/modules/admin/domain"
	"ielts-practice/app/modules/admin/domain/entity"
	"ielts-practice/app/modules/database"
)

var _ domain.IAdminRepository = &AdminRepository{}

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) domain.IAdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

func windowArgs(w entity.Window) map[string]any {
	return map[string]any{
		"tz": w.Timezone, "today": w.Today, "since7": w.Since7d, "since30": w.Since30d,
	}
}

// activityCTE: one row per activity of a (not deleted) user since @since30. Active means the user
// studied a card, wrote, generated cards, asked for a word list or signed in (Epic 6 §6).
const activityCTE = `
activity AS (
	SELECT a.user_id, a.at FROM (
		SELECT user_id, reviewed_at AS at FROM review_logs WHERE undone_at IS NULL AND reviewed_at >= @since30
		UNION ALL SELECT user_id, updated_at FROM writing_attempts WHERE updated_at >= @since30
		UNION ALL SELECT user_id, created_at FROM generation_jobs WHERE created_at >= @since30
		UNION ALL SELECT user_id, created_at FROM topic_suggestion_requests WHERE created_at >= @since30
		UNION ALL SELECT id, last_login_at FROM users WHERE last_login_at >= @since30
	) a
	JOIN users u ON u.id = a.user_id AND u.deleted_at IS NULL
)`

type totalsRow struct {
	UsersTotal  int `gorm:"column:users_total"`
	UsersNew7d  int `gorm:"column:users_new_7d"`
	UsersNew30d int `gorm:"column:users_new_30d"`
	Active1d    int `gorm:"column:active_1d"`
	Active7d    int `gorm:"column:active_7d"`
	Active30d   int `gorm:"column:active_30d"`

	Decks           int `gorm:"column:decks"`
	Cards           int `gorm:"column:cards"`
	CardsNew        int `gorm:"column:cards_new"`
	CardsLearning   int `gorm:"column:cards_learning"`
	CardsReview     int `gorm:"column:cards_review"`
	CardsRelearning int `gorm:"column:cards_relearning"`
	ReviewsTotal    int `gorm:"column:reviews_total"`
	Reviews7d       int `gorm:"column:reviews_7d"`
	LexemesCached   int `gorm:"column:lexemes_cached"`

	WordsGeneratedTotal int `gorm:"column:words_generated_total"`
	WordsGenerated7d    int `gorm:"column:words_generated_7d"`
	WordsCachedTotal    int `gorm:"column:words_cached_total"`
	WordsFailed7d       int `gorm:"column:words_failed_7d"`
	TopicListsTotal     int `gorm:"column:topic_lists_total"`
	TopicLists7d        int `gorm:"column:topic_lists_7d"`
	WritingLessonsTotal int `gorm:"column:writing_lessons_total"`
	WritingChecksTotal  int `gorm:"column:writing_checks_total"`
	WritingChecks7d     int `gorm:"column:writing_checks_7d"`
	WritingReviewsTotal int `gorm:"column:writing_reviews_total"`
	WritingReviews7d    int `gorm:"column:writing_reviews_7d"`
	InputTokens30d      int `gorm:"column:input_tokens_30d"`
	OutputTokens30d     int `gorm:"column:output_tokens_30d"`

	AttemptsTotal     int      `gorm:"column:attempts_total"`
	AttemptsSubmitted int      `gorm:"column:attempts_submitted"`
	AvgBand30d        *float64 `gorm:"column:avg_band_30d"`

	FeedbackNew   int `gorm:"column:feedback_new"`
	FeedbackTotal int `gorm:"column:feedback_total"`

	ReportsTotal       int `gorm:"column:reports_total"`
	Reports7d          int `gorm:"column:reports_7d"`
	ReportsNeedsReview int `gorm:"column:reports_needs_review"`
}

// One round trip: every total is a scalar subquery.
const totalsSQL = `WITH ` + activityCTE + `,
live_cards AS (
	SELECT c.state FROM cards c JOIN decks d ON d.id = c.deck_id AND d.archived_at IS NULL
)
SELECT
	(SELECT count(*) FROM users WHERE deleted_at IS NULL) AS users_total,
	(SELECT count(*) FROM users WHERE deleted_at IS NULL AND created_at >= @since7) AS users_new_7d,
	(SELECT count(*) FROM users WHERE deleted_at IS NULL AND created_at >= @since30) AS users_new_30d,
	(SELECT count(DISTINCT user_id) FROM activity WHERE at >= @today) AS active_1d,
	(SELECT count(DISTINCT user_id) FROM activity WHERE at >= @since7) AS active_7d,
	(SELECT count(DISTINCT user_id) FROM activity) AS active_30d,

	(SELECT count(*) FROM decks WHERE archived_at IS NULL) AS decks,
	(SELECT count(*) FROM live_cards) AS cards,
	(SELECT count(*) FROM live_cards WHERE state = 0) AS cards_new,
	(SELECT count(*) FROM live_cards WHERE state = 1) AS cards_learning,
	(SELECT count(*) FROM live_cards WHERE state = 2) AS cards_review,
	(SELECT count(*) FROM live_cards WHERE state = 3) AS cards_relearning,
	(SELECT count(*) FROM review_logs WHERE undone_at IS NULL) AS reviews_total,
	(SELECT count(*) FROM review_logs WHERE undone_at IS NULL AND reviewed_at >= @since7) AS reviews_7d,
	(SELECT count(*) FROM lexemes WHERE status = 'ready') AS lexemes_cached,

	(SELECT count(*) FROM generation_job_items WHERE charged) AS words_generated_total,
	(SELECT count(*) FROM generation_job_items WHERE charged AND created_at >= @since7) AS words_generated_7d,
	(SELECT count(*) FROM generation_job_items WHERE cached) AS words_cached_total,
	(SELECT count(*) FROM generation_job_items WHERE status = 'failed' AND created_at >= @since7) AS words_failed_7d,
	(SELECT count(*) FROM topic_suggestion_requests) AS topic_lists_total,
	(SELECT count(*) FROM topic_suggestion_requests WHERE created_at >= @since7) AS topic_lists_7d,
	(SELECT count(*) FROM writing_ai_requests WHERE kind = 'lesson') AS writing_lessons_total,
	(SELECT count(*) FROM writing_ai_requests WHERE kind = 'check') AS writing_checks_total,
	(SELECT count(*) FROM writing_ai_requests WHERE kind = 'check' AND created_at >= @since7) AS writing_checks_7d,
	(SELECT count(*) FROM writing_ai_requests WHERE kind = 'submit') AS writing_reviews_total,
	(SELECT count(*) FROM writing_ai_requests WHERE kind = 'submit' AND created_at >= @since7) AS writing_reviews_7d,
	(SELECT COALESCE(sum(input_tokens), 0) FROM generation_jobs WHERE created_at >= @since30)
		+ (SELECT COALESCE(sum(input_tokens), 0) FROM topic_suggestion_requests WHERE created_at >= @since30)
		+ (SELECT COALESCE(sum(input_tokens), 0) FROM writing_ai_requests WHERE created_at >= @since30) AS input_tokens_30d,
	(SELECT COALESCE(sum(output_tokens), 0) FROM generation_jobs WHERE created_at >= @since30)
		+ (SELECT COALESCE(sum(output_tokens), 0) FROM topic_suggestion_requests WHERE created_at >= @since30)
		+ (SELECT COALESCE(sum(output_tokens), 0) FROM writing_ai_requests WHERE created_at >= @since30) AS output_tokens_30d,

	(SELECT count(*) FROM writing_attempts) AS attempts_total,
	(SELECT count(*) FROM writing_attempts WHERE status = 'submitted') AS attempts_submitted,
	(SELECT round(avg((band_low + band_high) / 2), 2)::float8 FROM writing_submissions WHERE created_at >= @since30) AS avg_band_30d,

	(SELECT count(*) FROM feedback WHERE status = 'new') AS feedback_new,
	(SELECT count(*) FROM feedback) AS feedback_total,

	(SELECT count(*) FROM lexeme_reports) AS reports_total,
	(SELECT count(*) FROM lexeme_reports WHERE created_at >= @since7) AS reports_7d,
	(SELECT count(*) FROM lexemes WHERE status = 'needs_review') AS reports_needs_review`

func (r *AdminRepository) Totals(ctx context.Context, w entity.Window) (entity.Totals, error) {
	var row totalsRow
	if err := r.conn(ctx).Raw(totalsSQL, windowArgs(w)).Scan(&row).Error; err != nil {
		return entity.Totals{}, err
	}
	return entity.Totals{
		Users: entity.UserTotals{
			Total: row.UsersTotal, New7d: row.UsersNew7d, New30d: row.UsersNew30d,
			Active1d: row.Active1d, Active7d: row.Active7d, Active30d: row.Active30d,
		},
		Vocab: entity.VocabTotals{
			Decks: row.Decks, Cards: row.Cards,
			CardsNew: row.CardsNew, CardsLearning: row.CardsLearning, CardsReview: row.CardsReview, CardsRelearning: row.CardsRelearning,
			ReviewsTotal: row.ReviewsTotal, Reviews7d: row.Reviews7d, LexemesCached: row.LexemesCached,
		},
		AI: entity.AITotals{
			WordsGeneratedTotal: row.WordsGeneratedTotal, WordsGenerated7d: row.WordsGenerated7d,
			WordsCachedTotal: row.WordsCachedTotal, WordsFailed7d: row.WordsFailed7d,
			TopicListsTotal: row.TopicListsTotal, TopicLists7d: row.TopicLists7d,
			WritingLessonsTotal: row.WritingLessonsTotal,
			WritingChecksTotal:  row.WritingChecksTotal, WritingChecks7d: row.WritingChecks7d,
			WritingReviewsTotal: row.WritingReviewsTotal, WritingReviews7d: row.WritingReviews7d,
			InputTokens30d: row.InputTokens30d, OutputTokens30d: row.OutputTokens30d,
		},
		Writing:        entity.WritingTotals{AttemptsTotal: row.AttemptsTotal, AttemptsSubmitted: row.AttemptsSubmitted, AvgBand30d: row.AvgBand30d},
		Feedback:       entity.FeedbackTotals{New: row.FeedbackNew, Total: row.FeedbackTotal},
		ContentReports: entity.ReportTotals{Total: row.ReportsTotal, Last7d: row.Reports7d, NeedsReview: row.ReportsNeedsReview},
	}, nil
}

// dailySQL: one row per (kind, local date). AI requests = charged generation items + word-list
// requests + writing AI requests; active = distinct users per day.
const dailySQL = `WITH ` + activityCTE + `
SELECT kind, day, count(*) AS n FROM (
	SELECT 'signups' AS kind, to_char(created_at AT TIME ZONE @tz, 'YYYY-MM-DD') AS day
		FROM users WHERE deleted_at IS NULL AND created_at >= @since30
	UNION ALL SELECT 'reviews', to_char(reviewed_at AT TIME ZONE @tz, 'YYYY-MM-DD')
		FROM review_logs WHERE undone_at IS NULL AND reviewed_at >= @since30
	UNION ALL SELECT 'ai', to_char(created_at AT TIME ZONE @tz, 'YYYY-MM-DD')
		FROM generation_job_items WHERE charged AND created_at >= @since30
	UNION ALL SELECT 'ai', to_char(created_at AT TIME ZONE @tz, 'YYYY-MM-DD')
		FROM topic_suggestion_requests WHERE created_at >= @since30
	UNION ALL SELECT 'ai', to_char(created_at AT TIME ZONE @tz, 'YYYY-MM-DD')
		FROM writing_ai_requests WHERE created_at >= @since30
	UNION ALL SELECT 'active', day FROM (
		SELECT DISTINCT user_id, to_char(at AT TIME ZONE @tz, 'YYYY-MM-DD') AS day FROM activity
	) a
) t
GROUP BY kind, day`

func (r *AdminRepository) Daily(ctx context.Context, w entity.Window) (map[string]entity.DayStats, error) {
	var rows []struct {
		Kind string
		Day  string
		N    int
	}
	if err := r.conn(ctx).Raw(dailySQL, windowArgs(w)).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := map[string]entity.DayStats{}
	for _, row := range rows {
		d := out[row.Day]
		switch row.Kind {
		case "signups":
			d.Signups = row.N
		case "reviews":
			d.Reviews = row.N
		case "ai":
			d.AIRequests = row.N
		case "active":
			d.ActiveUsers = row.N
		}
		out[row.Day] = d
	}
	return out, nil
}

func (r *AdminRepository) RecentReports(ctx context.Context, limit int) ([]entity.ContentReport, error) {
	var rows []entity.ContentReport
	err := r.conn(ctx).Raw(`
		SELECT r.id, l.headword, r.reason, r.note, COALESCE(u.email, '') AS user_email, r.created_at
		FROM lexeme_reports r
		JOIN lexemes l ON l.id = r.lexeme_id
		LEFT JOIN users u ON u.id = r.user_id
		ORDER BY r.created_at DESC, r.id DESC
		LIMIT @limit`, map[string]any{"limit": limit}).Scan(&rows).Error
	return rows, err
}

type userRow struct {
	ID              string     `gorm:"column:id"`
	Email           string     `gorm:"column:email"`
	DisplayName     string     `gorm:"column:display_name"`
	AvatarURL       string     `gorm:"column:avatar_url"`
	UILocale        string     `gorm:"column:ui_locale"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	LastLoginAt     *time.Time `gorm:"column:last_login_at"`
	LastActiveAt    *time.Time `gorm:"column:last_active_at"`
	Decks           int        `gorm:"column:decks"`
	Cards           int        `gorm:"column:cards"`
	ReviewsTotal    int        `gorm:"column:reviews_total"`
	Reviews7d       int        `gorm:"column:reviews_7d"`
	AIWords         int        `gorm:"column:ai_words"`
	TopicLists      int        `gorm:"column:topic_lists"`
	WritingAttempts int        `gorm:"column:writing_attempts"`
	WritingSubmit   int        `gorm:"column:writing_submitted"`
	WritingAICalls  int        `gorm:"column:writing_ai_calls"`
	Feedback        int        `gorm:"column:feedback"`
}

const usersBaseWhere = `deleted_at IS NULL AND (@q = '' OR email ILIKE @pattern OR display_name ILIKE @pattern)`

// usersSQL: per-user aggregates in one query (one GROUP BY per table, joined to the page's users).
// last_active_at is the latest activity of any kind (GREATEST ignores NULLs).
const usersSQL = `
WITH base AS (
	SELECT id, email, display_name, avatar_url, ui_locale, created_at, last_login_at
	FROM users WHERE ` + usersBaseWhere + `
),
deck_counts AS (
	SELECT user_id, count(*) AS n FROM decks WHERE archived_at IS NULL GROUP BY user_id
),
card_counts AS (
	SELECT c.user_id, count(*) AS n FROM cards c JOIN decks d ON d.id = c.deck_id AND d.archived_at IS NULL GROUP BY c.user_id
),
review_counts AS (
	SELECT user_id, count(*) AS total, count(*) FILTER (WHERE reviewed_at >= @since7) AS last_7d, max(reviewed_at) AS last_at
	FROM review_logs WHERE undone_at IS NULL GROUP BY user_id
),
word_counts AS (
	SELECT user_id, count(*) AS n FROM generation_job_items WHERE charged GROUP BY user_id
),
job_last AS (
	SELECT user_id, max(created_at) AS last_at FROM generation_jobs GROUP BY user_id
),
topic_counts AS (
	SELECT user_id, count(*) AS n, max(created_at) AS last_at FROM topic_suggestion_requests GROUP BY user_id
),
writing_counts AS (
	SELECT user_id, count(*) AS n, count(*) FILTER (WHERE status = 'submitted') AS submitted, max(updated_at) AS last_at
	FROM writing_attempts GROUP BY user_id
),
writing_ai AS (
	SELECT user_id, count(*) AS n FROM writing_ai_requests GROUP BY user_id
),
feedback_counts AS (
	SELECT user_id, count(*) AS n FROM feedback GROUP BY user_id
),
report AS (
	SELECT b.*,
		GREATEST(rc.last_at, jl.last_at, tc.last_at, wc.last_at, b.last_login_at) AS last_active_at,
		COALESCE(dc.n, 0) AS decks,
		COALESCE(cc.n, 0) AS cards,
		COALESCE(rc.total, 0) AS reviews_total,
		COALESCE(rc.last_7d, 0) AS reviews_7d,
		COALESCE(wo.n, 0) AS ai_words,
		COALESCE(tc.n, 0) AS topic_lists,
		COALESCE(wc.n, 0) AS writing_attempts,
		COALESCE(wc.submitted, 0) AS writing_submitted,
		COALESCE(wa.n, 0) AS writing_ai_calls,
		COALESCE(fc.n, 0) AS feedback
	FROM base b
	LEFT JOIN deck_counts dc ON dc.user_id = b.id
	LEFT JOIN card_counts cc ON cc.user_id = b.id
	LEFT JOIN review_counts rc ON rc.user_id = b.id
	LEFT JOIN word_counts wo ON wo.user_id = b.id
	LEFT JOIN job_last jl ON jl.user_id = b.id
	LEFT JOIN topic_counts tc ON tc.user_id = b.id
	LEFT JOIN writing_counts wc ON wc.user_id = b.id
	LEFT JOIN writing_ai wa ON wa.user_id = b.id
	LEFT JOIN feedback_counts fc ON fc.user_id = b.id
)
SELECT * FROM report
ORDER BY `

// userOrder: whitelisted ORDER BY clauses (never built from user input).
var userOrder = map[string]string{
	entity.SortLastActive: "last_active_at DESC NULLS LAST, created_at DESC, id",
	entity.SortCreated:    "created_at DESC, id",
	entity.SortReviews:    "reviews_total DESC, last_active_at DESC NULLS LAST, id",
	entity.SortAI:         "(ai_words + topic_lists + writing_ai_calls) DESC, last_active_at DESC NULLS LAST, id",
}

func (r *AdminRepository) ListUsers(ctx context.Context, f entity.UserFilter) ([]entity.UserActivity, int64, error) {
	order, ok := userOrder[f.Sort]
	if !ok {
		order = userOrder[entity.SortLastActive]
	}
	args := map[string]any{
		"q": f.Query, "pattern": "%" + escapeLike(f.Query) + "%", "since7": f.Since7d,
		"limit": f.PageSize, "offset": (f.Page - 1) * f.PageSize,
	}
	var total int64
	if err := r.conn(ctx).Raw(`SELECT count(*) FROM users WHERE `+usersBaseWhere, args).Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []userRow
	if err := r.conn(ctx).Raw(usersSQL+order+" LIMIT @limit OFFSET @offset", args).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]entity.UserActivity, len(rows))
	for i, row := range rows {
		out[i] = entity.UserActivity(row)
	}
	return out, total, nil
}

// escapeLike makes the search text literal inside ILIKE (backslash is Postgres's default escape).
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
