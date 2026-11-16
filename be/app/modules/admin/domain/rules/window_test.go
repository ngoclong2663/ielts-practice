package rules

import (
	"testing"
	"time"

	"ielts-practice/app/modules/admin/domain/entity"
)

func TestWindowAndFillDaily(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	// 2026-11-15 20:00 UTC is already 2026-11-16 03:00 in Vietnam.
	now := time.Date(2026, 11, 15, 20, 0, 0, 0, time.UTC)
	w := NewWindow(now, loc)
	if got := w.Today.Format(time.RFC3339); got != "2026-11-16T00:00:00+07:00" {
		t.Fatalf("today: %s", got)
	}
	if w.Since7d.Format("2006-01-02") != "2026-11-10" || w.Since30d.Format("2006-01-02") != "2026-10-18" {
		t.Fatalf("since: %s %s", w.Since7d, w.Since30d)
	}
	days := FillDaily(w, map[string]entity.DayStats{"2026-11-16": {Reviews: 3}, "2026-10-18": {Signups: 1}})
	if len(days) != 30 || days[0].Date != "2026-10-18" || days[29].Date != "2026-11-16" {
		t.Fatalf("days: %d %v … %v", len(days), days[0], days[len(days)-1])
	}
	if days[0].Signups != 1 || days[29].Reviews != 3 || days[15].Reviews != 0 {
		t.Fatalf("values: %v %v", days[0], days[29])
	}
}
