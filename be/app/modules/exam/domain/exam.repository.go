package domain

import (
	"context"

	"ielts-practice/app/modules/exam/domain/entity"
)

// IExamRepository stores the test bank (books, tests, parts and keys) and the learners' submitted attempts.
type IExamRepository interface {
	// Seed.
	FindBookBySlug(ctx context.Context, slug string) (*entity.Book, error)
	UpsertBook(ctx context.Context, b *entity.Book) error
	FindTestBySlug(ctx context.Context, slug string) (*entity.Test, error)
	UpsertTest(ctx context.Context, t *entity.Test) error
	FindPartBySlug(ctx context.Context, slug string) (*entity.Part, error)
	// UpsertPart writes the part and its key (the caller runs it in a transaction).
	UpsertPart(ctx context.Context, p *entity.Part, key entity.Key) error

	// Test bank.
	ListPublishedBooks(ctx context.Context) ([]entity.Book, error)
	FindBookByID(ctx context.Context, id string) (*entity.Book, error)
	ListTests(ctx context.Context, bookIDs []string) ([]entity.Test, error)
	FindTestByID(ctx context.Context, id string) (*entity.Test, error)
	// ListParts returns the published parts of the tests for one skill, ordered by test and part number.
	ListParts(ctx context.Context, testIDs []string, skill string) ([]entity.Part, error)
	ListPartsByIDs(ctx context.Context, ids []string) ([]entity.Part, error)
	FindKeys(ctx context.Context, partIDs []string) (map[string]entity.Key, error)

	// Attempts.
	CreateAttempt(ctx context.Context, a *entity.Attempt) error
	FindAttempt(ctx context.Context, userID, id string) (*entity.Attempt, error)
	// ListAttempts returns the learner's attempts for a skill, newest first. withDetail loads answers and
	// results; before (an attempt id) pages backwards; limit 0 means all.
	ListAttempts(ctx context.Context, userID, skill string, limit int, before string, withDetail bool) ([]entity.Attempt, error)
}
