package entity

// Entities (have an identity): Lesson, Attempt; ParagraphCheck and Submission belong to an Attempt.
// Value objects (defined only by their values): Hint, PlanQuestion, LessonSentence, LessonParagraph, LessonContent,
// VocabItem, VocabGroup, LessonVocab, SentenceFeedback, CriterionFeedback, Fix, EssayFeedback, AttemptSummary (read model), AIUsage.

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
	IdeasVI    []string          `json:"ideas_vi"`
	Questions  []PlanQuestion    `json:"questions"`
	Paragraphs []LessonParagraph `json:"paragraphs"`
	// UsefulLanguage was the Epic 5 phrase list. Lessons written since Epic 7 leave it empty: the lesson's
	// LessonVocab replaces it. Kept so stored lessons still decode.
	UsefulLanguage []Hint `json:"useful_language,omitempty"`
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

// ---- useful vocabulary (Epic 7) ------------------------------------------------

// Vocabulary groups, in display order.
const (
	VocabTopic      = "topic"      // words and collocations about the task's topic
	VocabParaphrase = "paraphrase" // words of the question and how to say them differently
	VocabTask       = "task"       // language this task type needs: trends and comparisons, or opinions and causes
)

// VocabGroups lists the groups in display order.
var VocabGroups = []string{VocabTopic, VocabParaphrase, VocabTask}

type VocabItem struct {
	Term      string   `json:"term"` // a word or collocation, e.g. "workforce", "rise steadily"
	POS       string   `json:"pos"`  // noun | verb | adjective | adverb | phrase
	VI        string   `json:"vi"`
	Synonyms  []string `json:"synonyms"`   // close alternatives a learner can paraphrase with
	ExampleEN string   `json:"example_en"` // one sentence on this task's topic
	NoteVI    string   `json:"note_vi"`    // how the synonyms differ, or a common mistake; may be empty
}

type VocabGroup struct {
	Key   string      `json:"key"`
	Items []VocabItem `json:"items"`
}

type LessonVocab struct {
	Groups []VocabGroup `json:"groups"`
}

// StoredVocab is a lesson's vocabulary with the AI call that wrote it.
type StoredVocab struct {
	LessonID string
	Content  LessonVocab
	Usage    AIUsage
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
