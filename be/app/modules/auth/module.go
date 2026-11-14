package auth

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/auth/infrastructure"
	authInternal "ielts-practice/app/modules/auth/internal"
	"ielts-practice/app/modules/auth/internal/mapper"
	usecases "ielts-practice/app/modules/auth/internal/use-cases"
)

func NewAuthModule() fx.Option {
	return fx.Options(
		fx.Provide(
			authInternal.NewAuthRouter,
			authInternal.NewAuthPresenter,
			usecases.NewAuthUseCases,
			mapper.NewAuthMapper,
			infrastructure.NewRefreshTokenRepository,
			infrastructure.NewGoogleIdentityProvider,
		),
		fx.Invoke(
			authInternal.RegisterRoutes,
		),
	)
}
