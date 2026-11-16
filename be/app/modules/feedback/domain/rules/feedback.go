package rules

import (
	"strings"
	"time"
	"unicode/utf8"

	"ielts-practice/app/modules/feedback/domain/entity"
)

const (
	MinMessageRunes = 10
	MaxMessageRunes = 2000
	MaxPageRunes    = 300
	MaxAgentRunes   = 400
)

// NormalizeMessage trims the message and checks its length (10–2000 characters).
// It returns the validation code for the `message` field, or "" when the message is fine.
func NormalizeMessage(raw string) (string, string) {
	msg := strings.TrimSpace(raw)
	switch n := utf8.RuneCountInString(msg); {
	case n < MinMessageRunes:
		return msg, "VALIDATION_MIN"
	case n > MaxMessageRunes:
		return msg, "VALIDATION_MAX"
	}
	return msg, ""
}

// Truncate cuts s to at most n characters (runes).
func Truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

// ResolvedAt applies a status change: resolving sets the time (kept when it was already resolved),
// any other status clears it.
func ResolvedAt(current *time.Time, oldStatus, newStatus string, now time.Time) *time.Time {
	if newStatus != entity.StatusResolved {
		return nil
	}
	if oldStatus == entity.StatusResolved && current != nil {
		return current
	}
	return &now
}
