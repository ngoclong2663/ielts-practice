package users

import (
	"go.uber.org/fx"
	"ielts-practice/app/modules/users/infrastructure"
	usersInternal "ielts-practice/app/modules/users/internal"
)

func NewUserModule() fx.Option {
	return fx.Options(
		fx.Provide(
			usersInternal.NewUserRouter,
			usersInternal.NewUserPresenter,
			usersInternal.NewUserUseCase,
			usersInternal.NewUserService,
			infrastructure.NewUserRepository,
			usersInternal.NewUserMapper,
		),
		fx.Invoke(
			usersInternal.RegisterRoutes,
		),
	)
}
