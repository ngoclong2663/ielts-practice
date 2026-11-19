package usecases

import (
	"context"
	"net/http"

	decksin "ielts-practice/app/modules/decks/ports/in"
	"ielts-practice/app/modules/library/dto"
	"ielts-practice/pkg/exceptions"
)

// AddWords copies reviewed library cards into a deck (Epic 8 §3.5): no generation job, no AI call and no
// daily limit. Words already in the deck are skipped. Ids that are not library cards are ignored.
func (uc *LibraryUseCases) AddWords(ctx context.Context, userID string, req dto.AddWordsRequestDto) (dto.AddWordsResponseDto, error) {
	ids := make([]string, 0, len(req.LexemeIDs))
	seen := map[string]bool{}
	for _, id := range req.LexemeIDs {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	found, err := uc.repo.FindCards(ctx, ids)
	if err != nil {
		return dto.AddWordsResponseDto{}, internalErr("Failed to load the cards", err)
	}
	cards := make([]decksin.GeneratedCard, 0, len(found))
	for _, id := range ids { // keep the learner's order
		content, ok := found[id]
		if !ok {
			continue
		}
		lexemeID := id
		cards = append(cards, decksin.GeneratedCard{LexemeID: &lexemeID, Content: content})
	}
	if len(cards) == 0 {
		return dto.AddWordsResponseDto{}, appErr(http.StatusBadRequest, exceptions.CodeVocabNothingToSave, "No library words selected")
	}
	res, err := uc.decks.SaveGeneratedCards(ctx, decksin.SaveGeneratedInput{
		UserID: userID, DeckID: req.DeckID, NewDeckName: req.NewDeckName, Cards: cards,
	})
	if err != nil {
		return dto.AddWordsResponseDto{}, err
	}
	skipped := res.Skipped
	if skipped == nil {
		skipped = []string{}
	}
	return dto.AddWordsResponseDto{DeckID: res.DeckID, DeckName: res.DeckName, Created: res.Created, Skipped: skipped}, nil
}
