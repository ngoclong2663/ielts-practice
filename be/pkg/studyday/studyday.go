// Package studyday computes Anki-style study-day boundaries: a day starts at the user's
// configured hour (default 04:00) in their own time zone, not at midnight UTC.
package studyday

import "time"

// Bounds returns the start of the current study day and the start of the next one.
func Bounds(now time.Time, timezone string, dayStartHour int) (todayStart, tomorrowStart time.Time) {
	loc, err := time.LoadLocation(timezone)
	if err != nil || timezone == "" {
		loc = time.UTC
	}
	local := now.In(loc)
	start := time.Date(local.Year(), local.Month(), local.Day(), dayStartHour, 0, 0, 0, loc)
	if local.Before(start) {
		start = start.AddDate(0, 0, -1)
	}
	return start, start.AddDate(0, 0, 1)
}

// DayIndex returns how many study days `t` is after the study day containing `todayStart`
// (0 = today, 1 = tomorrow, -1 = yesterday).
func DayIndex(t, todayStart time.Time, timezone string) int {
	loc, err := time.LoadLocation(timezone)
	if err != nil || timezone == "" {
		loc = time.UTC
	}
	ts := todayStart.In(loc)
	tl := t.In(loc)
	dayStart := time.Date(tl.Year(), tl.Month(), tl.Day(), ts.Hour(), 0, 0, 0, loc)
	if tl.Before(dayStart) {
		dayStart = dayStart.AddDate(0, 0, -1)
	}
	// Count calendar days between the two day starts (DST safe).
	a := time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC)
	b := time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, time.UTC)
	return int(b.Sub(a).Hours() / 24)
}
