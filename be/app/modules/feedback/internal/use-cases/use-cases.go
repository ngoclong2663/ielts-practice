package usecases

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/feedback/domain"
	"ielts-practice/app/modules/feedback/domain/entity"
	"ielts-practice/app/modules/feedback/dto"
	notifyin "ielts-practice/app/modules/notify/ports/in"
	usersin "ielts-practice/app/modules/users/ports/in"
	"ielts-practice/configs"
	"ielts-practice/pkg/exceptions"
)

const (
	limitWindow     = 24 * time.Hour
	notifyTimeout   = 30 * time.Second
	defaultPageSize = 25
)

// IFeedbackUseCases is what the HTTP presenter needs. One use case per file in this folder.
type IFeedbackUseCases interface {
	SendFeedback(ctx context.Context, userID, userAgent string, req dto.CreateFeedbackRequestDto) (dto.CreatedFeedbackDto, error)
	ListFeedback(ctx context.Context, q dto.ListFeedbackQueryDto) (dto.FeedbackListDto, error)
	UpdateFeedback(ctx context.Context, id string, req dto.UpdateFeedbackRequestDto) (dto.FeedbackItemDto, error)
}

var _ IFeedbackUseCases = &FeedbackUseCases{}

type FeedbackUseCases struct {
	repo     domain.IFeedbackRepository
	users    usersin.IUserService
	notifier notifyin.IAdminNotifier
	cfg      *configs.Config
	logger   zerolog.Logger
	now      func() time.Time
	// inFlight tracks background notifications so shutdown waits for them.
	inFlight sync.WaitGroup
}

func NewFeedbackUseCases(
	repo domain.IFeedbackRepository,
	users usersin.IUserService,
	notifier notifyin.IAdminNotifier,
	cfg *configs.Config,
	logger zerolog.Logger,
) *FeedbackUseCases {
	return &FeedbackUseCases{repo: repo, users: users, notifier: notifier, cfg: cfg, logger: logger, now: time.Now}
}

// Wait blocks until background notifications finish or ctx ends (fx OnStop).
func (uc *FeedbackUseCases) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() { uc.inFlight.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
	}
	return nil
}

func internalErr(msg string, err error) *exceptions.AppError {
	return exceptions.ErrInternalServerError(msg, err)
}

func notFound() *exceptions.AppError {
	return exceptions.NewException(http.StatusNotFound, exceptions.CodeFeedbackNotFound, "Feedback not found", nil)
}

func toItemDto(fb *entity.Feedback) dto.FeedbackItemDto {
	return dto.FeedbackItemDto{
		ID: fb.ID, Category: fb.Category, Message: fb.Message, Page: fb.Page, UILocale: fb.UILocale,
		UserAgent: fb.UserAgent, Status: fb.Status, AdminNote: fb.AdminNote, Notified: fb.NotifiedAt != nil,
		CreatedAt: fb.CreatedAt, UpdatedAt: fb.UpdatedAt, ResolvedAt: fb.ResolvedAt,
		User: dto.FeedbackUserDto{ID: fb.User.ID, Email: fb.User.Email, DisplayName: fb.User.DisplayName, AvatarURL: fb.User.AvatarURL},
	}
}
