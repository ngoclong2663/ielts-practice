package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/vocab/dto"
	"ielts-practice/pkg/exceptions"
)

// ReportLexeme records "wrong content" once per user; after reportThreshold reports the shared
// content is regenerated the next time someone adds the word.
func (uc *VocabUseCases) ReportLexeme(ctx context.Context, userID, lexemeID string, req dto.ReportRequestDto) error {
	lex, err := uc.repo.FindLexemeByID(ctx, lexemeID)
	if err != nil {
		return internalErr("Failed to load word", err)
	}
	if lex == nil {
		return appErr(http.StatusNotFound, exceptions.CodeLexemeNotFound, "Word not found")
	}
	if _, err := uc.repo.ReportLexeme(ctx, lexemeID, userID, req.Reason, req.Note, reportThreshold); err != nil {
		return internalErr("Failed to save report", err)
	}
	return nil
}
