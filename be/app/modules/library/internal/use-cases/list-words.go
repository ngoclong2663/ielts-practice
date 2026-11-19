package usecases

import (
	"context"

	"ielts-practice/app/modules/library/domain/entity"
	"ielts-practice/app/modules/library/domain/rules"
	"ielts-practice/app/modules/library/dto"
)

// ListWords returns a collection's words for a topic, a level or a search, with the learner's flags
// (already in a deck, marked as known).
func (uc *LibraryUseCases) ListWords(ctx context.Context, userID, slug string, q dto.ListWordsQueryDto) (dto.WordsResponseDto, error) {
	col, err := uc.findCollection(ctx, slug)
	if err != nil {
		return dto.WordsResponseDto{}, err
	}
	query := rules.Fold(q.Q)
	if r := []rune(query); len(r) > rules.MaxQueryRunes {
		query = string(r[:rules.MaxQueryRunes])
	}
	filter := entity.WordFilter{TopicID: q.Topic, Level: q.Level, Query: query, Limit: rules.MaxWordsResult + 1}
	words, err := uc.repo.ListWords(ctx, userID, col.ID, filter)
	if err != nil {
		return dto.WordsResponseDto{}, internalErr("Failed to load the library words", err)
	}
	out := dto.WordsResponseDto{Collection: toCollectionDto(col), Words: make([]dto.WordDto, 0, len(words))}
	if len(words) > rules.MaxWordsResult {
		words, out.Truncated = words[:rules.MaxWordsResult], true
	}
	for _, w := range words {
		out.Words = append(out.Words, dto.WordDto{
			ID: w.ID, Topic: w.TopicID, LexemeID: w.LexemeID, CEFR: w.CEFR, Source: w.Source,
			Card: w.Card, InDeck: w.InDeck, Known: w.Known,
		})
	}
	return out, nil
}

// findCollection loads a published collection with its topics.
func (uc *LibraryUseCases) findCollection(ctx context.Context, slug string) (*entity.Collection, error) {
	cols, err := uc.repo.ListCollections(ctx)
	if err != nil {
		return nil, internalErr("Failed to load the library", err)
	}
	for i := range cols {
		if cols[i].Slug == slug {
			return &cols[i], nil
		}
	}
	return nil, notFound()
}
