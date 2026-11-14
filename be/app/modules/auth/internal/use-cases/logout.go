package usecases

import (
	"context"

	"ielts-practice/pkg/exceptions"
	"ielts-practice/pkg/tokens"
)

func (uc *AuthUseCases) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return uc.transaction.Execute(ctx, func(txCtx context.Context) error {
		current, err := uc.refreshRepo.FindByHashForUpdate(txCtx, tokens.HashToken(rawToken))
		if err != nil {
			return exceptions.ErrInternalServerError("Failed to load refresh token", err)
		}
		if current == nil || current.RevokedAt != nil {
			return nil
		}
		return uc.refreshRepo.Revoke(txCtx, current.ID, uc.now(), nil)
	})
}
