package integration

// Epic 6 integration tests: learner feedback (+ Telegram notification through a fake notifier),
// the admin account seed and sign-in, and the admin board endpoints.

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	adminin "ielts-practice/app/modules/admin/ports/in"
	notifyEntity "ielts-practice/app/modules/notify/domain/entity"
	notifyin "ielts-practice/app/modules/notify/ports/in"
)

const (
	testAdminUser = "owner"
	testAdminPass = "correct-horse-battery-staple"
)

// fakeNotifier replaces the shared notify channel (Telegram), so the feedback module's own adapter
// (feedback/infrastructure/notifier) still runs and its message is checked. The channel makes the
// background send deterministic.
type fakeNotifier struct {
	calls chan notifyEntity.Notification
}

func (f *fakeNotifier) Notify(_ context.Context, n notifyEntity.Notification) error {
	select {
	case f.calls <- n:
	default:
	}
	return nil
}

type adminHarness struct {
	*harness
	seeder   adminin.IAdminSeeder
	notifier *fakeNotifier
}

// newAdminHarness starts the app with the fake notifier and an empty admin_accounts table.
func newAdminHarness(t *testing.T) *adminHarness {
	t.Helper()
	t.Setenv("FRONTEND_URL", "https://app.example.com")
	ah := &adminHarness{notifier: &fakeNotifier{calls: make(chan notifyEntity.Notification, 50)}}
	ah.harness = newHarness(t,
		fx.Decorate(func(notifyin.IAdminNotifier) notifyin.IAdminNotifier { return ah.notifier }),
		fx.Populate(&ah.seeder),
	)
	if _, err := testDB.Exec("TRUNCATE admin_accounts"); err != nil {
		t.Fatal(err)
	}
	return ah
}

func bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// adminToken seeds the admin account and signs in.
func (ah *adminHarness) adminToken() string {
	ah.t.Helper()
	if _, err := ah.seeder.SeedAdmin(context.Background(), testAdminUser, testAdminPass); err != nil {
		ah.t.Fatal(err)
	}
	r := ah.mustStatus(ah.do("POST", "/api/v1/admin/auth/login", map[string]any{"username": testAdminUser, "password": testAdminPass}, ""), 200, "admin login")
	return r.Body["access_token"].(string)
}

func (ah *adminHarness) admin(method, path string, body any, token string) resp {
	ah.t.Helper()
	return ah.doWithHeaders(method, path, body, "", bearer(token))
}

func (ah *adminHarness) sendFeedback(access string, body map[string]any) resp {
	ah.t.Helper()
	return ah.doWithHeaders("POST", "/api/v1/feedback", body, access, map[string]string{"User-Agent": "TestAgent/1.0 (iPhone)"})
}

func fieldCode(r resp, field string) string {
	e, _ := r.Body["error"].(map[string]any)
	fields, _ := e["fields"].(map[string]any)
	code, _ := fields[field].(string)
	return code
}

func items(r resp) []map[string]any {
	raw, _ := r.Body["items"].([]any)
	out := make([]map[string]any, len(raw))
	for i, it := range raw {
		out[i] = it.(map[string]any)
	}
	return out
}

func num(m map[string]any, key string) int {
	v, ok := m[key].(float64)
	if !ok {
		panic("not a number: " + key)
	}
	return int(v)
}

func adminPasswordHash(t *testing.T) string {
	t.Helper()
	var hash string
	if err := testDB.QueryRow("SELECT password_hash FROM admin_accounts WHERE lower(username) = 'owner'").Scan(&hash); err != nil {
		t.Fatal(err)
	}
	return hash
}

