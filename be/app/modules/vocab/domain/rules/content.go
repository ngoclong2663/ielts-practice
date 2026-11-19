package rules

import (
	"slices"
	"strings"

	"ielts-practice/app/modules/vocab/domain/entity"
)

// Quality flags stored on the lexeme. The card is still usable; flags drive review and the eval set.
const (
	FlagFewCollocations    = "few_collocations"
	FlagNoWordForms        = "no_word_forms"
	FlagExampleMissingForm = "example_missing_form"
	// Security flags (Epic 8 §6): the card is shown to the learner who asked for it but never shared through
	// the lexeme cache, so text slipped into a word or sense hint cannot reach other learners.
	FlagHeadwordMismatch = "headword_mismatch" // the card is about another word than the input
	FlagUnsafeText       = "unsafe_text"       // a link, markup or instruction-like text in the card
)

const (
	maxCollocations  = 6
	minCollocations  = 3
	maxExtraExamples = 2
	maxSynonyms      = 3
	maxOtherSenses   = 3
	maxWordForms     = 6
)

// SanitizeResult trims strings, caps list sizes and records quality flags (Epic 3 §5.3 validator).
func SanitizeResult(r *entity.WordResult) {
	c := &r.Content
	c.Headword = strings.TrimSpace(c.Headword)
	if c.Headword == "" {
		c.Headword = strings.TrimSpace(r.Input)
	}
	c.Lemma = strings.TrimSpace(c.Lemma)
	c.MeaningVI = strings.TrimSpace(c.MeaningVI)
	c.DefinitionEN = strings.TrimSpace(c.DefinitionEN)
	c.PartOfSpeech = strings.TrimSpace(c.PartOfSpeech)
	if c.CEFR == "unknown" {
		c.CEFR = ""
	}
	if c.IPA != nil && c.IPA.UK == "" && c.IPA.US == "" {
		c.IPA = nil
	}
	r.Suggestion = strings.TrimSpace(r.Suggestion)
	if r.Topic = strings.TrimSpace(strings.ToLower(r.Topic)); !slices.Contains(TopicIDs(), r.Topic) {
		r.Topic = TopicOther
	}
	r.ImageQuery = strings.TrimSpace(r.ImageQuery)

	forms := c.WordForms[:0]
	for _, f := range c.WordForms {
		f.Form = strings.TrimSpace(f.Form)
		if f.Form == "" {
			continue
		}
		if f.Example != nil && strings.TrimSpace(f.Example.EN) == "" {
			f.Example = nil
		}
		forms = append(forms, f)
	}
	c.WordForms = capList(forms, maxWordForms)

	collocations := c.Collocations[:0]
	for _, col := range c.Collocations {
		if col.Phrase = strings.TrimSpace(col.Phrase); col.Phrase != "" {
			collocations = append(collocations, col)
		}
	}
	c.Collocations = capList(collocations, maxCollocations)

	examples := c.ExtraExamples[:0]
	for _, ex := range c.ExtraExamples {
		if ex.EN = strings.TrimSpace(ex.EN); ex.EN != "" {
			examples = append(examples, ex)
		}
	}
	c.ExtraExamples = capList(examples, maxExtraExamples)

	synonyms := c.Synonyms[:0]
	for _, s := range c.Synonyms {
		if s = strings.TrimSpace(s); s != "" && !strings.EqualFold(s, c.Headword) {
			synonyms = append(synonyms, s)
		}
	}
	c.Synonyms = capList(synonyms, maxSynonyms)
	c.OtherSenses = capList(c.OtherSenses, maxOtherSenses)

	if !r.IsValid {
		return
	}
	if !aboutInput(r.Input, c.Headword, c.Lemma) {
		r.QualityFlags = append(r.QualityFlags, FlagHeadwordMismatch)
	}
	if contentUnsafe(c) || looksUnsafe(r.ImageQuery) {
		r.QualityFlags = append(r.QualityFlags, FlagUnsafeText)
	}
	if len(c.WordForms) == 0 {
		r.QualityFlags = append(r.QualityFlags, FlagNoWordForms)
	}
	if len(c.Collocations) < minCollocations {
		r.QualityFlags = append(r.QualityFlags, FlagFewCollocations)
	}
	for _, f := range c.WordForms {
		if f.Example != nil && !containsForm(f.Example.EN, f.Form) {
			r.QualityFlags = append(r.QualityFlags, FlagExampleMissingForm)
			break
		}
	}
}

// Shareable: a card without security flags may be cached for every learner.
func Shareable(flags []string) bool {
	for _, f := range flags {
		if f == FlagHeadwordMismatch || f == FlagUnsafeText {
			return false
		}
	}
	return true
}

// aboutInput: the headword or lemma starts like the input (or the other way round), so inflections and
// spelling fixes ("studys" → "study") pass while a card about an unrelated word does not.
func aboutInput(input, headword, lemma string) bool {
	in := foldWord(input)
	if in == "" {
		return true
	}
	for _, w := range []string{headword, lemma} {
		w = foldWord(w)
		if w == "" {
			continue
		}
		n := min(4, len(in), len(w))
		if strings.HasPrefix(w, in) || strings.HasPrefix(in, w) || in[:n] == w[:n] {
			return true
		}
	}
	return false
}

func foldWord(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var unsafeMarkers = []string{
	"http://", "https://", "www.", "<", ">", "```", "{{", "javascript:",
	"ignore previous", "ignore all", "ignore the above", "system prompt", "you are now", "as an ai",
}

func looksUnsafe(s string) bool {
	s = strings.ToLower(s)
	for _, m := range unsafeMarkers {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}

// contentUnsafe checks every text a learner reads on the card.
func contentUnsafe(c *entity.CardContent) bool {
	texts := []string{c.Headword, c.Lemma, c.MeaningVI, c.DefinitionEN, c.PartOfSpeech}
	texts = append(texts, c.Synonyms...)
	for _, f := range c.WordForms {
		texts = append(texts, f.Form, f.MeaningVI)
		if f.Example != nil {
			texts = append(texts, f.Example.EN, f.Example.VI)
		}
	}
	for _, col := range c.Collocations {
		texts = append(texts, col.Phrase, col.MeaningVI)
	}
	for _, ex := range c.ExtraExamples {
		texts = append(texts, ex.EN, ex.VI)
	}
	for _, o := range c.OtherSenses {
		texts = append(texts, o.MeaningVI)
	}
	for _, t := range texts {
		if looksUnsafe(t) {
			return true
		}
	}
	return false
}

// containsForm accepts inflections by matching the first word's stem (decide → decided, decides).
func containsForm(sentence, form string) bool {
	s := strings.ToLower(sentence)
	f := strings.ToLower(strings.Fields(form + " ")[0])
	if strings.Contains(s, f) {
		return true
	}
	if r := []rune(f); len(r) > 4 {
		return strings.Contains(s, string(r[:len(r)-2]))
	}
	return false
}

func capList[T any](list []T, n int) []T {
	if len(list) > n {
		return list[:n]
	}
	return list
}
