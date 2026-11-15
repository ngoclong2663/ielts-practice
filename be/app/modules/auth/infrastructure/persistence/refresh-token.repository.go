package persistence

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/auth/domain"
	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/app/modules/database"
)

var _ domain.IRefreshTokenRepository = &RefreshTokenRepository{}

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) domain.IRefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, t *entity.RefreshTokenEntity) error {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx).Create(&models.RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt,
		UserAgent: t.UserAgent,
		IP:        t.IP,
	}).Error
}

func (r *RefreshTokenRepository) FindByHashForUpdate(ctx context.Context, tokenHash string) (*entity.RefreshTokenEntity, error) {
	var m models.RefreshToken
	err := database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&m, "token_hash = ?", tokenHash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity.RefreshTokenEntity{
		ID:         m.ID,
		UserID:     m.UserID,
		TokenHash:  m.TokenHash,
		ExpiresAt:  m.ExpiresAt,
		RevokedAt:  m.RevokedAt,
		ReplacedBy: m.ReplacedBy,
		UserAgent:  m.UserAgent,
		IP:         m.IP,
		CreatedAt:  m.CreatedAt,
	}, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string, revokedAt time.Time, replacedBy *string) error {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Updates(map[string]any{"revoked_at": revokedAt, "replaced_by": replacedBy}).Error
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, revokedAt time.Time) error {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).Error
}
