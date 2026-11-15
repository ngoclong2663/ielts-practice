package integration

// Epic 2 integration tests: decks, cards and the Anki-style study flow over HTTP.

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func (h *harness) mustStatus(r resp, status int, what string) resp {
	h.t.Helper()
	if r.Status != status {
		h.t.Fatalf("%s: want %d, got %d %v", what, status, r.Status, r.Body)
	}
	return r
}

func (h *harness) createDeck(access, name string) string {
	h.t.Helper()
	r := h.mustStatus(h.do("POST", "/api/v1/decks", map[string]any{"name": name}, access), http.StatusCreated, "create deck")
	return r.Body["id"].(string)
}

func (h *harness) createCard(access, deckID, word, meaning string) string {
	h.t.Helper()
	r := h.mustStatus(h.do("POST", "/api/v1/decks/"+deckID+"/cards", map[string]any{
		"headword": word, "meaning_vi": meaning, "ipa": "/x/", "example_en": "An example with " + word + ".",
	}, access), http.StatusCreated, "create card")
	return r.Body["id"].(string)
}

func (h *harness) answer(access, deckID, cardID string, rating int) resp {
	h.t.Helper()
	return h.mustStatus(h.do("POST", "/api/v1/study/answer", map[string]any{
		"card_id": cardID, "rating": rating, "duration_ms": 4200,
		"client_review_id": uuid.NewString(), "deck_id": deckID,
	}, access), http.StatusOK, fmt.Sprintf("answer %d", rating))
}

func counts(r resp) (n, l, rv int) {
	c := r.Body["counts"].(map[string]any)
	return int(c["new"].(float64)), int(c["learning"].(float64)), int(c["review"].(float64))
}

func queueCards(r resp) []map[string]any {
	raw, _ := r.Body["cards"].([]any)
	out := make([]map[string]any, len(raw))
	for i, c := range raw {
		out[i] = c.(map[string]any)
	}
	return out
}

