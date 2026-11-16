package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/feedback/domain"
	"ielts-practice/app/modules/feedback/domain/entity"
)

var _ domain.IFeedbackRepository = &FeedbackRepository{}

type FeedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) domain.IFeedbackRepository {
	return &FeedbackRepository{db: db}
}

func (r *FeedbackRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

func toEntity(m *models.Feedback) entity.Feedback {
	return entity.Feedback{
		ID: m.ID, UserID: m.UserID, Category: m.Category, Message: m.Message, Page: m.Page,
		UILocale: m.UILocale, UserAgent: m.UserAgent, Status: m.Status, AdminNote: m.AdminNote,
		NotifiedAt: m.NotifiedAt, ResolvedAt: m.ResolvedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *FeedbackRepository) Create(ctx context.Context, fb *entity.Feedback) error {
	m := models.Feedback{
		UserID: fb.UserID, Category: fb.Category, Message: fb.Message, Page: fb.Page,
		UILocale: fb.UILocale, UserAgent: fb.UserAgent, Status: entity.StatusNew,
	}
	if err := r.conn(ctx).Create(&m).Error; err != nil {
		return err
	}
	*fb = toEntity(&m)
	return nil
}

func (r *FeedbackRepository) CountSince(ctx context.Context, userID string, since time.Time) (int, error) {
	var n int64
	err := r.conn(ctx).Model(&models.Feedback{}).Where("user_id = ? AND created_at > ?", userID, since).Count(&n).Error
	return int(n), err
}

func (r *FeedbackRepository) MarkNotified(ctx context.Context, id string, at time.Time) error {
	// updated_at is left alone: it tracks the owner's changes.
	return r.conn(ctx).Model(&models.Feedback{}).Where("id = ?", id).UpdateColumn("notified_at", at).Error
}

type feedbackRow struct {
	models.Feedback
	UserEmail       string
	UserDisplayName string
	UserAvatarURL   string
}

func (r *FeedbackRepository) withUser(ctx context.Context) *gorm.DB {
	return r.conn(ctx).Table("feedback f").
		Select(`f.*, COALESCE(u.email, '') AS user_email, COALESCE(u.display_name, '') AS user_display_name,
			COALESCE(u.avatar_url, '') AS user_avatar_url`).
		Joins("LEFT JOIN users u ON u.id = f.user_id")
}

func rowToEntity(row *feedbackRow) entity.Feedback {
	e := toEntity(&row.Feedback)
	e.User = entity.FeedbackUser{ID: row.UserID, Email: row.UserEmail, DisplayName: row.UserDisplayName, AvatarURL: row.UserAvatarURL}
	return e
}

func (r *FeedbackRepository) List(ctx context.Context, f entity.FeedbackFilter) ([]entity.Feedback, int64, error) {
	scope := func(q *gorm.DB) *gorm.DB {
		if f.Status != "" {
			q = q.Where("f.status = ?", f.Status)
		}
		if f.Category != "" {
			q = q.Where("f.category = ?", f.Category)
		}
		return q
	}
	var total int64
	if err := scope(r.conn(ctx).Table("feedback f")).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []feedbackRow
	err := scope(r.withUser(ctx)).Order("f.created_at DESC, f.id DESC").
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]entity.Feedback, len(rows))
	for i := range rows {
		out[i] = rowToEntity(&rows[i])
	}
	return out, total, nil
}

func (r *FeedbackRepository) CountByStatus(ctx context.Context, category string) (entity.StatusCounts, error) {
	var out entity.StatusCounts
	q := r.conn(ctx).Table("feedback").Select(`
		count(*) FILTER (WHERE status = 'new') AS "new",
		count(*) FILTER (WHERE status = 'seen') AS seen,
		count(*) FILTER (WHERE status = 'resolved') AS resolved`)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	return out, q.Scan(&out).Error
}

func (r *FeedbackRepository) FindByID(ctx context.Context, id string) (*entity.Feedback, error) {
	var rows []feedbackRow
	if err := r.withUser(ctx).Where("f.id = ?", id).Limit(1).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	e := rowToEntity(&rows[0])
	return &e, nil
}

func (r *FeedbackRepository) UpdateReview(ctx context.Context, id, status, adminNote string, resolvedAt *time.Time) error {
	return r.conn(ctx).Model(&models.Feedback{}).Where("id = ?", id).Updates(map[string]any{
		"status": status, "admin_note": adminNote, "resolved_at": resolvedAt, "updated_at": time.Now(),
	}).Error
}
