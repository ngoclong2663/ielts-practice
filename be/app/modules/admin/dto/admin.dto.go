package dto

import "time"

// ---- GET /api/v1/admin/overview (TAdminOverview) ----

type UsersOverviewDto struct {
	Total     int `json:"total"`
	New7d     int `json:"new_7d"`
	New30d    int `json:"new_30d"`
	Active1d  int `json:"active_1d"`
	Active7d  int `json:"active_7d"`
	Active30d int `json:"active_30d"`
}

type CardsByStateDto struct {
	New        int `json:"new"`
	Learning   int `json:"learning"`
	Review     int `json:"review"`
	Relearning int `json:"relearning"`
}

type VocabOverviewDto struct {
	Decks         int             `json:"decks"`
	Cards         int             `json:"cards"`
	CardsByState  CardsByStateDto `json:"cards_by_state"`
	ReviewsTotal  int             `json:"reviews_total"`
	Reviews7d     int             `json:"reviews_7d"`
	LexemesCached int             `json:"lexemes_cached"`
}

type AIOverviewDto struct {
	WordsGeneratedTotal int `json:"words_generated_total"`
	WordsGenerated7d    int `json:"words_generated_7d"`
	WordsCachedTotal    int `json:"words_cached_total"`
	WordsFailed7d       int `json:"words_failed_7d"`
	TopicListsTotal     int `json:"topic_lists_total"`
	TopicLists7d        int `json:"topic_lists_7d"`
	WritingLessonsTotal int `json:"writing_lessons_total"`
	WritingChecksTotal  int `json:"writing_checks_total"`
	WritingChecks7d     int `json:"writing_checks_7d"`
	WritingReviewsTotal int `json:"writing_reviews_total"`
	WritingReviews7d    int `json:"writing_reviews_7d"`
	InputTokens30d      int `json:"input_tokens_30d"`
	OutputTokens30d     int `json:"output_tokens_30d"`
}

type WritingOverviewDto struct {
	AttemptsTotal     int      `json:"attempts_total"`
	AttemptsSubmitted int      `json:"attempts_submitted"`
	AvgBand30d        *float64 `json:"avg_band_30d"`
}

type FeedbackOverviewDto struct {
	New   int `json:"new"`
	Total int `json:"total"`
}

type ContentReportsOverviewDto struct {
	Total       int `json:"total"`
	Last7d      int `json:"last_7d"`
	NeedsReview int `json:"needs_review"`
}

type DailyDto struct {
	Date        string `json:"date"`
	Signups     int    `json:"signups"`
	ActiveUsers int    `json:"active_users"`
	Reviews     int    `json:"reviews"`
	AIRequests  int    `json:"ai_requests"`
}

type RecentReportDto struct {
	ID        string    `json:"id"`
	Headword  string    `json:"headword"`
	Reason    string    `json:"reason"`
	Note      string    `json:"note"`
	UserEmail string    `json:"user_email"`
	CreatedAt time.Time `json:"created_at"`
}

type OverviewDto struct {
	GeneratedAt    time.Time                 `json:"generated_at"`
	Timezone       string                    `json:"timezone"`
	Users          UsersOverviewDto          `json:"users"`
	Vocab          VocabOverviewDto          `json:"vocab"`
	AI             AIOverviewDto             `json:"ai"`
	Writing        WritingOverviewDto        `json:"writing"`
	Feedback       FeedbackOverviewDto       `json:"feedback"`
	ContentReports ContentReportsOverviewDto `json:"content_reports"`
	Daily          []DailyDto                `json:"daily"`
	RecentReports  []RecentReportDto         `json:"recent_reports"`
}

// ---- GET /api/v1/admin/users ----

type ListUsersQueryDto struct {
	Q        string `form:"q" binding:"max=100"`
	Sort     string `form:"sort" binding:"omitempty,oneof=last_active created reviews ai"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type AdminUserDto struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	DisplayName      string     `json:"display_name"`
	AvatarURL        string     `json:"avatar_url"`
	UILocale         string     `json:"ui_locale"`
	CreatedAt        time.Time  `json:"created_at"`
	LastLoginAt      *time.Time `json:"last_login_at"`
	LastActiveAt     *time.Time `json:"last_active_at"`
	Decks            int        `json:"decks"`
	Cards            int        `json:"cards"`
	ReviewsTotal     int        `json:"reviews_total"`
	Reviews7d        int        `json:"reviews_7d"`
	AIWords          int        `json:"ai_words"`
	TopicLists       int        `json:"topic_lists"`
	WritingAttempts  int        `json:"writing_attempts"`
	WritingSubmitted int        `json:"writing_submitted"`
	WritingAICalls   int        `json:"writing_ai_calls"`
	Feedback         int        `json:"feedback"`
}

type AdminUserListDto struct {
	Items    []AdminUserDto `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ---- Admin sign-in ----

type LoginRequestDto struct {
	Username string `json:"username" binding:"required,max=64"`
	Password string `json:"password" binding:"required,max=200"`
}

type AdminDto struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type LoginResponseDto struct {
	AccessToken string   `json:"access_token"`
	ExpiresAt   string   `json:"expires_at"` // RFC3339
	Admin       AdminDto `json:"admin"`
}
