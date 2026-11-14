package app

// Epic 4 integration tests: topic word suggestions, deck de-duplication, known words, daily limit.

import (
	"net/http"
	"testing"
)

func suggestionWords(r resp) []string {
	var out []string
	for _, raw := range r.Body["words"].([]any) {
		out = append(out, raw.(map[string]any)["lemma"].(string))
	}
	return out
}

func TestTopicWordSuggestions(t *testing.T) {
	t.Setenv("VOCAB_DAILY_SUGGESTIONS", "3")
	h := newHarness(t)
	access, _ := h.signIn()

	topics := h.mustStatus(h.do("GET", "/api/v1/vocab/topics", nil, access), http.StatusOK, "topics")
	if len(topics.Body["items"].([]any)) != 14 {
		t.Fatalf("expected 14 presets: %v", topics.Body)
	}

	// The deck already has "diagnosis"; the fake model always proposes it, so the server must drop it.
	deckID := h.createDeck(access, "Hospital words")
	h.createCard(access, deckID, "diagnosis", "chẩn đoán")
	h.createCard(access, deckID, "Treatment", "điều trị")

	first := h.mustStatus(h.do("POST", "/api/v1/vocab/suggestions", map[string]any{
		"topic_id": "health", "count": 5, "deck_id": deckID,
	}, access), http.StatusCreated, "suggest")
	words := suggestionWords(first)
	want := []string{"symptom", "surgery", "vaccine", "therapy", "recovery"}
	if len(words) != 5 || words[0] != want[0] || words[4] != want[4] {
		t.Fatalf("words = %v, want %v", words, want)
	}
	if first.Body["skipped_existing"].(float64) != 1 || first.Body["remaining_requests"].(float64) != 2 {
		t.Fatalf("unexpected counters: %v", first.Body)
	}
	id := first.Body["id"].(string)

	// "I know this" hides the word now and in later suggestions.
	h.mustStatus(h.do("POST", "/api/v1/vocab/known-words", map[string]any{"lemma": "Surgery"}, access), http.StatusNoContent, "known")
	again := h.mustStatus(h.do("GET", "/api/v1/vocab/suggestions/"+id, nil, access), http.StatusOK, "get")
	for _, w := range suggestionWords(again) {
		if w == "surgery" {
			t.Fatalf("known word still listed: %v", suggestionWords(again))
		}
	}

	// "Suggest more" never repeats the list, the deck or known words.
	more := h.mustStatus(h.do("POST", "/api/v1/vocab/suggestions/"+id+"/more", map[string]any{"count": 3}, access), http.StatusOK, "more")
	got := suggestionWords(more)
	if more.Body["added"].(float64) != 3 || len(got) != 7 {
		t.Fatalf("more: added=%v words=%v", more.Body["added"], got)
	}
	seen := map[string]bool{}
	for _, w := range got {
		if seen[w] || w == "diagnosis" || w == "treatment" || w == "surgery" {
			t.Fatalf("repeated or excluded word %q in %v", w, got)
		}
		seen[w] = true
	}

	// Undo "I know this".
	h.mustStatus(h.do("DELETE", "/api/v1/vocab/known-words/surgery", nil, access), http.StatusNoContent, "unknown")

	// Typed topics work; a missing topic is rejected (and not counted).
	if r := h.do("POST", "/api/v1/vocab/suggestions", map[string]any{"count": 5}, access); errCode(r) != "VOCAB_TOPIC_REQUIRED" {
		t.Fatalf("missing topic: %v", r.Body)
	}
	typed := h.mustStatus(h.do("POST", "/api/v1/vocab/suggestions", map[string]any{
		"topic_text": "  banking   apps ", "count": 5,
	}, access), http.StatusCreated, "typed topic")
	if typed.Body["topic_text"] != "banking apps" || typed.Body["remaining_requests"].(float64) != 0 {
		t.Fatalf("typed topic: %v", typed.Body)
	}

	// Daily limit: 3 requests (2 lists + 1 "more") used.
	if r := h.do("POST", "/api/v1/vocab/suggestions", map[string]any{"topic_id": "money", "count": 5}, access); errCode(r) != "VOCAB_SUGGESTION_LIMIT" || r.Status != 429 {
		t.Fatalf("limit not enforced: %d %v", r.Status, r.Body)
	}
	status := h.mustStatus(h.do("GET", "/api/v1/vocab/status", nil, access), http.StatusOK, "status")
	if s := status.Body["suggestions"].(map[string]any); s["remaining"].(float64) != 0 || s["limit"].(float64) != 3 {
		t.Fatalf("status suggestions: %v", s)
	}

	// Lists are private.
	h.google.identity.Subject, h.google.identity.Email = "google-sub-2", "second@example.com"
	access2, _ := h.signIn()
	if r := h.do("GET", "/api/v1/vocab/suggestions/"+id, nil, access2); errCode(r) != "VOCAB_SUGGESTION_NOT_FOUND" {
		t.Fatalf("suggestions must be private: %v", r.Body)
	}
	if r := h.do("POST", "/api/v1/vocab/suggestions", map[string]any{"topic_id": "money", "count": 5, "deck_id": deckID}, access2); errCode(r) != "DECK_NOT_FOUND" {
		t.Fatalf("other users' decks must be rejected: %v", r.Body)
	}
}
