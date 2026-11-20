package exam

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/exam/infrastructure/persistence"
	examInternal "ielts-practice/app/modules/exam/internal"
	usecases "ielts-practice/app/modules/exam/internal/use-cases"
	examin "ielts-practice/app/modules/exam/ports/in"
)

// NewExamModule: Epic 10 Reading and Listening practice (test bank, marking, history).
func NewExamModule() fx.Option {
	return fx.Options(
		NewExamSeedModule(),
		fx.Provide(
			usecases.NewExamUseCases,
			examInternal.NewExamPresenter,
			examInternal.NewExamRouter,
		),
		fx.Invoke(examInternal.RegisterRoutes),
	)
}

// NewExamSeedModule wires only what `api seed` needs.
func NewExamSeedModule() fx.Option {
	return fx.Provide(
		usecases.NewExamSeeder,
		func(s *usecases.ExamSeeder) examin.IExamSeeder { return s },
		persistence.NewExamRepository,
	)
}
