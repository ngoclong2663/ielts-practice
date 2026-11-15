// Package images adapts stock-photo APIs (Pixabay, Unsplash) to vocab's IImageProvider.
package images

import (
	vocabout "ielts-practice/app/modules/vocab/ports/out"
	"ielts-practice/configs"
)

// NewImageProviders returns the providers in the configured search order (VOCAB_IMAGE_PROVIDERS).
func NewImageProviders(cfg *configs.Config) vocabout.ImageProviders {
	if cfg.FakeProviders {
		return vocabout.ImageProviders{NewFakeImageProvider()}
	}
	providers := vocabout.ImageProviders{}
	for _, name := range cfg.Vocab.ImageProviders {
		switch name {
		case "pixabay":
			providers = append(providers, NewPixabayProvider(cfg.Vocab.PixabayAPIKey))
		case "unsplash":
			providers = append(providers, NewUnsplashProvider(cfg.Vocab.UnsplashKey))
		}
	}
	return providers
}
