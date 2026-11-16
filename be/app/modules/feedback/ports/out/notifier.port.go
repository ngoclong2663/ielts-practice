// Package out lists the external services the feedback module needs.
package out

import (
	"context"
	"errors"

	"ielts-practice/app/modules/feedback/domain/entity"
)

// ErrNotifierDisabled: owner notifications are off (not configured, tests or fake providers).
// Nothing was sent, so the feedback must not be marked as delivered.
var ErrNotifierDisabled = errors.New("feedback: owner notifications are off")

// IFeedbackNotifier tells the app owner about a new feedback (Epic 6 §4: Telegram today).
// fb.User holds the author when it could be loaded. Best effort: callers log errors.
type IFeedbackNotifier interface {
	NotifyNewFeedback(ctx context.Context, fb entity.Feedback) error
}
