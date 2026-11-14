package internal

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"ielts-practice/app/modules/auth/domain"
	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/app/modules/auth/dto"
	"ielts-practice/app/modules/auth/infrastructure"
	"ielts-practice/app/modules/database"
	userEntity "ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/app/modules/users/ports"
	"ielts-practice/configs"
	"ielts-practice/pkg/exceptions"
	"ielts-practice/pkg/tokens"
)

// refreshReuseGrace: a rotated refresh token presented again within this window is treated as a
// benign race (e.g. two tabs refreshing at once), not as token theft.
const refreshReuseGrace = 30 * time.Second

type AuthUseCase struct {
	transaction *database.TransactionStruct
	logger      zerolog.Logger
	users       ports.IUserService
	refreshRepo domain.IRefreshTokenRepository
	google      domain.IGoogleIdentityProvider
	jwt         *tokens.JWTManager
	cfg         *configs.Config
	mapper      *AuthMapper
	now         func() time.Time
}

func NewAuthUseCase(
	tx *database.TransactionStruct,
	logger zerolog.Logger,
	userService ports.IUserService,
	refreshRepo domain.IRefreshTokenRepository,
	google domain.IGoogleIdentityProvider,
	jwt *tokens.JWTManager,
	cfg *configs.Config,
	mapper *AuthMapper,
) *AuthUseCase {
	return &AuthUseCase{
		transaction: tx,
		logger:      logger,
		users:       userService,
		refreshRepo: refreshRepo,
		google:      google,
		jwt:         jwt,
		cfg:         cfg,
		mapper:      mapper,
		now:         time.Now,
	}
}

func unauthorized(code, message string, err error) *exceptions.AppError {
	return exceptions.NewException(http.StatusUnauthorized, code, message, err)
}

func (uc *AuthUseCase) SignInWithGoogle(ctx context.Context, req dto.GoogleSignInRequestDto, meta entity.ClientMeta) (dto.SessionResponseDto, error) {
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

// TestLogin creates a session without Google. Only available when APP_ENV=test.
func (uc *AuthUseCase) TestLogin(ctx context.Context, req dto.TestLoginRequestDto, meta entity.ClientMeta) (dto.SessionResponseDto, error) {
	if !uc.cfg.IsTest() {
		return dto.SessionResponseDto{}, exceptions.ErrNotFound("Not found", nil)
	}
	return uc.signIn(ctx, ports.GoogleUserInput{
		GoogleSub:   "test:" + req.Email,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		UILocale:    req.UILocale,
	}, meta)
}

func (uc *AuthUseCase) signIn(ctx context.Context, input ports.GoogleUserInput, meta entity.ClientMeta) (dto.SessionResponseDto, error) {
	var session dto.SessionResponseDto
	err := uc.transaction.Execute(ctx, func(txCtx context.Context) error {
		user, err := uc.users.UpsertGoogleUser(txCtx, input)
		if err != nil {
			return err
		}
		session, _, err = uc.issueSession(txCtx, user, meta)
		return err
	})
	return session, err
}

func (uc *AuthUseCase) Refresh(ctx context.Context, rawToken string, meta entity.ClientMeta) (dto.SessionResponseDto, error) {
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

func (uc *AuthUseCase) Logout(ctx context.Context, rawToken string) error {
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

// issueSession creates a new refresh token row and a signed access token.
// It returns the session and the ID of the new refresh token row.
func (uc *AuthUseCase) issueSession(ctx context.Context, user *userEntity.UserEntity, meta entity.ClientMeta) (dto.SessionResponseDto, string, error) {
	rawRefresh, err := tokens.NewOpaqueToken()
	if err != nil {
		return dto.SessionResponseDto{}, "", exceptions.ErrInternalServerError("Failed to create refresh token", err)
	}
	refreshExpiresAt := uc.now().Add(uc.cfg.Auth.RefreshTokenTTL)
	record := &entity.RefreshTokenEntity{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: tokens.HashToken(rawRefresh),
		ExpiresAt: refreshExpiresAt,
		UserAgent: truncate(meta.UserAgent, 512),
		IP:        truncate(meta.IP, 64),
	}
	if err := uc.refreshRepo.Create(ctx, record); err != nil {
		return dto.SessionResponseDto{}, "", exceptions.ErrInternalServerError("Failed to store refresh token", err)
	}

	access, accessExpiresAt, err := uc.jwt.Issue(user.ID)
	if err != nil {
		return dto.SessionResponseDto{}, "", exceptions.ErrInternalServerError("Failed to sign access token", err)
	}

	return dto.SessionResponseDto{
		AccessToken:           access,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          rawRefresh,
		RefreshTokenExpiresAt: refreshExpiresAt,
		User:                  uc.mapper.MapUserToSessionUser(user),
	}, record.ID, nil
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}
