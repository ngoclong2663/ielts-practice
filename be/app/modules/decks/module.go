package decks

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/decks/infrastructure"
	decksInternal "ielts-practice/app/modules/decks/internal"
	"ielts-practice/app/modules/decks/internal/mapper"
	usecases "ielts-practice/app/modules/decks/internal/use-cases"
	"ielts-practice/app/modules/decks/ports"
)

func NewDecksModule() fx.Option {
	return fx.Options(
		fx.Provide(
			decksInternal.NewDeckRouter,
			decksInternal.NewDeckPresenter,
			usecases.NewDeckUseCases,
			// Other modules (vocab) save cards through the port.
			func(uc *usecases.DeckUseCases) ports.IDeckService { return uc },
			mapper.NewDeckMapper,
			infrastructure.NewDeckRepository,
			infrastructure.NewCardRepository,
		),
		fx.Invoke(decksInternal.RegisterRoutes),
	)
}
