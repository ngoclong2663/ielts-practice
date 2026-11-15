package entity

import "time"

// ---- lesson (Epic 5 §5) ------------------------------------------------------

type Hint struct {
	EN string `json:"en"`
	VI string `json:"vi"`
}

type PlanQuestion struct {
	QuestionVI string   `json:"question_vi"`
	OptionsVI  []string `json:"options_vi"`
	Answer     int      `json:"answer"` // index into OptionsVI
	ExplainVI  string   `json:"explain_vi"`
}

type LessonSentence struct {
	VI        string `json:"vi"`         // level 1: the full Vietnamese sentence
	NoteVI    string `json:"note_vi"`    // level 2: a short Vietnamese idea note
	OutlineEN string `json:"outline_en"` // level 3: English key words
	Hints     []Hint `json:"hints"`
	ModelEN   string `json:"model_en"`
}

type LessonParagraph struct {
	Key       string           `json:"key"` // introduction | overview | body1 | body2 | conclusion
	GoalVI    string           `json:"goal_vi"`
	Sentences []LessonSentence `json:"sentences"`
}

type LessonContent struct {
	IdeasVI        []string          `json:"ideas_vi"`
	Questions      []PlanQuestion    `json:"questions"`
	Paragraphs     []LessonParagraph `json:"paragraphs"`
	UsefulLanguage []Hint            `json:"useful_language"`
}

type Lesson struct {
	ID           string
	TaskID       string
	Position     string
	TaskHash     string
	Content      LessonContent
	Model        string
	InputTokens  int
	OutputTokens int
	CreatedAt    time.Time
}

// ---- feedback ----------------------------------------------------------------

// SentenceFeedback is the verdict on one sentence of a checked paragraph.
type SentenceFeedback struct {
	Verdict     string   `json:"verdict"` // correct | understandable | wrong | empty
	ErrorTypes  []string `json:"error_types"`
	ExplainVI   string   `json:"explain_vi"`
	CorrectedEN string   `json:"corrected_en"`
}

type CriterionFeedback struct {
	Key       string  `json:"key"` // task_achievement | task_response | coherence_cohesion | lexical_resource | grammatical_range
	Band      float64 `json:"band"`
	CommentVI string  `json:"comment_vi"`
}

type Fix struct {
	Quote     string `json:"quote"`
	Better    string `json:"better"`
	ExplainVI string `json:"explain_vi"`
}

type EssayFeedback struct {
	Criteria  []CriterionFeedback `json:"criteria"`
	SummaryVI string              `json:"summary_vi"`
	Fixes     []Fix               `json:"fixes"`
	Model     string              `json:"model"`
}

// ---- attempts ----------------------------------------------------------------

type ParagraphCheck struct {
	Key          string
	Answers      []string
	Feedback     []SentenceFeedback
	CheckedCount int
	UpdatedAt    time.Time
}

type Submission struct {
	Text         string
	WordCount    int
	SecondsSpent int
	Feedback     EssayFeedback
	BandLow      float64
	BandHigh     float64
	CreatedAt    time.Time
}

type Attempt struct {
	ID          string
	UserID      string
	TaskID      string
	LessonID    string
	Level       int
	Position    string
	PlanAnswers []int // nil until answered
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SubmittedAt *time.Time
}

// AttemptSummary is one row of the history and of a task's attempt list.
type AttemptSummary struct {
	Attempt
	TaskTitle string
	TaskNo    int
	TaskType  string
	BandLow   *float64
	BandHigh  *float64
}

// Usage of one AI call.
type AIUsage struct {
	Model        string
	InputTokens  int
	OutputTokens int
}
