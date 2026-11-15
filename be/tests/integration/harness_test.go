// Package integration holds the API integration tests: real PostgreSQL (embedded, downloaded on
// first run), the real HTTP stack and modules (app.CoreOptions), Google replaced by a fake identity
// provider. Skipped with `go test -short` or SKIP_INTEGRATION=1.
package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"

	authout "ielts-practice/app/modules/auth/ports/out"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"ielts-practice/app"
	"ielts-practice/app/modules/auth/domain/entity"
)

const testDSN = "postgres://postgres:postgres@localhost:54329/ielts_test?sslmode=disable"

var testDB *sql.DB

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() || os.Getenv("SKIP_INTEGRATION") == "1" {
		os.Exit(m.Run())
	}
	tmp, _ := os.MkdirTemp("", "ielts-pg-*")
	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(54329).
		Database("ielts_test").
		RuntimePath(filepath.Join(tmp, "runtime")).
		DataPath(filepath.Join(tmp, "data")).
		Logger(nil))
	if err := pg.Start(); err != nil {
		fmt.Println("embedded postgres failed to start:", err)
		os.Exit(1)
	}
	code := func() int {
		defer func() { _ = pg.Stop(); _ = os.RemoveAll(tmp) }()
		var err error
		testDB, err = sql.Open("pgx", testDSN)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		if err := applyMigrations(testDB); err != nil {
			fmt.Println("migrations failed:", err)
			return 1
		}
		return m.Run()
	}()
	os.Exit(code)
}

func applyMigrations(db *sql.DB) error {
	files, _ := filepath.Glob("../../migrations/*.up.sql")
	sort.Strings(files)
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(b)); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
	}
	return nil
}

// ---- fakes & helpers -------------------------------------------------------

type fakeGoogle struct{ identity *entity.GoogleIdentity }

func (f *fakeGoogle) Exchange(_ context.Context, code, _ string) (*entity.GoogleIdentity, error) {
	id := *f.identity
	return &id, nil
}

type harness struct {
	t      *testing.T
	engine *gin.Engine
	google *fakeGoogle
}

func newHarness(t *testing.T, extra ...fx.Option) *harness {
	if testDB == nil {
		t.Skip("integration tests disabled")
	}
	_, err := testDB.Exec("TRUNCATE refresh_tokens, user_study_settings, users, lexemes, image_links CASCADE")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", testDSN)
	t.Setenv("JWT_SECRET", "integration-secret-integration-secret")
	t.Setenv("RATE_LIMIT_DISABLED", "true")
	t.Setenv("S3_BUCKET_NAME", "") // keep tests off the real bucket (be/.env is loaded too)
	if _, set := os.LookupEnv("VOCAB_FAKE_PROVIDERS"); !set {
		t.Setenv("VOCAB_FAKE_PROVIDERS", "true") // never call Claude, Pixabay or Unsplash from tests
	}

	h := &harness{t: t, google: &fakeGoogle{identity: &entity.GoogleIdentity{
		Subject: "google-sub-1", Email: "learner@example.com", EmailVerified: true,
		Name: "Learner One", Picture: "https://example.com/a.png", Nonce: "nonce-1",
	}}}
	gin.SetMode(gin.TestMode)
	opts := []fx.Option{
		app.CoreOptions(),
		fx.Decorate(func(authout.IGoogleIdentityProvider) authout.IGoogleIdentityProvider { return h.google }),
		fx.Populate(&h.engine),
		fx.NopLogger,
	}
	app := fxtest.New(t, append(opts, extra...)...)
	app.RequireStart()
	t.Cleanup(app.RequireStop)
	return h
}

type resp struct {
	Status int
	Body   map[string]any
}

func (h *harness) do(method, path string, body any, accessToken string) resp {
	h.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	}
	w := httptest.NewRecorder()
	h.engine.ServeHTTP(w, req)
	out := resp{Status: w.Code, Body: map[string]any{}}
	_ = json.Unmarshal(w.Body.Bytes(), &out.Body)
	return out
}

func errCode(r resp) string {
	if e, ok := r.Body["error"].(map[string]any); ok {
		code, _ := e["code"].(string)
		return code
	}
	return ""
}

