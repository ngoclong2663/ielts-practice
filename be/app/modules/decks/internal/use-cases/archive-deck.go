package usecases

import "context"

func (uc *DeckUseCases) ArchiveDeck(ctx context.Context, userID, deckID string) error {
	if err := uc.decks.Archive(ctx, userID, deckID, uc.now()); err != nil {
		if isNotFound(err) {
			return deckNotFound()
		}
		return internalErr("Failed to delete deck", err)
	}
	return nil
}
