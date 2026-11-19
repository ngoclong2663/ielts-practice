package integration

import (
	"context"
	"net/http"
	"testing"

	"go.uber.org/fx"

	libraryin "ielts-practice/app/modules/library/ports/in"
	"ielts-practice/seeds"
)

func TestLibrary(t *testing.T) {
	var seeder libraryin.ILibrarySeeder
	h := newHarness(t, fx.Populate(&seeder))
	if _, err := testDB.Exec("TRUNCATE library_collections CASCADE"); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// The committed seed: idempotent.
	res, err := seeder.SeedLibrary(ctx, seeds.Library, "library")
	if err != nil || res.Topics < 1 || res.Words < 30 {
		t.Fatalf("first seed: %+v %v", res, err)
	}
	if again, err := seeder.SeedLibrary(ctx, seeds.Library, "library"); err != nil || again.Topics != 0 || again.Unchanged != res.Topics {
		t.Fatalf("second seed: %+v %v", again, err)
	}

	access, _ := h.signIn()
	if r := h.do("GET", "/api/v1/library/collections", nil, ""); r.Status != http.StatusUnauthorized {
		t.Fatalf("collections without sign-in: %d", r.Status)
	}
	cols := h.mustStatus(h.do("GET", "/api/v1/library/collections", nil, access), http.StatusOK, "collections").Body["collections"].([]any)
	col := cols[0].(map[string]any)
	if len(cols) != 1 || col["slug"] != "ielts-academic" || col["words"].(float64) != float64(res.Words) {
		t.Fatalf("collections: %v", cols)
	}
	if topics := col["topics"].([]any); topics[0].(map[string]any)["id"] != "environment" {
		t.Fatalf("topics in collection order: %v", topics)
	}

	words := func(query string) []map[string]any {
		t.Helper()
		r := h.mustStatus(h.do("GET", "/api/v1/library/collections/ielts-academic/words"+query, nil, access), http.StatusOK, "words "+query)
		out := []map[string]any{}
		for _, w := range r.Body["words"].([]any) {
			out = append(out, w.(map[string]any))
		}
		return out
	}
	env := words("?topic=environment")
	if len(env) != 30 || env[0]["card"].(map[string]any)["headword"] != "environment" || env[0]["in_deck"] != false {
		t.Fatalf("environment words: %d %v", len(env), env[0])
	}
	// Search across topics, by English or by Vietnamese with or without accents.
	for _, q := range []string{"pollut", "o%20nhiem", "%C3%B4%20nhi%E1%BB%85m"} {
		found := words("?q=" + q)
		if len(found) == 0 || found[0]["card"].(map[string]any)["headword"] != "pollution" {
			t.Fatalf("search %q: %v", q, found)
		}
	}
	for _, w := range words("?topic=environment&level=C1") {
		if w["cefr"] != "C1" {
			t.Fatalf("level filter: %v", w["cefr"])
		}
	}
	if r := h.do("GET", "/api/v1/library/collections/toeic/words", nil, access); errCode(r) != "LIBRARY_NOT_FOUND" {
		t.Fatalf("unknown collection: %d %v", r.Status, r.Body)
	}

	// Add two words to a new deck at once (no AI); again, they are skipped.
	ids := []string{env[0]["lexeme_id"].(string), env[1]["lexeme_id"].(string)}
	added := h.mustStatus(h.do("POST", "/api/v1/library/add", map[string]any{"lexeme_ids": ids, "new_deck_name": "Environment"}, access),
		http.StatusCreated, "add")
	if added.Body["created"].(float64) != 2 || added.Body["deck_name"] != "Environment" {
		t.Fatalf("add: %v", added.Body)
	}
	deckID := added.Body["deck_id"].(string)
	again := h.mustStatus(h.do("POST", "/api/v1/library/add", map[string]any{"lexeme_ids": ids, "deck_id": deckID}, access),
		http.StatusCreated, "add again")
	if again.Body["created"].(float64) != 0 || len(again.Body["skipped"].([]any)) != 2 {
		t.Fatalf("add again: %v", again.Body)
	}
	if env = words("?topic=environment"); env[0]["in_deck"] != true || env[2]["in_deck"] != false {
		t.Fatalf("in_deck flags: %v %v", env[0]["in_deck"], env[2]["in_deck"])
	}
	// Only library cards can be added this way.
	var generated string
	if err := testDB.QueryRow(`INSERT INTO lexemes (key, headword, content, status) VALUES ('zz||vi', 'zz', '{"headword":"zz","meaning_vi":"x"}', 'ready') RETURNING id`).Scan(&generated); err != nil {
		t.Fatal(err)
	}
	if r := h.do("POST", "/api/v1/library/add", map[string]any{"lexeme_ids": []string{generated}, "deck_id": deckID}, access); errCode(r) != "VOCAB_NOTHING_TO_SAVE" {
		t.Fatalf("a generated lexeme must not be added through the library: %d %v", r.Status, r.Body)
	}

	// Pasting a library word reuses the reviewed card; a report never sends it back to the AI.
	job := h.mustStatus(h.do("POST", "/api/v1/vocab/jobs", entriesBody("pollution"), access), http.StatusCreated, "job")
	it := job.Body["items"].([]any)[0].(map[string]any)
	if it["status"] != "ready" || it["cached"] != true || it["lexeme_id"] != env[1]["lexeme_id"] {
		t.Fatalf("library card not reused: %v", it)
	}
	if _, err := testDB.Exec(`UPDATE lexemes SET report_count = 99 WHERE id = $1`, it["lexeme_id"]); err != nil {
		t.Fatal(err)
	}
	h.mustStatus(h.do("POST", "/api/v1/lexemes/"+it["lexeme_id"].(string)+"/report", map[string]any{"reason": "other"}, access), http.StatusNoContent, "report")
	var status string
	_ = testDB.QueryRow(`SELECT status FROM lexemes WHERE id = $1`, it["lexeme_id"]).Scan(&status)
	if status != "ready" {
		t.Fatalf("a reported library card stays ready, got %s", status)
	}
}
