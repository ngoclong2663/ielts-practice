package usecases

import (
	"context"
	"net/http"
	"strings"

	"ielts-practice/app/modules/decks/domain/entity"
	decksin "ielts-practice/app/modules/decks/ports/in"
	"ielts-practice/pkg/exceptions"
)

var _ decksin.IDeckService = (*DeckUseCases)(nil)

func (uc *DeckUseCases) DeckWords(ctx context.Context, userID, deckID string) ([]string, error) {
	if _, err := uc.mustOwnDeck(ctx, userID, deckID); err != nil {
		return nil, err
	}
	words, err := uc.cards.Words(ctx, deckID)
	if err != nil {
		return nil, internalErr("Failed to load deck words", err)
	}
	return words, nil
}

func (uc *DeckUseCases) OwnsDeck(ctx context.Context, userID, deckID string) (bool, error) {
	deck, err := uc.decks.FindByID(ctx, userID, deckID)
	if err != nil {
		return false, internalErr("Failed to load deck", err)
	}
	return deck != nil, nil
}

func (uc *DeckUseCases) SaveGeneratedCards(ctx context.Context, in decksin.SaveGeneratedInput) (decksin.SaveGeneratedResult, error) {
	var out decksin.SaveGeneratedResult
	err := uc.transaction.Execute(ctx, func(txCtx context.Context) error {
		deck, err := uc.resolveTargetDeck(txCtx, in)
		if err != nil {
			return err
		}
		out.DeckID, out.DeckName = deck.ID, deck.Name

		position, err := uc.cards.NextPosition(txCtx, deck.ID)
		if err != nil {
			return internalErr("Failed to compute position", err)
		}
		now := uc.now()
		for _, c := range in.Cards {
			headword := strings.TrimSpace(c.Content.Headword)
			if headword == "" {
				continue
			}
			exists, err := uc.cards.HeadwordExists(txCtx, deck.ID, headword, "")
			if err == nil && !exists && c.LexemeID != nil {
				exists, err = uc.cards.LexemeExists(txCtx, deck.ID, *c.LexemeID)
			}
			if err != nil {
				return internalErr("Failed to check duplicates", err)
			}
			if exists {
				out.Skipped = append(out.Skipped, headword)
				continue
			}
			if _, err := uc.cards.Create(txCtx, &entity.CardEntity{
				UserID:   in.UserID,
				DeckID:   deck.ID,
				LexemeID: c.LexemeID,
				Headword: headword,
				Content:  c.Content,
				Due:      now,
				Position: position,
			}); err != nil {
				return internalErr("Failed to create card", err)
			}
			position++
			out.Created++
		}
		return nil
	})
	return out, err
}

func (uc *DeckUseCases) resolveTargetDeck(ctx context.Context, in decksin.SaveGeneratedInput) (*entity.DeckEntity, error) {
	if in.DeckID != "" {
		return uc.mustOwnDeck(ctx, in.UserID, in.DeckID)
	}
	name := strings.TrimSpace(in.NewDeckName)
	if name == "" {
		e := exceptions.NewException(http.StatusBadRequest, exceptions.CodeValidationFailed, "Choose a deck", nil)
		e.Fields = map[string]string{"deck_id": "VALIDATION_REQUIRED"}
		return nil, e
	}
	taken, err := uc.decks.NameTaken(ctx, in.UserID, name, "")
	if err != nil {
		return nil, internalErr("Failed to check deck name", err)
	}
	if taken {
		e := nameTaken()
		e.Fields = map[string]string{"new_deck_name": exceptions.CodeDeckNameTaken}
		return nil, e
	}
	deck, err := uc.decks.Create(ctx, &entity.DeckEntity{UserID: in.UserID, Name: name, Source: "paste"})
	if err != nil {
		return nil, internalErr("Failed to create deck", err)
	}
	return deck, nil
}
