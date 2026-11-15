package rules

import (
	"regexp"
	"strings"
	"unicode"

	"ielts-practice/app/modules/vocab/domain/entity"
)

const (
	MaxEntriesPerJob = 30
	maxEntryChars    = 60
	maxEntryWords    = 6
	maxSenseHint     = 40
)

var (
	splitter      = regexp.MustCompile(`[\n\r,;]+`)
	listMarker    = regexp.MustCompile(`^\s*(?:[-*•·]+|\d+[.)])\s*`)
	spaces        = regexp.MustCompile(`\s+`)
	pipeHint      = regexp.MustCompile(`^(.+?)\s*\|\s*(.+)$`)
	parenHint     = regexp.MustCompile(`^(.+?)\s*\(([^()]+)\)$`)
	allowedRunes  = regexp.MustCompile(`^[\p{L}][\p{L}\p{M}' .\-’]*$`)
	curlyApostrop = strings.NewReplacer("’", "'", "‘", "'")
)

// Normalize turns pasted text into unique entries (Epic 3 §3 step 2). It never calls the LLM.
func Normalize(text string) ([]entity.Entry, []entity.RejectedEntry) {
	entries := []entity.Entry{}
	rejected := []entity.RejectedEntry{}
	seen := map[string]bool{}
	for _, raw := range splitter.Split(text, -1) {
		raw = strings.TrimSpace(listMarker.ReplaceAllString(raw, ""))
		raw = spaces.ReplaceAllString(curlyApostrop.Replace(raw), " ")
		if raw == "" {
			continue
		}
		word, hint := raw, ""
		if m := pipeHint.FindStringSubmatch(raw); m != nil {
			word, hint = m[1], m[2]
		} else if m := parenHint.FindStringSubmatch(raw); m != nil {
			word, hint = m[1], m[2]
		}
		word = strings.Trim(strings.TrimSpace(word), ".")
		hint = strings.ToLower(strings.TrimSpace(hint))
		if len([]rune(hint)) > maxSenseHint {
			hint = string([]rune(hint)[:maxSenseHint])
		}
		switch {
		case len([]rune(word)) > maxEntryChars:
			rejected = append(rejected, entity.RejectedEntry{Text: raw, Reason: entity.RejectTooLong})
			continue
		case len(strings.Fields(word)) > maxEntryWords:
			rejected = append(rejected, entity.RejectedEntry{Text: raw, Reason: entity.RejectTooManyWords})
			continue
		case !allowedRunes.MatchString(word):
			rejected = append(rejected, entity.RejectedEntry{Text: raw, Reason: entity.RejectInvalidChars})
			continue
		}
		word = lowerUnlessAcronym(word)
		key := LexemeKey(word, hint)
		if seen[key] {
			rejected = append(rejected, entity.RejectedEntry{Text: raw, Reason: entity.RejectDuplicate})
			continue
		}
		seen[key] = true
		entries = append(entries, entity.Entry{Text: word, SenseHint: hint})
	}
	return entries, rejected
}

// lowerUnlessAcronym lowercases sentence-case words ("Decide" → "decide") but keeps
// acronyms and mixed-case names with two or more capitals ("UNESCO", "McDonald").
func lowerUnlessAcronym(word string) string {
	upper := 0
	for _, r := range word {
		if unicode.IsUpper(r) {
			upper++
		}
	}
	if upper >= 2 {
		return word
	}
	return strings.ToLower(word)
}

// LexemeKey identifies shared content: headword, sense hint and meaning language.
func LexemeKey(headword, senseHint string) string {
	return strings.ToLower(strings.TrimSpace(headword)) + "|" + strings.ToLower(strings.TrimSpace(senseHint)) + "|" + entity.MeaningLangVI
}
