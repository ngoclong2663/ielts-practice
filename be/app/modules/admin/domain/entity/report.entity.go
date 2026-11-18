package entity

// Value objects only (read-only reports, no identity of their own): Window, Totals, DayStats,
// ContentReport, UserActivity, UserFilter, AIUsageRow, ModelPrice, AICost.

import "time"

const DailyDays = 30

// Window is "now" on the board: day boundaries in ADMIN_TIMEZONE. Since7d and Since30d start
// 6 and 29 days before today, so "7 days" means today and the 6 days before.
type Window struct {
	Now      time.Time
	Timezone string
	Today    time.Time
	Since7d  time.Time
	Since30d time.Time
}

type UserTotals struct {
	Total, New7d, New30d, Active1d, Active7d, Active30d int
}

type VocabTotals struct {
	Decks, Cards                                          int
	CardsNew, CardsLearning, CardsReview, CardsRelearning int
	ReviewsTotal, Reviews7d, LexemesCached                int
}

type AITotals struct {
	WordsGeneratedTotal, WordsGenerated7d, WordsCachedTotal, WordsFailed7d int
	TopicListsTotal, TopicLists7d                                          int
	WritingLessonsTotal                                                    int
	WritingChecksTotal, WritingChecks7d                                    int
	WritingReviewsTotal, WritingReviews7d                                  int
	InputTokens30d, OutputTokens30d                                        int
}

type WritingTotals struct {
	AttemptsTotal, AttemptsSubmitted int
	AvgBand30d                       *float64
}

type FeedbackTotals struct {
	New, Total int
}

type ReportTotals struct {
	Total, Last7d, NeedsReview int
}

// Totals: every number of the overview except the daily chart and the latest reports.
type Totals struct {
	Users          UserTotals
	Vocab          VocabTotals
	AI             AITotals
	Writing        WritingTotals
	Feedback       FeedbackTotals
	ContentReports ReportTotals
}

// DayStats is one day of the 30-day chart. Date is YYYY-MM-DD in the board's time zone.
type DayStats struct {
	Date        string
	Signups     int
	ActiveUsers int
	Reviews     int
	AIRequests  int
}

type ContentReport struct {
	ID        string
	Headword  string
	Reason    string
	Note      string
	UserEmail string
	CreatedAt time.Time
}

const (
	SortLastActive = "last_active"
	SortCreated    = "created"
	SortReviews    = "reviews"
	SortAI         = "ai"
)

type UserFilter struct {
	Query    string
	Sort     string
	Page     int
	PageSize int
	Since7d  time.Time
}

type UserActivity struct {
	ID              string
	Email           string
	DisplayName     string
	AvatarURL       string
	UILocale        string
	CreatedAt       time.Time
	LastLoginAt     *time.Time
	LastActiveAt    *time.Time
	Decks           int
	Cards           int
	ReviewsTotal    int
	Reviews7d       int
	AIWords         int
	TopicLists      int
	WritingAttempts int
	WritingSubmit   int
	WritingAICalls  int
	Feedback        int
}

// ---- AI cost (Epic 7) -----------------------------------------------------------

// AIUsageRow is the ledger (ai_usage) summed per provider, model and feature, for all time and since Since30d.
type AIUsageRow struct {
	Provider, Model, Feature string
	Calls, Failed            int
	InputTokens              int
	OutputTokens             int
	CacheReadTokens          int
	Calls30d, Failed30d      int
	InputTokens30d           int
	OutputTokens30d          int
	CacheReadTokens30d       int
}

// ModelPrice is a model's list price in USD per million tokens.
type ModelPrice struct {
	Input, Output, CacheRead float64
}

// Tokens is a token count by kind.
type Tokens struct {
	Input, Output, CacheRead int
}
