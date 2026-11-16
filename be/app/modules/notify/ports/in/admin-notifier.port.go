// Package in is what the notify module offers to other modules: messages to the app owner
// (Telegram today).
package in

import (
	"context"
	"errors"

	"ielts-practice/app/modules/notify/domain/entity"
)

// ErrDisabled: notifications are off (not configured, tests or fake providers). Nothing was sent,
// so callers must not record a delivery.
var ErrDisabled = errors.New("notify: admin notifications are off")

// IAdminNotifier sends one message to the owner. Callers treat it as best effort: an error is
// logged, never shown to the learner.
type IAdminNotifier interface {
	Notify(ctx context.Context, n entity.Notification) error
}