func TestDecksAndCardsCRUD(t *testing.T) {
	h := newHarness(t)
	access, _ := h.signIn()

	deckID := h.createDeck(access, "IELTS Day 1")
	if r := h.do("POST", "/api/v1/decks", map[string]any{"name": "ielts day 1"}, access); r.Status != 409 || errCode(r) != "DECK_NAME_TAKEN" {
		t.Fatalf("names are unique per user (case-insensitive): %d %v", r.Status, r.Body)
	}

	c1 := h.createCard(access, deckID, "resilient", "kiên cường")
	h.createCard(access, deckID, "decide", "quyết định")
	if r := h.do("POST", "/api/v1/decks/"+deckID+"/cards", map[string]any{"headword": "Resilient", "meaning_vi": "x"}, access); errCode(r) != "CARD_DUPLICATE_HEADWORD" {
		t.Fatalf("duplicate headword should warn: %d %v", r.Status, r.Body)
	}
	h.mustStatus(h.do("POST", "/api/v1/decks/"+deckID+"/cards", map[string]any{"headword": "Resilient", "meaning_vi": "x", "allow_duplicate": true}, access), 201, "allow duplicate")
	// The web form sends empty strings for untouched optional fields.
	h.mustStatus(h.do("POST", "/api/v1/decks/"+deckID+"/cards", map[string]any{
		"headword": "empty optionals", "meaning_vi": "x", "ipa": "", "part_of_speech": "", "example_en": "", "example_vi": "", "image_url": "",
	}, access), 201, "empty optional fields")
	if r := h.do("POST", "/api/v1/decks/"+deckID+"/cards", map[string]any{"headword": "bad url", "meaning_vi": "x", "image_url": "javascript:alert(1)"}, access); r.Status != 400 {
		t.Fatalf("non-http image URL must be rejected: %d %v", r.Status, r.Body)
	}
	if r := h.do("POST", "/api/v1/decks/"+deckID+"/cards", map[string]any{"headword": "word"}, access); r.Status != 400 {
		t.Fatalf("meaning is required: %d %v", r.Status, r.Body)
	}

	list := h.mustStatus(h.do("GET", "/api/v1/decks", nil, access), 200, "list decks")
	deck := list.Body["items"].([]any)[0].(map[string]any)
	if deck["card_count"].(float64) != 4 {
		t.Fatalf("card_count: %v", deck)
	}

	search := h.mustStatus(h.do("GET", "/api/v1/decks/"+deckID+"/cards?q=quyết", nil, access), 200, "search")
	if search.Body["total"].(float64) != 1 {
		t.Fatalf("search by Vietnamese meaning: %v", search.Body)
	}

	card := h.mustStatus(h.do("PATCH", "/api/v1/cards/"+c1, map[string]any{"meaning_vi": "kiên cường, mau phục hồi", "image_url": "https://images.unsplash.com/photo-1"}, access), 200, "update card")
	content := card.Body["content"].(map[string]any)
	if content["meaning_vi"] != "kiên cường, mau phục hồi" || content["image"].(map[string]any)["url"] == "" || content["extra_examples"] == nil {
		t.Fatalf("edit should merge into content: %v", content)
	}

	other := h.createDeck(access, "Other")
	h.mustStatus(h.do("POST", "/api/v1/cards/bulk", map[string]any{"card_ids": []string{c1}, "action": "suspend"}, access), 200, "suspend")
	if r := h.do("GET", "/api/v1/decks/"+deckID+"/cards?state=suspended", nil, access); r.Body["total"].(float64) != 1 {
		t.Fatalf("suspended filter: %v", r.Body)
	}
	moved := h.mustStatus(h.do("POST", "/api/v1/cards/bulk", map[string]any{"card_ids": []string{c1}, "action": "move", "deck_id": other}, access), 200, "move")
	if moved.Body["affected"].(float64) != 1 {
		t.Fatalf("move: %v", moved.Body)
	}

	h.mustStatus(h.do("PATCH", "/api/v1/decks/"+deckID, map[string]any{"new_per_day": 5, "name": "IELTS Day 1 (renamed)"}, access), 200, "update deck")
	h.mustStatus(h.do("DELETE", "/api/v1/decks/"+deckID, nil, access), 204, "archive deck")
	if r := h.do("GET", "/api/v1/decks/"+deckID, nil, access); errCode(r) != "DECK_NOT_FOUND" {
		t.Fatalf("archived deck must be gone: %v", r.Body)
	}
	if r := h.do("GET", "/api/v1/decks/not-a-uuid", nil, access); r.Status != 404 {
		t.Fatalf("malformed id: %d", r.Status)
	}

	// Another user cannot see or touch these cards.
	h.google.identity.Subject, h.google.identity.Email = "intruder", "intruder@example.com"
	intruder, _ := h.signIn()
	if r := h.do("GET", "/api/v1/cards/"+c1, nil, intruder); r.Status != 404 {
		t.Fatalf("cards are private: %d", r.Status)
	}
	if r := h.do("POST", "/api/v1/cards/bulk", map[string]any{"card_ids": []string{c1}, "action": "delete"}, intruder); r.Body["affected"].(float64) != 0 {
		t.Fatalf("bulk on someone else's card must affect nothing: %v", r.Body)
	}
}

