package study

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/study/infrastructure"
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
			infrastructure.NewStudyRepository,
		),
		fx.Invoke(studyInternal.RegisterRoutes),
	)
}
