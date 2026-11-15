package usecases

import (
	"context"

	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/app/modules/auth/dto"
	usersin "ielts-practice/app/modules/users/ports/in"
	"ielts-practice/pkg/exceptions"
)

// TestLogin creates a session without Google. Only available when APP_ENV=test.
func (uc *AuthUseCases) TestLogin(ctx context.Context, req dto.TestLoginRequestDto, meta entity.ClientMeta) (dto.SessionResponseDto, error) {
	if !uc.cfg.IsTest() {
		return dto.SessionResponseDto{}, exceptions.ErrNotFound("Not found", nil)
	}
	return uc.signIn(ctx, usersin.GoogleUserInput{
		GoogleSub:   "test:" + req.Email,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		UILocale:    req.UILocale,
	}, meta)
}
