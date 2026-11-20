package usecases

import (
	"context"
	"errors"
	"net/http"
	"time"

	"ielts-practice/app/modules/writing/domain/rules"

	writingout "ielts-practice/app/modules/writing/ports/out"

	"github.com/rs/zerolog"
	"golang.org/x/sync/singleflight"

	storagein "ielts-practice/app/modules/storage/ports/in"
	"ielts-practice/app/modules/writing/domain"
	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/app/modules/writing/dto"
	"ielts-practice/configs"
	"ielts-practice/pkg/exceptions"
)

const (
	limitWindow   = 24 * time.Hour
	historyLimit  = 50
	lessonTimeout = 3 * time.Minute
	vocabTimeout  = 90 * time.Second
	checkTimeout  = 90 * time.Second
	reviewTimeout = 2 * time.Minute
)

// WritingUseCases implements Epic 5: the task bank, guided attempts and AI feedback.
type WritingUseCases struct {
	repo    domain.IWritingRepository
	coach   writingout.ICoach
	storage storagein.IObjectStorage
	cfg     *configs.Config
	logger  zerolog.Logger
	now     func() time.Time
	lessons singleflight.Group
}

func NewWritingUseCases(
	repo domain.IWritingRepository,
	coach writingout.ICoach,
	storage storagein.IObjectStorage,
	cfg *configs.Config,
	logger zerolog.Logger,
) *WritingUseCases {
	return &WritingUseCases{
		repo: repo, coach: coach, storage: storage, cfg: cfg,
		logger: logger.With().Str("module", "writing").Logger(), now: time.Now,
	}
}

func appErr(status int, code, msg string) *exceptions.AppError {
	return exceptions.NewException(status, code, msg, nil)
}

func internalErr(msg string, err error) *exceptions.AppError {
	return exceptions.ErrInternalServerError(msg, err)
}

func taskNotFound() *exceptions.AppError {
	return appErr(http.StatusNotFound, exceptions.CodeWritingTaskNotFound, "Writing task not found")
}

func attemptNotFound() *exceptions.AppError {
	return appErr(http.StatusNotFound, exceptions.CodeWritingAttemptNotFound, "Attempt not found")
}

// coachErr maps a coach failure to the error the learner sees (their text is never lost: nothing was saved yet).
func coachErr(err error) *exceptions.AppError {
	if errors.Is(err, writingout.ErrCoachNotConfigured) {
		return exceptions.NewException(http.StatusServiceUnavailable, exceptions.CodeWritingNotConfigured, "AI feedback is not configured", err)
	}
	return exceptions.NewException(http.StatusBadGateway, exceptions.CodeWritingAIFailed, "The AI could not answer; please try again", err)
}

// checkLimit fails with 429 when the user has used today's allowance for this kind of AI call.
func (uc *WritingUseCases) checkLimit(ctx context.Context, userID, kind string) error {
	limit, code := uc.cfg.Writing.DailyChecks, exceptions.CodeWritingCheckLimit
	if kind == entity.KindSubmit {
		limit, code = uc.cfg.Writing.DailySubmissions, exceptions.CodeWritingSubmitLimit
	}
	used, err := uc.repo.CountAIRequests(ctx, userID, kind, uc.now().Add(-limitWindow))
	if err != nil {
		return internalErr("Failed to load the daily limit", err)
	}
	if used >= limit {
		e := appErr(http.StatusTooManyRequests, code, "Daily limit reached")
		e.Params = map[string]any{"limit": limit}
		return e
	}
	return nil
}

func (uc *WritingUseCases) record(ctx context.Context, userID string, attemptID *string, kind string, u entity.AIUsage) {
	if err := uc.repo.RecordAIRequest(ctx, userID, attemptID, kind, u); err != nil {
		uc.logger.Error().Err(err).Str("kind", kind).Msg("Failed to record AI usage")
	}
}

// ---- task mapping ----------------------------------------------------------------

func figureDto(f entity.Figure) dto.FigureDto {
	// The bucket is private: the API redirects to a short-lived presigned link (resolve-figure.go).
	return dto.FigureDto{URL: "/api/v1/writing/figures/" + f.Key[len(FigureKeyPrefix):], Alt: f.Alt, Width: f.Width, Height: f.Height}
}

func taskSummary(t *entity.Task, p domain.TaskProgress) dto.TaskSummaryDto {
	s := dto.TaskSummaryDto{
		ID: t.ID, TaskNo: t.TaskNo, TaskType: t.TaskType, Topic: t.Topic, Title: t.Title,
		Attempts: p.Attempts, BestBand: p.BestBand,
	}
	if len(t.Figures) > 0 {
		f := figureDto(t.Figures[0])
		s.Figure = &f
	}
	return s
}

func taskDto(t *entity.Task, p domain.TaskProgress) dto.TaskDto {
	d := dto.TaskDto{
		TaskSummaryDto: taskSummary(t, p), Prompt: t.Prompt, MinWords: t.MinWords, Minutes: t.Minutes,
		Figures: make([]dto.FigureDto, 0, len(t.Figures)), Positions: []string{},
	}
	for _, f := range t.Figures {
		d.Figures = append(d.Figures, figureDto(f))
	}
	if t.TaskNo == 2 {
		d.Positions = rules.Positions(t)
	}
	return d
}

func attemptSummary(a entity.AttemptSummary) dto.AttemptSummaryDto {
	return dto.AttemptSummaryDto{
		ID: a.ID, TaskID: a.TaskID, TaskTitle: a.TaskTitle, TaskNo: a.TaskNo, TaskType: a.TaskType,
		Level: a.Level, Position: a.Position, Status: a.Status, BandLow: a.BandLow, BandHigh: a.BandHigh,
		CreatedAt: a.CreatedAt, SubmittedAt: a.SubmittedAt,
	}
}
