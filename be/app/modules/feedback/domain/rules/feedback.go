package rules

import (
	"strings"
	"time"
	"unicode/utf8"

	"ielts-practice/app/modules/feedback/domain/entity"
	notifyEntity "ielts-practice/app/modules/notify/domain/entity"
)

const (
	MinMessageRunes = 10
	MaxMessageRunes = 2000
	MaxPageRunes    = 300
	MaxAgentRunes   = 400
)

var categoryLabels = map[string]string{
	entity.CategoryBug:     "Something is broken",
	entity.CategoryIdea:    "An idea",
	entity.CategoryContent: "Wrong word content",
	entity.CategoryOther:   "Other",
}

var categoryEmoji = map[string]string{
	entity.CategoryBug:     "🐞",
	entity.CategoryIdea:    "💡",
	entity.CategoryContent: "📝",
	entity.CategoryOther:   "💬",
}

// CategoryLabel is the English label shown to the owner (Telegram).
func CategoryLabel(category string) string {
	if l, ok := categoryLabels[category]; ok {
		return l
	}
	return category
}

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

// AdminLink is the board link sent in the notification (the owner reads the board in English).
func AdminLink(frontendURL, feedbackID string) string {
	return strings.TrimRight(frontendURL, "/") + "/en/admin/feedback?id=" + feedbackID
}

// BuildNotification is the owner's message for a new feedback (Epic 6 §4).
func BuildNotification(fb entity.Feedback, displayName, email, frontendURL string) notifyEntity.Notification {
	from := email
	if displayName != "" && email != "" {
		from = displayName + " (" + email + ")"
	} else if displayName != "" {
		from = displayName
	}
	if from == "" {
		from = fb.UserID
	}
	var where []string
	if fb.Page != "" {
		where = append(where, fb.Page)
	}
	if fb.UILocale != "" {
		where = append(where, fb.UILocale)
	}
	fields := []notifyEntity.Field{{Label: "From", Value: from}}
	if len(where) > 0 {
		fields = append(fields, notifyEntity.Field{Label: "Page", Value: strings.Join(where, "  ·  ")})
	}
	return notifyEntity.Notification{
		Emoji:  categoryEmoji[fb.Category],
		Title:  "New feedback · " + CategoryLabel(fb.Category),
		Fields: fields,
		Body:   fb.Message,
		Link:   AdminLink(frontendURL, fb.ID),
	}
}
