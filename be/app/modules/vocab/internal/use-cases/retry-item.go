package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/vocab/domain/rules"

	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/dto"
	"ielts-practice/pkg/exceptions"
)

// RetryItem retries a failed word, accepts a suggestion (new input text / sense hint),
// or regenerates fresh content for this item only.
func (uc *VocabUseCases) RetryItem(ctx context.Context, userID, jobID, itemID string, req dto.RetryItemRequestDto) (dto.JobItemDto, error) {
	item, err := uc.repo.FindItem(ctx, userID, jobID, itemID)
	if err != nil {
		return dto.JobItemDto{}, internalErr("Failed to load word", err)
	}
	if item == nil {
		return dto.JobItemDto{}, itemNotFound()
	}
	if !item.IsFinal() {
		return dto.JobItemDto{}, appErr(http.StatusConflict, exceptions.CodeVocabItemNotReady, "This word is still being generated")
	}

	text, hint := item.InputText, item.SenseHint
	if req.InputText != nil {
		text = *req.InputText
	}
	if req.SenseHint != nil {
		hint = *req.SenseHint
	}
	line := text
	if hint != "" {
		line += " | " + hint
	}
	entries, _ := rules.Normalize(line)
	if len(entries) != 1 {
		e := appErr(http.StatusBadRequest, exceptions.CodeValidationFailed, "Enter one word or phrase")
		e.Fields = map[string]string{"input_text": "VALIDATION_INVALID"}
		return dto.JobItemDto{}, e
	}
	item.InputText, item.SenseHint = entries[0].Text, entries[0].SenseHint
	item.Regenerate = req.Regenerate
	if err := uc.prepareItem(ctx, item); err != nil {
		return dto.JobItemDto{}, err
	}
	if item.Charged {
		limit, used, err := uc.quota(ctx, userID)
		if err != nil {
			return dto.JobItemDto{}, internalErr("Failed to load quota", err)
		}
		if used+1 > limit {
			return dto.JobItemDto{}, quotaExceeded(limit, used)
		}
	}
	item.Removed = false
	if err := uc.repo.UpdateItem(ctx, item); err != nil {
		return dto.JobItemDto{}, internalErr("Failed to update word", err)
	}
	if err := uc.repo.RefreshJobProgress(ctx, jobID, uc.now()); err != nil {
		return dto.JobItemDto{}, internalErr("Failed to update generation", err)
	}
	if item.Status == entity.ItemQueued {
		uc.notifyWorker()
	}
	return uc.mapper.ItemEntityToDto(item), nil
}
