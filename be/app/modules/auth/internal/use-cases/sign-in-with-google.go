package usecases

import (
	"context"
	"errors"
	"net/http"

	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/app/modules/auth/dto"
	"ielts-practice/app/modules/auth/infrastructure"
	"ielts-practice/pkg/exceptions"
)

func (uc *AuthUseCases) SignInWithGoogle(ctx context.Context, req dto.GoogleSignInRequestDto, meta entity.ClientMeta) (dto.SessionResponseDto, error) {
	identity, err := uc.google.Exchange(ctx, req.Code, req.CodeVerifier)
	if err != nil {
		switch {
		case errors.Is(err, infrastructure.ErrGoogleNotConfigured):
			return dto.SessionResponseDto{}, exceptions.NewException(http.StatusServiceUnavailable, exceptions.CodeGoogleNotConfigured, "Google sign-in is not configured", err)
		case errors.Is(err, infrastructure.ErrGoogleIDToken):
			return dto.SessionResponseDto{}, unauthorized(exceptions.CodeGoogleInvalidToken, "Invalid Google ID token", err)
		default:
			uc.logger.Warn().Err(err).Msg("google code exchange failed")
			return dto.SessionResponseDto{}, unauthorized(exceptions.CodeGoogleExchangeFailed, "Google sign-in failed", err)
		}
	}

	if identity.Nonce == "" || identity.Nonce != req.Nonce {
		return dto.SessionResponseDto{}, unauthorized(exceptions.CodeGoogleInvalidToken, "Nonce mismatch", nil)
	}
	if !identity.EmailVerified || identity.Email == "" {
		return dto.SessionResponseDto{}, unauthorized(exceptions.CodeGoogleEmailUnverified, "Google email is not verified", nil)
	}

	return uc.signIn(ctx, uc.mapper.MapGoogleIdentityToUserInput(identity, req), meta)
}
