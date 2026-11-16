// Package noop is the notifier used when Telegram is not configured, in tests and with fake providers.
package noop

import (
	"context"

	"ielts-practice/app/modules/notify/domain/entity"
	notifyin "ielts-practice/app/modules/notify/ports/in"
)

var _ notifyin.IAdminNotifier = Notifier{}

type Notifier struct{}

// Notify sends nothing and returns notifyin.ErrDisabled, so no delivery is recorded.
func (Notifier) Notify(context.Context, entity.Notification) error { return notifyin.ErrDisabled }
