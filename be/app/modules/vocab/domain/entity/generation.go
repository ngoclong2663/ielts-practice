package entity

// Data exchanged with the content generator and word suggester (vocab/ports/out).

// WordRequest is one entry sent to the content generator.
type WordRequest struct {
	Input     string
	SenseHint string
}

// WordResult is the generator's answer for one entry.
type WordResult struct {
	Input        string
	IsValid      bool
	Suggestion   string
	Content      CardContent
	ImageQuery   string
	Visualizable bool
	QualityFlags []string
	// Topic: the IELTS topic the card belongs to (rules.TopicIDs), for library candidates (Epic 9).
	Topic string
}

// SuggestRequest asks for topic words (Epic 4). Exclude lists lemmas the model should avoid.
type SuggestRequest struct {
	Topic   string
	Count   int
	Exclude []string
}
