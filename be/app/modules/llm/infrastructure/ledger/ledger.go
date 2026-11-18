// Package ledger is the PostgreSQL (GORM) adapter that stores every model call (table ai_usage).
package ledger

import (
	"context"

	"gorm.io/gorm"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	llmout "ielts-practice/app/modules/llm/ports/out"
)

type UsageLedger struct {
	db *gorm.DB
}

var _ llmout.IUsageLedger = (*UsageLedger)(nil)

func NewUsageLedger(db *gorm.DB) llmout.IUsageLedger {
	return &UsageLedger{db: db}
}

func (l *UsageLedger) Record(ctx context.Context, r llmout.UsageRecord) error {
	return database.GetDBFromContext(ctx, l.db).WithContext(ctx).Create(&models.AIUsage{
		Provider: r.Provider, Model: r.Model, Feature: r.Feature, OK: r.OK,
		InputTokens: r.InputTokens, OutputTokens: r.OutputTokens, CacheReadTokens: r.CacheReadTokens,
	}).Error
}
