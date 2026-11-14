package usecases

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/study/domain"
	"ielts-practice/app/modules/study/domain/entity"
	"ielts-practice/app/modules/study/dto"
	userEntity "ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/app/modules/users/ports"
	"ielts-practice/configs"
	"ielts-practice/pkg/exceptions"
	"ielts-practice/pkg/steps"
	"ielts-practice/pkg/studyday"
)

const (
	defaultQueueSize = 5
	// Like Anki: when nothing else is left, learning cards due within this window are shown early.
	learnAhead = 20 * time.Minute
	// Anki caps the time counted for one answer.
	maxAnswerDuration = 60_000
)

// IStudyUseCases is what the HTTP presenter needs. One use case per file in this folder.
type IStudyUseCases interface {
	GetQueue(ctx context.Context, userID string, q dto.QueueQueryDto) (dto.StudyQueueDto, error)
	Answer(ctx context.Context, userID string, req dto.AnswerRequestDto) (dto.AnswerResponseDto, error)
	Undo(ctx context.Context, userID string, req dto.UndoRequestDto) (dto.UndoResponseDto, error)
	GetCounts(ctx context.Context, userID string) (dto.CountsResponseDto, error)
	GetStats(ctx context.Context, userID string) (dto.StatsOverviewDto, error)
}

var _ IStudyUseCases = &StudyUseCases{}

type StudyUseCases struct {
	repo        domain.IStudyRepository
	users       ports.IUserService
	transaction *database.TransactionStruct
	logger      zerolog.Logger
	enableFuzz  bool
	now         func() time.Time
}

func NewStudyUseCases(
	repo domain.IStudyRepository,
	users ports.IUserService,
	tx *database.TransactionStruct,
	cfg *configs.Config,
	logger zerolog.Logger,
) *StudyUseCases {
	return &StudyUseCases{
		repo:        repo,
		users:       users,
		transaction: tx,
		logger:      logger,
		enableFuzz:  !cfg.IsTest(), // deterministic intervals in tests
		now:         time.Now,
	}
}

// studyContext is everything about "now" for one user: day boundaries, limits, scheduler.
type studyContext struct {
	now           time.Time
	user          *userEntity.UserEntity
	settings      *userEntity.StudySettingsEntity
	todayStart    time.Time
	tomorrowStart time.Time
	scheduler     *domain.Scheduler
}

func (uc *StudyUseCases) loadContext(ctx context.Context, userID string) (*studyContext, error) {
	user, settings, err := uc.users.GetStudyProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := uc.now()
	todayStart, tomorrowStart := studyday.Bounds(now, user.Timezone, user.DayStartHour)
	learning, err := steps.ParseMinutes(settings.LearningSteps)
	if err != nil {
		learning = nil // fall back to FSRS defaults
	}
	relearning, err := steps.ParseMinutes(settings.RelearningSteps)
	if err != nil {
		relearning = nil
	}
	return &studyContext{
		now:           now,
		user:          user,
		settings:      settings,
		todayStart:    todayStart,
		tomorrowStart: tomorrowStart,
		scheduler: domain.NewScheduler(domain.SchedulerConfig{
			DesiredRetention: settings.DesiredRetention,
			LearningSteps:    learning,
			RelearningSteps:  relearning,
			MaxIntervalDays:  float64(settings.MaxIntervalDays),
			EnableFuzz:       uc.enableFuzz,
		}),
	}, nil
}

func internalErr(msg string, err error) *exceptions.AppError {
	return exceptions.ErrInternalServerError(msg, err)
}

func stateChanged() *exceptions.AppError {
	return exceptions.NewException(http.StatusConflict, exceptions.CodeStudyStateChanged, "The card changed since it was shown", nil)
}

func sizeOrDefault(size int) int {
	if size <= 0 {
		return defaultQueueSize
	}
	return size
}

// limitsInScope applies daily limits per deck and sums them for the requested scope.
type limitsInScope struct {
	newCaps, reviewCaps map[string]int
	counts              entity.Counts
	perDeck             []dto.DeckCountsDto
	newDone, reviewDone int
}

