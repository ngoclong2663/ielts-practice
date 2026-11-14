package usecases

import (
	"context"
	"time"

	"ielts-practice/app/modules/study/dto"
)

const forecastDays = 7

// GetStats returns the dashboard numbers: today's work, streak, 7-day forecast, retention.
func (uc *StudyUseCases) GetStats(ctx context.Context, userID string) (dto.StatsOverviewDto, error) {
	sc, err := uc.loadContext(ctx, userID)
	if err != nil {
		return dto.StatsOverviewDto{}, err
	}
	tz, hour := sc.user.Timezone, sc.user.DayStartHour
	todayDate := sc.todayStart.Format(time.DateOnly) // todayStart is local, so this is the study-day date

	today, err := uc.repo.TodaySummary(ctx, userID, sc.todayStart)
	if err != nil {
		return dto.StatsOverviewDto{}, internalErr("Failed to load today's stats", err)
	}

	days, err := uc.repo.StudyDays(ctx, userID, tz, hour, sc.todayStart.AddDate(-1, 0, -1))
	if err != nil {
		return dto.StatsOverviewDto{}, internalErr("Failed to load study days", err)
	}
	studied := make(map[string]bool, len(days))
	for _, d := range days {
		studied[d.Format(time.DateOnly)] = true
	}
	// The streak survives until today ends, so start counting from yesterday if today is empty.
	streak := 0
	cursor := sc.todayStart
	if !studied[todayDate] {
		cursor = cursor.AddDate(0, 0, -1)
	}
	for studied[cursor.Format(time.DateOnly)] {
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}

	forecastRows, err := uc.repo.DueForecast(ctx, userID, tz, hour, sc.todayStart.AddDate(0, 0, forecastDays))
	if err != nil {
		return dto.StatsOverviewDto{}, internalErr("Failed to load forecast", err)
	}
	byDay := map[string]int{}
	for _, row := range forecastRows {
		key := row.Day.Format(time.DateOnly)
		if key < todayDate {
			key = todayDate // overdue cards count for today
		}
		byDay[key] += row.Count
	}
	forecast := make([]dto.ForecastDayDto, 0, forecastDays)
	for i := 0; i < forecastDays; i++ {
		date := sc.todayStart.AddDate(0, 0, i).Format(time.DateOnly)
		forecast = append(forecast, dto.ForecastDayDto{Date: date, Count: byDay[date]})
	}

	passed, total, err := uc.repo.Retention(ctx, userID, sc.now.AddDate(0, 0, -30))
	if err != nil {
		return dto.StatsOverviewDto{}, internalErr("Failed to load retention", err)
	}
	var retention *float64
	if total > 0 {
		r := float64(passed) / float64(total)
		retention = &r
	}

	totals, err := uc.repo.CardTotals(ctx, userID)
	if err != nil {
		return dto.StatsOverviewDto{}, internalErr("Failed to count cards", err)
	}

	return dto.StatsOverviewDto{
		Today:           dto.TodayDto{Reviews: today.Reviews, NewCards: today.NewCards, Again: today.Again, DurationMs: today.DurationMs},
		StreakDays:      streak,
		StudiedToday:    studied[todayDate],
		Forecast:        forecast,
		Retention30d:    retention,
		RetentionSample: total,
		Totals:          dto.CardTotalsDto{New: totals.New, Learning: totals.Learning, Review: totals.Review, Suspended: totals.Suspended},
	}, nil
}