func TestStudyFlow(t *testing.T) {
	h := newHarness(t)
	access, _ := h.signIn()
	deckID := h.createDeck(access, "Study")
	h.mustStatus(h.do("PATCH", "/api/v1/decks/"+deckID, map[string]any{"new_per_day": 2}, access), 200, "limit")
	c1 := h.createCard(access, deckID, "alpha", "a")
	c2 := h.createCard(access, deckID, "beta", "b")
	h.createCard(access, deckID, "gamma", "c")

	q := h.mustStatus(h.do("GET", "/api/v1/study/queue?deck_id="+deckID, nil, access), 200, "queue")
	if n, l, r := counts(q); n != 2 || l != 0 || r != 0 {
		t.Fatalf("deck limit 2 new: got %d/%d/%d", n, l, r)
	}
	cards := queueCards(q)
	if len(cards) != 2 || cards[0]["id"] != c1 || cards[1]["id"] != c2 {
		t.Fatalf("new cards in deck order: %v", cards)
	}
	iv := cards[0]["intervals"].(map[string]any)
	if iv["again"].(float64) != 60 || iv["hard"].(float64) != 360 || iv["good"].(float64) != 600 || iv["easy"].(float64) < 86400 {
		t.Fatalf("button previews should be 1m / 6m / 10m / days: %v", iv)
	}
	if cards[0]["content"].(map[string]any)["headword"] != "alpha" {
		t.Fatalf("content is included: %v", cards[0])
	}

	// Good on a new card: learning again in 10 minutes.
	a := h.answer(access, deckID, c1, 3)
	if ans := a.Body["answered"].(map[string]any); ans["state"] != "learning" {
		t.Fatalf("answered: %v", ans)
	}
	if n, l, _ := counts(a); n != 1 || l != 1 {
		t.Fatalf("after Good: want 1 new, 1 learning, got %v", a.Body["counts"])
	}
	if next := queueCards(a); next[0]["id"] != c2 {
		t.Fatalf("the remaining new card comes before a learning card that is not due yet: %v", next)
	}

	// Easy on the second card graduates it; the new limit (2) is now used up.
	a = h.answer(access, deckID, c2, 4)
	if n, l, r := counts(a); n != 0 || l != 1 || r != 0 {
		t.Fatalf("limit reached: got %d/%d/%d", n, l, r)
	}
	if next := queueCards(a); len(next) != 1 || next[0]["id"] != c1 {
		t.Fatalf("with nothing else left, the learning card is shown early (learn-ahead): %v", next)
	}

	// Idempotent retries and stale answers.
	reviewID := uuid.NewString()
	body := map[string]any{"card_id": c1, "rating": 3, "client_review_id": reviewID, "expected_reps": 1, "deck_id": deckID}
	h.mustStatus(h.do("POST", "/api/v1/study/answer", body, access), 200, "answer once")
	h.mustStatus(h.do("POST", "/api/v1/study/answer", body, access), 200, "same answer again")
	var logs int
	if err := testDB.QueryRow("SELECT count(*) FROM review_logs WHERE card_id = $1", c1).Scan(&logs); err != nil || logs != 2 {
		t.Fatalf("retry must not create a second log: %d %v", logs, err)
	}
	stale := h.do("POST", "/api/v1/study/answer", map[string]any{"card_id": c1, "rating": 3, "client_review_id": uuid.NewString(), "expected_reps": 0}, access)
	if stale.Status != 409 || errCode(stale) != "STUDY_CARD_STATE_CHANGED" {
		t.Fatalf("stale answer: %d %v", stale.Status, stale.Body)
	}

	// Undo the last answer: the card comes back first, in its previous state.
	u := h.mustStatus(h.do("POST", "/api/v1/study/undo", map[string]any{"deck_id": deckID}, access), 200, "undo")
	if u.Body["undone_card_id"] != c1 || queueCards(u)[0]["id"] != c1 || queueCards(u)[0]["state"] != "learning" {
		t.Fatalf("undo: %v", u.Body)
	}

	// Next day: make the graduated card due and check it appears as a review.
	if _, err := testDB.Exec("UPDATE cards SET due = now() - interval '1 hour' WHERE id = $1", c2); err != nil {
		t.Fatal(err)
	}
	q = h.mustStatus(h.do("GET", "/api/v1/study/queue?deck_id="+deckID, nil, access), 200, "queue with review")
	if _, _, r := counts(q); r != 1 {
		t.Fatalf("review due: %v", q.Body["counts"])
	}

	// Leech: the 8th lapse marks the card.
	if _, err := testDB.Exec("UPDATE cards SET lapses = 7 WHERE id = $1", c2); err != nil {
		t.Fatal(err)
	}
	a = h.answer(access, deckID, c2, 1)
	if ans := a.Body["answered"].(map[string]any); ans["became_leech"] != true || ans["state"] != "relearning" {
		t.Fatalf("leech: %v", ans)
	}

	// Counts and stats.
	cnt := h.mustStatus(h.do("GET", "/api/v1/study/counts", nil, access), 200, "counts")
	if len(cnt.Body["decks"].([]any)) != 1 {
		t.Fatalf("per-deck counts: %v", cnt.Body)
	}
	st := h.mustStatus(h.do("GET", "/api/v1/stats/overview", nil, access), 200, "stats")
	today := st.Body["today"].(map[string]any)
	if today["reviews"].(float64) != 3 || today["new_cards"].(float64) != 2 || st.Body["streak_days"].(float64) != 1 || st.Body["studied_today"] != true {
		t.Fatalf("stats: %v", st.Body)
	}
	if len(st.Body["forecast"].([]any)) != 7 {
		t.Fatalf("forecast has 7 days: %v", st.Body["forecast"])
	}
	if st.Body["retention_30d"].(float64) != 0 || st.Body["retention_sample"].(float64) != 1 {
		t.Fatalf("one failed review out of one: %v", st.Body)
	}
}

