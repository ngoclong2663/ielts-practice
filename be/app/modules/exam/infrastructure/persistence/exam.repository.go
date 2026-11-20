// Package persistence is the PostgreSQL (GORM) adapter of the exam module.
package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/exam/domain"
	"ielts-practice/app/modules/exam/domain/entity"
)

var _ domain.IExamRepository = &ExamRepository{}

type ExamRepository struct {
	db *gorm.DB
}

func NewExamRepository(db *gorm.DB) domain.IExamRepository {
	return &ExamRepository{db: db}
}

func (r *ExamRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

// ---- books and tests --------------------------------------------------------

func bookEntity(m *models.ExamBook) (*entity.Book, error) {
	b := &entity.Book{ID: m.ID, Slug: m.Slug, SourceRef: m.SourceRef, Sort: int(m.Sort), Status: m.Status}
	if err := json.Unmarshal(m.Title, &b.Title); err != nil {
		return nil, fmt.Errorf("book %s title: %w", m.Slug, err)
	}
	return b, nil
}

func (r *ExamRepository) FindBookBySlug(ctx context.Context, slug string) (*entity.Book, error) {
	var rows []models.ExamBook
	if err := r.conn(ctx).Where("slug = ?", slug).Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return nil, err
	}
	return bookEntity(&rows[0])
}

func (r *ExamRepository) FindBookByID(ctx context.Context, id string) (*entity.Book, error) {
	var rows []models.ExamBook
	if err := r.conn(ctx).Where("id = ?", id).Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return nil, err
	}
	return bookEntity(&rows[0])
}

func (r *ExamRepository) UpsertBook(ctx context.Context, b *entity.Book) error {
	title, _ := json.Marshal(b.Title)
	existing, err := r.FindBookBySlug(ctx, b.Slug)
	if err != nil {
		return err
	}
	if existing == nil {
		row := models.ExamBook{Slug: b.Slug, SourceRef: b.SourceRef, Title: title, Sort: int16(b.Sort), Status: b.Status}
		if err := r.conn(ctx).Create(&row).Error; err != nil {
			return err
		}
		b.ID = row.ID
		return nil
	}
	b.ID = existing.ID
	return r.conn(ctx).Model(&models.ExamBook{}).Where("id = ?", b.ID).Updates(map[string]any{
		"source_ref": b.SourceRef, "title": title, "sort": b.Sort, "status": b.Status, "updated_at": gorm.Expr("now()"),
	}).Error
}

func (r *ExamRepository) ListPublishedBooks(ctx context.Context) ([]entity.Book, error) {
	var rows []models.ExamBook
	if err := r.conn(ctx).Where("status = ?", entity.StatusPublished).Order("sort, slug").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.Book, 0, len(rows))
	for i := range rows {
		b, err := bookEntity(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *b)
	}
	return out, nil
}

func testEntity(m *models.ExamTest) entity.Test {
	return entity.Test{ID: m.ID, BookID: m.BookID, TestNo: int(m.TestNo), Slug: m.Slug}
}

func (r *ExamRepository) FindTestBySlug(ctx context.Context, slug string) (*entity.Test, error) {
	var rows []models.ExamTest
	if err := r.conn(ctx).Where("slug = ?", slug).Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return nil, err
	}
	t := testEntity(&rows[0])
	return &t, nil
}

func (r *ExamRepository) FindTestByID(ctx context.Context, id string) (*entity.Test, error) {
	var rows []models.ExamTest
	if err := r.conn(ctx).Where("id = ?", id).Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return nil, err
	}
	t := testEntity(&rows[0])
	return &t, nil
}

func (r *ExamRepository) UpsertTest(ctx context.Context, t *entity.Test) error {
	existing, err := r.FindTestBySlug(ctx, t.Slug)
	if err != nil {
		return err
	}
	if existing != nil {
		t.ID = existing.ID
		return nil
	}
	row := models.ExamTest{BookID: t.BookID, TestNo: int16(t.TestNo), Slug: t.Slug}
	if err := r.conn(ctx).Create(&row).Error; err != nil {
		return err
	}
	t.ID = row.ID
	return nil
}

