package usecases

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	deckports "ielts-practice/app/modules/decks/ports"
	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/internal/mapper"
	"ielts-practice/configs"
	"ielts-practice/pkg/exceptions"
)

const (
	imageCandidates  = 6
	maxStoredImages  = 30
	maxAttempts      = 3
	reportThreshold  = 3
	pendingClaimTTL  = 2 * time.Minute
	waitForOtherItem = 2 * time.Second
	quotaWindow      = 24 * time.Hour
)

// VocabUseCases implements Epic 3: paste words → generated, reviewable cards.
type VocabUseCases struct {
	repo      domain.IVocabRepository
	generator domain.IContentGenerator
	suggester domain.IWordSuggester
	images    domain.ImageProviders
	decks     deckports.IDeckService
	mapper    *mapper.VocabMapper
	cfg       *configs.Config
	logger    zerolog.Logger
	now       func() time.Time

	wake      chan struct{}
	refreshMu sync.Mutex
	refreshes map[string]*refreshCall
}

func NewVocabUseCases(
	repo domain.IVocabRepository,
	generator domain.IContentGenerator,
	suggester domain.IWordSuggester,
	images domain.ImageProviders,
	decks deckports.IDeckService,
	m *mapper.VocabMapper,
	cfg *configs.Config,
	logger zerolog.Logger,
) *VocabUseCases {
	return &VocabUseCases{
		repo: repo, generator: generator, suggester: suggester, images: images, decks: decks, mapper: m, cfg: cfg,
		logger: logger.With().Str("module", "vocab").Logger(), now: time.Now,
		wake: make(chan struct{}, 1), refreshes: map[string]*refreshCall{},
	}
}

// Wake returns the channel the worker listens on for new work.
func (uc *VocabUseCases) Wake() <-chan struct{} { return uc.wake }

func (uc *VocabUseCases) notifyWorker() {
	select {
	case uc.wake <- struct{}{}:
	default:
	}
}

func appErr(status int, code, msg string) *exceptions.AppError {
	return exceptions.NewException(status, code, msg, nil)
}

func internalErr(msg string, err error) *exceptions.AppError {
	return exceptions.ErrInternalServerError(msg, err)
}

func jobNotFound() *exceptions.AppError {
	return appErr(http.StatusNotFound, exceptions.CodeVocabJobNotFound, "Generation not found")
}

func itemNotFound() *exceptions.AppError {
	return appErr(http.StatusNotFound, exceptions.CodeVocabItemNotFound, "Word not found in this generation")
}

func (uc *VocabUseCases) provider(name string) domain.IImageProvider {
	for _, p := range uc.images {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

func (uc *VocabUseCases) providerNames() []string {
	out := []string{}
	for _, p := range uc.images {
		if p.Configured() {
			out = append(out, p.Name())
		}
	}
	return out
}

// quota returns the user's generated-word usage in the last 24 h.
func (uc *VocabUseCases) quota(ctx context.Context, userID string) (limit, used int, err error) {
	used, err = uc.repo.ChargedWordsSince(ctx, userID, uc.now().Add(-quotaWindow))
	return uc.cfg.Vocab.DailyQuota, used, err
}

func quotaExceeded(limit, used int) *exceptions.AppError {
	return appErr(http.StatusTooManyRequests, exceptions.CodeVocabQuotaExceeded, "Daily generation limit reached").
		WithParams(map[string]any{"limit": limit, "remaining": max(0, limit-used)})
}

// searchImages queries one provider, or each configured provider in order until one has results.
// Photos with expiring links are registered and rewritten to our refresh endpoint.
func (uc *VocabUseCases) searchImages(ctx context.Context, query, headword, only string) ([]entity.CardImage, string) {
	query = strings.TrimSpace(query)
	if query == "" {
		query = headword
	}
	for _, p := range uc.images {
		if !p.Configured() || (only != "" && p.Name() != only) {
			continue
		}
		images, err := p.Search(ctx, query, imageCandidates)
		if err != nil {
			uc.logger.Warn().Err(err).Str("provider", p.Name()).Str("query", query).Msg("Image search failed")
			continue
		}
		if len(images) == 0 {
			continue
		}
		return uc.linkImages(ctx, p, images, query, headword), p.Name()
	}
	return []entity.CardImage{}, ""
}

func imagePath(provider, id string) string {
	return "/api/v1/images/" + provider + "/" + id
}

func (uc *VocabUseCases) linkImages(ctx context.Context, p domain.IImageProvider, images []entity.CardImage, query, headword string) []entity.CardImage {
	if !p.Expiring() {
		return images
	}
	out := make([]entity.CardImage, 0, len(images))
	for _, img := range images {
		expires := img.LinkExpiresAt
		if expires.IsZero() {
			expires = uc.now().Add(time.Hour)
		}
		err := uc.repo.UpsertImageLink(ctx, &entity.ImageLink{
			Provider: img.Provider, ProviderID: img.ProviderID, URL: img.URL, ThumbURL: img.ThumbURL,
			ExpiresAt: expires, Query: query, Headword: headword,
		})
		if err != nil {
			uc.logger.Warn().Err(err).Msg("Failed to store image link")
			continue
		}
		img.URL = imagePath(img.Provider, img.ProviderID)
		img.ThumbURL = img.URL + "?size=thumb"
		out = append(out, img)
	}
	return out
}

func mergeImages(existing, added []entity.CardImage) []entity.CardImage {
	seen := map[string]bool{}
	out := make([]entity.CardImage, 0, len(existing)+len(added))
	for _, img := range append(append([]entity.CardImage{}, added...), existing...) {
		k := img.Provider + "/" + img.ProviderID
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, img)
	}
	if len(out) > maxStoredImages {
		out = out[:maxStoredImages]
	}
	return out
}
