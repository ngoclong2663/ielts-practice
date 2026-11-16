package rules

import (
	"time"

	"ielts-practice/app/modules/admin/domain/entity"
)

// NewWindow computes the board's day boundaries for `now` in `loc`.
func NewWindow(now time.Time, loc *time.Location) entity.Window {
	local := now.In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	return entity.Window{
		Now:      now,
		Timezone: loc.String(),
		Today:    today,
		Since7d:  today.AddDate(0, 0, -6),
		Since30d: today.AddDate(0, 0, -(entity.DailyDays - 1)),
	}
}

// FillDaily returns exactly 30 days, oldest first, ending today; days without data are zeros.
func FillDaily(w entity.Window, byDate map[string]entity.DayStats) []entity.DayStats {
	out := make([]entity.DayStats, 0, entity.DailyDays)
	for i := entity.DailyDays - 1; i >= 0; i-- {
		date := w.Today.AddDate(0, 0, -i).Format("2006-01-02")
		d := byDate[date]
		d.Date = date
		out = append(out, d)
	}
	return out
}
