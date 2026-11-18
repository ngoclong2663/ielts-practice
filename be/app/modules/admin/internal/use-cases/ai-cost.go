package usecases

import (
	"context"
	"slices"
	"time"

	"ielts-practice/app/modules/admin/domain/entity"
	"ielts-practice/app/modules/admin/domain/rules"
	"ielts-practice/app/modules/admin/dto"
)

// aiCost prices the model-call ledger per provider, model and feature (Epic 7), part of the overview.
func (uc *AdminUseCases) aiCost(ctx context.Context, w entity.Window) (dto.AICostDto, error) {
	rows, err := uc.repo.AIUsage(ctx, w.Since30d)
	if err != nil {
		return dto.AICostDto{}, internalErr("Failed to load AI usage", err)
	}
	since, err := uc.repo.FirstAIUsage(ctx)
	if err != nil {
		return dto.AICostDto{}, internalErr("Failed to load AI usage", err)
	}
	return summarizeAICost(rows, since), nil
}

func summarizeAICost(rows []entity.AIUsageRow, since *time.Time) dto.AICostDto {
	out := dto.AICostDto{Since: since, Providers: []dto.AIProviderCostDto{}, Models: []dto.AIModelCostDto{}, Features: []dto.AIFeatureCostDto{}}
	providers := map[string]*dto.AIProviderCostDto{}
	models := map[string]*dto.AIModelCostDto{}
	features := map[string]*dto.AIFeatureCostDto{}
	for _, r := range rows {
		cost30 := rules.Cost(r.Model, entity.Tokens{Input: r.InputTokens30d, Output: r.OutputTokens30d, CacheRead: r.CacheReadTokens30d})
		costAll := rules.Cost(r.Model, entity.Tokens{Input: r.InputTokens, Output: r.OutputTokens, CacheRead: r.CacheReadTokens})
		out.Cost30dUSD += cost30
		out.CostTotalUSD += costAll

		p := providers[r.Provider]
		if p == nil {
			p = &dto.AIProviderCostDto{Provider: r.Provider}
			providers[r.Provider] = p
		}
		p.Calls30d += r.Calls30d
		p.Cost30dUSD += cost30
		p.CostTotalUSD += costAll

		m := models[r.Model]
		if m == nil {
			m = &dto.AIModelCostDto{Provider: r.Provider, Model: r.Model}
			if price, ok := rules.PriceOf(r.Model); ok {
				m.Price = &dto.ModelPriceDto{Input: price.Input, Output: price.Output, CacheRead: price.CacheRead}
			}
			models[r.Model] = m
		}
		m.Calls30d += r.Calls30d
		m.Failed30d += r.Failed30d
		m.InputTokens30d += r.InputTokens30d
		m.OutputTokens30d += r.OutputTokens30d
		m.CacheReadTokens30d += r.CacheReadTokens30d
		m.Cost30dUSD += cost30
		m.CallsTotal += r.Calls
		m.CostTotalUSD += costAll

		f := features[r.Feature]
		if f == nil {
			f = &dto.AIFeatureCostDto{Feature: r.Feature}
			features[r.Feature] = f
		}
		f.Calls30d += r.Calls30d
		f.Cost30dUSD += cost30
	}
	for _, p := range providers {
		out.Providers = append(out.Providers, *p)
	}
	for _, m := range models {
		out.Models = append(out.Models, *m)
	}
	for _, f := range features {
		out.Features = append(out.Features, *f)
	}
	// Most expensive first; ties (e.g. no calls this month) by name so the order is stable.
	slices.SortFunc(out.Providers, func(a, b dto.AIProviderCostDto) int {
		return byCost(a.CostTotalUSD, b.CostTotalUSD, a.Provider, b.Provider)
	})
	slices.SortFunc(out.Models, func(a, b dto.AIModelCostDto) int { return byCost(a.CostTotalUSD, b.CostTotalUSD, a.Model, b.Model) })
	slices.SortFunc(out.Features, func(a, b dto.AIFeatureCostDto) int { return byCost(a.Cost30dUSD, b.Cost30dUSD, a.Feature, b.Feature) })
	return out
}

func byCost(a, b float64, nameA, nameB string) int {
	switch {
	case a > b:
		return -1
	case a < b:
		return 1
	case nameA < nameB:
		return -1
	case nameA > nameB:
		return 1
	}
	return 0
}
