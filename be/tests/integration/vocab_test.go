package integration

// Epic 3 integration tests: paste → generate (fake Claude) → review → save to deck,
// the shared lexeme cache, quotas, and the Pixabay link refresh / replacement.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/fx"

	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/infrastructure"
)

func (h *harness) waitJob(access, jobID string) resp {
	h.t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		r := h.mustStatus(h.do("GET", "/api/v1/vocab/jobs/"+jobID, nil, access), http.StatusOK, "get job")
		switch r.Body["status"] {
		case "done", "partial", "failed":
			return r
		}
		if time.Now().After(deadline) {
			h.t.Fatalf("job did not finish: %v", r.Body)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func jobItems(r resp) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, raw := range r.Body["items"].([]any) {
		it := raw.(map[string]any)
		out[it["input_text"].(string)] = it
	}
	return out
}

func entriesBody(words ...string) map[string]any {
	entries := make([]map[string]any, len(words))
	for i, w := range words {
		entries[i] = map[string]any{"text": w}
	}
	return map[string]any{"entries": entries}
}

func TestVocabGenerateReviewAndSave(t *testing.T) {
	var gen domain.IContentGenerator
	h := newHarness(t, fx.Populate(&gen))
	fake := gen.(*infrastructure.FakeGenerator)
	access, _ := h.signIn()

	status := h.mustStatus(h.do("GET", "/api/v1/vocab/status", nil, access), http.StatusOK, "status")
	if status.Body["configured"] != true || status.Body["max_entries"].(float64) != 30 {
		t.Fatalf("unexpected status: %v", status.Body)
	}
	norm := h.mustStatus(h.do("POST", "/api/v1/vocab/normalize", map[string]any{"text": "Decide, decide; bank | river\nnot a word at all because too long here"}, access), http.StatusOK, "normalize")
	if len(norm.Body["entries"].([]any)) != 2 || len(norm.Body["rejected"].([]any)) != 2 {
		t.Fatalf("unexpected normalize: %v", norm.Body)
	}

	created := h.mustStatus(h.do("POST", "/api/v1/vocab/jobs", entriesBody("decide", "take into account", "zzbuzz", "Decide"), access), http.StatusCreated, "create job")
	if created.Body["total"].(float64) != 3 {
		t.Fatalf("duplicates should be merged: %v", created.Body)
	}
	jobID := created.Body["id"].(string)
	job := h.waitJob(access, jobID)
	items := jobItems(job)
	decide := items["decide"]
	if decide["status"] != "ready" || decide["cached"] != false || decide["lexeme_id"] == nil {
		t.Fatalf("decide not ready: %v", decide)
	}
	content := decide["content"].(map[string]any)
	if content["meaning_vi"] != "nghĩa của decide" || content["image"] == nil {
		t.Fatalf("unexpected content: %v", content)
	}
	if items["zzbuzz"]["status"] != "invalid" || items["zzbuzz"]["suggestion"] != "buzz" {
		t.Fatalf("typo not detected: %v", items["zzbuzz"])
	}
	if job.Body["status"] != "partial" {
		t.Fatalf("job with an invalid word should be partial: %v", job.Body["status"])
	}

	// Accept the suggestion: the item is regenerated as "buzz".
	typo := items["zzbuzz"]["id"].(string)
	h.mustStatus(h.do("POST", fmt.Sprintf("/api/v1/vocab/jobs/%s/items/%s/retry", jobID, typo), map[string]any{"input_text": "buzz"}, access), http.StatusOK, "retry")
	job = h.waitJob(access, jobID)
	items = jobItems(job)
	if items["buzz"]["status"] != "ready" || job.Body["status"] != "done" {
		t.Fatalf("retry did not finish: %v", job.Body)
	}

	// Image picker: more candidates, choose the second one, then edit the meaning.
	decideID := decide["id"].(string)
	itemPath := fmt.Sprintf("/api/v1/vocab/jobs/%s/items/%s", jobID, decideID)
	imgs := h.mustStatus(h.do("GET", itemPath+"/images?q=crossroads", nil, access), http.StatusOK, "images")
	picked := imgs.Body["images"].([]any)[1].(map[string]any)
	upd := h.mustStatus(h.do("PATCH", itemPath, map[string]any{
		"image": map[string]any{"provider": picked["provider"], "provider_id": picked["provider_id"]},
	}, access), http.StatusOK, "pick image")
	if upd.Body["content"].(map[string]any)["image"].(map[string]any)["provider_id"] != picked["provider_id"] {
		t.Fatalf("image not chosen: %v", upd.Body)
	}
	if r := h.do("PATCH", itemPath, map[string]any{"image": map[string]any{"provider": "fake", "provider_id": "not-offered"}}, access); r.Status != 400 {
		t.Fatalf("arbitrary images must be rejected: %d", r.Status)
	}
	edited := upd.Body["content"].(map[string]any)
	edited["meaning_vi"] = "quyết định (đã sửa)"
	h.mustStatus(h.do("PATCH", itemPath, map[string]any{"content": edited}, access), http.StatusOK, "edit content")
	if r := h.do("PATCH", itemPath, map[string]any{"content": map[string]any{"headword": "decide", "meaning_vi": " "}}, access); errCode(r) != "VALIDATION_FAILED" {
		t.Fatalf("empty meaning must be rejected: %v", r.Body)
	}

	// Save into a new deck; saving again skips every word.
	ids := []string{}
	for _, it := range items {
		ids = append(ids, it["id"].(string))
	}
	commit := h.mustStatus(h.do("POST", "/api/v1/vocab/jobs/"+jobID+"/commit", map[string]any{"new_deck_name": "IELTS words", "item_ids": ids}, access), http.StatusCreated, "commit")
	if commit.Body["created"].(float64) != 3 || len(commit.Body["skipped"].([]any)) != 0 {
		t.Fatalf("unexpected commit: %v", commit.Body)
	}
	deckID := commit.Body["deck_id"].(string)
	again := h.mustStatus(h.do("POST", "/api/v1/vocab/jobs/"+jobID+"/commit", map[string]any{"deck_id": deckID, "item_ids": ids}, access), http.StatusCreated, "commit again")
	if again.Body["created"].(float64) != 0 || len(again.Body["skipped"].([]any)) != 3 {
		t.Fatalf("duplicates should be skipped: %v", again.Body)
	}
	cards := h.mustStatus(h.do("GET", "/api/v1/decks/"+deckID+"/cards", nil, access), http.StatusOK, "list cards")
	var decideCard map[string]any
	for _, raw := range cards.Body["items"].([]any) {
		c := raw.(map[string]any)
		if c["headword"] == "decide" {
			decideCard = c
		}
	}
	if decideCard == nil || decideCard["meaning_vi"] != "quyết định (đã sửa)" || decideCard["lexeme_id"] == nil {
		t.Fatalf("saved card should keep the user's edits: %v", decideCard)
	}
	queue := h.mustStatus(h.do("GET", "/api/v1/study/queue?deck_id="+deckID, nil, access), http.StatusOK, "queue")
	if len(queueCards(queue)) == 0 || queueCards(queue)[0]["lexeme_id"] == nil {
		t.Fatalf("generated cards should be studyable: %v", queue.Body)
	}

	// A second user gets the shared lexeme instantly, with no generator call.
	wordsBefore := fake.Words.Load()
	h.google.identity.Subject, h.google.identity.Email = "google-sub-2", "second@example.com"
	access2, _ := h.signIn()
	cached := h.mustStatus(h.do("POST", "/api/v1/vocab/jobs", entriesBody("decide"), access2), http.StatusCreated, "cached job")
	it := cached.Body["items"].([]any)[0].(map[string]any)
	if cached.Body["status"] != "done" || it["status"] != "ready" || it["cached"] != true {
		t.Fatalf("cache hit should be ready at once: %v", cached.Body)
	}
	if it["content"].(map[string]any)["meaning_vi"] != "nghĩa của decide" {
		t.Fatalf("another user's edits must not leak into the shared lexeme: %v", it["content"])
	}
	if fake.Words.Load() != wordsBefore {
		t.Fatalf("cache hit must not call the generator")
	}

	// Reports: once per user.
	lexemeID := it["lexeme_id"].(string)
	h.mustStatus(h.do("POST", "/api/v1/lexemes/"+lexemeID+"/report", map[string]any{"reason": "wrong_meaning"}, access2), http.StatusNoContent, "report")
	h.mustStatus(h.do("POST", "/api/v1/lexemes/"+lexemeID+"/report", map[string]any{"reason": "wrong_meaning"}, access2), http.StatusNoContent, "report twice")
	var reports int
	_ = testDB.QueryRow("SELECT report_count FROM lexemes WHERE id = $1", lexemeID).Scan(&reports)
	if reports != 1 {
		t.Fatalf("report_count = %d, want 1", reports)
	}

	// The other user's job is not visible.
	if r := h.do("GET", "/api/v1/vocab/jobs/"+jobID, nil, access2); errCode(r) != "VOCAB_JOB_NOT_FOUND" {
		t.Fatalf("jobs must be private: %v", r.Body)
	}
	list := h.mustStatus(h.do("GET", "/api/v1/vocab/jobs", nil, access), http.StatusOK, "list jobs")
	if len(list.Body["items"].([]any)) != 1 {
		t.Fatalf("unexpected job list: %v", list.Body)
	}
}

func TestVocabQuotaAndLimits(t *testing.T) {
	t.Setenv("VOCAB_DAILY_QUOTA", "2")
	h := newHarness(t)
	access, _ := h.signIn()

	if r := h.do("POST", "/api/v1/vocab/jobs", entriesBody("alpha", "beta", "gamma"), access); errCode(r) != "VOCAB_QUOTA_EXCEEDED" || r.Status != 429 {
		t.Fatalf("quota should block: %d %v", r.Status, r.Body)
	}
	job := h.mustStatus(h.do("POST", "/api/v1/vocab/jobs", entriesBody("alpha", "beta"), access), http.StatusCreated, "within quota")
	h.waitJob(access, job.Body["id"].(string))
	if r := h.do("POST", "/api/v1/vocab/jobs", entriesBody("gamma"), access); errCode(r) != "VOCAB_QUOTA_EXCEEDED" {
		t.Fatalf("quota should be used up: %v", r.Body)
	}
	// Cached words are free.
	h.mustStatus(h.do("POST", "/api/v1/vocab/jobs", entriesBody("alpha"), access), http.StatusCreated, "cache hit")

	many := make([]string, 31)
	for i := range many {
		many[i] = "word" + strings.Repeat("a", i+1)
	}
	if r := h.do("POST", "/api/v1/vocab/jobs", entriesBody(many...), access); errCode(r) != "VOCAB_TOO_MANY_ENTRIES" {
		t.Fatalf("31 words should be rejected: %v", r.Body)
	}
	if r := h.do("POST", "/api/v1/vocab/jobs", entriesBody("123"), access); errCode(r) != "VOCAB_NO_ENTRIES" {
		t.Fatalf("invalid entries should be rejected: %v", r.Body)
	}
}

func TestVocabGeneratorNotConfigured(t *testing.T) {
	t.Setenv("VOCAB_FAKE_PROVIDERS", "false")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "") // the fallback model must not be called either
	h := newHarness(t)
	access, _ := h.signIn()
	if r := h.do("GET", "/api/v1/vocab/status", nil, access); r.Body["configured"] != false {
		t.Fatalf("status should report not configured: %v", r.Body)
	}
	job := h.mustStatus(h.do("POST", "/api/v1/vocab/jobs", entriesBody("decide"), access), http.StatusCreated, "job")
	done := h.waitJob(access, job.Body["id"].(string))
	it := done.Body["items"].([]any)[0].(map[string]any)
	if it["status"] != "failed" || it["error_code"] != "VOCAB_GENERATOR_NOT_CONFIGURED" {
		t.Fatalf("expected a clear failure: %v", it)
	}
	if r := h.do("POST", "/api/v1/vocab/suggestions", map[string]any{"topic_id": "health", "count": 5}, access); errCode(r) != "VOCAB_GENERATOR_NOT_CONFIGURED" {
		t.Fatalf("suggestions need a model too: %v", r.Body)
	}
	// Failed words do not use the quota.
	if r := h.do("GET", "/api/v1/vocab/status", nil, access); r.Body["quota"].(map[string]any)["used"].(float64) != 0 {
		t.Fatalf("failed words must not count: %v", r.Body)
	}
}

