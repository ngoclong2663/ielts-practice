package entity

// Entities: GuideVideo. Value objects: Localized.

const (
	TopicVocab   = "vocab"
	TopicWriting = "writing"

	StatusPublished = "published"

	VideoKeyPrefix  = "guides/videos/"
	PosterKeyPrefix = "guides/posters/"
)

// Topics a guide can belong to (where the app shows it).
var Topics = []string{TopicVocab, TopicWriting}

// Localized is a text in the app's UI languages.
type Localized struct {
	VI string `json:"vi"`
	EN string `json:"en"`
}

type GuideVideo struct {
	ID              string
	Slug            string
	Topic           string
	Sort            int
	Lang            string // language of the narration
	Title           Localized
	Description     Localized
	VideoKey        string
	PosterKey       string
	DurationSeconds float64
	Status          string
	ContentHash     string
}
