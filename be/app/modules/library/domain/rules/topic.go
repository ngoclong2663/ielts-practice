package rules

import (
	"strings"

	"ielts-practice/app/modules/library/domain/entity"
)

// SuggestTopic picks the topic a candidate card most likely belongs to (Epic 9), without an AI call: the
// English words of the card (headword, definition, forms, collocations, synonyms, examples) are matched
// against each topic's vocabulary (its library words and its description). The best topic wins when it shares
// at least minTopicMatches words with the card; otherwise "other". topics keeps the display order, which
// breaks ties.
func SuggestTopic(card *entity.CardContent, topics []string, vocab map[string]string) string {
	words := cardWords(card)
	best, bestScore := entity.OtherTopic, minTopicMatches-1
	for _, t := range topics {
		set := tokenSet(vocab[t])
		score := 0
		for w := range words {
			if set[w] {
				score++
			}
		}
		if score > bestScore {
			best, bestScore = t, score
		}
	}
	return best
}

const minTopicMatches = 2

// cardWords collects the card's English content words.
func cardWords(c *entity.CardContent) map[string]bool {
	parts := []string{c.Headword, c.Lemma, c.DefinitionEN}
	parts = append(parts, c.Synonyms...)
	for _, f := range c.WordForms {
		parts = append(parts, f.Form)
		if f.Example != nil {
			parts = append(parts, f.Example.EN)
		}
	}
	for _, col := range c.Collocations {
		parts = append(parts, col.Phrase)
	}
	for _, ex := range c.ExtraExamples {
		parts = append(parts, ex.EN)
	}
	return tokenSet(strings.Join(parts, " "))
}

// tokenSet folds text into content words of 4+ letters, without a plural "s" and without common words.
func tokenSet(text string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.Fields(Fold(strings.ReplaceAll(text, "|", " "))) {
		if len(w) < 4 || commonWords[w] {
			continue
		}
		if len(w) > 4 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") {
			w = strings.TrimSuffix(w, "s")
		}
		out[w] = true
	}
	return out
}

var commonWords = func() map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(`that this with have from they their them there these those which what when where
		will would could should been were more most many much some such also than then into onto over very only other
		people person thing things make made makes being does done just like well even many every each both after before
		about around while because often still again always never ever your yours ours someone something time times
		year years good better best great small large long high take takes taken used using uses help helps`) {
		m[w] = true
	}
	return m
}()
