package usecases

import (
	"time"

	"github.com/rs/zerolog"

	storageports "ielts-practice/app/modules/storage/ports"
	"ielts-practice/app/modules/writing/domain"
	"ielts-practice/configs"
)

// WritingUseCases implements Epic 5 (guided writing practice). For now: the task bank seed.
type WritingUseCases struct {
	repo    domain.IWritingRepository
	storage storageports.IObjectStorage
	cfg     *configs.Config
	logger  zerolog.Logger
	now     func() time.Time
}

func NewWritingUseCases(
	repo domain.IWritingRepository,
	storage storageports.IObjectStorage,
	cfg *configs.Config,
	logger zerolog.Logger,
) *WritingUseCases {
	return &WritingUseCases{
		repo: repo, storage: storage, cfg: cfg,
		logger: logger.With().Str("module", "writing").Logger(), now: time.Now,
	}
}
