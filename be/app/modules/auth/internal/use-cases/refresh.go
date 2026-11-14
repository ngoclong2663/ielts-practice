package usecases

import (
	"context"

	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/app/modules/auth/dto"
	"ielts-practice/pkg/exceptions"
	"ielts-practice/pkg/tokens"
)

func (uc *AuthUseCases) Refresh(ctx context.Context, rawToken string, meta entity.ClientMeta) (dto.SessionResponseDto, error) {
	var (
		session  dto.SessionResponseDto
		failCode string
	)
	now := uc.now()

	err := uc.transaction.Execute(ctx, func(txCtx context.Context) error {
		current, err := uc.refreshRepo.FindByHashForUpdate(txCtx, tokens.HashToken(rawToken))
		if err != nil {
			return exceptions.ErrInternalServerError("Failed to load refresh token", err)
		}
		if current == nil || now.After(current.ExpiresAt) {
			failCode = exceptions.CodeRefreshInvalid
			return nil
		}
		if current.RevokedAt != nil {
			if now.Sub(*current.RevokedAt) <= refreshReuseGrace {
				failCode = exceptions.CodeRefreshInvalid
				return nil
			}
			// A rotated token was replayed: assume theft and end every session of this user.
			uc.logger.Warn().Str("user_id", current.UserID).Msg("refresh token reuse detected; revoking all sessions")
			if err := uc.refreshRepo.RevokeAllForUser(txCtx, current.UserID, now); err != nil {
				return exceptions.ErrInternalServerError("Failed to revoke sessions", err)
			}
			failCode = exceptions.CodeRefreshReused
			return nil // commit the revocation
		}

		user, err := uc.users.GetUserByID(txCtx, current.UserID)
		if err != nil {
			return err
		}
		var newID string
		session, newID, err = uc.issueSession(txCtx, user, meta)
		if err != nil {
			return err
		}
		return uc.refreshRepo.Revoke(txCtx, current.ID, now, &newID)
	})
	if err != nil {
		return dto.SessionResponseDto{}, err
	}
	if failCode != "" {
		return dto.SessionResponseDto{}, unauthorized(failCode, "Refresh token is invalid", nil)
	}
	return session, nil
}
