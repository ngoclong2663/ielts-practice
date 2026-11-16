package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"ielts-practice/app/modules/notify/domain/entity"
)

type captured struct {
	path string
	body map[string]any
}

func newServer(t *testing.T, status int, reply string) (*httptest.Server, chan captured) {
	t.Helper()
	ch := make(chan captured, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		ch <- captured{path: r.URL.Path, body: body}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)
	return srv, ch
}

func TestNotifySendsHTMLMessage(t *testing.T) {
	srv, ch := newServer(t, http.StatusOK, `{"ok":true}`)
	n := NewNotifier("123:secret", "-1001234567890", 42, srv.URL+"/")

	long := strings.Repeat("é", 2000)
	err := n.Notify(context.Background(), entity.Notification{
		Emoji:  "🐞",
		Title:  "New feedback · Something is broken",
		Fields: []entity.Field{{Label: "From", Value: "Tom & <Jerry> (tom@example.com)"}},
		Body:   "<script>alert(1)</script> & more " + long,
		Link:   "https://app.example.com/en/admin/feedback?id=1&x=2",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := <-ch
	if got.path != "/bot123:secret/sendMessage" {
		t.Fatalf("path: %s", got.path)
	}
	if got.body["chat_id"] != "-1001234567890" || got.body["parse_mode"] != "HTML" || got.body["message_thread_id"].(float64) != 42 {
		t.Fatalf("body: %v", got.body)
	}
	text := got.body["text"].(string)
	for _, want := range []string{
		"🐞 <b>New feedback · Something is broken</b>",
		"From: Tom &amp; &lt;Jerry&gt; (tom@example.com)",
		"&lt;script&gt;alert(1)&lt;/script&gt; &amp; more",
		"Open: https://app.example.com/en/admin/feedback?id=1&amp;x=2",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("text misses %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "<script>") {
		t.Fatal("user text must be escaped")
	}
	// The body is cut to 1500 characters (before escaping), ending with an ellipsis.
	body := strings.Split(text, "\n\n")[1]
	unescaped := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">").Replace(body)
	if utf8.RuneCountInString(unescaped) != MaxBodyRunes || !strings.HasSuffix(unescaped, "…") {
		t.Fatalf("body not truncated to %d: %d", MaxBodyRunes, utf8.RuneCountInString(unescaped))
	}
}

func TestNotifyWithoutThreadAndErrors(t *testing.T) {
	srv, ch := newServer(t, http.StatusBadRequest, `{"ok":false,"description":"Bad Request: chat not found"}`)
	n := NewNotifier("123:secret", "-1", 0, srv.URL)
	err := n.Notify(context.Background(), entity.Notification{Title: "Hi", Body: "short"})
	got := <-ch
	if _, ok := got.body["message_thread_id"]; ok {
		t.Fatalf("message_thread_id must be omitted when unset: %v", got.body)
	}
	if err == nil || !strings.Contains(err.Error(), "chat not found") {
		t.Fatalf("expected API error, got %v", err)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatal("errors must not contain the bot token")
	}

	// Network errors must not leak the token either (url.Error holds the URL).
	srv.Close()
	err = NewNotifier("123:secret", "-1", 0, srv.URL).Notify(context.Background(), entity.Notification{Title: "Hi"})
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("expected a sanitised error, got %v", err)
	}
}
