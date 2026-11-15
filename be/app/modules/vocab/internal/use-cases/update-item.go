package usecases

import (
	"context"
	"net/http"
	"strings"

	"ielts-practice/app/modules/vocab/domain/rules"

	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/dto"
	"ielts-practice/pkg/exceptions"
)

// UpdateItem stores the user's edits before saving: text fields, the chosen image, or removal.
// The shared lexeme is never changed (Epic 3 §5.1).
func (uc *VocabUseCases) UpdateItem(ctx context.Context, userID, jobID, itemID string, req dto.UpdateItemRequestDto) (dto.JobItemDto, error) {
	item, err := uc.repo.FindItem(ctx, userID, jobID, itemID)
	if err != nil {
		return dto.JobItemDto{}, internalErr("Failed to load word", err)
	}
	if item == nil {
		return dto.JobItemDto{}, itemNotFound()
	}
	editing := req.Content != nil || req.Image != nil || req.ClearImage
	if editing && (item.Status != entity.ItemReady || item.Content == nil) {
		return dto.JobItemDto{}, appErr(http.StatusConflict, exceptions.CodeVocabItemNotReady, "This word is not ready yet")
	}

	if req.Content != nil {
		next := *req.Content
		if fields := validateContent(&next); len(fields) > 0 {
			e := appErr(http.StatusBadRequest, exceptions.CodeValidationFailed, "Validation failed")
			e.Fields = fields
			return dto.JobItemDto{}, e
		}
		next.Image = item.Content.Image // the image changes only through `image`
		item.Content = &next
	}
	if req.ClearImage {
		item.Content.Image = nil
	}
	if req.Image != nil {
		var chosen *entity.CardImage
		for _, img := range item.ImageCandidates {
			if img.Provider == req.Image.Provider && img.ProviderID == req.Image.ProviderID {
				c := img
				chosen = &c
				break
			}
		}
		if chosen == nil {
			e := appErr(http.StatusBadRequest, exceptions.CodeValidationFailed, "Pick one of the suggested images")
			e.Fields = map[string]string{"image": "VALIDATION_INVALID"}
			return dto.JobItemDto{}, e
		}
		item.Content.Image = chosen
	}
	if req.Removed != nil {
		item.Removed = *req.Removed
	}
	if err := uc.repo.UpdateItem(ctx, item); err != nil {
		return dto.JobItemDto{}, internalErr("Failed to update word", err)
	}
	if req.Removed != nil {
		if err := uc.repo.RefreshJobProgress(ctx, jobID, uc.now()); err != nil {
			return dto.JobItemDto{}, internalErr("Failed to update generation", err)
		}
	}
	return uc.mapper.ItemEntityToDto(item), nil
}

// validateContent trims user edits and checks the required fields.
func validateContent(c *entity.CardContent) map[string]string {
	c.Headword = strings.TrimSpace(c.Headword)
	c.MeaningVI = strings.TrimSpace(c.MeaningVI)
	fields := map[string]string{}
	if c.Headword == "" {
		fields["headword"] = "VALIDATION_REQUIRED"
	} else if len([]rune(c.Headword)) > 100 {
		fields["headword"] = "VALIDATION_MAX"
	}
	if c.MeaningVI == "" {
		fields["meaning_vi"] = "VALIDATION_REQUIRED"
	} else if len([]rune(c.MeaningVI)) > 300 {
		fields["meaning_vi"] = "VALIDATION_MAX"
	}
	if len(fields) > 0 {
		return fields
	}
	r := entity.WordResult{Input: c.Headword, IsValid: true, Content: *c}
	rules.SanitizeResult(&r)
	*c = r.Content
	return nil
}
