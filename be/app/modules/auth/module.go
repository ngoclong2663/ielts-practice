package auth

import (
	"go.uber.org/fx"
	"ielts-practice/app/modules/auth/infrastructure"
	authInternal "ielts-practice/app/modules/auth/internal"
)

func NewAuthModule() fx.Option {
	return fx.Options(
		fx.Provide(
			authInternal.NewAuthRouter,
			authInternal.NewAuthPresenter,
			authInternal.NewAuthUseCase,
			authInternal.NewAuthMapper,
			infrastructure.NewRefreshTokenRepository,
			infrastructure.NewGoogleIdentityProvider,
		),
		fx.Invoke(
			authInternal.RegisterRoutes,
		),
	)
}
