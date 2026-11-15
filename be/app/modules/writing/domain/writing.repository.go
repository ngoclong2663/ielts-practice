package domain

import (
	"context"

	"ielts-practice/app/modules/writing/domain/entity"
)

type IWritingRepository interface {
	FindTaskBySlug(ctx context.Context, slug string) (*entity.Task, error)
	// UpsertTask inserts by slug or updates the content of an existing task (keeping its id and status).
	UpsertTask(ctx context.Context, t *entity.Task) error
}
