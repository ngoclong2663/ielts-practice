// Package notifier implements feedback/ports/out.IFeedbackNotifier on top of the shared notify
// module: it turns a feedback into the owner's message (Epic 6 §4) and hands it to the channel.
package notifier

import (
	"context"
	"errors"
	"strings"

	"ielts-practice/app/modules/feedback/domain/entity"
	feedbackout "ielts-practice/app/modules/feedback/ports/out"
	notifyEntity "ielts-practice/app/modules/notify/domain/entity"
	notifyin "ielts-practice/app/modules/notify/ports/in"
	"ielts-practice/configs"
)

var _ feedbackout.IFeedbackNotifier = &Notifier{}

// English labels: the owner reads the board and the messages in English.
var categoryLabels = map[string]string{
	entity.CategoryBug:     "Something is broken",
	entity.CategoryIdea:    "An idea",
	entity.CategoryContent: "Wrong word content",
	entity.CategoryOther:   "Other",
}

var categoryEmoji = map[string]string{
	entity.CategoryBug:     "🐞",
	entity.CategoryIdea:    "💡",
	entity.CategoryContent: "📝",
	entity.CategoryOther:   "💬",
}

type Notifier struct {
	admin       notifyin.IAdminNotifier
	frontendURL string
}

func NewFeedbackNotifier(admin notifyin.IAdminNotifier, cfg *configs.Config) feedbackout.IFeedbackNotifier {
	return &Notifier{admin: admin, frontendURL: cfg.Server.FrontendURL}
}

func (n *Notifier) NotifyNewFeedback(ctx context.Context, fb entity.Feedback) error {
	err := n.admin.Notify(ctx, BuildNotification(fb, n.frontendURL))
	if errors.Is(err, notifyin.ErrDisabled) {
		return feedbackout.ErrNotifierDisabled
	}
	return err
}

func categoryLabel(category string) string {
	if l, ok := categoryLabels[category]; ok {
		return l
	}
	return category
}

// AdminLink is the board link to one feedback.
func AdminLink(frontendURL, feedbackID string) string {
	return strings.TrimRight(frontendURL, "/") + "/en/admin/feedback?id=" + feedbackID
}

// BuildNotification is the owner's message for a new feedback.
func BuildNotification(fb entity.Feedback, frontendURL string) notifyEntity.Notification {
	name, email := fb.User.DisplayName, fb.User.Email
	from := email
	if name != "" && email != "" {
		from = name + " (" + email + ")"
	} else if name != "" {
		from = name
	}
	if from == "" {
		from = fb.UserID
	}
	var where []string
	if fb.Page != "" {
		where = append(where, fb.Page)
	}
	if fb.UILocale != "" {
		where = append(where, fb.UILocale)
	}
	fields := []notifyEntity.Field{{Label: "From", Value: from}}
	if len(where) > 0 {
		fields = append(fields, notifyEntity.Field{Label: "Page", Value: strings.Join(where, "  ·  ")})
	}
	return notifyEntity.Notification{
		Emoji:  categoryEmoji[fb.Category],
		Title:  "New feedback · " + categoryLabel(fb.Category),
		Fields: fields,
		Body:   fb.Message,
		Link:   AdminLink(frontendURL, fb.ID),
	}
}
