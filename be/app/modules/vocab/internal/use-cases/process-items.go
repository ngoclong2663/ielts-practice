package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	"ielts-practice/app/modules/vocab/domain/rules"

	vocabout "ielts-practice/app/modules/vocab/ports/out"

	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/pkg/exceptions"
)

// BatchSize: words per LLM request (Epic 3 §5.3 batching).
const BatchSize = 10

// ClaimBatch takes the next due items off the queue.
func (uc *VocabUseCases) ClaimBatch(ctx context.Context) ([]entity.JobItem, error) {
	return uc.repo.ClaimQueuedItems(ctx, BatchSize, uc.now())
}

// RequeueStuck returns items abandoned by a crashed worker to the queue.
func (uc *VocabUseCases) RequeueStuck(ctx context.Context, olderThan time.Duration) {
	n, err := uc.repo.RequeueStuckItems(ctx, uc.now().Add(-olderThan))
	if err != nil {
		uc.logger.Error().Err(err).Msg("Failed to requeue stuck items")
	} else if n > 0 {
		uc.logger.Warn().Int64("items", n).Msg("Requeued stuck generation items")
	}
}

type pendingWord struct {
	item   *entity.JobItem
	lexeme *entity.Lexeme // owned lexeme to fill; nil for "regenerate" items
}

// ProcessBatch runs Epic 3 §5: cache lookup → content (one LLM call for the batch) → images.
func (uc *VocabUseCases) ProcessBatch(ctx context.Context, items []entity.JobItem) {
	jobs := map[string]bool{}
	var toGenerate []pendingWord
	for i := range items {
		it := &items[i]
		jobs[it.JobID] = true
		if it.Regenerate {
			toGenerate = append(toGenerate, pendingWord{item: it})
			continue
		}
		key := rules.LexemeKey(it.InputText, it.SenseHint)
		claim, err := uc.repo.ClaimLexeme(ctx, key, it.InputText, it.SenseHint, uc.now().Add(-pendingClaimTTL))
		if err != nil {
			uc.retryOrFail(ctx, pendingWord{item: it}, exceptions.CodeVocabGenerationFailed, err)
			continue
		}
		switch {
		case claim.Owned:
			toGenerate = append(toGenerate, pendingWord{item: it, lexeme: claim.Lexeme})
		case claim.Lexeme != nil && uc.applyCachedLexeme(it, claim.Lexeme):
			uc.saveItem(ctx, it)
		default:
			// Another worker is generating this word right now: check again shortly, free of charge.
			it.Status, it.Attempts = entity.ItemQueued, max(0, it.Attempts-1)
			it.NextAttemptAt = uc.now().Add(waitForOtherItem)
			uc.saveItem(ctx, it)
		}
	}
	if len(toGenerate) > 0 {
		uc.generate(ctx, toGenerate)
	}
	for jobID := range jobs {
		if err := uc.repo.RefreshJobProgress(ctx, jobID, uc.now()); err != nil {
			uc.logger.Error().Err(err).Str("job_id", jobID).Msg("Failed to refresh job progress")
		}
	}
}

func (uc *VocabUseCases) saveItem(ctx context.Context, it *entity.JobItem) {
	if err := uc.repo.UpdateItem(ctx, it); err != nil {
		uc.logger.Error().Err(err).Str("item_id", it.ID).Msg("Failed to save generation item")
	}
}

func (uc *VocabUseCases) generate(ctx context.Context, words []pendingWord) {
	if !uc.generator.Configured() {
		for _, w := range words {
			uc.fail(ctx, w, exceptions.CodeVocabNotConfigured)
		}
		return
	}
	requests := make([]entity.WordRequest, len(words))
	for i, w := range words {
		requests[i] = entity.WordRequest{Input: w.item.InputText, SenseHint: w.item.SenseHint}
	}
	start := uc.now()
	results, usage, err := uc.generator.Generate(ctx, requests)
	uc.logger.Info().Int("words", len(words)).Int("input_tokens", usage.InputTokens).
		Int("output_tokens", usage.OutputTokens).Int("cache_read_tokens", usage.CacheReadTokens).
		Dur("latency", uc.now().Sub(start)).Err(err).Msg("Generated card content")
	uc.addUsage(ctx, words, usage)

	byInput := map[string]entity.WordResult{}
	for _, r := range results {
		byInput[strings.ToLower(strings.TrimSpace(r.Input))] = r
	}
	for _, w := range words {
		res, ok := byInput[strings.ToLower(w.item.InputText)]
		switch {
		case ok:
			uc.applyResult(ctx, w, res)
		case errors.Is(err, vocabout.ErrGeneratorRefused):
			uc.fail(ctx, w, exceptions.CodeVocabGenerationFailed)
		case errors.Is(err, vocabout.ErrGeneratorNotConfigured):
			uc.fail(ctx, w, exceptions.CodeVocabNotConfigured)
		default:
			if err == nil {
				err = errors.New("word missing from the generator response")
			}
			uc.retryOrFail(ctx, w, exceptions.CodeVocabGenerationFailed, err)
		}
	}
}

