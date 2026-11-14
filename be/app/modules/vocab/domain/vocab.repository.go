package domain

import (
	"context"
	"time"

	"ielts-practice/app/modules/vocab/domain/entity"
)

// ClaimResult: what happened when the worker tried to own the generation of a lexeme key.
type ClaimResult struct {
	Lexeme *entity.Lexeme
	// Owned: this worker must generate the content now.
	Owned bool
}

type IVocabRepository interface {
	// ---- lexemes
	FindLexemeByKey(ctx context.Context, key string) (*entity.Lexeme, error)
	FindLexemeByID(ctx context.Context, id string) (*entity.Lexeme, error)
	// ClaimLexeme inserts a pending lexeme or takes over a failed/stale one. When another worker
	// is generating it right now, Owned is false and Lexeme.Status is pending.
	ClaimLexeme(ctx context.Context, key, headword, senseHint string, staleBefore time.Time) (ClaimResult, error)
	SaveLexeme(ctx context.Context, lexeme *entity.Lexeme) error
	SetLexemeStatus(ctx context.Context, id, status string) error
	// ReportLexeme stores one report per user; past the threshold the lexeme needs review.
	ReportLexeme(ctx context.Context, lexemeID, userID, reason, note string, threshold int) (bool, error)

	// ---- jobs and items
	CreateJob(ctx context.Context, job *entity.Job) (*entity.Job, error)
	FindJob(ctx context.Context, userID, jobID string) (*entity.Job, error)
	ListJobs(ctx context.Context, userID string, limit int) ([]entity.Job, error)
	RefreshJobProgress(ctx context.Context, jobID string, now time.Time) error
	AddJobUsage(ctx context.Context, jobID string, usage entity.Usage) error
	SetJobCommitted(ctx context.Context, jobID, deckID string) error

	FindItem(ctx context.Context, userID, jobID, itemID string) (*entity.JobItem, error)
	UpdateItem(ctx context.Context, item *entity.JobItem) error
	// ClaimQueuedItems moves up to `limit` due items to status "content" and increments attempts.
	ClaimQueuedItems(ctx context.Context, limit int, now time.Time) ([]entity.JobItem, error)
	// RequeueStuckItems puts items back in the queue when a worker died mid-way.
	RequeueStuckItems(ctx context.Context, before time.Time) (int64, error)
	// ChargedWordsSince counts generated (uncached) words for the daily quota.
	ChargedWordsSince(ctx context.Context, userID string, since time.Time) (int, error)

	// ---- image links
	UpsertImageLink(ctx context.Context, link *entity.ImageLink) error
	FindImageLink(ctx context.Context, provider, providerID string) (*entity.ImageLink, error)
	// ReplaceImageEverywhere swaps a removed photo in lexemes, job items and cards (nil removes the image).
	ReplaceImageEverywhere(ctx context.Context, provider, providerID string, replacement *entity.CardImage) error
}
