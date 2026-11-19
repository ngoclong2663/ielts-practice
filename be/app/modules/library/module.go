package library

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/library/infrastructure/persistence"
	libraryInternal "ielts-practice/app/modules/library/internal"
	usecases "ielts-practice/app/modules/library/internal/use-cases"
	libraryin "ielts-practice/app/modules/library/ports/in"
)

// NewLibraryModule: Epic 8 §3 vocabulary library (browse, search, add reviewed cards to a deck).
func NewLibraryModule() fx.Option {
	return fx.Options(
		NewLibrarySeedModule(),
		fx.Provide(
			usecases.NewLibraryUseCases,
			func(uc *usecases.LibraryUseCases) libraryin.ITopicWords { return uc },
			libraryInternal.NewLibraryPresenter,
			libraryInternal.NewLibraryRouter,
		),
		fx.Invoke(libraryInternal.RegisterRoutes),
	)
}

// NewLibrarySeedModule wires only what `api seed` needs.
func NewLibrarySeedModule() fx.Option {
	return fx.Provide(
		usecases.NewLibrarySeeder,
		func(s *usecases.LibrarySeeder) libraryin.ILibrarySeeder { return s },
		func(s *usecases.LibrarySeeder) libraryin.ICardStore { return s },
		persistence.NewLibraryRepository,
	)
}