// addUsage splits the token usage across jobs in proportion to their words in the batch.
func (uc *VocabUseCases) addUsage(ctx context.Context, words []pendingWord, usage entity.Usage) {
	perJob := map[string]int{}
	for _, w := range words {
		perJob[w.item.JobID]++
	}
	for jobID, n := range perJob {
		share := entity.Usage{
			InputTokens:     usage.InputTokens * n / len(words),
			OutputTokens:    usage.OutputTokens * n / len(words),
			CacheReadTokens: usage.CacheReadTokens * n / len(words),
		}
		if err := uc.repo.AddJobUsage(ctx, jobID, share); err != nil {
			uc.logger.Warn().Err(err).Msg("Failed to record token usage")
		}
	}
}

func (uc *VocabUseCases) applyResult(ctx context.Context, w pendingWord, res entity.WordResult) {
	rules.SanitizeResult(&res)
	it := w.item
	it.Charged, it.Cached, it.ErrorCode = true, false, ""

	if !res.IsValid {
		it.Status, it.Suggestion = entity.ItemInvalid, res.Suggestion
		if w.lexeme != nil {
			w.lexeme.Status, w.lexeme.Suggestion = entity.LexemeInvalid, res.Suggestion
			uc.saveLexeme(ctx, w.lexeme)
			it.LexemeID = &w.lexeme.ID
		}
		uc.saveItem(ctx, it)
		return
	}
	if res.Content.MeaningVI == "" {
		uc.retryOrFail(ctx, w, exceptions.CodeVocabGenerationFailed, errors.New("empty meaning"))
		return
	}

	it.Status = entity.ItemImage
	uc.saveItem(ctx, it)

	content := res.Content
	content.SenseHint = it.SenseHint
	candidates, _ := uc.searchImages(ctx, res.ImageQuery, content.Headword, "")
	// Words with no sensible photo keep the candidates for the picker but no image by default.
	if len(candidates) > 0 && res.Visualizable {
		img := candidates[0]
		content.Image = &img
	}

	if w.lexeme != nil {
		lex := w.lexeme
		lex.Headword, lex.Lemma = content.Headword, content.Lemma
		lex.Content, lex.Status, lex.Suggestion = &content, entity.LexemeReady, ""
		if !rules.Shareable(res.QualityFlags) {
			// Kept for review but not reused: the next learner asking for this word gets a fresh card.
			lex.Status = entity.LexemeNeedsReview
			uc.logger.Warn().Str("lexeme_id", lex.ID).Strs("flags", res.QualityFlags).Msg("Generated card not shared")
		}
		lex.ImageQuery, lex.ImageCandidates, lex.QualityFlags = res.ImageQuery, candidates, res.QualityFlags
		uc.saveLexeme(ctx, lex)
		it.LexemeID = &lex.ID
	}
	copied := content
	it.Status, it.Content, it.ImageCandidates, it.Suggestion = entity.ItemReady, &copied, candidates, res.Suggestion
	uc.saveItem(ctx, it)
}

func (uc *VocabUseCases) saveLexeme(ctx context.Context, lex *entity.Lexeme) {
	if err := uc.repo.SaveLexeme(ctx, lex); err != nil {
		uc.logger.Error().Err(err).Str("lexeme_id", lex.ID).Msg("Failed to save lexeme")
	}
}

// retryOrFail retries with exponential backoff (2 s, 4 s, …) up to maxAttempts.
func (uc *VocabUseCases) retryOrFail(ctx context.Context, w pendingWord, code string, cause error) {
	uc.logger.Warn().Err(cause).Str("item_id", w.item.ID).Int("attempt", w.item.Attempts).Msg("Generation attempt failed")
	if w.item.Attempts >= maxAttempts {
		uc.fail(ctx, w, code)
		return
	}
	uc.releaseLexeme(ctx, w)
	w.item.Status = entity.ItemQueued
	w.item.NextAttemptAt = uc.now().Add(time.Duration(1<<w.item.Attempts) * time.Second)
	uc.saveItem(ctx, w.item)
}

func (uc *VocabUseCases) fail(ctx context.Context, w pendingWord, code string) {
	uc.releaseLexeme(ctx, w)
	w.item.Status, w.item.ErrorCode, w.item.Charged = entity.ItemFailed, code, false
	uc.saveItem(ctx, w.item)
}

// releaseLexeme marks an owned lexeme failed so the next attempt (or another user) can claim it.
func (uc *VocabUseCases) releaseLexeme(ctx context.Context, w pendingWord) {
	if w.lexeme == nil {
		return
	}
	if err := uc.repo.SetLexemeStatus(ctx, w.lexeme.ID, entity.LexemeFailed); err != nil {
		uc.logger.Error().Err(err).Msg("Failed to release lexeme")
	}
}
