package entity

// Value objects (defined only by their values): CardContent, IPA, Example, WordForm, Collocation, Sense, CardImage, CardAudio (the content of a card).

import "time"

// CardContent is the JSON stored in cards.content (schema v1, see Epic 3 §4).
// Manual cards (Epic 2) fill only a few fields; generated cards (Epic 3) fill all.
type CardContent struct {
	Headword      string        `json:"headword"`
	Lemma         string        `json:"lemma,omitempty"`
	SenseHint     string        `json:"sense_hint,omitempty"`
	IPA           *IPA          `json:"ipa,omitempty"`
	PartOfSpeech  string        `json:"part_of_speech,omitempty"`
	CEFR          string        `json:"cefr,omitempty"`
	MeaningVI     string        `json:"meaning_vi"`
	DefinitionEN  string        `json:"definition_en,omitempty"`
	WordForms     []WordForm    `json:"word_forms,omitempty"`
	Collocations  []Collocation `json:"collocations,omitempty"`
	ExtraExamples []Example     `json:"extra_examples,omitempty"`
	Synonyms      []string      `json:"synonyms,omitempty"`
	OtherSenses   []Sense       `json:"other_senses,omitempty"`
	Image         *CardImage    `json:"image,omitempty"`
	// Audio is unused since Epic 3 (the browser reads text aloud); kept so older content still parses.
	Audio *CardAudio `json:"audio,omitempty"`
}

type IPA struct {
	UK string `json:"uk,omitempty"`
	US string `json:"us,omitempty"`
}

type Example struct {
	EN string `json:"en"`
	VI string `json:"vi,omitempty"`
}

type WordForm struct {
	Form      string   `json:"form"`
	POS       string   `json:"pos,omitempty"`
	IPA       string   `json:"ipa,omitempty"`
	MeaningVI string   `json:"meaning_vi,omitempty"`
	Example   *Example `json:"example,omitempty"`
}

type Collocation struct {
	Phrase    string `json:"phrase"`
	MeaningVI string `json:"meaning_vi,omitempty"`
}

type Sense struct {
	POS       string `json:"pos,omitempty"`
	MeaningVI string `json:"meaning_vi"`
}

type CardImage struct {
	Provider   string `json:"provider"` // unsplash | pixabay | manual
	ProviderID string `json:"provider_id,omitempty"`
	URL        string `json:"url"`
	ThumbURL   string `json:"thumb_url,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Alt        string `json:"alt,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
	AuthorURL  string `json:"author_url,omitempty"`
	SourceURL  string `json:"source_url,omitempty"`
	// DownloadLocation: Unsplash endpoint we must call when the photo is chosen for a card.
	DownloadLocation string `json:"download_location,omitempty"`
	// LinkExpiresAt is transient (never stored in content): when a provider's direct URL stops working.
	LinkExpiresAt time.Time `json:"-"`
}

type CardAudio struct {
	Headword map[string]string `json:"headword,omitempty"` // accent (uk|us) -> URL
	Examples map[string]string `json:"examples,omitempty"` // example key -> URL
}
