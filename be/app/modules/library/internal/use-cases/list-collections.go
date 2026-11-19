package usecases

import (
	"context"

	"ielts-practice/app/modules/library/dto"
)

// ListCollections returns the published collections that have words.
func (uc *LibraryUseCases) ListCollections(ctx context.Context) (dto.CollectionsResponseDto, error) {
	cols, err := uc.repo.ListCollections(ctx)
	if err != nil {
		return dto.CollectionsResponseDto{}, internalErr("Failed to load the library", err)
	}
	out := dto.CollectionsResponseDto{Collections: make([]dto.CollectionDto, 0, len(cols))}
	for i := range cols {
		if len(cols[i].Topics) == 0 {
			continue
		}
		out.Collections = append(out.Collections, toCollectionDto(&cols[i]))
	}
	return out, nil
}
