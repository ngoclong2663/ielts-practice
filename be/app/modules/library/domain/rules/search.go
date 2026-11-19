// Package rules holds the library's pure rules (Epic 8 §3).
package rules

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"ielts-practice/app/modules/library/domain/entity"
)

const (
	MaxQueryRunes  = 60
	MaxWordsResult = 120 // a search across topics; a topic has 30-50 words
	MaxAddWords    = 100
)

// Fold lowercases and drops accents and punctuation, so "Ô nhiễm" matches "o nhiem" and "ô nhiễm".
func Fold(s string) string {
	var b strings.Builder
	space := false
	for _, r := range norm.NFD.String(strings.ToLower(s)) {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue
		case r == 'đ':
			r = 'd'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			space = false
		} else if !space && b.Len() > 0 {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

// SearchText is what a library word is searched by: headword, its forms, synonyms and Vietnamese meanings.
func SearchText(c *entity.CardContent) string {
	parts := []string{c.Headword, c.Lemma, c.MeaningVI}
	parts = append(parts, c.Synonyms...)
	for _, f := range c.WordForms {
		parts = append(parts, f.Form, f.MeaningVI)
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if f := Fold(p); f != "" && !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	// "|" separates entries so a query never matches across two of them.
	return "|" + strings.Join(out, "|") + "|"
}

// LikePattern escapes a folded query for ILIKE.
func LikePattern(q string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(q) + "%"
}
