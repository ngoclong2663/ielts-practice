package guides

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/guides/infrastructure/persistence"
	guidesInternal "ielts-practice/app/modules/guides/internal"
	usecases "ielts-practice/app/modules/guides/internal/use-cases"
	guidesin "ielts-practice/app/modules/guides/ports/in"
)

// NewGuidesModule: Epic 7 §5 guide videos (list and media links).
func NewGuidesModule() fx.Option {
	return fx.Options(
		NewGuidesSeedModule(),
		fx.Provide(
			usecases.NewGuidesUseCases,
			guidesInternal.NewGuidesPresenter,
			guidesInternal.NewGuidesRouter,
		),
		fx.Invoke(guidesInternal.RegisterRoutes),
	)
}

// NewGuidesSeedModule wires only what `api seed` needs.
func NewGuidesSeedModule() fx.Option {
	return fx.Provide(
		usecases.NewGuideSeeder,
		func(s *usecases.GuideSeeder) guidesin.IGuideSeeder { return s },
		persistence.NewGuidesRepository,
	)
}
