package rules

import "ielts-practice/app/modules/study/domain/entity"

// QueueInput holds candidate cards (already sorted) and today's limit counters.
type QueueInput struct {
	LearningNow   []entity.StudyCard // learning/relearning due now, by due time
	Reviews       []entity.StudyCard // reviews due today, by due time (already capped per deck)
	News          []entity.StudyCard // new cards in deck order (already capped per deck)
	LearningAhead []entity.StudyCard // learning cards due within the learn-ahead window

	ReviewRemaining int // reviews still allowed and due today
	NewRemaining    int // new cards still allowed today
	ReviewDoneToday int
	NewDoneToday    int
	Size            int
}

// BuildQueue orders the next cards like Anki: due learning cards first, then reviews and new
// cards mixed so new cards are spread evenly through the day, then learning cards that are
// due soon (learn-ahead) only when nothing else is left.
func BuildQueue(in QueueInput) []entity.StudyCard {
	out := make([]entity.StudyCard, 0, in.Size)
	add := func(c entity.StudyCard) bool {
		out = append(out, c)
		return len(out) >= in.Size
	}

	for _, c := range in.LearningNow {
		if add(c) {
			return out
		}
	}

	newTotal := in.NewRemaining + in.NewDoneToday
	reviewTotal := in.ReviewRemaining + in.ReviewDoneToday
	newRem, revRem := in.NewRemaining, in.ReviewRemaining
	ni, ri := 0, 0
	for newRem+revRem > 0 && len(out) < in.Size {
		// Pick a new card while the share of new cards left is at least today's overall share.
		pickNew := newRem > 0 && (revRem == 0 || newRem*(reviewTotal+newTotal) >= newTotal*(revRem+newRem))
		if pickNew && ni < len(in.News) {
			out = append(out, in.News[ni])
			ni, newRem = ni+1, newRem-1
			continue
		}
		if ri < len(in.Reviews) && revRem > 0 {
			out = append(out, in.Reviews[ri])
			ri, revRem = ri+1, revRem-1
			continue
		}
		if ni < len(in.News) && newRem > 0 {
			out = append(out, in.News[ni])
			ni, newRem = ni+1, newRem-1
			continue
		}
		break
	}
	if len(out) > 0 {
		return out
	}

	for _, c := range in.LearningAhead {
		if add(c) {
			return out
		}
	}
	return out
}

// ApplyLimits turns raw deck counts into what is still available today.
func ApplyLimits(stat entity.DeckStat, today entity.TodayStat, defaultNew, defaultReviews int) (newRemaining, reviewRemaining int) {
	newLimit, reviewLimit := defaultNew, defaultReviews
	if stat.NewPerDay != nil {
		newLimit = *stat.NewPerDay
	}
	if stat.ReviewsPerDay != nil {
		reviewLimit = *stat.ReviewsPerDay
	}
	newRemaining = min(stat.NewAvailable, max(0, newLimit-today.Introduced))
	reviewRemaining = min(stat.ReviewDue, max(0, reviewLimit-today.Reviewed))
	return newRemaining, reviewRemaining
}
