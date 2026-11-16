package admin

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/admin/infrastructure/persistence"
	adminInternal "ielts-practice/app/modules/admin/internal"
	usecases "ielts-practice/app/modules/admin/internal/use-cases"
	adminin "ielts-practice/app/modules/admin/ports/in"
)

// NewAdminModule: Epic 6 admin board: sign-in, overview and users reports (read-only).
func NewAdminModule() fx.Option {
	return fx.Options(
		NewAdminSeedModule(),
		fx.Provide(
			adminInternal.NewAdminRouter,
			adminInternal.NewAdminPresenter,
			usecases.NewAdminUseCases,
			func(uc *usecases.AdminUseCases) usecases.IAdminUseCases { return uc },
			// Other modules' admin routes check sessions through the port.
			func(uc *usecases.AdminUseCases) adminin.IAdminSession { return uc },
			persistence.NewAdminRepository,
		),
		fx.Invoke(adminInternal.RegisterRoutes),
	)
}

// NewAdminSeedModule wires only what `api seed` needs (no HTTP routes).
func NewAdminSeedModule() fx.Option {
	return fx.Provide(
		usecases.NewAdminSeeder,
		func(s *usecases.AdminSeeder) adminin.IAdminSeeder { return s },
		persistence.NewAdminAccountRepository,
	)
}