// expiringProvider mimics Pixabay: links expire, photos can disappear.
type expiringProvider struct {
	mu      sync.Mutex
	version int
	removed map[string]bool
	empty   bool
}

func (p *expiringProvider) Name() string     { return "pixabay" }
func (p *expiringProvider) Configured() bool { return true }
func (p *expiringProvider) Expiring() bool   { return true }

func (p *expiringProvider) image(id string) entity.CardImage {
	return entity.CardImage{
		Provider: "pixabay", ProviderID: id, AuthorName: "pix",
		URL:           fmt.Sprintf("https://cdn.example/%s-v%d.jpg", id, p.version),
		ThumbURL:      fmt.Sprintf("https://cdn.example/%s-v%d-thumb.jpg", id, p.version),
		LinkExpiresAt: time.Now().Add(23 * time.Hour),
	}
}

func (p *expiringProvider) Search(_ context.Context, query string, limit int) ([]entity.CardImage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.empty {
		return nil, nil
	}
	out := []entity.CardImage{}
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("%d%d", len(query), i)
		if !p.removed[id] {
			out = append(out, p.image(id))
		}
	}
	return out, nil
}

func (p *expiringProvider) Lookup(_ context.Context, id string) (*entity.CardImage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.removed[id] {
		return nil, nil
	}
	img := p.image(id)
	return &img, nil
}

