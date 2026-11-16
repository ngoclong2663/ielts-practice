// Package telegram posts admin notifications to a Telegram group with the Bot API (sendMessage).
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ielts-practice/app/modules/notify/domain/entity"
	notifyin "ielts-practice/app/modules/notify/ports/in"
)

const (
	DefaultBaseURL = "https://api.telegram.org"
	// MaxBodyRunes: the user text is cut to this many characters (Telegram allows 4096 in total).
	MaxBodyRunes = 1500
	timeout      = 10 * time.Second
)

var _ notifyin.IAdminNotifier = &Notifier{}

type Notifier struct {
	token    string
	chatID   string
	threadID int64
	baseURL  string
	client   *http.Client
}

// NewNotifier: baseURL is DefaultBaseURL in the app and an httptest server in tests.
func NewNotifier(token, chatID string, threadID int64, baseURL string) *Notifier {
	return &Notifier{
		token:    token,
		chatID:   chatID,
		threadID: threadID,
		baseURL:  strings.TrimRight(baseURL, "/"),
		client:   &http.Client{Timeout: timeout},
	}
}

type linkPreviewOptions struct {
	IsDisabled bool `json:"is_disabled"`
}

type sendMessageRequest struct {
	ChatID             string             `json:"chat_id"`
	MessageThreadID    int64              `json:"message_thread_id,omitempty"`
	Text               string             `json:"text"`
	ParseMode          string             `json:"parse_mode"`
	LinkPreviewOptions linkPreviewOptions `json:"link_preview_options"`
}

type apiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func (n *Notifier) Notify(ctx context.Context, msg entity.Notification) error {
	body, err := json.Marshal(sendMessageRequest{
		ChatID:             n.chatID,
		MessageThreadID:    n.threadID,
		Text:               Format(msg),
		ParseMode:          "HTML",
		LinkPreviewOptions: linkPreviewOptions{IsDisabled: true},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.baseURL+"/bot"+n.token+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return errors.New("telegram: cannot build request")
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := n.client.Do(req)
	if err != nil {
		// url.Error includes the URL, which holds the bot token: keep only the cause.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return fmt.Errorf("telegram: request failed: %w", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	var out apiResponse
	_ = json.Unmarshal(raw, &out)
	if res.StatusCode != http.StatusOK || !out.OK {
		return fmt.Errorf("telegram: sendMessage failed: http %d: %s", res.StatusCode, out.Description)
	}
	return nil
}

// Format renders the notification as Telegram HTML:
//
//	🐞 <b>New feedback · Something is broken</b>
//	From: Learner One (learner@example.com)
//	Page: /vi/decks  ·  vi
//
//	<the message, cut to 1500 characters>
//
//	Open: https://…/en/admin/feedback?id=…
func Format(msg entity.Notification) string {
	var b strings.Builder
	if msg.Emoji != "" {
		b.WriteString(escape(msg.Emoji))
		b.WriteString(" ")
	}
	b.WriteString("<b>")
	b.WriteString(escape(msg.Title))
	b.WriteString("</b>")
	for _, f := range msg.Fields {
		b.WriteString("\n")
		b.WriteString(escape(f.Label))
		b.WriteString(": ")
		b.WriteString(escape(f.Value))
	}
	if body := strings.TrimSpace(msg.Body); body != "" {
		b.WriteString("\n\n")
		b.WriteString(escape(truncate(body, MaxBodyRunes)))
	}
	if msg.Link != "" {
		b.WriteString("\n\nOpen: ")
		b.WriteString(escape(msg.Link))
	}
	return b.String()
}

// escape: Telegram's HTML mode needs only these three characters escaped.
func escape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// truncate cuts s to at most n characters (runes), ending with "…" when it was longer.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
