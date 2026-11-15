package rules

import (
	"strings"

	"ielts-practice/app/modules/vocab/domain/entity"
)

// Quality flags stored on the lexeme. The card is still usable; flags drive review and the eval set.
const (
	FlagFewCollocations    = "few_collocations"
	FlagNoWordForms        = "no_word_forms"
	FlagExampleMissingForm = "example_missing_form"
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
