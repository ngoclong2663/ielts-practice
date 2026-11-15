// Package out lists what the writing module needs from outside: the AI coach that writes lessons and
// gives feedback. Adapters live in writing/infrastructure/coach.
package out

import (
	"context"
	"errors"

	"ielts-practice/app/modules/writing/domain/entity"
)

var (
	ErrCoachNotConfigured = errors.New("writing: no AI model is configured")
	ErrCoachFailed        = errors.New("writing: the AI answer was not usable")
)

// ICoach writes lessons and gives feedback (Claude first, Gemini fallback; a fake offline).
type ICoach interface {
	Configured() bool
	WriteLesson(ctx context.Context, req entity.LessonRequest) (entity.LessonContent, entity.AIUsage, error)
	CheckParagraph(ctx context.Context, req entity.CheckRequest) ([]entity.SentenceFeedback, entity.AIUsage, error)
	ReviewEssay(ctx context.Context, req entity.ReviewRequest) (entity.EssayFeedback, entity.AIUsage, error)
}
