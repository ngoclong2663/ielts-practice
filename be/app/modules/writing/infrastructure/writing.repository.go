package infrastructure

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/writing/domain"
	"ielts-practice/app/modules/writing/domain/entity"
)

var _ domain.IWritingRepository = &WritingRepository{}

type WritingRepository struct {
	db *gorm.DB
}

func NewWritingRepository(db *gorm.DB) domain.IWritingRepository {
	return &WritingRepository{db: db}
}

func (r *WritingRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

func (r *WritingRepository) FindTaskBySlug(ctx context.Context, slug string) (*entity.Task, error) {
	// Find + Limit instead of First: a missing slug is normal during seeding and should not be logged.
	var rows []models.WritingTask
	if err := r.conn(ctx).Where("slug = ?", slug).Limit(1).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return taskToEntity(&rows[0])
}

func (r *WritingRepository) UpsertTask(ctx context.Context, t *entity.Task) error {
	figures, err := json.Marshal(t.Figures)
	if err != nil {
		return err
	}
	facts := []byte(t.Facts)
	if len(facts) == 0 {
		facts = []byte("{}")
	}
	row := &models.WritingTask{
		Slug: t.Slug, Pack: t.Pack, SourceRef: t.SourceRef, TestNo: int16(t.TestNo), TaskNo: int16(t.TaskNo),
		Module: t.Module, TaskType: t.TaskType, Topic: t.Topic, Title: t.Title, Prompt: t.Prompt,
		Figures: figures, Facts: facts, MinWords: int16(t.MinWords), Minutes: int16(t.Minutes),
		Status: t.Status, ContentHash: t.ContentHash,
	}
	// Status is set on insert only, so a task archived by hand stays archived when its pack changes.
	return r.conn(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "slug"}},
		DoUpdates: clause.Assignments(map[string]any{
			"pack": row.Pack, "source_ref": row.SourceRef, "test_no": row.TestNo, "task_no": row.TaskNo,
			"module": row.Module, "task_type": row.TaskType, "topic": row.Topic, "title": row.Title,
			"prompt": row.Prompt, "figures": row.Figures, "facts": row.Facts, "min_words": row.MinWords,
			"minutes": row.Minutes, "content_hash": row.ContentHash, "updated_at": time.Now(),
		}),
	}).Create(row).Error
}

func taskToEntity(m *models.WritingTask) (*entity.Task, error) {
	t := &entity.Task{
		ID: m.ID, Slug: m.Slug, Pack: m.Pack, SourceRef: m.SourceRef, TestNo: int(m.TestNo), TaskNo: int(m.TaskNo),
		Module: m.Module, TaskType: m.TaskType, Topic: m.Topic, Title: m.Title, Prompt: m.Prompt,
		Facts: json.RawMessage(m.Facts), MinWords: int(m.MinWords), Minutes: int(m.Minutes),
		Status: m.Status, ContentHash: m.ContentHash, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
	if err := json.Unmarshal(m.Figures, &t.Figures); err != nil {
		return nil, err
	}
	return t, nil
}
