package rules

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"ielts-practice/app/modules/writing/domain/entity"
)

// Checks shared by the AI coach and the seed loader (Epic 8 §4): a lesson or vocabulary set the practice flow
// can use, whoever wrote it.

// LessonWords counts the words of the model answer.
func LessonWords(l entity.LessonContent) int {
	n := 0
	for _, p := range l.Paragraphs {
		for _, s := range p.Sentences {
			n += CountWords(s.ModelEN)
		}
	}
	return n
}

// CleanLesson trims the model's lesson and rejects one the practice flow can't use.
func CleanLesson(l *entity.LessonContent, keys []string) error {
	if len(l.Paragraphs) != len(keys) {
		return fmt.Errorf("lesson: %d paragraphs, want %d", len(l.Paragraphs), len(keys))
	}
	for i := range l.Paragraphs {
		p := &l.Paragraphs[i]
		if p.Key != keys[i] {
			return fmt.Errorf("lesson: paragraph %d is %q, want %q", i, p.Key, keys[i])
		}
		p.GoalVI = strings.TrimSpace(p.GoalVI)
		if len(p.Sentences) == 0 || len(p.Sentences) > 6 {
			return fmt.Errorf("lesson: paragraph %s has %d sentences", p.Key, len(p.Sentences))
		}
		for j := range p.Sentences {
			s := &p.Sentences[j]
			s.VI, s.NoteVI = strings.TrimSpace(s.VI), strings.TrimSpace(s.NoteVI)
			s.OutlineEN, s.ModelEN = strings.TrimSpace(s.OutlineEN), strings.TrimSpace(s.ModelEN)
			if s.VI == "" || s.NoteVI == "" || s.OutlineEN == "" || s.ModelEN == "" {
				return fmt.Errorf("lesson: empty field in %s sentence %d", p.Key, j+1)
			}
			s.Hints = cleanHints(s.Hints, 3)
		}
	}
	questions := l.Questions[:0]
	for _, q := range l.Questions {
		q.QuestionVI = strings.TrimSpace(q.QuestionVI)
		if q.QuestionVI == "" || len(q.OptionsVI) < 2 || q.Answer < 0 || q.Answer >= len(q.OptionsVI) {
			continue // a broken question is dropped, not fatal
		}
		questions = append(questions, q)
	}
	l.Questions = questions
	ideas := l.IdeasVI[:0]
	for _, s := range l.IdeasVI {
		if s = strings.TrimSpace(s); s != "" {
			ideas = append(ideas, s)
		}
	}
	l.IdeasVI = ideas
	l.UsefulLanguage = nil // replaced by the lesson's vocabulary (WriteVocab)
	if len(l.IdeasVI) == 0 {
		return errors.New("lesson: no ideas")
	}
	return nil
}

func cleanHints(hints []entity.Hint, limit int) []entity.Hint {
	out := make([]entity.Hint, 0, len(hints))
	for _, h := range hints {
		h.EN, h.VI = strings.TrimSpace(h.EN), strings.TrimSpace(h.VI)
		if h.EN != "" && len(out) < limit {
			out = append(out, h)
		}
	}
	return out
}

// VocabPOS are the parts of speech a vocabulary item may have.
var VocabPOS = []string{"noun", "verb", "adjective", "adverb", "phrase"}

const (
	maxVocabPerGroup = 10
	maxSynonyms      = 4
)

// CleanVocab trims the model's vocabulary, puts the groups in display order, drops broken or repeated items
// and rejects a list too thin to be worth showing.
func CleanVocab(v *entity.LessonVocab) error {
	byKey := map[string][]entity.VocabItem{}
	seen := map[string]bool{}
	for _, g := range v.Groups {
		for _, it := range g.Items {
			it.Term = strings.TrimSpace(it.Term)
			it.VI, it.NoteVI = strings.TrimSpace(it.VI), strings.TrimSpace(it.NoteVI)
			it.ExampleEN = strings.TrimSpace(it.ExampleEN)
			key := strings.ToLower(it.Term)
			if it.Term == "" || it.VI == "" || seen[key] || len(byKey[g.Key]) >= maxVocabPerGroup {
				continue
			}
			if !slices.Contains(VocabPOS, it.POS) {
				it.POS = "phrase"
			}
			synonyms := make([]string, 0, maxSynonyms)
			for _, syn := range it.Synonyms {
				syn = strings.TrimSpace(syn)
				if syn != "" && !strings.EqualFold(syn, it.Term) && !slices.Contains(synonyms, syn) && len(synonyms) < maxSynonyms {
					synonyms = append(synonyms, syn)
				}
			}
			it.Synonyms = synonyms
			seen[key] = true
			byKey[g.Key] = append(byKey[g.Key], it)
		}
	}
	v.Groups = v.Groups[:0]
	total := 0
	for _, key := range entity.VocabGroups {
		if items := byKey[key]; len(items) > 0 {
			v.Groups = append(v.Groups, entity.VocabGroup{Key: key, Items: items})
			total += len(items)
		}
	}
	if total < 6 {
		return fmt.Errorf("vocab: only %d usable items", total)
	}
	return nil
}