func (h *harness) signIn() (access, refresh string) {
	h.t.Helper()
	r := h.do("POST", "/api/v1/auth/google", map[string]any{
		"code": "auth-code", "code_verifier": "v123456789012345678901234567890123456789012345",
		"nonce": "nonce-1", "ui_locale": "vi", "timezone": "Asia/Ho_Chi_Minh",
	}, "")
	if r.Status != http.StatusOK {
		h.t.Fatalf("sign in failed: %d %v", r.Status, r.Body)
	}
	return r.Body["access_token"].(string), r.Body["refresh_token"].(string)
}

// ---- tests -------------------------------------------------------------------

func TestHealth(t *testing.T) {
	h := newHarness(t)
	r := h.do("GET", "/api/v1/health", nil, "")
	if r.Status != 200 || r.Body["db"] != "ok" || r.Body["storage"] != "not_configured" {
		t.Fatalf("unexpected health: %d %v", r.Status, r.Body)
	}
}

func TestGoogleSignInCreatesUserAndMeWorks(t *testing.T) {
	h := newHarness(t)
	access, _ := h.signIn()

	me := h.do("GET", "/api/v1/me", nil, access)
	if me.Status != 200 {
		t.Fatalf("me failed: %d %v", me.Status, me.Body)
	}
	user := me.Body["user"].(map[string]any)
	if user["email"] != "learner@example.com" || user["ui_locale"] != "vi" || user["timezone"] != "Asia/Ho_Chi_Minh" {
		t.Fatalf("unexpected user: %v", user)
	}
	settings := me.Body["study_settings"].(map[string]any)
	if settings["new_cards_per_day"].(float64) != 20 || settings["desired_retention"].(float64) != 0.9 {
		t.Fatalf("unexpected default settings: %v", settings)
	}

	// Second sign-in updates the same user (matched by Google sub) and keeps the chosen locale.
	h.google.identity.Email = "new-email@example.com"
	access2, _ := h.signIn()
	me2 := h.do("GET", "/api/v1/me", nil, access2)
	u2 := me2.Body["user"].(map[string]any)
	if u2["id"] != user["id"] || u2["email"] != "new-email@example.com" {
		t.Fatalf("expected same user with updated email: %v", u2)
	}
}

func TestGoogleSignInRejections(t *testing.T) {
	h := newHarness(t)
	body := map[string]any{"code": "c", "code_verifier": "v123456789012345678901234567890123456789012345", "nonce": "other"}

	if r := h.do("POST", "/api/v1/auth/google", body, ""); errCode(r) != "AUTH_GOOGLE_INVALID_TOKEN" {
		t.Fatalf("nonce mismatch should fail: %d %v", r.Status, r.Body)
	}

	h.google.identity.EmailVerified = false
	body["nonce"] = "nonce-1"
	if r := h.do("POST", "/api/v1/auth/google", body, ""); errCode(r) != "AUTH_GOOGLE_EMAIL_UNVERIFIED" {
		t.Fatalf("unverified email should fail: %d %v", r.Status, r.Body)
	}

	r := h.do("POST", "/api/v1/auth/google", map[string]any{"code": "c"}, "")
	if r.Status != 400 || errCode(r) != "VALIDATION_FAILED" {
		t.Fatalf("expected validation error: %d %v", r.Status, r.Body)
	}
	fields := r.Body["error"].(map[string]any)["fields"].(map[string]any)
	if fields["code_verifier"] != "VALIDATION_REQUIRED" || fields["nonce"] != "VALIDATION_REQUIRED" {
		t.Fatalf("expected json field names in validation errors: %v", fields)
	}
}

func TestMeRequiresAuth(t *testing.T) {
	h := newHarness(t)
	if r := h.do("GET", "/api/v1/me", nil, ""); r.Status != 401 || errCode(r) != "AUTH_UNAUTHORIZED" {
		t.Fatalf("expected 401: %d %v", r.Status, r.Body)
	}
	if r := h.do("GET", "/api/v1/me", nil, "garbage"); r.Status != 401 {
		t.Fatalf("expected 401 for invalid token: %d", r.Status)
	}
}

