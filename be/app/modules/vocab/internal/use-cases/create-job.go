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

func (uc *VocabUseCases) CreateJob(ctx context.Context, userID string, req dto.CreateJobRequestDto) (dto.JobDto, error) {
	lines := make([]string, 0, len(req.Entries))
	for _, e := range req.Entries {
		line := strings.ReplaceAll(e.Text, "|", " ")
		if hint := strings.TrimSpace(e.SenseHint); hint != "" {
			line += " | " + strings.ReplaceAll(hint, "|", " ")
		}
		lines = append(lines, line)
	}
	entries, _ := rules.Normalize(strings.Join(lines, "\n"))
	if len(entries) == 0 {
		return dto.JobDto{}, appErr(http.StatusBadRequest, exceptions.CodeVocabNoEntries, "No valid words to generate")
	}
	if len(entries) > rules.MaxEntriesPerJob {
		return dto.JobDto{}, appErr(http.StatusBadRequest, exceptions.CodeVocabTooManyEntries, "Too many words").
			WithParams(map[string]any{"max": rules.MaxEntriesPerJob})
	}
	if req.TargetDeckID != nil {
		owns, err := uc.decks.OwnsDeck(ctx, userID, *req.TargetDeckID)
		if err != nil {
			return dto.JobDto{}, err
		}
		if !owns {
			return dto.JobDto{}, appErr(http.StatusNotFound, exceptions.CodeDeckNotFound, "Deck not found")
		}
	}

	job := &entity.Job{UserID: userID, TargetDeckID: req.TargetDeckID, Status: entity.JobQueued}
	charged := 0
	for i, e := range entries {
		item := entity.JobItem{Position: i, InputText: e.Text, SenseHint: e.SenseHint}
		if err := uc.prepareItem(ctx, &item); err != nil {
			return dto.JobDto{}, err
		}
		if item.Charged {
			charged++
		}
		job.Items = append(job.Items, item)
	}
	if charged > 0 {
		limit, used, err := uc.quota(ctx, userID)
		if err != nil {
			return dto.JobDto{}, internalErr("Failed to load quota", err)
		}
		if used+charged > limit {
			return dto.JobDto{}, quotaExceeded(limit, used)
		}
	}

	created, err := uc.repo.CreateJob(ctx, job)
	if err != nil {
		return dto.JobDto{}, internalErr("Failed to create generation", err)
	}
	if charged > 0 {
		uc.notifyWorker()
	}
	return uc.mapper.JobEntityToDto(created), nil
}

// prepareItem resolves an item from the shared cache when possible (ready instantly, no quota),
// otherwise queues it for generation and marks it as charged.
func (uc *VocabUseCases) prepareItem(ctx context.Context, item *entity.JobItem) error {
	item.Status, item.ErrorCode, item.Suggestion = entity.ItemQueued, "", ""
	item.Content, item.ImageCandidates, item.LexemeID = nil, []entity.CardImage{}, nil
	item.Cached, item.Charged, item.Attempts = false, false, 0
	item.NextAttemptAt = uc.now()
	if !item.Regenerate {
		lex, err := uc.repo.FindLexemeByKey(ctx, rules.LexemeKey(item.InputText, item.SenseHint))
		if err != nil {
			return internalErr("Failed to look up word", err)
		}
		if lex != nil && uc.applyCachedLexeme(item, lex) {
			return nil
		}
	}
	item.Charged = true
	return nil
}

// applyCachedLexeme fills the item from a finished lexeme; false when the lexeme is not usable.
func (uc *VocabUseCases) applyCachedLexeme(item *entity.JobItem, lex *entity.Lexeme) bool {
	switch {
	case lex.Status == entity.LexemeReady && lex.Content != nil:
		content := *lex.Content
		item.Status, item.Content = entity.ItemReady, &content
		item.ImageCandidates = lex.ImageCandidates
	case lex.Status == entity.LexemeInvalid:
		item.Status, item.Suggestion = entity.ItemInvalid, lex.Suggestion
	default:
		return false
	}
	id := lex.ID
	item.LexemeID, item.Cached, item.Charged = &id, true, false
	return true
}
