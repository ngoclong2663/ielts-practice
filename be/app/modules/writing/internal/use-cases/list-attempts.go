package usecases

import (
	"context"

	"ielts-practice/app/modules/writing/domain/entity"

	"ielts-practice/app/modules/writing/dto"
)

func (uc *WritingUseCases) ListAttempts(ctx context.Context, userID string) ([]dto.AttemptSummaryDto, error) {
	rows, err := uc.repo.ListAttempts(ctx, userID, "", historyLimit)
	if err != nil {
		return nil, internalErr("Failed to load attempts", err)
	}
	out := make([]dto.AttemptSummaryDto, 0, len(rows))
	for _, a := range rows {
		out = append(out, attemptSummary(a))
	}
	return out, nil
}

func (uc *WritingUseCases) GetStatus(ctx context.Context, userID string) (dto.StatusDto, error) {
	since := uc.now().Add(-limitWindow)
	checks, err := uc.repo.CountAIRequests(ctx, userID, entity.KindCheck, since)
	if err != nil {
		return dto.StatusDto{}, internalErr("Failed to load the daily limit", err)
	}
	submits, err := uc.repo.CountAIRequests(ctx, userID, entity.KindSubmit, since)
	if err != nil {
		return dto.StatusDto{}, internalErr("Failed to load the daily limit", err)
	}
	quota := func(limit, used int) dto.QuotaDto {
		return dto.QuotaDto{Limit: limit, Used: used, Remaining: max(0, limit-used)}
	}
	return dto.StatusDto{
		Configured:  uc.coach.Configured(),
		Checks:      quota(uc.cfg.Writing.DailyChecks, checks),
		Submissions: quota(uc.cfg.Writing.DailySubmissions, submits),
	}, nil
}
