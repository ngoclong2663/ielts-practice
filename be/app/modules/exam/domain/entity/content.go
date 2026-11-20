package entity

import "encoding/json"

// Group types (Epic 10 §8.1). They are also the categories of the accuracy-per-type stats.
const (
	TypeTFNG          = "tfng"
	TypeYNNG          = "ynng"
	TypeMCQ           = "mcq"
	TypeChooseN       = "choose_n"
	TypeMatchHeadings = "match_headings"
	TypeMatchInfo     = "match_info"
	TypeMatchFeatures = "match_features"
	TypeMatchEndings  = "match_endings"
	TypeSummaryBank   = "summary_bank"
	TypeSentence      = "sentence"
	TypeNotes         = "notes"
	TypeTable         = "table"
	TypeFlowchart     = "flowchart"
	TypeForm          = "form"
	TypeSummary       = "summary"
	TypeShortAnswer   = "short_answer"
	TypeDiagram       = "diagram"
	TypeMap           = "map"
)

// Answer modes: how a group's questions are answered and marked.
const (
	AnswerChoice  = "choice"   // one value per question from a fixed set (T/F/NG, a letter)
	AnswerChooseN = "choose_n" // N letters for N questions, any order
	AnswerLetter  = "letter"   // one letter per question from the group's options
	AnswerGap     = "gap"      // typed words
)

// AnswerMode maps a group type to its answer mode ("" for an unknown type).
func AnswerMode(groupType string) string {
	switch groupType {
	case TypeTFNG, TypeYNNG, TypeMCQ:
		return AnswerChoice
	case TypeChooseN:
		return AnswerChooseN
	case TypeMatchHeadings, TypeMatchInfo, TypeMatchFeatures, TypeMatchEndings, TypeSummaryBank, TypeMap:
		return AnswerLetter
	case TypeSentence, TypeNotes, TypeTable, TypeFlowchart, TypeForm, TypeSummary, TypeShortAnswer, TypeDiagram:
		return AnswerGap
	}
	return ""
}

// Content is what the player shows: the passage (reading) or transcript (listening) and the question groups.
// It never holds the answers; those are in Key.
type Content struct {
	Subtitle   string      `json:"subtitle,omitempty"`
	Paragraphs []Paragraph `json:"paragraphs,omitempty"` // reading
	Lines      []Line      `json:"lines,omitempty"`      // listening transcript, shown after submit only
	Groups     []Group     `json:"groups"`
}

// Paragraph is one block of the passage. ID is its letter ("A") or "p1", "p2" when the passage is unlettered;
// Label is shown in the margin ("A", or "" for none).
type Paragraph struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
	Text  string `json:"text"`
}

// Line is one speaker turn of a listening transcript, with its start time in the audio.
type Line struct {
	ID      string  `json:"id"`
	Speaker string  `json:"speaker,omitempty"`
	Text    string  `json:"text"`
	At      float64 `json:"at"`
	// Say is what the TTS reads when it must differ from the transcript ("M, A, R, S, H" for "M-A-R-S-H",
	// "oh seven nine four six" for a phone number). Used by the audio build only.
	Say string `json:"say,omitempty"`
}

// Option is one lettered choice ("A", "the first paragraph") or a heading ("iv", "…").
type Option struct {
	Key  string `json:"key"`
	Text string `json:"text,omitempty"`
}

// Item is one numbered question inside a group (a statement, an MCQ stem, a matching item).
type Item struct {
	N       int      `json:"n"`
	Text    string   `json:"text"`
	Options []Option `json:"options,omitempty"` // mcq: the item's own options
}

// Group is a block of questions sharing one instruction. Gap questions sit in Lines, Rows or Steps as {n}
// placeholders, e.g. "After midnight, lights run at {5} per cent of full power.".
type Group struct {
	Type         string     `json:"type"`
	From         int        `json:"from"`
	To           int        `json:"to"`
	Instructions string     `json:"instructions"`
	Title        string     `json:"title,omitempty"`
	WordLimit    int        `json:"word_limit,omitempty"`   // gap groups: words allowed per answer
	AllowNumber  bool       `json:"allow_number,omitempty"` // "AND/OR A NUMBER"
	Pick         int        `json:"pick,omitempty"`         // choose_n
	Stem         string     `json:"stem,omitempty"`         // choose_n question
	Options      []Option   `json:"options,omitempty"`      // choose_n, letter groups, summary bank
	ReuseOptions bool       `json:"reuse_options,omitempty"`
	Items        []Item     `json:"items,omitempty"`
	Lines        []string   `json:"lines,omitempty"` // notes, sentences, summary, flow chart steps
	Header       []string   `json:"header,omitempty"`
	Rows         [][]string `json:"rows,omitempty"` // table and form cells
	// Figure is a map, plan or diagram: in a seed file its path in the set folder (figures/x.svg), stored as
	// its content name (<sha256>.svg) and sent to the browser as a URL.
	Figure string `json:"figure,omitempty"`
}

// Audio is a listening part's recording: the S3 key, its length and the reading pauses of exam mode.
type Audio struct {
	Key      string  `json:"key"`
	Duration float64 `json:"duration"`
	Cues     []Cue   `json:"cues,omitempty"`
}

// Cue is a reading pause of exam mode: the player stops the audio at At for Seconds and shows Label. In a
// seed file After names the transcript line it follows; the audio build turns that into At.
type Cue struct {
	At      float64 `json:"at"`
	Seconds int     `json:"seconds"`
	Label   string  `json:"label"`
	After   string  `json:"after,omitempty"`
}

// KeyItem is the answer and explanation of one question.
type KeyItem struct {
	Answer   []string    `json:"answer"`
	Evidence *Evidence   `json:"evidence,omitempty"`
	Pairs    [][2]string `json:"pairs,omitempty"`
	Why      string      `json:"why"`
	Trap     string      `json:"trap,omitempty"`
}

// Evidence points to the words that give the answer: a paragraph or transcript line and an exact quote.
type Evidence struct {
	Block string `json:"block"`
	Quote string `json:"quote"`
}

// Key maps a question number ("1") to its answer.
type Key map[string]KeyItem

// Answers maps a question number to what the learner gave: a string, or for a choose_n group (keyed by the
// group's first number) a list of letters.
type Answers map[string]json.RawMessage
