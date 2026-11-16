package notifier

import (
	"context"
	"errors"
	"testing"

	"ielts-practice/app/modules/feedback/domain/entity"
	feedbackout "ielts-practice/app/modules/feedback/ports/out"
	notifyEntity "ielts-practice/app/modules/notify/domain/entity"
	notifyin "ielts-practice/app/modules/notify/ports/in"
)

func TestBuildNotification(t *testing.T) {
	fb := entity.Feedback{ID: "f1", Category: "bug", Message: "It broke", Page: "/vi/decks", UILocale: "vi",
		User: entity.FeedbackUser{DisplayName: "Learner One", Email: "learner@example.com"}}
	n := BuildNotification(fb, "https://app.example.com/")
	if n.Title != "New feedback · Something is broken" || n.Emoji != "🐞" {
		t.Fatalf("title: %q %q", n.Emoji, n.Title)
	}
	if n.Fields[0].Value != "Learner One (learner@example.com)" || n.Fields[1].Value != "/vi/decks  ·  vi" {
		t.Fatalf("fields: %v", n.Fields)
	}
	if n.Link != "https://app.example.com/en/admin/feedback?id=f1" || n.Body != "It broke" {
		t.Fatalf("link/body: %q %q", n.Link, n.Body)
	}
}

func TestBuildNotificationWithoutAuthor(t *testing.T) {
	n := BuildNotification(entity.Feedback{ID: "f2", UserID: "u1", Category: "other", Message: "Hi there"}, "https://x")
	if n.Fields[0].Value != "u1" || len(n.Fields) != 1 {
		t.Fatalf("fields: %v", n.Fields)
	}
}

type stubAdmin struct{ err error }

func (s stubAdmin) Notify(context.Context, notifyEntity.Notification) error { return s.err }

func TestDisabledChannelMapsToPortError(t *testing.T) {
	n := &Notifier{admin: stubAdmin{err: notifyin.ErrDisabled}}
	if err := n.NotifyNewFeedback(context.Background(), entity.Feedback{}); !errors.Is(err, feedbackout.ErrNotifierDisabled) {
		t.Fatalf("want ErrNotifierDisabled, got %v", err)
	}
	boom := errors.New("telegram down")
	n = &Notifier{admin: stubAdmin{err: boom}}
	if err := n.NotifyNewFeedback(context.Background(), entity.Feedback{}); !errors.Is(err, boom) {
		t.Fatalf("want the channel error, got %v", err)
	}
}