func TestAdminSeedAndLogin(t *testing.T) {
	ah := newAdminHarness(t)
	ctx := context.Background()

	if res, err := ah.seeder.SeedAdmin(ctx, "", testAdminPass); err != nil || res != adminin.SeedAdminSkipped {
		t.Fatalf("missing username should skip: %q %v", res, err)
	}
	if res, err := ah.seeder.SeedAdmin(ctx, testAdminUser, testAdminPass); err != nil || res != adminin.SeedAdminCreated {
		t.Fatalf("first seed: %q %v", res, err)
	}
	hash1 := adminPasswordHash(t)
	if !strings.HasPrefix(hash1, "$2a$12$") || strings.Contains(hash1, testAdminPass) {
		t.Fatalf("password must be stored as a bcrypt cost-12 hash: %s", hash1)
	}
	if res, err := ah.seeder.SeedAdmin(ctx, "OWNER", testAdminPass); err != nil || res != adminin.SeedAdminUnchanged {
		t.Fatalf("re-seed (case-insensitive) should change nothing: %q %v", res, err)
	}
	if adminPasswordHash(t) != hash1 {
		t.Fatal("re-running the seed must keep the hash")
	}

	// Wrong credentials: same error for a wrong password and an unknown user.
	for _, body := range []map[string]any{
		{"username": testAdminUser, "password": "wrong-password"},
		{"username": "nobody", "password": testAdminPass},
	} {
		if r := ah.do("POST", "/api/v1/admin/auth/login", body, ""); r.Status != 401 || errCode(r) != "ADMIN_INVALID_CREDENTIALS" {
			t.Fatalf("expected 401 ADMIN_INVALID_CREDENTIALS: %d %v", r.Status, r.Body)
		}
	}
	if r := ah.do("POST", "/api/v1/admin/auth/login", map[string]any{"username": testAdminUser}, ""); r.Status != 400 {
		t.Fatalf("password is required: %d", r.Status)
	}

	r := ah.mustStatus(ah.do("POST", "/api/v1/admin/auth/login", map[string]any{"username": "Owner", "password": testAdminPass}, ""), 200, "login")
	token, _ := r.Body["access_token"].(string)
	admin := r.Body["admin"].(map[string]any)
	if token == "" || admin["username"] != testAdminUser || admin["id"] == "" {
		t.Fatalf("login response: %v", r.Body)
	}
	if exp, err := time.Parse(time.RFC3339, r.Body["expires_at"].(string)); err != nil || exp.Sub(time.Now()) < 11*time.Hour {
		t.Fatalf("expires_at should be RFC3339, ~12 h ahead: %v %v", r.Body["expires_at"], err)
	}
	if len(r.Body) != 3 {
		t.Fatalf("unexpected fields: %v", r.Body)
	}
	var lastLogin *time.Time
	_ = testDB.QueryRow("SELECT last_login_at FROM admin_accounts").Scan(&lastLogin)
	if lastLogin == nil {
		t.Fatal("login must set last_login_at")
	}

	me := ah.mustStatus(ah.admin("GET", "/api/v1/admin/auth/me", nil, token), 200, "me")
	if me.Body["id"] != admin["id"] || me.Body["username"] != testAdminUser {
		t.Fatalf("me: %v", me.Body)
	}
	// The browser goes through the Next.js rewrite with the admin_token cookie.
	cookie := ah.doWithHeaders("GET", "/api/v1/admin/auth/me", nil, "", map[string]string{"Cookie": "admin_token=" + token})
	ah.mustStatus(cookie, 200, "me with cookie")

	// A changed ADMIN_PASSWORD rotates the hash; the old password stops working.
	if res, err := ah.seeder.SeedAdmin(ctx, testAdminUser, "a-brand-new-password"); err != nil || res != adminin.SeedAdminUpdated {
		t.Fatalf("password change: %q %v", res, err)
	}
	if adminPasswordHash(t) == hash1 {
		t.Fatal("hash must change with the password")
	}
	if r := ah.do("POST", "/api/v1/admin/auth/login", map[string]any{"username": testAdminUser, "password": testAdminPass}, ""); r.Status != 401 {
		t.Fatalf("old password must fail: %d", r.Status)
	}
	ah.mustStatus(ah.do("POST", "/api/v1/admin/auth/login", map[string]any{"username": testAdminUser, "password": "a-brand-new-password"}, ""), 200, "new password")

	// Deleting the account revokes its tokens at once.
	if _, err := testDB.Exec("DELETE FROM admin_accounts"); err != nil {
		t.Fatal(err)
	}
	if r := ah.admin("GET", "/api/v1/admin/auth/me", nil, token); r.Status != 401 || errCode(r) != "ADMIN_UNAUTHORIZED" {
		t.Fatalf("deleted account must be rejected: %d %v", r.Status, r.Body)
	}
}