func (r *ExamRepository) ListTests(ctx context.Context, bookIDs []string) ([]entity.Test, error) {
	if len(bookIDs) == 0 {
		return nil, nil
	}
	var rows []models.ExamTest
	if err := r.conn(ctx).Where("book_id IN ?", bookIDs).Order("test_no").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.Test, 0, len(rows))
	for i := range rows {
		out = append(out, testEntity(&rows[i]))
	}
	return out, nil
}

// ---- parts and keys ---------------------------------------------------------

func partEntity(m *models.ExamPart) (*entity.Part, error) {
	p := &entity.Part{ID: m.ID, TestID: m.TestID, Skill: m.Skill, PartNo: int(m.PartNo), Slug: m.Slug, SourceRef: m.SourceRef,
		Title: m.Title, Topic: m.Topic, QFrom: int(m.QFrom), QTo: int(m.QTo), Status: m.Status, ContentHash: m.ContentHash}
	if err := json.Unmarshal(m.Content, &p.Content); err != nil {
		return nil, fmt.Errorf("part %s content: %w", m.Slug, err)
	}
	if len(m.QuestionTypes) > 0 {
		if err := json.Unmarshal(m.QuestionTypes, &p.QuestionTypes); err != nil {
			return nil, fmt.Errorf("part %s types: %w", m.Slug, err)
		}
	}
	if len(m.Audio) > 0 && string(m.Audio) != "null" {
		p.Audio = &entity.Audio{}
		if err := json.Unmarshal(m.Audio, p.Audio); err != nil {
			return nil, fmt.Errorf("part %s audio: %w", m.Slug, err)
		}
	}
	return p, nil
}

