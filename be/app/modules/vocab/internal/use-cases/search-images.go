package usecases

import (
	"context"

	"ielts-practice/app/modules/vocab/dto"
)

// SearchImages returns more candidates (optionally a custom keyword or one provider) for the picker.
// Results are stored on the item so the user can pick one of them.
func (uc *VocabUseCases) SearchImages(ctx context.Context, userID, jobID, itemID string, q dto.ImageSearchQueryDto) (dto.ImageSearchResponseDto, error) {
	item, err := uc.repo.FindItem(ctx, userID, jobID, itemID)
	if err != nil {
		return dto.ImageSearchResponseDto{}, internalErr("Failed to load word", err)
	}
	if item == nil {
		return dto.ImageSearchResponseDto{}, itemNotFound()
	}
	headword := item.InputText
	if item.Content != nil && item.Content.Headword != "" {
		headword = item.Content.Headword
	}
	query := q.Q
	if query == "" && item.LexemeID != nil {
		if lex, err := uc.repo.FindLexemeByID(ctx, *item.LexemeID); err == nil && lex != nil {
			query = lex.ImageQuery
		}
	}
	images, provider := uc.searchImages(ctx, query, headword, q.Provider)
	if len(images) > 0 {
		item.ImageCandidates = mergeImages(item.ImageCandidates, images)
		if err := uc.repo.UpdateItem(ctx, item); err != nil {
			return dto.ImageSearchResponseDto{}, internalErr("Failed to store images", err)
		}
	}
	return dto.ImageSearchResponseDto{Images: images, Provider: provider, Providers: uc.providerNames()}, nil
}
