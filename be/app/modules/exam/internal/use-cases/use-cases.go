package usecases

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/exam/domain"
	"ielts-practice/app/modules/exam/domain/entity"
	"ielts-practice/app/modules/exam/dto"
	storagein "ielts-practice/app/modules/storage/ports/in"
	"ielts-practice/pkg/exceptions"
)

// IExamUseCases is what the presenter needs (Epic 10 §10).
type IExamUseCases interface {
	ListBooks(ctx context.Context, userID, skill string) ([]dto.BookDto, error)
	GetTest(ctx context.Context, slug, skill string) (*dto.TestDto, error)
	SubmitAttempt(ctx context.Context, userID string, req dto.SubmitAttemptRequestDto) (*dto.SubmitAttemptResponseDto, error)
	GetAttempt(ctx context.Context, userID, id string) (*dto.AttemptDto, error)
	ListAttempts(ctx context.Context, userID string, q dto.ListAttemptsQueryDto) (*dto.AttemptListDto, error)
	GetStats(ctx context.Context, userID, skill string) (*dto.StatsDto, error)
	ResolveMedia(ctx context.Context, file string) (string, error)
	ResolveFigure(name string) ([]byte, bool)
}

var _ IExamUseCases = &ExamUseCases{}

// ExamUseCases implements Epic 10: the Reading and Listening test bank, marking and history.
type ExamUseCases struct {
	repo    domain.IExamRepository
	tx      *database.TransactionStruct
	storage storagein.IObjectStorage
	logger  zerolog.Logger
}

func NewExamUseCases(repo domain.IExamRepository, tx *database.TransactionStruct, storage storagein.IObjectStorage, logger zerolog.Logger) *ExamUseCases {
	return &ExamUseCases{repo: repo, tx: tx, storage: storage, logger: logger.With().Str("module", "exam").Logger()}
}

func errTestNotFound() error {
	return exceptions.NewException(http.StatusNotFound, exceptions.CodeExamTestNotFound, "Test not found", nil)
}

func errAttemptNotFound() error {
	return exceptions.NewException(http.StatusNotFound, exceptions.CodeExamAttemptNotFound, "Attempt not found", nil)
}

func errSelection(msg string) error {
	return exceptions.NewException(http.StatusBadRequest, exceptions.CodeExamInvalidSelection, msg, nil)
}

func internal(msg string, err error) error {
	return exceptions.ErrInternalServerError(msg, err)
}

// partDto builds the player's view of a part. key is nil outside review.
func (uc *ExamUseCases) partDto(p *entity.Part, key entity.Key) dto.PartDto {
	d := dto.PartDto{PartNo: p.PartNo, Title: p.Title, QFrom: p.QFrom, QTo: p.QTo, QuestionTypes: p.QuestionTypes, Content: p.Content, Key: key}
	if d.QuestionTypes == nil {
		d.QuestionTypes = []string{}
	}
	if key == nil {
		// The transcript gives the answers away, so it is only sent with the key (after submitting).
		d.Content.Lines = nil
	}
	// Groups are copied so the figure URL never leaks into the stored content.
	d.Content.Groups = append([]entity.Group(nil), p.Content.Groups...)
	for i := range d.Content.Groups {
		if f := d.Content.Groups[i].Figure; f != "" {
			d.Content.Groups[i].Figure = FigureRoute + f
		}
	}
	if p.Audio != nil {
		cues := p.Audio.Cues
		if cues == nil {
			cues = []entity.Cue{}
		}
		d.Audio = &dto.AudioDto{URL: MediaURL(p.Audio.Key), Duration: p.Audio.Duration, Cues: cues}
	}
	return d
}

func (uc *ExamUseCases) testDto(ctx context.Context, t *entity.Test, skill string, parts []entity.Part, keys map[string]entity.Key) (*dto.TestDto, error) {
	book, err := uc.repo.FindBookByID(ctx, t.BookID)
	if err != nil {
		return nil, internal("Failed to load book", err)
	}
	if book == nil {
		return nil, errTestNotFound()
	}
	out := &dto.TestDto{Slug: t.Slug, TestNo: t.TestNo, Skill: skill, BookSlug: book.Slug, BookTitle: book.Title, Parts: []dto.PartDto{}}
	for i := range parts {
		var key entity.Key
		if keys != nil {
			key = keys[parts[i].ID]
			if key == nil {
				key = entity.Key{}
			}
		}
		out.Parts = append(out.Parts, uc.partDto(&parts[i], key))
	}
	return out, nil
}
