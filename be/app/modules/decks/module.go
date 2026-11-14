package decks

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/decks/infrastructure"
	decksInternal "ielts-practice/app/modules/decks/internal"
	"ielts-practice/app/modules/decks/internal/mapper"
	usecases "ielts-practice/app/modules/decks/internal/use-cases"
)

func NewDecksModule() fx.Option {
	return fx.Options(
		fx.Provide(
			decksInternal.NewDeckRouter,
			decksInternal.NewDeckPresenter,
			usecases.NewDeckUseCases,
			mapper.NewDeckMapper,
			infrastructure.NewDeckRepository,
			infrastructure.NewCardRepository,
		),
		fx.Invoke(decksInternal.RegisterRoutes),
	)
}
