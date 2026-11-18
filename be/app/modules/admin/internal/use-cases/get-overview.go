package usecases

import (
	"context"

	"ielts-practice/app/modules/admin/domain/rules"
	"ielts-practice/app/modules/admin/dto"
)

// GetOverview: the key numbers of the app on one page (Epic 6 §6), computed on request.
func (uc *AdminUseCases) GetOverview(ctx context.Context) (dto.OverviewDto, error) {
	w := uc.window()
	t, err := uc.repo.Totals(ctx, w)
	if err != nil {
		return dto.OverviewDto{}, internalErr("Failed to load totals", err)
	}
	byDate, err := uc.repo.Daily(ctx, w)
	if err != nil {
		return dto.OverviewDto{}, internalErr("Failed to load daily numbers", err)
	}
	reports, err := uc.repo.RecentReports(ctx, recentReports)
	if err != nil {
		return dto.OverviewDto{}, internalErr("Failed to load content reports", err)
	}
	aiCost, err := uc.aiCost(ctx, w)
	if err != nil {
		return dto.OverviewDto{}, err
	}

	out := dto.OverviewDto{
		GeneratedAt: w.Now.UTC(),
		Timezone:    w.Timezone,
		Users: dto.UsersOverviewDto{
			Total: t.Users.Total, New7d: t.Users.New7d, New30d: t.Users.New30d,
			Active1d: t.Users.Active1d, Active7d: t.Users.Active7d, Active30d: t.Users.Active30d,
		},
		Vocab: dto.VocabOverviewDto{
			Decks: t.Vocab.Decks, Cards: t.Vocab.Cards,
			CardsByState: dto.CardsByStateDto{
				New: t.Vocab.CardsNew, Learning: t.Vocab.CardsLearning, Review: t.Vocab.CardsReview, Relearning: t.Vocab.CardsRelearning,
			},
			ReviewsTotal: t.Vocab.ReviewsTotal, Reviews7d: t.Vocab.Reviews7d, LexemesCached: t.Vocab.LexemesCached,
		},
		AI: dto.AIOverviewDto{
			WordsGeneratedTotal: t.AI.WordsGeneratedTotal, WordsGenerated7d: t.AI.WordsGenerated7d,
			WordsCachedTotal: t.AI.WordsCachedTotal, WordsFailed7d: t.AI.WordsFailed7d,
			TopicListsTotal: t.AI.TopicListsTotal, TopicLists7d: t.AI.TopicLists7d,
			WritingLessonsTotal: t.AI.WritingLessonsTotal,
			WritingChecksTotal:  t.AI.WritingChecksTotal, WritingChecks7d: t.AI.WritingChecks7d,
			WritingReviewsTotal: t.AI.WritingReviewsTotal, WritingReviews7d: t.AI.WritingReviews7d,
			InputTokens30d: t.AI.InputTokens30d, OutputTokens30d: t.AI.OutputTokens30d,
		},
		Writing: dto.WritingOverviewDto{
			AttemptsTotal: t.Writing.AttemptsTotal, AttemptsSubmitted: t.Writing.AttemptsSubmitted, AvgBand30d: t.Writing.AvgBand30d,
		},
		Feedback: dto.FeedbackOverviewDto{New: t.Feedback.New, Total: t.Feedback.Total},
		ContentReports: dto.ContentReportsOverviewDto{
			Total: t.ContentReports.Total, Last7d: t.ContentReports.Last7d, NeedsReview: t.ContentReports.NeedsReview,
		},
		AICost:        aiCost,
		Daily:         []dto.DailyDto{},
		RecentReports: make([]dto.RecentReportDto, 0, len(reports)),
	}
	for _, d := range rules.FillDaily(w, byDate) {
		out.Daily = append(out.Daily, dto.DailyDto{
			Date: d.Date, Signups: d.Signups, ActiveUsers: d.ActiveUsers, Reviews: d.Reviews, AIRequests: d.AIRequests,
		})
	}
	for _, r := range reports {
		out.RecentReports = append(out.RecentReports, dto.RecentReportDto{
			ID: r.ID, Headword: r.Headword, Reason: r.Reason, Note: r.Note, UserEmail: r.UserEmail, CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}
