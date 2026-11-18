// Package persistence is the PostgreSQL (GORM) adapter of the writing module.
package persistence

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

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err) // only plain structs and slices are marshalled here
	}
	return b
}

// first loads at most one row; a missing row is (nil, nil) and is not logged as an error.
func first[M any](q *gorm.DB) (*M, error) {
	var rows []M
	if err := q.Limit(1).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// ---- tasks -------------------------------------------------------------------

func (r *WritingRepository) FindTaskBySlug(ctx context.Context, slug string) (*entity.Task, error) {
	row, err := first[models.WritingTask](r.conn(ctx).Where("slug = ?", slug))
	if err != nil || row == nil {
		return nil, err
	}
	return taskToEntity(row)
}

func (r *WritingRepository) FindTaskByID(ctx context.Context, id string) (*entity.Task, error) {
	row, err := first[models.WritingTask](r.conn(ctx).Where("id = ? AND status = ?", id, entity.StatusPublished))
	if err != nil || row == nil {
		return nil, err
	}
	return taskToEntity(row)
}

func (r *WritingRepository) ListTasks(ctx context.Context, f domain.TaskFilter) ([]entity.Task, error) {
	q := r.conn(ctx).Where("status = ?", entity.StatusPublished)
	if f.TaskNo != 0 {
		q = q.Where("task_no = ?", f.TaskNo)
	}
	if f.TaskType != "" {
		q = q.Where("task_type = ?", f.TaskType)
	}
	if f.Topic != "" {
		q = q.Where("topic = ?", f.Topic)
	}
	var rows []models.WritingTask
	if err := q.Order("pack DESC, test_no, task_no").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.Task, 0, len(rows))
	for i := range rows {
		t, err := taskToEntity(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, nil
}

func (r *WritingRepository) UpsertTask(ctx context.Context, t *entity.Task) error {
	facts := []byte(t.Facts)
	if len(facts) == 0 {
		facts = []byte("{}")
	}
	positions := t.Positions
	if positions == nil {
		positions = []string{}
	}
	row := &models.WritingTask{
		Slug: t.Slug, Pack: t.Pack, SourceRef: t.SourceRef, TestNo: int16(t.TestNo), TaskNo: int16(t.TaskNo),
		Module: t.Module, TaskType: t.TaskType, Topic: t.Topic, Title: t.Title, Prompt: t.Prompt,
		Figures: mustJSON(t.Figures), Facts: facts, Positions: mustJSON(positions),
		MinWords: int16(t.MinWords), Minutes: int16(t.Minutes), Status: t.Status, ContentHash: t.ContentHash,
	}
	// Status is set on insert only, so a task archived by hand stays archived when its pack changes.
	return r.conn(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "slug"}},
		DoUpdates: clause.Assignments(map[string]any{
			"pack": row.Pack, "source_ref": row.SourceRef, "test_no": row.TestNo, "task_no": row.TaskNo,
			"module": row.Module, "task_type": row.TaskType, "topic": row.Topic, "title": row.Title,
			"prompt": row.Prompt, "figures": row.Figures, "facts": row.Facts, "positions": row.Positions,
			"min_words": row.MinWords, "minutes": row.Minutes, "content_hash": row.ContentHash, "updated_at": time.Now(),
		}),
	}).Create(row).Error
}

