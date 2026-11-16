package rules

import (
	"strings"
	"testing"
	"time"
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
