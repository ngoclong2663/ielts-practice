package usecases

import (
	"context"

	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/app/modules/vocab/dto"
)

func (uc *VocabUseCases) GetStatus(ctx context.Context, userID string) (dto.StatusDto, error) {
	limit, used, err := uc.quota(ctx, userID)
	if err != nil {
		sLimit, sUsed, err := uc.suggestionQuota(ctx, userID)
		if err != nil {
			return dto.StatusDto{}, internalErr("Failed to load suggestion limit", err)
		}
		return dto.StatusDto{
			Suggestions: dto.QuotaDto{Limit: sLimit, Used: sUsed, Remaining: max(0, sLimit-sUsed)}}, internalErr("Failed to load quota", err)
	}
	sLimit, sUsed, err := uc.suggestionQuota(ctx, userID)
	if err != nil {
		return dto.StatusDto{}, internalErr("Failed to load suggestion limit", err)
	}
	return dto.StatusDto{
		Suggestions:    dto.QuotaDto{Limit: sLimit, Used: sUsed, Remaining: max(0, sLimit-sUsed)},
		Configured:     uc.generator.Configured(),
		MaxEntries:     domain.MaxEntriesPerJob,
		Quota:          dto.QuotaDto{Limit: limit, Used: used, Remaining: max(0, limit-used)},
		ImageProviders: uc.providerNames(),
	}, nil
}