func TestRefreshRotationAndReuseDetection(t *testing.T) {
	h := newHarness(t)
	_, refresh1 := h.signIn()

	r := h.do("POST", "/api/v1/auth/refresh", map[string]any{"refresh_token": refresh1}, "")
	if r.Status != 200 {
		t.Fatalf("refresh failed: %d %v", r.Status, r.Body)
	}
	refresh2 := r.Body["refresh_token"].(string)
	if refresh2 == refresh1 {
		t.Fatal("refresh token must rotate")
	}

	// Replaying the old token within the grace window: rejected, but other sessions survive.
	if r := h.do("POST", "/api/v1/auth/refresh", map[string]any{"refresh_token": refresh1}, ""); errCode(r) != "AUTH_REFRESH_INVALID" {
		t.Fatalf("expected AUTH_REFRESH_INVALID: %d %v", r.Status, r.Body)
	}

	// Replay after the grace window: treated as theft, every session of the user is revoked.
	if _, err := testDB.Exec("UPDATE refresh_tokens SET revoked_at = now() - interval '5 minutes' WHERE revoked_at IS NOT NULL"); err != nil {
		t.Fatal(err)
	}
	if r := h.do("POST", "/api/v1/auth/refresh", map[string]any{"refresh_token": refresh1}, ""); errCode(r) != "AUTH_REFRESH_REUSED" {
		t.Fatalf("expected AUTH_REFRESH_REUSED: %d %v", r.Status, r.Body)
	}
	if r := h.do("POST", "/api/v1/auth/refresh", map[string]any{"refresh_token": refresh2}, ""); r.Status != 401 {
		t.Fatalf("newest token should be revoked after reuse detection: %d %v", r.Status, r.Body)
	}
}

func TestLogoutRevokesRefreshToken(t *testing.T) {
	h := newHarness(t)
	_, refresh := h.signIn()
	if r := h.do("POST", "/api/v1/auth/logout", map[string]any{"refresh_token": refresh}, ""); r.Status != 204 {
		t.Fatalf("logout failed: %d", r.Status)
	}
	if r := h.do("POST", "/api/v1/auth/refresh", map[string]any{"refresh_token": refresh}, ""); r.Status != 401 {
		t.Fatalf("refresh after logout should fail: %d", r.Status)
	}
}

func TestUpdateProfileAndStudySettings(t *testing.T) {
	h := newHarness(t)
	access, _ := h.signIn()

	r := h.do("PATCH", "/api/v1/me", map[string]any{"timezone": "Mars/Olympus"}, access)
	if r.Status != 400 || errCode(r) != "USER_INVALID_TIMEZONE" {
		t.Fatalf("expected invalid timezone: %d %v", r.Status, r.Body)
	}

	r = h.do("PATCH", "/api/v1/me", map[string]any{"display_name": "Long", "ui_locale": "en", "timezone": "Europe/London", "day_start_hour": 0}, access)
	user := r.Body["user"].(map[string]any)
	if r.Status != 200 || user["display_name"] != "Long" || user["ui_locale"] != "en" || user["day_start_hour"].(float64) != 0 {
		t.Fatalf("profile not updated: %d %v", r.Status, r.Body)
	}

	r = h.do("PATCH", "/api/v1/me/study-settings", map[string]any{"new_cards_per_day": 501}, access)
	if r.Status != 400 {
		t.Fatalf("expected validation error for 501 new cards: %d", r.Status)
	}
	r = h.do("PATCH", "/api/v1/me/study-settings", map[string]any{"new_cards_per_day": 30, "accent": "us", "desired_retention": 0.85}, access)
	if r.Status != 200 || r.Body["new_cards_per_day"].(float64) != 30 || r.Body["accent"] != "us" || r.Body["max_reviews_per_day"].(float64) != 200 {
		t.Fatalf("settings not updated: %d %v", r.Status, r.Body)
	}
}

func TestTestLoginOnlyInTestEnv(t *testing.T) {
	h := newHarness(t)
	r := h.do("POST", "/api/v1/auth/test-login", map[string]any{"email": "e2e@example.com", "display_name": "E2E"}, "")
	if r.Status != 200 || r.Body["access_token"] == nil {
		t.Fatalf("test-login should work in test env: %d %v", r.Status, r.Body)
	}
}
