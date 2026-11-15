package users

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/users/infrastructure/persistence"
	usersInternal "ielts-practice/app/modules/users/internal"
	"ielts-practice/app/modules/users/internal/mapper"
	usecases "ielts-practice/app/modules/users/internal/use-cases"
	usersin "ielts-practice/app/modules/users/ports/in"
)

func NewUserModule() fx.Option {
	return fx.Options(
		fx.Provide(
			usersInternal.NewUserRouter,
			usersInternal.NewUserPresenter,
			usecases.NewUserUseCases,
			// One set of use cases, two views: HTTP presenters and other modules (via the port).
			func(uc *usecases.UserUseCases) usecases.IUserUseCases { return uc },
			func(uc *usecases.UserUseCases) usersin.IUserService { return uc },
			persistence.NewUserRepository,
			mapper.NewUserMapper,
		),
		fx.Invoke(
			usersInternal.RegisterRoutes,
		),
	)
}