func TestAdminRoutesNeedAdminToken(t *testing.T) {
	ah := newAdminHarness(t)
	access, _ := ah.signIn()
	token := ah.adminToken()

	for _, path := range []string{"/api/v1/admin/overview", "/api/v1/admin/users", "/api/v1/admin/feedback", "/api/v1/admin/auth/me"} {
		if r := ah.do("GET", path, nil, ""); r.Status != 401 || errCode(r) != "ADMIN_UNAUTHORIZED" {
			t.Fatalf("%s without token: %d %v", path, r.Status, r.Body)
		}
		// A learner token (header or cookie) never opens the admin API.
		if r := ah.doWithHeaders("GET", path, nil, access, bearer(access)); r.Status != 401 || errCode(r) != "ADMIN_UNAUTHORIZED" {
			t.Fatalf("%s with learner token: %d %v", path, r.Status, r.Body)
		}
		if r := ah.admin("GET", path, nil, "garbage"); r.Status != 401 {
			t.Fatalf("%s with garbage token: %d", path, r.Status)
		}
		ah.mustStatus(ah.admin("GET", path, nil, token), 200, path+" with admin token")
	}
	if r := ah.admin("PATCH", "/api/v1/admin/feedback/"+uuid.NewString(), map[string]any{"status": "seen"}, access); r.Status != 401 {
		t.Fatalf("PATCH with learner token: %d", r.Status)
	}
	// And an admin token is not a learner session.
	if r := ah.doWithHeaders("GET", "/api/v1/me", nil, "", bearer(token)); r.Status != 401 {
		t.Fatalf("admin token must not open learner routes: %d", r.Status)
	}
}