func (uc *StudyUseCases) computeLimits(ctx context.Context, sc *studyContext, userID, deckID string) (*limitsInScope, error) {
	stats, err := uc.repo.DeckStats(ctx, userID, deckID, sc.tomorrowStart)
	if err != nil {
		return nil, internalErr("Failed to count cards", err)
	}
	if deckID != "" && len(stats) == 0 {
		return nil, exceptions.NewException(http.StatusNotFound, exceptions.CodeDeckNotFound, "Deck not found", nil)
	}
	today, err := uc.repo.TodayStats(ctx, userID, sc.todayStart)
	if err != nil {
		return nil, internalErr("Failed to count today's reviews", err)
	}
	out := &limitsInScope{newCaps: map[string]int{}, reviewCaps: map[string]int{}}
	for _, st := range stats {
		t := today[st.DeckID]
		n, r := domain.ApplyLimits(st, t, sc.settings.NewCardsPerDay, sc.settings.MaxReviewsPerDay)
		if n > 0 {
			out.newCaps[st.DeckID] = n
		}
		if r > 0 {
			out.reviewCaps[st.DeckID] = r
		}
		c := entity.Counts{New: n, Learning: st.LearningDue, Review: r}
		out.perDeck = append(out.perDeck, dto.DeckCountsDto{DeckID: st.DeckID, Counts: c})
		out.counts.New += n
		out.counts.Learning += st.LearningDue
		out.counts.Review += r
		out.newDone += t.Introduced
		out.reviewDone += t.Reviewed
	}
	return out, nil
}

// buildQueue returns the next cards for the scope. `first`, when set, is shown first (undo).
func (uc *StudyUseCases) buildQueue(ctx context.Context, sc *studyContext, userID, deckID string, size int, first *entity.StudyCard) (dto.StudyQueueDto, error) {
	size = sizeOrDefault(size)
	lim, err := uc.computeLimits(ctx, sc, userID, deckID)
	if err != nil {
		return dto.StudyQueueDto{}, err
	}

	learning, err := uc.repo.LearningCards(ctx, userID, deckID, sc.now.Add(learnAhead), size*2)
	if err != nil {
		return dto.StudyQueueDto{}, internalErr("Failed to load learning cards", err)
	}
	var learningNow, learningAhead []entity.StudyCard
	for _, c := range learning {
		if !c.Due.After(sc.now) {
			learningNow = append(learningNow, c)
		} else {
			learningAhead = append(learningAhead, c)
		}
	}
	reviews, err := uc.repo.ReviewCards(ctx, userID, lim.reviewCaps, sc.tomorrowStart, size)
	if err != nil {
		return dto.StudyQueueDto{}, internalErr("Failed to load reviews", err)
	}
	news, err := uc.repo.NewCards(ctx, userID, lim.newCaps, size)
	if err != nil {
		return dto.StudyQueueDto{}, internalErr("Failed to load new cards", err)
	}

	cards := domain.BuildQueue(domain.QueueInput{
		LearningNow:     learningNow,
		Reviews:         reviews,
		News:            news,
		LearningAhead:   learningAhead,
		ReviewRemaining: lim.counts.Review,
		NewRemaining:    lim.counts.New,
		ReviewDoneToday: lim.reviewDone,
		NewDoneToday:    lim.newDone,
		Size:            size,
	})
	if first != nil {
		rest := make([]entity.StudyCard, 0, len(cards))
		for _, c := range cards {
			if c.ID != first.ID {
				rest = append(rest, c)
			}
		}
		cards = append([]entity.StudyCard{*first}, rest...)
		if len(cards) > size {
			cards = cards[:size]
		}
	}

	out := dto.StudyQueueDto{Cards: make([]dto.StudyCardDto, 0, len(cards)), Counts: lim.counts, DayEndsAt: sc.tomorrowStart}
	for _, c := range cards {
		cardDto, err := uc.toCardDto(sc, c)
		if err != nil {
			return dto.StudyQueueDto{}, err
		}
		out.Cards = append(out.Cards, cardDto)
	}
	if len(out.Cards) == 0 {
		next, err := uc.repo.NextLearningDue(ctx, userID, deckID, sc.now, sc.tomorrowStart)
		if err != nil {
			return dto.StudyQueueDto{}, internalErr("Failed to find next due card", err)
		}
		out.NextDueAt = next
	}
	return out, nil
}

func stateName(s int16) string {
	switch s {
	case entity.StateLearning:
		return "learning"
	case entity.StateReview:
		return "review"
	case entity.StateRelearning:
		return "relearning"
	default:
		return "new"
	}
}

func (uc *StudyUseCases) toCardDto(sc *studyContext, c entity.StudyCard) (dto.StudyCardDto, error) {
	preview, err := sc.scheduler.Preview(c.SchedulingState, sc.now)
	if err != nil {
		return dto.StudyCardDto{}, internalErr("Failed to preview intervals", err)
	}
	secs := func(r entity.Rating) int64 { return int64(preview[r].Round(time.Second).Seconds()) }
	return dto.StudyCardDto{
		ID:       c.ID,
		DeckID:   c.DeckID,
		DeckName: c.DeckName,
		LexemeID: c.LexemeID,
		State:    stateName(c.State),
		Content:  c.Content,
		IsLeech:  c.IsLeech,
		Reps:     c.Reps,
		Lapses:   c.Lapses,
		Intervals: dto.IntervalsDto{
			Again: secs(entity.Again), Hard: secs(entity.Hard), Good: secs(entity.Good), Easy: secs(entity.Easy),
		},
	}, nil
}
