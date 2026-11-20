// Package rules holds the exam module's pure rules: answer matching and word limits, marking, the band
// tables and the seed checks (Epic 10 §7.3, §8.2).
package rules

import (
	"regexp"
	"strings"
)

var (
	spaces   = regexp.MustCompile(`\s+`)
	numberRe = regexp.MustCompile(`^[£$€]?\d[\d,.:/]*(%|[a-z]{1,2})?$`)
	quotes   = strings.NewReplacer("’", "'", "‘", "'", "“", `"`, "”", `"`, "–", "-", "—", "-")
)

// Normalize prepares an answer for comparison: case, curly quotes and dashes, inner and surrounding spaces,
// and a trailing full stop or comma are ignored.
func Normalize(s string) string {
	s = strings.ToLower(quotes.Replace(s))
	s = spaces.ReplaceAllString(strings.TrimSpace(s), " ")
	return strings.TrimRight(s, ".,;")
}

// ExpandVariants turns one key entry written in book notation into every accepted form: "/" separates
// alternatives ("10/ten") and brackets mark optional letters or words ("sport(s)", "(the) harbour").
func ExpandVariants(entry string) []string {
	var out []string
	seen := map[string]bool{}
	for _, alt := range strings.Split(entry, "/") {
		for _, v := range expandOptional(alt) {
			v = Normalize(v)
			if v != "" && !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
	}
	return out
}

func expandOptional(s string) []string {
	open := strings.Index(s, "(")
	if open < 0 {
		return []string{s}
	}
	closeAt := strings.Index(s[open:], ")")
	if closeAt < 0 {
		return []string{s}
	}
	closeAt += open
	with := s[:open] + s[open+1:closeAt] + s[closeAt+1:]
	without := s[:open] + s[closeAt+1:]
	return append(expandOptional(with), expandOptional(without)...)
}

// CountWords splits an answer into words and numbers. A hyphenated word is one word; "3,000", "10.25",
// "£35" and "20%" are numbers.
func CountWords(answer string) (words, numbers int) {
	for _, tok := range strings.Fields(Normalize(answer)) {
		if numberRe.MatchString(tok) {
			numbers++
		} else {
			words++
		}
	}
	return words, numbers
}

// WithinLimit reports whether an answer obeys a gap group's limit: "ONE WORD ONLY" is limit 1;
// "NO MORE THAN TWO WORDS AND/OR A NUMBER" is limit 2 with allowNumber (up to two words plus one number);
// "A NUMBER" alone is limit 0 with allowNumber. Without allowNumber a number counts as a word. Limit 0
// without allowNumber means no limit.
func WithinLimit(answer string, limit int, allowNumber bool) bool {
	if limit == 0 && !allowNumber {
		return true
	}
	w, n := CountWords(answer)
	if allowNumber {
		return w <= limit && n <= 1
	}
	return w+n <= limit
}

// MatchesAny reports whether a typed answer equals one of the key's accepted forms.
func MatchesAny(given string, key []string) bool {
	g := Normalize(given)
	if g == "" {
		return false
	}
	for _, entry := range key {
		for _, v := range ExpandVariants(entry) {
			if g == v {
				return true
			}
		}
	}
	return false
}
