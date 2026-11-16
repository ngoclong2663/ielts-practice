package rules

import (
	"strings"
	"testing"
	"time"

	"ielts-practice/app/modules/feedback/domain/entity"
)

func TestNormalizeMessage(t *testing.T) {
	if msg, code := NormalizeMessage("   short   "); code != "VALIDATION_MIN" || msg != "short" {
		t.Fatalf("got %q %q", msg, code)
	}
	if _, code := NormalizeMessage("  " + strings.Repeat("ă", 2000) + "  "); code != "" {
		t.Fatalf("2000 characters after trimming are allowed: %q", code)
	}
	if _, code := NormalizeMessage(strings.Repeat("a", 2001)); code != "VALIDATION_MAX" {
		t.Fatalf("got %q", code)
	}
}

func TestResolvedAt(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)
	if got := ResolvedAt(nil, "new", "resolved", now); got == nil || !got.Equal(now) {
		t.Fatalf("resolving sets the time: %v", got)
	}
	if got := ResolvedAt(&earlier, "resolved", "resolved", now); got == nil || !got.Equal(earlier) {
		t.Fatalf("resolving again keeps the time: %v", got)
	}
	if got := ResolvedAt(&earlier, "resolved", "seen", now); got != nil {
		t.Fatalf("reopening clears the time: %v", got)
	}
}

func TestBuildNotification(t *testing.T) {
	n := BuildNotification(entity.Feedback{ID: "f1", Category: "bug", Message: "It broke", Page: "/vi/decks", UILocale: "vi"},
		"Learner One", "learner@example.com", "https://app.example.com/")
	if n.Title != "New feedback · Something is broken" || n.Emoji != "🐞" {
		t.Fatalf("title: %q %q", n.Emoji, n.Title)
	}
	if n.Fields[0].Value != "Learner One (learner@example.com)" || n.Fields[1].Value != "/vi/decks  ·  vi" {
		t.Fatalf("fields: %v", n.Fields)
	}
	if n.Link != "https://app.example.com/en/admin/feedback?id=f1" || n.Body != "It broke" {
		t.Fatalf("link/body: %q %q", n.Link, n.Body)
	}
}
