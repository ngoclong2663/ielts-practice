package writing

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/writing/infrastructure"
	writingInternal "ielts-practice/app/modules/writing/internal"
	usecases "ielts-practice/app/modules/writing/internal/use-cases"
	"ielts-practice/app/modules/writing/ports"
)

// NewWritingModule: Epic 5 guided writing practice (task bank and figure links for now).
func NewWritingModule() fx.Option {
	return fx.Options(
		NewWritingSeedModule(),
		fx.Provide(writingInternal.NewWritingPresenter),
		fx.Invoke(writingInternal.RegisterRoutes),
	)
}

// NewWritingSeedModule wires only what `api seed` needs (no HTTP routes).
func NewWritingSeedModule() fx.Option {
	return fx.Provide(
		usecases.NewWritingUseCases,
		// `api seed` loads the task bank through the port.
		func(uc *usecases.WritingUseCases) ports.ITaskSeeder { return uc },
		infrastructure.NewWritingRepository,
	)
}
