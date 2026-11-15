package usecases

import (
	"context"
	"net/http"
	"time"

	authout "ielts-practice/app/modules/auth/ports/out"

	userEntity "ielts-practice/app/modules/users/domain/entity"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"ielts-practice/app/modules/auth/domain"
	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/app/modules/auth/dto"
	"ielts-practice/app/modules/auth/internal/mapper"
	"ielts-practice/app/modules/database"
	usersin "ielts-practice/app/modules/users/ports/in"
	"ielts-practice/configs"
	"ielts-practice/pkg/exceptions"
	"ielts-practice/pkg/tokens"
)

// refreshReuseGrace: a rotated refresh token presented again within this window is treated as a
// benign race (e.g. two tabs refreshing at once), not as token theft.
const refreshReuseGrace = 30 * time.Second

type AuthUseCases struct {
	transaction *database.TransactionStruct
	logger      zerolog.Logger
	users       usersin.IUserService
	refreshRepo domain.IRefreshTokenRepository
	google      authout.IGoogleIdentityProvider
	jwt         *tokens.JWTManager
	cfg         *configs.Config
	mapper      *mapper.AuthMapper
	now         func() time.Time
}

func NewAuthUseCases(
	tx *database.TransactionStruct,
	logger zerolog.Logger,
	userService usersin.IUserService,
	refreshRepo domain.IRefreshTokenRepository,
	google authout.IGoogleIdentityProvider,
	jwt *tokens.JWTManager,
	cfg *configs.Config,
	mapper *mapper.AuthMapper,
) *AuthUseCases {
	return &AuthUseCases{
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

func (uc *AuthUseCases) signIn(ctx context.Context, input usersin.GoogleUserInput, meta entity.ClientMeta) (dto.SessionResponseDto, error) {
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

// issueSession creates a new refresh token row and a signed access token.
// It returns the session and the ID of the new refresh token row.
func (uc *AuthUseCases) issueSession(ctx context.Context, user *userEntity.UserEntity, meta entity.ClientMeta) (dto.SessionResponseDto, string, error) {
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
