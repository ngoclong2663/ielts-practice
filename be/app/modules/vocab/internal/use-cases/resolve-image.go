package usecases

import (
	"context"
	"net/http"
	"time"

	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/pkg/exceptions"
)

// ImageRedirect is where /api/v1/images/:provider/:id sends the browser.
type ImageRedirect struct {
	Location string
	MaxAge   time.Duration
}

type refreshCall struct {
	done   chan struct{}
	result *entity.ImageLink
	err    error
}

func imageNotFound() *exceptions.AppError {
	return appErr(http.StatusNotFound, exceptions.CodeImageNotFound, "Image not found")
}

// ResolveImage implements the Pixabay link refresh (Epic 3 §5.4): serve the stored direct link while it
// is valid, refresh it by id when it has expired, and find a replacement when the photo was removed.
func (uc *VocabUseCases) ResolveImage(ctx context.Context, provider, providerID string, thumb bool) (ImageRedirect, error) {
	p := uc.provider(provider)
	if p == nil || !p.Expiring() {
		return ImageRedirect{}, imageNotFound()
	}
	link, err := uc.repo.FindImageLink(ctx, provider, providerID)
	if err != nil {
		return ImageRedirect{}, internalErr("Failed to load image", err)
	}
	if link == nil {
		return ImageRedirect{}, imageNotFound()
	}
	if link.Replacement == nil && !link.Gone && uc.now().After(link.ExpiresAt) {
		link, err = uc.refreshOnce(ctx, link)
		if err != nil {
			return ImageRedirect{}, err
		}
	}
	switch {
	case link.Replacement != nil:
		return ImageRedirect{Location: pickURL(*link.Replacement, thumb), MaxAge: time.Hour}, nil
	case link.Gone:
		return ImageRedirect{}, imageNotFound()
	}
	maxAge := min(time.Hour, max(0, link.ExpiresAt.Sub(uc.now())))
	return ImageRedirect{Location: pickURL(entity.CardImage{URL: link.URL, ThumbURL: link.ThumbURL}, thumb), MaxAge: maxAge}, nil
}

func pickURL(img entity.CardImage, thumb bool) string {
	if thumb && img.ThumbURL != "" {
		return img.ThumbURL
	}
	return img.URL
}

// refreshOnce collapses concurrent refreshes of the same photo into one provider call.
func (uc *VocabUseCases) refreshOnce(ctx context.Context, link *entity.ImageLink) (*entity.ImageLink, error) {
	key := link.Provider + "/" + link.ProviderID
	uc.refreshMu.Lock()
	if call, ok := uc.refreshes[key]; ok {
		uc.refreshMu.Unlock()
		select {
		case <-call.done:
			return call.result, call.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	call := &refreshCall{done: make(chan struct{})}
	uc.refreshes[key] = call
	uc.refreshMu.Unlock()

	call.result, call.err = uc.refreshLink(context.WithoutCancel(ctx), link)
	close(call.done)
	uc.refreshMu.Lock()
	delete(uc.refreshes, key)
	uc.refreshMu.Unlock()
	return call.result, call.err
}

func (uc *VocabUseCases) refreshLink(ctx context.Context, link *entity.ImageLink) (*entity.ImageLink, error) {
	p := uc.provider(link.Provider)
	fresh, err := p.Lookup(ctx, link.ProviderID)
	if err != nil {
		// Provider unavailable: try the old link rather than failing the image outright.
		uc.logger.Warn().Err(err).Str("provider", link.Provider).Str("id", link.ProviderID).Msg("Image refresh failed")
		return link, nil
	}
	if fresh != nil {
		link.URL, link.ThumbURL = fresh.URL, fresh.ThumbURL
		link.ExpiresAt = fresh.LinkExpiresAt
		if link.ExpiresAt.IsZero() {
			link.ExpiresAt = uc.now().Add(time.Hour)
		}
		if err := uc.repo.UpsertImageLink(ctx, link); err != nil {
			return nil, internalErr("Failed to store image link", err)
		}
		return link, nil
	}

	// The photo was removed upstream: search again (same provider first, then the others).
	replacement := uc.findReplacement(ctx, link)
	if replacement == nil {
		link.Gone = true
	} else {
		link.Replacement = replacement
	}
	if err := uc.repo.ReplaceImageEverywhere(ctx, link.Provider, link.ProviderID, replacement); err != nil {
		return nil, internalErr("Failed to replace image", err)
	}
	if err := uc.repo.UpsertImageLink(ctx, link); err != nil {
		return nil, internalErr("Failed to store image link", err)
	}
	uc.logger.Info().Str("provider", link.Provider).Str("id", link.ProviderID).Bool("replaced", replacement != nil).
		Msg("Image removed upstream")
	return link, nil
}

func (uc *VocabUseCases) findReplacement(ctx context.Context, link *entity.ImageLink) *entity.CardImage {
	query := link.Query
	if query == "" {
		query = link.Headword
	}
	if query == "" {
		return nil
	}
	for _, p := range uc.images {
		if !p.Configured() {
			continue
		}
		images, err := p.Search(ctx, query, imageCandidates)
		if err != nil {
			uc.logger.Warn().Err(err).Str("provider", p.Name()).Msg("Replacement image search failed")
			continue
		}
		for _, img := range images {
			if img.Provider == link.Provider && img.ProviderID == link.ProviderID {
				continue
			}
			linked := uc.linkImages(ctx, p, []entity.CardImage{img}, query, link.Headword)
			if len(linked) == 1 {
				return &linked[0]
			}
		}
	}
	return nil
}
