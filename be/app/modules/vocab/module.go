package vocab

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/vocab/infrastructure"
	vocabInternal "ielts-practice/app/modules/vocab/internal"
	"ielts-practice/app/modules/vocab/internal/mapper"
	usecases "ielts-practice/app/modules/vocab/internal/use-cases"
)

// NewVocabModule: Epic 3 smart card generation.
func NewVocabModule() fx.Option {
	return fx.Options(
		fx.Provide(
			vocabInternal.NewVocabRouter,
			vocabInternal.NewVocabPresenter,
			vocabInternal.NewWorker,
			usecases.NewVocabUseCases,
			mapper.NewVocabMapper,
			infrastructure.NewVocabRepository,
			infrastructure.NewContentGenerator,
			infrastructure.NewWordSuggester,
			infrastructure.NewImageProviders,
		),
		fx.Invoke(vocabInternal.RegisterRoutes, vocabInternal.RegisterWorker),
	)
}