func (r *WritingRepository) TaskProgress(ctx context.Context, userID string) (map[string]domain.TaskProgress, error) {
	var rows []struct {
		TaskID   string
		Attempts int
		BestBand *float64
	}
	err := r.conn(ctx).Raw(`
		SELECT a.task_id, count(*) AS attempts, max(s.band_high) AS best_band
		FROM writing_attempts a LEFT JOIN writing_submissions s ON s.attempt_id = a.id
		WHERE a.user_id = ? GROUP BY a.task_id`, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]domain.TaskProgress, len(rows))
	for _, row := range rows {
		out[row.TaskID] = domain.TaskProgress{Attempts: row.Attempts, BestBand: row.BestBand}
	}
	return out, nil
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
	if len(m.Positions) > 0 {
		if err := json.Unmarshal(m.Positions, &t.Positions); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// ---- lessons -----------------------------------------------------------------

func (r *WritingRepository) FindLesson(ctx context.Context, taskID, position, taskHash string) (*entity.Lesson, error) {
	row, err := first[models.WritingLesson](r.conn(ctx).
		Where("task_id = ? AND position = ? AND task_hash = ?", taskID, position, taskHash))
	if err != nil || row == nil {
		return nil, err
	}
	return lessonToEntity(row)
}

func (r *WritingRepository) FindLessonByID(ctx context.Context, id string) (*entity.Lesson, error) {
	row, err := first[models.WritingLesson](r.conn(ctx).Where("id = ?", id))
	if err != nil || row == nil {
		return nil, err
	}
	return lessonToEntity(row)
}

func (r *WritingRepository) InsertLesson(ctx context.Context, l *entity.Lesson) (*entity.Lesson, error) {
	row := &models.WritingLesson{
		TaskID: l.TaskID, Position: l.Position, TaskHash: l.TaskHash, Content: mustJSON(l.Content),
		Model: l.Model, InputTokens: l.InputTokens, OutputTokens: l.OutputTokens,
	}
	err := r.conn(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(row).Error
	if err != nil {
		return nil, err
	}
	return r.FindLesson(ctx, l.TaskID, l.Position, l.TaskHash)
}

func (r *WritingRepository) FindLessonVocab(ctx context.Context, lessonID string) (*entity.LessonVocab, error) {
	row, err := first[models.WritingLessonVocab](r.conn(ctx).Where("lesson_id = ?", lessonID))
	if err != nil || row == nil {
		return nil, err
	}
	var v entity.LessonVocab
	return &v, json.Unmarshal(row.Content, &v)
}

func (r *WritingRepository) InsertLessonVocab(ctx context.Context, v *entity.StoredVocab) (*entity.LessonVocab, error) {
	row := &models.WritingLessonVocab{
		LessonID: v.LessonID, Content: mustJSON(v.Content),
		Model: v.Usage.Model, InputTokens: v.Usage.InputTokens, OutputTokens: v.Usage.OutputTokens,
	}
	if err := r.conn(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(row).Error; err != nil {
		return nil, err
	}
	return r.FindLessonVocab(ctx, v.LessonID)
}

func lessonToEntity(m *models.WritingLesson) (*entity.Lesson, error) {
	l := &entity.Lesson{
		ID: m.ID, TaskID: m.TaskID, Position: m.Position, TaskHash: m.TaskHash, Model: m.Model,
		InputTokens: m.InputTokens, OutputTokens: m.OutputTokens, CreatedAt: m.CreatedAt,
	}
	return l, json.Unmarshal(m.Content, &l.Content)
}

// ---- attempts ----------------------------------------------------------------

func (r *WritingRepository) CreateAttempt(ctx context.Context, a *entity.Attempt) error {
	row := &models.WritingAttempt{
		UserID: a.UserID, TaskID: a.TaskID, LessonID: a.LessonID, Level: int16(a.Level),
		Position: a.Position, Status: entity.StatusInProgress,
	}
	if err := r.conn(ctx).Create(row).Error; err != nil {
		return err
	}
	a.ID, a.Status, a.CreatedAt, a.UpdatedAt = row.ID, row.Status, row.CreatedAt, row.UpdatedAt
	return nil
}

func (r *WritingRepository) FindAttempt(ctx context.Context, userID, id string) (*entity.Attempt, error) {
	row, err := first[models.WritingAttempt](r.conn(ctx).Where("id = ? AND user_id = ?", id, userID))
	if err != nil || row == nil {
		return nil, err
	}
	return attemptToEntity(row)
}

func (r *WritingRepository) SavePlanAnswers(ctx context.Context, attemptID string, answers []int) error {
	return r.conn(ctx).Model(&models.WritingAttempt{}).Where("id = ?", attemptID).
		Updates(map[string]any{"plan_answers": mustJSON(answers), "updated_at": time.Now()}).Error
}

func (r *WritingRepository) ListAttempts(ctx context.Context, userID, taskID string, limit int) ([]entity.AttemptSummary, error) {
	q := r.conn(ctx).Table("writing_attempts a").
		Select("a.*, t.title AS task_title, t.task_no, t.task_type, s.band_low, s.band_high").
		Joins("JOIN writing_tasks t ON t.id = a.task_id").
		Joins("LEFT JOIN writing_submissions s ON s.attempt_id = a.id").
		Where("a.user_id = ?", userID)
	if taskID != "" {
		q = q.Where("a.task_id = ?", taskID)
	}
	var rows []struct {
		models.WritingAttempt
		TaskTitle string
		TaskNo    int
		TaskType  string
		BandLow   *float64
		BandHigh  *float64
	}
	if err := q.Order("a.created_at DESC").Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.AttemptSummary, 0, len(rows))
	for i := range rows {
		a, err := attemptToEntity(&rows[i].WritingAttempt)
		if err != nil {
			return nil, err
		}
		out = append(out, entity.AttemptSummary{
			Attempt: *a, TaskTitle: rows[i].TaskTitle, TaskNo: rows[i].TaskNo, TaskType: rows[i].TaskType,
			BandLow: rows[i].BandLow, BandHigh: rows[i].BandHigh,
		})
	}
	return out, nil
}

func attemptToEntity(m *models.WritingAttempt) (*entity.Attempt, error) {
	a := &entity.Attempt{
		ID: m.ID, UserID: m.UserID, TaskID: m.TaskID, LessonID: m.LessonID, Level: int(m.Level),
		Position: m.Position, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, SubmittedAt: m.SubmittedAt,
	}
	if len(m.PlanAnswers) > 0 && string(m.PlanAnswers) != "null" {
		if err := json.Unmarshal(m.PlanAnswers, &a.PlanAnswers); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (r *WritingRepository) UpsertParagraphCheck(ctx context.Context, attemptID string, p *entity.ParagraphCheck) error {
	row := &models.WritingAttemptParagraph{
		AttemptID: attemptID, ParagraphKey: p.Key, Answers: mustJSON(p.Answers), Feedback: mustJSON(p.Feedback), CheckedCount: 1,
	}
	err := r.conn(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "attempt_id"}, {Name: "paragraph_key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"answers": row.Answers, "feedback": row.Feedback, "updated_at": time.Now(),
			"checked_count": gorm.Expr("writing_attempt_paragraphs.checked_count + 1"),
		}),
	}).Create(row).Error
	if err != nil {
		return err
	}
	return r.conn(ctx).Model(&models.WritingAttempt{}).Where("id = ?", attemptID).Update("updated_at", time.Now()).Error
}

func (r *WritingRepository) ParagraphChecks(ctx context.Context, attemptID string) ([]entity.ParagraphCheck, error) {
	var rows []models.WritingAttemptParagraph
	if err := r.conn(ctx).Where("attempt_id = ?", attemptID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.ParagraphCheck, 0, len(rows))
	for _, row := range rows {
		p := entity.ParagraphCheck{Key: row.ParagraphKey, CheckedCount: row.CheckedCount, UpdatedAt: row.UpdatedAt}
		if err := json.Unmarshal(row.Answers, &p.Answers); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(row.Feedback, &p.Feedback); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *WritingRepository) SaveSubmission(ctx context.Context, attemptID string, s *entity.Submission) (bool, error) {
	saved := false
	err := r.conn(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		res := tx.Model(&models.WritingAttempt{}).
			Where("id = ? AND status = ?", attemptID, entity.StatusInProgress).
			Updates(map[string]any{"status": entity.StatusSubmitted, "submitted_at": now, "updated_at": now})
		if res.Error != nil || res.RowsAffected == 0 {
			return res.Error
		}
		saved = true
		return tx.Create(&models.WritingSubmission{
			AttemptID: attemptID, Text: s.Text, WordCount: s.WordCount, SecondsSpent: s.SecondsSpent,
			Feedback: mustJSON(s.Feedback), BandLow: s.BandLow, BandHigh: s.BandHigh,
		}).Error
	})
	return saved, err
}

func (r *WritingRepository) FindSubmission(ctx context.Context, attemptID string) (*entity.Submission, error) {
	row, err := first[models.WritingSubmission](r.conn(ctx).Where("attempt_id = ?", attemptID))
	if err != nil || row == nil {
		return nil, err
	}
	s := &entity.Submission{
		Text: row.Text, WordCount: row.WordCount, SecondsSpent: row.SecondsSpent,
		BandLow: row.BandLow, BandHigh: row.BandHigh, CreatedAt: row.CreatedAt,
	}
	return s, json.Unmarshal(row.Feedback, &s.Feedback)
}

// ---- AI usage ----------------------------------------------------------------

func (r *WritingRepository) RecordAIRequest(ctx context.Context, userID string, attemptID *string, kind string, u entity.AIUsage) error {
	return r.conn(ctx).Create(&models.WritingAIRequest{
		UserID: userID, AttemptID: attemptID, Kind: kind, Model: u.Model,
		InputTokens: u.InputTokens, OutputTokens: u.OutputTokens,
	}).Error
}

func (r *WritingRepository) CountAIRequests(ctx context.Context, userID, kind string, since time.Time) (int, error) {
	var n int64
	err := r.conn(ctx).Model(&models.WritingAIRequest{}).
		Where("user_id = ? AND kind = ? AND created_at >= ?", userID, kind, since).Count(&n).Error
	return int(n), err
}
