package usecases

import (
	"context"
	"strings"

	"ielts-practice/app/modules/decks/domain/entity"
	"ielts-practice/app/modules/decks/dto"
)

func (uc *DeckUseCases) ListCards(ctx context.Context, userID, deckID string, q dto.ListCardsQueryDto) (dto.CardListDto, error) {
	if _, err := uc.mustOwnDeck(ctx, userID, deckID); err != nil {
		return dto.CardListDto{}, err
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}
	cards, total, err := uc.cards.List(ctx, userID, entity.CardFilter{
		DeckID: deckID, Query: strings.TrimSpace(q.Q), State: q.State, Page: q.Page, PageSize: q.PageSize,
	})
	if err != nil {
		return dto.CardListDto{}, internalErr("Failed to list cards", err)
	}
	out := dto.CardListDto{Items: make([]dto.CardSummaryDto, 0, len(cards)), Page: q.Page, PageSize: q.PageSize, Total: total}
	for i := range cards {
		out.Items = append(out.Items, uc.mapper.CardEntityToSummaryDto(&cards[i]))
	}
	return out, nil
}