func TestSuspendedAndArchivedCardsAreNotStudied(t *testing.T) {
	h := newHarness(t)
	access, _ := h.signIn()
	d1 := h.createDeck(access, "One")
	d2 := h.createDeck(access, "Two")
	c1 := h.createCard(access, d1, "one", "1")
	h.createCard(access, d2, "two", "2")

	h.mustStatus(h.do("POST", "/api/v1/cards/bulk", map[string]any{"card_ids": []string{c1}, "action": "suspend"}, access), 200, "suspend")
	h.mustStatus(h.do("DELETE", "/api/v1/decks/"+d2, nil, access), 204, "archive")

	q := h.mustStatus(h.do("GET", "/api/v1/study/queue", nil, access), 200, "queue all")
	if len(queueCards(q)) != 0 {
		t.Fatalf("nothing to study: %v", q.Body)
	}
	if r := h.do("POST", "/api/v1/study/answer", map[string]any{"card_id": c1, "rating": 3, "client_review_id": uuid.NewString()}, access); errCode(r) != "STUDY_CARD_STATE_CHANGED" {
		t.Fatalf("suspended cards cannot be answered: %v", r.Body)
	}
	if r := h.do("POST", "/api/v1/study/undo", map[string]any{}, access); errCode(r) != "STUDY_NOTHING_TO_UNDO" {
		t.Fatalf("undo with no answers: %v", r.Body)
	}
	if r := h.do("GET", "/api/v1/study/queue?deck_id="+d2, nil, access); errCode(r) != "DECK_NOT_FOUND" {
		t.Fatalf("archived deck queue: %v", r.Body)
	}
}

func TestCustomLearningSteps(t *testing.T) {
	h := newHarness(t)
	access, _ := h.signIn()
	if r := h.do("PATCH", "/api/v1/me/study-settings", map[string]any{"learning_steps": "5x"}, access); errCode(r) != "USER_INVALID_STEPS" {
		t.Fatalf("invalid steps: %v", r.Body)
	}
	h.mustStatus(h.do("PATCH", "/api/v1/me/study-settings", map[string]any{"learning_steps": "2m  15m"}, access), 200, "steps")
	deckID := h.createDeck(access, "Steps")
	h.createCard(access, deckID, "word", "từ")
	q := h.mustStatus(h.do("GET", "/api/v1/study/queue", nil, access), 200, "queue")
	iv := queueCards(q)[0]["intervals"].(map[string]any)
	if iv["again"].(float64) != 120 || iv["good"].(float64) != 900 {
		t.Fatalf("custom steps should drive previews: %v", iv)
	}
}
