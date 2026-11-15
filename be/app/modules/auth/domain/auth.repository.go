package domain

import (
	"context"
	"time"

	"ielts-practice/app/modules/auth/domain/entity"
)

type IRefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshTokenEntity) error
	// FindByHashForUpdate locks the row (SELECT ... FOR UPDATE); call it inside a transaction.
	// Returns (nil, nil) when no token matches.
	FindByHashForUpdate(ctx context.Context, tokenHash string) (*entity.RefreshTokenEntity, error)
	Revoke(ctx context.Context, id string, revokedAt time.Time, replacedBy *string) error
	RevokeAllForUser(ctx context.Context, userID string, revokedAt time.Time) error
}
