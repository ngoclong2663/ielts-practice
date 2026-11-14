package usecases

import (
	"context"

	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/app/modules/vocab/dto"
)

func (uc *VocabUseCases) GetStatus(ctx context.Context, userID string) (dto.StatusDto, error) {
	limit, used, err := uc.quota(ctx, userID)
	if err != nil {
		return dto.StatusDto{}, internalErr("Failed to load quota", err)
	}
	return dto.StatusDto{
		Configured:     uc.generator.Configured(),
		MaxEntries:     domain.MaxEntriesPerJob,
		Quota:          dto.QuotaDto{Limit: limit, Used: used, Remaining: max(0, limit-used)},
		ImageProviders: uc.providerNames(),
	}, nil
}
