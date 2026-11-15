package decks

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/decks/infrastructure/persistence"
	decksInternal "ielts-practice/app/modules/decks/internal"
	"ielts-practice/app/modules/decks/internal/mapper"
	usecases "ielts-practice/app/modules/decks/internal/use-cases"
	decksin "ielts-practice/app/modules/decks/ports/in"
)

func NewDecksModule() fx.Option {
	return fx.Options(
		fx.Provide(
			decksInternal.NewDeckRouter,
			decksInternal.NewDeckPresenter,
			usecases.NewDeckUseCases,
			// Other modules (vocab) save cards through the port.
			func(uc *usecases.DeckUseCases) decksin.IDeckService { return uc },
			mapper.NewDeckMapper,
			persistence.NewDeckRepository,
			persistence.NewCardRepository,
		),
		fx.Invoke(decksInternal.RegisterRoutes),
	)
}