func TestFeedbackSendLimitAndNotify(t *testing.T) {
	t.Setenv("FEEDBACK_DAILY_LIMIT", "2")
	ah := newAdminHarness(t)
	access, _ := ah.signIn()
	token := ah.adminToken()

	if r := ah.do("POST", "/api/v1/feedback", map[string]any{"category": "bug", "message": "The audio does not play"}, ""); r.Status != 401 {
		t.Fatalf("feedback needs a learner session: %d", r.Status)
	}
	// Validation.
	cases := []struct {
		body  map[string]any
		field string
		code  string
	}{
		{map[string]any{"category": "bug", "message": "   too short   "}, "message", "VALIDATION_MIN"},
		{map[string]any{"category": "bug", "message": strings.Repeat("x", 2001)}, "message", "VALIDATION_MAX"},
		{map[string]any{"category": "bug"}, "message", "VALIDATION_REQUIRED"},
		{map[string]any{"category": "praise", "message": "A long enough message"}, "category", "VALIDATION_ONEOF"},
		{map[string]any{"category": "bug", "message": "A long enough message", "ui_locale": "fr"}, "ui_locale", "VALIDATION_ONEOF"},
		{map[string]any{"category": "bug", "message": "A long enough message", "page": "/" + strings.Repeat("p", 300)}, "page", "VALIDATION_MAX"},
	}
	for _, c := range cases {
		r := ah.sendFeedback(access, c.body)
		if r.Status != 400 || errCode(r) != "VALIDATION_FAILED" || fieldCode(r, c.field) != c.code {
			t.Fatalf("%v: want %s=%s, got %d %v", c.body, c.field, c.code, r.Status, r.Body)
		}
	}

	r := ah.mustStatus(ah.sendFeedback(access, map[string]any{
		"category": "bug", "message": "  The audio button does not play on iPhone <b>  ", "page": "/vi/decks/123", "ui_locale": "vi",
	}), http.StatusCreated, "send feedback")
	id, _ := r.Body["id"].(string)
	if _, err := uuid.Parse(id); err != nil || r.Body["created_at"] == nil || len(r.Body) != 2 {
		t.Fatalf("201 body should be {id, created_at}: %v", r.Body)
	}

	// The owner is notified in the background.
	select {
	case n := <-ah.notifier.calls:
		if n.Title != "New feedback · Something is broken" || n.Body != "The audio button does not play on iPhone <b>" {
			t.Fatalf("notification: %+v", n)
		}
		if n.Link != "https://app.example.com/en/admin/feedback?id="+id {
			t.Fatalf("link: %s", n.Link)
		}
		joined := ""
		for _, f := range n.Fields {
			joined += f.Label + ": " + f.Value + "\n"
		}
		if !strings.Contains(joined, "From: Learner One (learner@example.com)") || !strings.Contains(joined, "Page: /vi/decks/123  ·  vi") {
			t.Fatalf("fields: %s", joined)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("notifier was not called")
	}

	// notified_at is saved right after the send.
	var item map[string]any
	deadline := time.Now().Add(5 * time.Second)
	for {
		list := ah.mustStatus(ah.admin("GET", "/api/v1/admin/feedback", nil, token), 200, "list feedback")
		if its := items(list); len(its) == 1 {
			item = its[0]
		}
		if item != nil && item["notified"] == true {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("feedback never marked notified: %v", item)
		}
		time.Sleep(20 * time.Millisecond)
	}
	user := item["user"].(map[string]any)
	if item["id"] != id || item["category"] != "bug" || item["page"] != "/vi/decks/123" || item["ui_locale"] != "vi" ||
		item["user_agent"] != "TestAgent/1.0 (iPhone)" || item["status"] != "new" || item["admin_note"] != "" ||
		item["resolved_at"] != nil || user["email"] != "learner@example.com" || user["display_name"] != "Learner One" {
		t.Fatalf("feedback item: %v", item)
	}

	// Daily limit (2 in this test, rolling 24 h).
	ah.mustStatus(ah.sendFeedback(access, map[string]any{"category": "idea", "message": "Please add dark mode"}), 201, "second feedback")
	r = ah.sendFeedback(access, map[string]any{"category": "other", "message": "One message too many"})
	if r.Status != 429 || errCode(r) != "FEEDBACK_LIMIT" {
		t.Fatalf("expected 429 FEEDBACK_LIMIT: %d %v", r.Status, r.Body)
	}
	// Older than 24 h no longer counts.
	if _, err := testDB.Exec("UPDATE feedback SET created_at = now() - interval '25 hours'"); err != nil {
		t.Fatal(err)
	}
	ah.mustStatus(ah.sendFeedback(access, map[string]any{"category": "other", "message": "A day later, one more"}), 201, "after 24 h")
}

func TestAdminFeedbackStatusFlow(t *testing.T) {
	ah := newAdminHarness(t)
	access, _ := ah.signIn()
	token := ah.adminToken()

	bugID := ah.mustStatus(ah.sendFeedback(access, map[string]any{"category": "bug", "message": "Something broke here"}), 201, "bug").Body["id"].(string)
	time.Sleep(5 * time.Millisecond) // distinct created_at for the order check
	ideaID := ah.mustStatus(ah.sendFeedback(access, map[string]any{"category": "idea", "message": "A small idea for you"}), 201, "idea").Body["id"].(string)

	list := ah.mustStatus(ah.admin("GET", "/api/v1/admin/feedback", nil, token), 200, "list")
	its := items(list)
	if len(its) != 2 || its[0]["id"] != ideaID || its[1]["id"] != bugID || num(list.Body, "total") != 2 ||
		num(list.Body, "page") != 1 || num(list.Body, "page_size") != 25 {
		t.Fatalf("default list (status new, newest first): %v", list.Body)
	}
	byCategory := ah.mustStatus(ah.admin("GET", "/api/v1/admin/feedback?category=bug&status=all", nil, token), 200, "by category")
	if num(byCategory.Body, "total") != 1 || num(byCategory.Body["counts"].(map[string]any), "new") != 1 {
		t.Fatalf("category filter: %v", byCategory.Body)
	}

	seen := ah.mustStatus(ah.admin("PATCH", "/api/v1/admin/feedback/"+ideaID, map[string]any{"status": "seen"}, token), 200, "seen")
	if seen.Body["status"] != "seen" || seen.Body["resolved_at"] != nil || seen.Body["id"] != ideaID {
		t.Fatalf("seen: %v", seen.Body)
	}
	resolved := ah.mustStatus(ah.admin("PATCH", "/api/v1/admin/feedback/"+bugID, map[string]any{"status": "resolved", "admin_note": "Fixed in v2"}, token), 200, "resolve")
	if resolved.Body["status"] != "resolved" || resolved.Body["resolved_at"] == nil || resolved.Body["admin_note"] != "Fixed in v2" {
		t.Fatalf("resolved: %v", resolved.Body)
	}
	if resolved.Body["user"].(map[string]any)["email"] != "learner@example.com" {
		t.Fatalf("PATCH returns the full item: %v", resolved.Body)
	}

	for _, status := range []string{"all", "new"} {
		l := ah.mustStatus(ah.admin("GET", "/api/v1/admin/feedback?status="+status, nil, token), 200, "list "+status)
		c := l.Body["counts"].(map[string]any)
		if num(c, "new") != 0 || num(c, "seen") != 1 || num(c, "resolved") != 1 {
			t.Fatalf("counts ignore the status filter (%s): %v", status, c)
		}
		want := map[string]int{"all": 2, "new": 0}[status]
		if num(l.Body, "total") != want || len(items(l)) != want {
			t.Fatalf("status=%s: %v", status, l.Body)
		}
	}
	if l := ah.mustStatus(ah.admin("GET", "/api/v1/admin/feedback?status=resolved&page_size=1", nil, token), 200, "resolved"); len(items(l)) != 1 || items(l)[0]["id"] != bugID {
		t.Fatalf("status=resolved: %v", l.Body)
	}

	// Reopen clears resolved_at; the note stays unless sent.
	reopened := ah.mustStatus(ah.admin("PATCH", "/api/v1/admin/feedback/"+bugID, map[string]any{"status": "new"}, token), 200, "reopen")
	if reopened.Body["resolved_at"] != nil || reopened.Body["admin_note"] != "Fixed in v2" || reopened.Body["status"] != "new" {
		t.Fatalf("reopen: %v", reopened.Body)
	}

	if r := ah.admin("PATCH", "/api/v1/admin/feedback/"+bugID, map[string]any{"admin_note": strings.Repeat("n", 1001)}, token); r.Status != 400 || fieldCode(r, "admin_note") != "VALIDATION_MAX" {
		t.Fatalf("note too long: %d %v", r.Status, r.Body)
	}
	if r := ah.admin("PATCH", "/api/v1/admin/feedback/"+bugID, map[string]any{"status": "done"}, token); r.Status != 400 {
		t.Fatalf("bad status: %d", r.Status)
	}
	if r := ah.admin("GET", "/api/v1/admin/feedback?page_size=101", nil, token); r.Status != 400 {
		t.Fatalf("page_size max is 100: %d", r.Status)
	}
	for _, id := range []string{uuid.NewString(), "not-a-uuid"} {
		if r := ah.admin("PATCH", "/api/v1/admin/feedback/"+id, map[string]any{"status": "seen"}, token); r.Status != 404 || errCode(r) != "FEEDBACK_NOT_FOUND" {
			t.Fatalf("unknown id %s: %d %v", id, r.Status, r.Body)
		}
	}
}

func TestAdminOverviewAndUsers(t *testing.T) {
	ah := newAdminHarness(t)
	access, _ := ah.signIn()
	token := ah.adminToken()

	// Learner One: a deck, two cards, one review, one feedback.
	deckID := ah.createDeck(access, "IELTS Day 1")
	cardID := ah.createCard(access, deckID, "resilient", "kiên cường")
	ah.createCard(access, deckID, "decide", "quyết định")
	ah.answer(access, deckID, cardID, 3)
	ah.mustStatus(ah.sendFeedback(access, map[string]any{"category": "content", "message": "The meaning of resilient is wrong"}), 201, "feedback")

	var learnerID, lexemeID string
	_ = testDB.QueryRow("SELECT id FROM users WHERE email = 'learner@example.com'").Scan(&learnerID)
	if err := testDB.QueryRow(`INSERT INTO lexemes (key, headword, status) VALUES ('resilient||vi', 'resilient', 'needs_review') RETURNING id`).Scan(&lexemeID); err != nil {
		t.Fatal(err)
	}
	if _, err := testDB.Exec(`INSERT INTO lexeme_reports (lexeme_id, user_id, reason, note) VALUES ($1, $2, 'wrong_meaning', 'not this sense')`, lexemeID, learnerID); err != nil {
		t.Fatal(err)
	}

	// A second user who only signed in.
	ah.google.identity.Subject = "google-sub-2"
	ah.google.identity.Email = "second@example.com"
	ah.google.identity.Name = "Second Person"
	ah.signIn()

	o := ah.mustStatus(ah.admin("GET", "/api/v1/admin/overview", nil, token), 200, "overview").Body
	users := o["users"].(map[string]any)
	if num(users, "total") != 2 || num(users, "new_7d") != 2 || num(users, "new_30d") != 2 ||
		num(users, "active_1d") != 2 || num(users, "active_7d") != 2 || num(users, "active_30d") != 2 {
		t.Fatalf("users: %v", users)
	}
	vocab := o["vocab"].(map[string]any)
	states := vocab["cards_by_state"].(map[string]any)
	if num(vocab, "decks") != 1 || num(vocab, "cards") != 2 || num(vocab, "reviews_total") != 1 || num(vocab, "reviews_7d") != 1 ||
		num(states, "new") != 1 || num(states, "learning") != 1 || num(states, "review") != 0 || num(states, "relearning") != 0 {
		t.Fatalf("vocab: %v", vocab)
	}
	if fb := o["feedback"].(map[string]any); num(fb, "new") != 1 || num(fb, "total") != 1 {
		t.Fatalf("feedback: %v", fb)
	}
	if cr := o["content_reports"].(map[string]any); num(cr, "total") != 1 || num(cr, "last_7d") != 1 || num(cr, "needs_review") != 1 {
		t.Fatalf("content_reports: %v", cr)
	}
	writing := o["writing"].(map[string]any)
	if num(writing, "attempts_total") != 0 || writing["avg_band_30d"] != nil {
		t.Fatalf("writing: %v", writing)
	}
	ai := o["ai"].(map[string]any)
	for _, k := range []string{"words_generated_total", "words_generated_7d", "words_cached_total", "words_failed_7d", "topic_lists_total",
		"topic_lists_7d", "writing_lessons_total", "writing_checks_total", "writing_checks_7d", "writing_reviews_total",
		"writing_reviews_7d", "input_tokens_30d", "output_tokens_30d"} {
		if num(ai, k) != 0 {
			t.Fatalf("ai.%s: %v", k, ai)
		}
	}
	if o["timezone"] != "Asia/Ho_Chi_Minh" {
		t.Fatalf("timezone: %v", o["timezone"])
	}
	if _, err := time.Parse(time.RFC3339, o["generated_at"].(string)); err != nil {
		t.Fatalf("generated_at: %v", o["generated_at"])
	}
	daily := o["daily"].([]any)
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	today := daily[len(daily)-1].(map[string]any)
	first := daily[0].(map[string]any)
	if len(daily) != 30 || today["date"] != time.Now().In(loc).Format("2006-01-02") ||
		first["date"] != time.Now().In(loc).AddDate(0, 0, -29).Format("2006-01-02") {
		t.Fatalf("daily: %d entries, %v … %v", len(daily), first, today)
	}
	if num(today, "signups") != 2 || num(today, "active_users") != 2 || num(today, "reviews") != 1 || num(today, "ai_requests") != 0 ||
		num(first, "signups") != 0 {
		t.Fatalf("today: %v", today)
	}
	reports := o["recent_reports"].([]any)
	if len(reports) != 1 {
		t.Fatalf("recent_reports: %v", reports)
	}
	rep := reports[0].(map[string]any)
	if rep["headword"] != "resilient" || rep["reason"] != "wrong_meaning" || rep["note"] != "not this sense" || rep["user_email"] != "learner@example.com" {
		t.Fatalf("report: %v", rep)
	}

	// Users list.
	list := ah.mustStatus(ah.admin("GET", "/api/v1/admin/users", nil, token), 200, "users")
	if num(list.Body, "total") != 2 || num(list.Body, "page") != 1 || num(list.Body, "page_size") != 25 || len(items(list)) != 2 {
		t.Fatalf("users list: %v", list.Body)
	}
	var learner map[string]any
	for _, it := range items(list) {
		if it["email"] == "learner@example.com" {
			learner = it
		}
	}
	if learner == nil || learner["display_name"] != "Learner One" || learner["ui_locale"] != "vi" || learner["last_active_at"] == nil ||
		learner["last_login_at"] == nil || num(learner, "decks") != 1 || num(learner, "cards") != 2 || num(learner, "reviews_total") != 1 ||
		num(learner, "reviews_7d") != 1 || num(learner, "ai_words") != 0 || num(learner, "topic_lists") != 0 ||
		num(learner, "writing_attempts") != 0 || num(learner, "writing_submitted") != 0 || num(learner, "writing_ai_calls") != 0 ||
		num(learner, "feedback") != 1 {
		t.Fatalf("learner row: %v", learner)
	}

	if l := ah.mustStatus(ah.admin("GET", "/api/v1/admin/users?sort=reviews", nil, token), 200, "sort reviews"); items(l)[0]["email"] != "learner@example.com" {
		t.Fatalf("sort=reviews: %v", items(l))
	}
	if l := ah.mustStatus(ah.admin("GET", "/api/v1/admin/users?sort=created", nil, token), 200, "sort created"); items(l)[0]["email"] != "second@example.com" {
		t.Fatalf("sort=created (newest first): %v", items(l))
	}
	ah.mustStatus(ah.admin("GET", "/api/v1/admin/users?sort=ai", nil, token), 200, "sort ai")
	if l := ah.mustStatus(ah.admin("GET", "/api/v1/admin/users?q=SECOND", nil, token), 200, "search"); num(l.Body, "total") != 1 || items(l)[0]["email"] != "second@example.com" {
		t.Fatalf("search by email: %v", l.Body)
	}
	if l := ah.mustStatus(ah.admin("GET", "/api/v1/admin/users?q=learner%20one", nil, token), 200, "search name"); num(l.Body, "total") != 1 {
		t.Fatalf("search by name: %v", l.Body)
	}
	if l := ah.mustStatus(ah.admin("GET", "/api/v1/admin/users?q=%25", nil, token), 200, "search %"); num(l.Body, "total") != 0 {
		t.Fatalf("%% must be literal: %v", l.Body)
	}
	if l := ah.mustStatus(ah.admin("GET", "/api/v1/admin/users?page=2&page_size=1", nil, token), 200, "page 2"); num(l.Body, "total") != 2 || len(items(l)) != 1 {
		t.Fatalf("paging: %v", l.Body)
	}
	if r := ah.admin("GET", "/api/v1/admin/users?sort=bogus", nil, token); r.Status != 400 {
		t.Fatalf("bad sort: %d", r.Status)
	}
	if r := ah.admin("GET", "/api/v1/admin/users?page_size=101", nil, token); r.Status != 400 {
		t.Fatalf("page_size max is 100: %d", r.Status)
	}

	// Deleted users disappear from the board.
	if _, err := testDB.Exec("UPDATE users SET deleted_at = now() WHERE email = 'second@example.com'"); err != nil {
		t.Fatal(err)
	}
	if l := ah.mustStatus(ah.admin("GET", "/api/v1/admin/users", nil, token), 200, "after delete"); num(l.Body, "total") != 1 {
		t.Fatalf("deleted users must be excluded: %v", l.Body)
	}
	o = ah.mustStatus(ah.admin("GET", "/api/v1/admin/overview", nil, token), 200, "overview").Body
	if u := o["users"].(map[string]any); num(u, "total") != 1 || num(u, "active_1d") != 1 {
		t.Fatalf("deleted users must be excluded from the overview: %v", u)
	}
}
