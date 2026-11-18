package usecases

import (
	"github.com/rs/zerolog"

	"ielts-practice/app/modules/guides/domain"
	storagein "ielts-practice/app/modules/storage/ports/in"
)

// GuidesUseCases implements Epic 7 §5: the guide videos list and their media links.
type GuidesUseCases struct {
	repo    domain.IGuidesRepository
	storage storagein.IObjectStorage
	logger  zerolog.Logger
}

func NewGuidesUseCases(repo domain.IGuidesRepository, storage storagein.IObjectStorage, logger zerolog.Logger) *GuidesUseCases {
	return &GuidesUseCases{repo: repo, storage: storage, logger: logger.With().Str("module", "guides").Logger()}
}
