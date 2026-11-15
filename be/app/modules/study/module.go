package study

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/study/infrastructure/persistence"
	studyInternal "ielts-practice/app/modules/study/internal"
	usecases "ielts-practice/app/modules/study/internal/use-cases"
)

func NewStudyModule() fx.Option {
	return fx.Options(
		fx.Provide(
			studyInternal.NewStudyRouter,
			studyInternal.NewStudyPresenter,
			usecases.NewStudyUseCases,
			func(uc *usecases.StudyUseCases) usecases.IStudyUseCases { return uc },
			persistence.NewStudyRepository,
		),
		fx.Invoke(studyInternal.RegisterRoutes),
	)
}