func (p *expiringProvider) TrackDownload(context.Context, entity.CardImage) error { return nil }

func TestImageLinkRefreshAndReplacement(t *testing.T) {
	pix := &expiringProvider{version: 1, removed: map[string]bool{}}
	h := newHarness(t, fx.Decorate(func(domain.ImageProviders) domain.ImageProviders {
		return domain.ImageProviders{pix, infrastructure.NewFakeImageProvider()}
	}))
	access, _ := h.signIn()

	job := h.mustStatus(h.do("POST", "/api/v1/vocab/jobs", entriesBody("ocean"), access), http.StatusCreated, "job")
	jobID := job.Body["id"].(string)
	done := h.waitJob(access, jobID)
	item := done.Body["items"].([]any)[0].(map[string]any)
	img := item["content"].(map[string]any)["image"].(map[string]any)
	path := img["url"].(string)
	if !strings.HasPrefix(path, "/api/v1/images/pixabay/") || img["thumb_url"] != path+"?size=thumb" {
		t.Fatalf("pixabay images must go through the refresh endpoint: %v", img)
	}
	oldID := img["provider_id"].(string)
	ids := []string{item["id"].(string)}
	commit := h.mustStatus(h.do("POST", "/api/v1/vocab/jobs/"+jobID+"/commit", map[string]any{"new_deck_name": "Images", "item_ids": ids}, access), http.StatusCreated, "commit")
	deckID := commit.Body["deck_id"].(string)

	location := func(p string) (int, string) {
		req, _ := http.NewRequest("GET", p, nil)
		w := httptest.NewRecorder()
		h.engine.ServeHTTP(w, req)
		return w.Code, w.Header().Get("Location")
	}

	// Valid link: straight redirect, no provider call needed.
	if code, loc := location(path); code != 302 || loc != fmt.Sprintf("https://cdn.example/%s-v1.jpg", oldID) {
		t.Fatalf("redirect = %d %s", code, loc)
	}
	if _, loc := location(path + "?size=thumb"); !strings.HasSuffix(loc, "-v1-thumb.jpg") {
		t.Fatalf("thumb redirect = %s", loc)
	}

	// Expired link: refreshed by id.
	expire := func() { _, _ = testDB.Exec("UPDATE image_links SET expires_at = now() - interval '1 hour'") }
	pix.version = 2
	expire()
	if _, loc := location(path); loc != fmt.Sprintf("https://cdn.example/%s-v2.jpg", oldID) {
		t.Fatalf("refreshed redirect = %s", loc)
	}

	// Removed upstream: replaced by another search result everywhere.
	pix.removed[oldID] = true
	expire()
	code, loc := location(path)
	if code != 302 || !strings.HasPrefix(loc, "/api/v1/images/pixabay/") || strings.HasSuffix(loc, "/"+oldID) {
		t.Fatalf("replacement redirect = %d %s", code, loc)
	}
	cards := h.mustStatus(h.do("GET", "/api/v1/decks/"+deckID+"/cards", nil, access), http.StatusOK, "cards")
	cardID := cards.Body["items"].([]any)[0].(map[string]any)["id"].(string)
	card := h.mustStatus(h.do("GET", "/api/v1/cards/"+cardID, nil, access), http.StatusOK, "card")
	newImg := card.Body["content"].(map[string]any)["image"].(map[string]any)
	if newImg["provider_id"] == oldID || newImg["url"] != loc {
		t.Fatalf("card image should be replaced: %v", newImg)
	}

	// Removed again and Pixabay has nothing: falls back to the next provider (direct link).
	newID := newImg["provider_id"].(string)
	pix.removed[newID] = true
	pix.empty = true
	expire()
	_, loc = location(newImg["url"].(string))
	if !strings.HasPrefix(loc, "data:image/svg+xml") {
		t.Fatalf("should fall back to the second provider: %s", loc)
	}
	card = h.mustStatus(h.do("GET", "/api/v1/cards/"+cardID, nil, access), http.StatusOK, "card")
	if p := card.Body["content"].(map[string]any)["image"].(map[string]any)["provider"]; p != "fake" {
		t.Fatalf("card should use the fallback provider: %v", p)
	}

	if code, _ := location("/api/v1/images/pixabay/999999"); code != 404 {
		t.Fatalf("unknown images must 404, got %d", code)
	}
}