func partEntities(rows []models.ExamPart) ([]entity.Part, error) {
	out := make([]entity.Part, 0, len(rows))
	for i := range rows {
		p, err := partEntity(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, nil
}

func (r *ExamRepository) FindPartBySlug(ctx context.Context, slug string) (*entity.Part, error) {
	var rows []models.ExamPart
	if err := r.conn(ctx).Where("slug = ?", slug).Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return nil, err
	}
	return partEntity(&rows[0])
}

func (r *ExamRepository) UpsertPart(ctx context.Context, p *entity.Part, key entity.Key) error {
	content, _ := json.Marshal(p.Content)
	types, _ := json.Marshal(p.QuestionTypes)
	var audio []byte
	if p.Audio != nil {
		audio, _ = json.Marshal(p.Audio)
	}
	keyJSON, _ := json.Marshal(key)
	existing, err := r.FindPartBySlug(ctx, p.Slug)
	if err != nil {
		return err
	}
	if existing == nil {
		row := models.ExamPart{TestID: p.TestID, Skill: p.Skill, PartNo: int16(p.PartNo), Slug: p.Slug, SourceRef: p.SourceRef,
			Title: p.Title, Topic: p.Topic, QFrom: int16(p.QFrom), QTo: int16(p.QTo), QuestionTypes: types, Content: content,
			Audio: audio, Status: p.Status, ContentHash: p.ContentHash}
		if err := r.conn(ctx).Create(&row).Error; err != nil {
			return err
		}
		p.ID = row.ID
		return r.conn(ctx).Create(&models.ExamKey{PartID: p.ID, Key: keyJSON}).Error
	}
	p.ID = existing.ID
	if err := r.conn(ctx).Model(&models.ExamPart{}).Where("id = ?", p.ID).Updates(map[string]any{
		"test_id": p.TestID, "skill": p.Skill, "part_no": p.PartNo, "source_ref": p.SourceRef, "title": p.Title, "topic": p.Topic,
		"q_from": p.QFrom, "q_to": p.QTo, "question_types": types, "content": content, "audio": audio, "status": p.Status,
		"content_hash": p.ContentHash, "updated_at": gorm.Expr("now()"),
	}).Error; err != nil {
		return err
	}
	return r.conn(ctx).Exec(`INSERT INTO exam_keys (part_id, key) VALUES (?, ?)
		ON CONFLICT (part_id) DO UPDATE SET key = EXCLUDED.key, updated_at = now()`, p.ID, keyJSON).Error
}

func (r *ExamRepository) ListParts(ctx context.Context, testIDs []string, skill string) ([]entity.Part, error) {
	if len(testIDs) == 0 {
		return nil, nil
	}
	var rows []models.ExamPart
	err := r.conn(ctx).Where("test_id IN ? AND skill = ? AND status = ?", testIDs, skill, entity.StatusPublished).
		Order("test_id, part_no").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return partEntities(rows)
}

func (r *ExamRepository) ListPartsByIDs(ctx context.Context, ids []string) ([]entity.Part, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []models.ExamPart
	if err := r.conn(ctx).Where("id IN ?", ids).Order("part_no").Find(&rows).Error; err != nil {
		return nil, err
	}
	return partEntities(rows)
}

func (r *ExamRepository) FindKeys(ctx context.Context, partIDs []string) (map[string]entity.Key, error) {
	out := map[string]entity.Key{}
	if len(partIDs) == 0 {
		return out, nil
	}
	var rows []models.ExamKey
	if err := r.conn(ctx).Where("part_id IN ?", partIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		var k entity.Key
		if err := json.Unmarshal(row.Key, &k); err != nil {
			return nil, fmt.Errorf("key of part %s: %w", row.PartID, err)
		}
		out[row.PartID] = k
	}
	return out, nil
}

// ---- attempts ---------------------------------------------------------------

func (r *ExamRepository) CreateAttempt(ctx context.Context, a *entity.Attempt) error {
	partIDs, _ := json.Marshal(a.PartIDs)
	questions, _ := json.Marshal(a.Questions)
	answers, _ := json.Marshal(a.Answers)
	results, _ := json.Marshal(a.Results)
	scores, _ := json.Marshal(a.PartScores)
	row := models.ExamAttempt{UserID: a.UserID, Skill: a.Skill, TestID: a.TestID, PartIDs: partIDs, Kind: a.Kind, Mode: a.Mode,
		TimeLimitSec: a.TimeLimitSec, DurationSec: a.DurationSec, Questions: questions, Answers: answers, Results: results,
		PartScores: scores, Correct: int16(a.Correct), Total: int16(a.Total), Band: a.Band}
	if a.ParentAttemptID != "" {
		row.ParentAttemptID = &a.ParentAttemptID
	}
	if err := r.conn(ctx).Create(&row).Error; err != nil {
		return err
	}
	a.ID, a.CreatedAt = row.ID, row.CreatedAt
	return nil
}

func attemptEntity(m *models.ExamAttempt, withDetail bool) (*entity.Attempt, error) {
	a := &entity.Attempt{ID: m.ID, UserID: m.UserID, Skill: m.Skill, TestID: m.TestID, Kind: m.Kind, Mode: m.Mode,
		TimeLimitSec: m.TimeLimitSec, DurationSec: m.DurationSec, Correct: int(m.Correct), Total: int(m.Total), Band: m.Band,
		CreatedAt: m.CreatedAt}
	if m.ParentAttemptID != nil {
		a.ParentAttemptID = *m.ParentAttemptID
	}
	if err := json.Unmarshal(m.PartIDs, &a.PartIDs); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(m.PartScores, &a.PartScores); err != nil {
		return nil, err
	}
	if len(m.Questions) > 0 {
		if err := json.Unmarshal(m.Questions, &a.Questions); err != nil {
			return nil, err
		}
	}
	if withDetail {
		if err := json.Unmarshal(m.Answers, &a.Answers); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(m.Results, &a.Results); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (r *ExamRepository) FindAttempt(ctx context.Context, userID, id string) (*entity.Attempt, error) {
	var rows []models.ExamAttempt
	if err := r.conn(ctx).Where("id = ? AND user_id = ?", id, userID).Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return nil, err
	}
	return attemptEntity(&rows[0], true)
}

func (r *ExamRepository) ListAttempts(ctx context.Context, userID, skill string, limit int, before string, withDetail bool) ([]entity.Attempt, error) {
	q := r.conn(ctx).Where("user_id = ? AND skill = ?", userID, skill)
	if !withDetail {
		q = q.Omit("answers", "results")
	}
	if before != "" {
		q = q.Where("(created_at, id) < (SELECT created_at, id FROM exam_attempts WHERE id = ? AND user_id = ?)", before, userID)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var rows []models.ExamAttempt
	if err := q.Order("created_at DESC, id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.Attempt, 0, len(rows))
	for i := range rows {
		a, err := attemptEntity(&rows[i], withDetail)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, nil
}
