package usecases

import (
	"context"
	"net/http"
	"time"

	decksin "ielts-practice/app/modules/decks/ports/in"
	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/dto"
	"ielts-practice/pkg/exceptions"
)

// CommitJob saves the chosen ready words as new cards in a deck (Epic 3 §3 step 6).
func (uc *VocabUseCases) CommitJob(ctx context.Context, userID, jobID string, req dto.CommitRequestDto) (dto.CommitResponseDto, error) {
	job, err := uc.repo.FindJob(ctx, userID, jobID)
	if err != nil {
		return dto.CommitResponseDto{}, internalErr("Failed to load generation", err)
	}
	if job == nil {
		return dto.CommitResponseDto{}, jobNotFound()
	}
	wanted := map[string]bool{}
	for _, id := range req.ItemIDs {
		wanted[id] = true
	}
	var cards []decksin.GeneratedCard
	var images []entity.CardImage
	for _, it := range job.Items {
		if !wanted[it.ID] || it.Removed || it.Status != entity.ItemReady || it.Content == nil {
			continue
		}
		cards = append(cards, decksin.GeneratedCard{LexemeID: it.LexemeID, Content: *it.Content})
		if it.Content.Image != nil {
			images = append(images, *it.Content.Image)
		}
	}
	if len(cards) == 0 {
		return dto.CommitResponseDto{}, appErr(http.StatusBadRequest, exceptions.CodeVocabNothingToSave, "No ready words selected")
	}

	deckID := req.DeckID
	if deckID == "" && req.NewDeckName == "" && job.TargetDeckID != nil {
		deckID = *job.TargetDeckID
	}
	res, err := uc.decks.SaveGeneratedCards(ctx, decksin.SaveGeneratedInput{
		UserID: userID, DeckID: deckID, NewDeckName: req.NewDeckName, Cards: cards,
	})
	if err != nil {
		return dto.CommitResponseDto{}, err
	}
	if err := uc.repo.SetJobCommitted(ctx, jobID, res.DeckID); err != nil {
		uc.logger.Warn().Err(err).Msg("Failed to mark generation as saved")
	}
	uc.trackDownloads(images)

	skipped := res.Skipped
	if skipped == nil {
		skipped = []string{}
	}
	return dto.CommitResponseDto{DeckID: res.DeckID, DeckName: res.DeckName, Created: res.Created, Skipped: skipped}, nil
}

// trackDownloads tells Unsplash which photos were chosen for cards (API guideline), in the background.
func (uc *VocabUseCases) trackDownloads(images []entity.CardImage) {
	if len(images) == 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		for _, img := range images {
			p := uc.provider(img.Provider)
			if p == nil {
				continue
			}
			if err := p.TrackDownload(ctx, img); err != nil {
				uc.logger.Warn().Err(err).Str("provider", img.Provider).Msg("Failed to track image download")
			}
		}
	}()
}
