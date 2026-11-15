package writing

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/writing/infrastructure/coach"
	"ielts-practice/app/modules/writing/infrastructure/persistence"
	writingInternal "ielts-practice/app/modules/writing/internal"
	usecases "ielts-practice/app/modules/writing/internal/use-cases"
	writingin "ielts-practice/app/modules/writing/ports/in"
)

// NewWritingModule: Epic 5 guided writing practice (task bank, attempts, AI coach, figure links).
func NewWritingModule() fx.Option {
	return fx.Options(
		NewWritingSeedModule(),
		fx.Provide(
			usecases.NewWritingUseCases,
			writingInternal.NewWritingPresenter,
			writingInternal.NewWritingRouter,
			coach.NewCoach,
		),
		fx.Invoke(writingInternal.RegisterRoutes),
	)
}

// NewWritingSeedModule wires only what `api seed` needs (no HTTP routes).
func NewWritingSeedModule() fx.Option {
	return fx.Provide(
		usecases.NewTaskSeeder,
		// `api seed` loads the task bank through the port.
		func(s *usecases.TaskSeeder) writingin.ITaskSeeder { return s },
		persistence.NewWritingRepository,
	)
}
