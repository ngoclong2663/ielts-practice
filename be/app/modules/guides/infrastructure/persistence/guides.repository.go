// Package persistence is the PostgreSQL (GORM) adapter of the guides module.
package persistence

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/guides/domain"
	"ielts-practice/app/modules/guides/domain/entity"
)

var _ domain.IGuidesRepository = &GuidesRepository{}

type GuidesRepository struct {
	db *gorm.DB
}

func NewGuidesRepository(db *gorm.DB) domain.IGuidesRepository {
	return &GuidesRepository{db: db}
}

func (r *GuidesRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

func (r *GuidesRepository) ListPublished(ctx context.Context, topic string) ([]entity.GuideVideo, error) {
	q := r.conn(ctx).Where("status = ?", entity.StatusPublished)
	if topic != "" {
		q = q.Where("topic = ?", topic)
	}
	var rows []models.GuideVideo
	if err := q.Order("topic, sort, slug").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.GuideVideo, 0, len(rows))
	for i := range rows {
		g, err := toEntity(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *g)
	}
	return out, nil
}

func (r *GuidesRepository) FindBySlug(ctx context.Context, slug string) (*entity.GuideVideo, error) {
	var rows []models.GuideVideo
	if err := r.conn(ctx).Where("slug = ?", slug).Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return nil, err
	}
	return toEntity(&rows[0])
}

func (r *GuidesRepository) Upsert(ctx context.Context, g *entity.GuideVideo) error {
	title, _ := json.Marshal(g.Title)
	desc, _ := json.Marshal(g.Description)
	row := models.GuideVideo{
		Slug: g.Slug, Topic: g.Topic, Sort: int16(g.Sort), Lang: g.Lang, Title: title, Description: desc,
		VideoKey: g.VideoKey, PosterKey: g.PosterKey, DurationSeconds: g.DurationSeconds, Status: g.Status, ContentHash: g.ContentHash,
	}
	existing, err := r.FindBySlug(ctx, g.Slug)
	if err != nil {
		return err
	}
	if existing == nil {
		if err := r.conn(ctx).Create(&row).Error; err != nil {
			return err
		}
		g.ID = row.ID
		return nil
	}
	g.ID = existing.ID
	return r.conn(ctx).Model(&models.GuideVideo{}).Where("id = ?", existing.ID).Updates(map[string]any{
		"topic": row.Topic, "sort": row.Sort, "lang": row.Lang, "title": row.Title, "description": row.Description,
		"video_key": row.VideoKey, "poster_key": row.PosterKey, "duration_seconds": row.DurationSeconds,
		"status": row.Status, "content_hash": row.ContentHash, "updated_at": gorm.Expr("now()"),
	}).Error
}

func (r *GuidesRepository) HasMedia(ctx context.Context, key string) (bool, error) {
	var n int64
	err := r.conn(ctx).Model(&models.GuideVideo{}).
		Where("status = ? AND (video_key = ? OR poster_key = ?)", entity.StatusPublished, key, key).Count(&n).Error
	return n > 0, err
}

func toEntity(m *models.GuideVideo) (*entity.GuideVideo, error) {
	g := &entity.GuideVideo{
		ID: m.ID, Slug: m.Slug, Topic: m.Topic, Sort: int(m.Sort), Lang: m.Lang, VideoKey: m.VideoKey, PosterKey: m.PosterKey,
		DurationSeconds: m.DurationSeconds, Status: m.Status, ContentHash: m.ContentHash,
	}
	if err := json.Unmarshal(m.Title, &g.Title); err != nil {
		return nil, err
	}
	return g, json.Unmarshal(m.Description, &g.Description)
}
