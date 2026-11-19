package domain

import (
	"context"
	"time"

	"ielts-practice/app/modules/writing/domain/entity"
)

type TaskFilter struct {
	TaskNo   int
	TaskType string
	Topic    string
}

// TaskProgress is a learner's best result on one task.
type TaskProgress struct {
	Attempts int
	BestBand *float64
}

type IWritingRepository interface {
	// ---- tasks
	FindTaskBySlug(ctx context.Context, slug string) (*entity.Task, error)
	FindTaskByID(ctx context.Context, id string) (*entity.Task, error)
	// ListTasks returns published tasks in pack/test order.
	ListTasks(ctx context.Context, f TaskFilter) ([]entity.Task, error)
	// UpsertTask inserts by slug or updates the content of an existing task (keeping its id and status).
	UpsertTask(ctx context.Context, t *entity.Task) error
	TaskProgress(ctx context.Context, userID string) (map[string]TaskProgress, error)

	// ---- lessons
	FindLesson(ctx context.Context, taskID, position, taskHash string) (*entity.Lesson, error)
	FindLessonByID(ctx context.Context, id string) (*entity.Lesson, error)
	// InsertLesson keeps the first lesson when two requests race; it returns the stored one.
	InsertLesson(ctx context.Context, l *entity.Lesson) (*entity.Lesson, error)
	// SeedLesson stores a reviewed lesson (Epic 8 §4). A lesson the AI wrote earlier is replaced only when no
	// attempt uses it (an attempt's sentences follow its lesson). It reports whether anything changed.
	SeedLesson(ctx context.Context, l *entity.Lesson) (changed bool, err error)

	FindTaskVocab(ctx context.Context, taskID string) (*entity.LessonVocab, error)
	// InsertTaskVocab keeps the first vocabulary when two requests race; it returns the stored one.
	InsertTaskVocab(ctx context.Context, v *entity.StoredVocab) (*entity.LessonVocab, error)
	TaskVocabHash(ctx context.Context, taskID string) (string, error)
	// SeedTaskVocab stores a reviewed vocabulary set, replacing what the AI wrote.
	SeedTaskVocab(ctx context.Context, v *entity.StoredVocab) error

	// ---- attempts
	CreateAttempt(ctx context.Context, a *entity.Attempt) error
	FindAttempt(ctx context.Context, userID, id string) (*entity.Attempt, error)
	SavePlanAnswers(ctx context.Context, attemptID string, answers []int) error
	ListAttempts(ctx context.Context, userID, taskID string, limit int) ([]entity.AttemptSummary, error)
	UpsertParagraphCheck(ctx context.Context, attemptID string, p *entity.ParagraphCheck) error
	ParagraphChecks(ctx context.Context, attemptID string) ([]entity.ParagraphCheck, error)
	// SaveSubmission stores the answer and marks the attempt submitted (false if it already was).
	SaveSubmission(ctx context.Context, attemptID string, s *entity.Submission) (bool, error)
	FindSubmission(ctx context.Context, attemptID string) (*entity.Submission, error)

	// ---- AI usage
	RecordAIRequest(ctx context.Context, userID string, attemptID *string, kind string, u entity.AIUsage) error
	CountAIRequests(ctx context.Context, userID, kind string, since time.Time) (int, error)
}
