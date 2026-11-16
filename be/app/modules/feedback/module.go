package feedback

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/feedback/infrastructure/persistence"
	feedbackInternal "ielts-practice/app/modules/feedback/internal"
	usecases "ielts-practice/app/modules/feedback/internal/use-cases"
)

// NewFeedbackModule: Epic 6 learner feedback and the admin feedback list.
func NewFeedbackModule() fx.Option {
	return fx.Options(
		fx.Provide(
			feedbackInternal.NewFeedbackRouter,
			feedbackInternal.NewFeedbackPresenter,
			usecases.NewFeedbackUseCases,
			func(uc *usecases.FeedbackUseCases) usecases.IFeedbackUseCases { return uc },
			persistence.NewFeedbackRepository,
		),
		fx.Invoke(
			feedbackInternal.RegisterRoutes,
			// Shutdown waits for background Telegram sends (bounded by the stop timeout).
			func(lc fx.Lifecycle, uc *usecases.FeedbackUseCases) {
				lc.Append(fx.Hook{OnStop: uc.Wait})
			},
		),
	)
}
