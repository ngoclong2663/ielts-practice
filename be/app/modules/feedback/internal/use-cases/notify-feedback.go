package usecases

import (
	"context"
	"errors"

	"ielts-practice/app/modules/feedback/domain/entity"
	feedbackout "ielts-practice/app/modules/feedback/ports/out"
)

// notifyInBackground tells the owner about the feedback through ports/out.IFeedbackNotifier after
// it is saved. Best effort: a failure is logged (never the message text) and leaves notified_at empty.
func (uc *FeedbackUseCases) notifyInBackground(fb entity.Feedback) {
	uc.inFlight.Add(1)
	go func() {
		defer uc.inFlight.Done()
		log := uc.logger.With().Str("feedback_id", fb.ID).Logger()
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Msg("Feedback notification panicked")
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
		defer cancel()

		if user, err := uc.users.GetUserByID(ctx, fb.UserID); err != nil {
			log.Warn().Err(err).Msg("Feedback notification: failed to load the user")
		} else {
			fb.User = entity.FeedbackUser{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, AvatarURL: user.AvatarURL}
		}
		if err := uc.notifier.NotifyNewFeedback(ctx, fb); err != nil {
			if !errors.Is(err, feedbackout.ErrNotifierDisabled) {
				log.Warn().Err(err).Msg("Feedback notification failed")
			}
			return
		}
		if err := uc.repo.MarkNotified(ctx, fb.ID, uc.now()); err != nil {
			log.Warn().Err(err).Msg("Feedback notification: failed to save notified_at")
		}
	}()
}
