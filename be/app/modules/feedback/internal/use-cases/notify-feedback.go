package usecases

import (
	"context"
	"errors"

	"ielts-practice/app/modules/feedback/domain/entity"
	"ielts-practice/app/modules/feedback/domain/rules"
	notifyin "ielts-practice/app/modules/notify/ports/in"
)

// notifyInBackground posts the feedback to the owner (Telegram) after it is saved. Best effort:
// a failure is logged (never the message text) and leaves notified_at empty.
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

		var name, email string
		if user, err := uc.users.GetUserByID(ctx, fb.UserID); err != nil {
			log.Warn().Err(err).Msg("Feedback notification: failed to load the user")
		} else {
			name, email = user.DisplayName, user.Email
		}
		n := rules.BuildNotification(fb, name, email, uc.cfg.Server.FrontendURL)
		if err := uc.notifier.Notify(ctx, n); err != nil {
			if !errors.Is(err, notifyin.ErrDisabled) {
				log.Warn().Err(err).Msg("Feedback notification failed")
			}
			return
		}
		if err := uc.repo.MarkNotified(ctx, fb.ID, uc.now()); err != nil {
			log.Warn().Err(err).Msg("Feedback notification: failed to save notified_at")
		}
	}()
}
