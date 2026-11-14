package studyday

import (
	"testing"
	"time"
)

func TestBounds(t *testing.T) {
	hcm := "Asia/Ho_Chi_Minh" // UTC+7
	cases := []struct {
		name      string
		nowUTC    string
		wantStart string // local
	}{
		// 07:00Z = 14:00 local.
		{"afternoon", "2026-11-14T07:00:00Z", "2026-11-14T04:00:00+07:00"},
		// 19:30Z = 02:30 local on the 15th: still the study day that started on the 14th.
		{"after midnight, before 4am", "2026-11-14T19:30:00Z", "2026-11-14T04:00:00+07:00"},
		// 21:00Z = 04:00 local on the 15th: a new study day.
		{"exactly 4am", "2026-11-14T21:00:00Z", "2026-11-15T04:00:00+07:00"},
	}
	for _, c := range cases {
		now, _ := time.Parse(time.RFC3339, c.nowUTC)
		want, _ := time.Parse(time.RFC3339, c.wantStart)
		start, end := Bounds(now, hcm, 4)
		if !start.Equal(want) {
			t.Errorf("%s: start = %v, want %v", c.name, start, want)
		}
		if !end.Equal(want.Add(24 * time.Hour)) {
			t.Errorf("%s: end = %v", c.name, end)
		}
	}
}

func TestBoundsInvalidZoneFallsBackToUTC(t *testing.T) {
	now := time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)
	start, _ := Bounds(now, "Nope/Zone", 4)
	if !start.Equal(time.Date(2026, 11, 13, 4, 0, 0, 0, time.UTC)) {
		t.Fatalf("got %v", start)
	}
}

func TestDayIndex(t *testing.T) {
	tz := "Asia/Ho_Chi_Minh"
	now := time.Date(2026, 11, 14, 7, 0, 0, 0, time.UTC) // 14:00 local
	today, _ := Bounds(now, tz, 4)
	if got := DayIndex(now.Add(2*time.Hour), today, tz); got != 0 {
		t.Errorf("same day: %d", got)
	}
	if got := DayIndex(today.Add(24*time.Hour), today, tz); got != 1 {
		t.Errorf("tomorrow: %d", got)
	}
	if got := DayIndex(today.Add(23*time.Hour+59*time.Minute), today, tz); got != 0 {
		t.Errorf("03:59 next morning is still today: %d", got)
	}
	if got := DayIndex(today.Add(-time.Minute), today, tz); got != -1 {
		t.Errorf("yesterday: %d", got)
	}
}
