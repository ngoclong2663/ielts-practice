package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/fx"

	examin "ielts-practice/app/modules/exam/ports/in"
	"ielts-practice/seeds"
)

// TestExamReading covers Epic 10 Phase 1: the seed, the test bank, the player content (no key), marking of
// part, full and redo attempts, the review data, history and stats.
func TestExamReading(t *testing.T) {
	var seeder examin.IExamSeeder
	h := newHarness(t, fx.Populate(&seeder))
	if _, err := testDB.Exec("TRUNCATE exam_attempts, exam_keys, exam_parts, exam_tests, exam_books CASCADE"); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	res, err := seeder.SeedExam(ctx, seeds.Exam, "exam")
	if err != nil || res.Inserted < 3 {
		t.Fatalf("first seed: %+v %v", res, err)
	}
	if again, _ := seeder.SeedExam(ctx, seeds.Exam, "exam"); again.Unchanged != res.Inserted || again.Inserted != 0 {
		t.Fatalf("second seed must change nothing: %+v", again)
	}

	access, _ := h.signIn()
	if r := h.do("GET", "/api/v1/exam/books?skill=reading", nil, ""); r.Status != 401 {
		t.Fatalf("books without sign-in: %d", r.Status)
	}
	if r := h.do("GET", "/api/v1/exam/books?skill=maths", nil, access); r.Status != 400 {
		t.Fatalf("unknown skill: %d", r.Status)
	}
	books := h.doList("/api/v1/exam/books?skill=reading", access)
	if len(books) == 0 {
		t.Fatal("no books")
	}
	book := books[0]
	tests := book["tests"].([]any)
	test1 := tests[0].(map[string]any)
	parts := test1["parts"].([]any)
	if book["slug"] != "s1" || test1["slug"] != "s1-t1" || len(parts) != 3 {
		t.Fatalf("bank: %v", book)
	}
	if _, leaked := book["source_ref"]; leaked {
		t.Fatal("source_ref must not be returned")
	}
	if p := parts[0].(map[string]any); p["best"] != nil || p["q_from"].(float64) != 1 || len(p["question_types"].([]any)) == 0 {
		t.Fatalf("part summary: %v", p)
	}

	// The player content never carries the key.
	tr := h.do("GET", "/api/v1/exam/tests/s1-t1?skill=reading", nil, access)
	if tr.Status != 200 || len(tr.Body["parts"].([]any)) != 3 {
		t.Fatalf("test: %d %v", tr.Status, tr.Body)
	}
	for _, p := range tr.Body["parts"].([]any) {
		if _, ok := p.(map[string]any)["key"]; ok {
			t.Fatal("the player content must not include the key")
		}
	}
	if r := h.do("GET", "/api/v1/exam/tests/nope?skill=reading", nil, access); r.Status != 404 || errCode(r) != "EXAM_TEST_NOT_FOUND" {
		t.Fatalf("missing test: %d %v", r.Status, r.Body)
	}
	// Listening: the player gets the audio and reading pauses, but neither the transcript nor the key.
	lt := h.do("GET", "/api/v1/exam/tests/s1-t1?skill=listening", nil, access)
	if lt.Status != 200 || len(lt.Body["parts"].([]any)) != 4 {
		t.Fatalf("listening test: %d %v", lt.Status, lt.Body)
	}
	lp := lt.Body["parts"].([]any)[0].(map[string]any)
	au := lp["audio"].(map[string]any)
	if !strings.HasPrefix(au["url"].(string), "/api/v1/exam/media/") || au["duration"].(float64) <= 0 || len(au["cues"].([]any)) == 0 {
		t.Fatalf("listening audio: %v", au)
	}
	if _, ok := lp["key"]; ok {
		t.Fatal("the listening player must not get the key")
	}
	if _, ok := lp["content"].(map[string]any)["lines"]; ok {
		t.Fatal("the transcript must only come with the review")
	}
	ls := h.do("POST", "/api/v1/exam/attempts", map[string]any{"test": "s1-t1", "skill": "listening", "kind": "part", "parts": []int{1},
		"mode": "exam", "answers": map[string]any{"1": "harness", "2": "eight"}}, access)
	la := h.do("GET", "/api/v1/exam/attempts/"+ls.Body["id"].(string), nil, access)
	lreview := la.Body["test"].(map[string]any)["parts"].([]any)[0].(map[string]any)
	if la.Body["correct"].(float64) != 2 || la.Body["mode"] != "exam" || len(lreview["content"].(map[string]any)["lines"].([]any)) == 0 || lreview["key"] == nil {
		t.Fatalf("listening review: correct=%v mode=%v", la.Body["correct"], la.Body["mode"])
	}
	// Media links are public (an <audio> element cannot refresh a session); without S3 they are not found.
	if r := h.do("GET", au["url"].(string), nil, ""); r.Status != 404 || errCode(r) != "EXAM_MEDIA_NOT_FOUND" {
		t.Fatalf("media without S3: %d %v", r.Status, r.Body)
	}

	// Part 1 only: 4 right (1, 2, 8 and 12); 6 is over the word limit and 9 is wrong.
	sub := h.do("POST", "/api/v1/exam/attempts", map[string]any{
		"test": "s1-t1", "skill": "reading", "kind": "part", "parts": []int{1}, "time_limit_sec": 1200, "duration_sec": 900,
		"answers": map[string]any{"1": "london", "2": " Flowers. ", "6": "Charles Darwin", "8": "TRUE", "9": "TRUE", "12": "FALSE", "30": "A"},
	}, access)
	if sub.Status != 201 {
		t.Fatalf("submit part: %d %v", sub.Status, sub.Body)
	}
	partID := sub.Body["id"].(string)
	a := h.do("GET", "/api/v1/exam/attempts/"+partID, nil, access)
	if a.Status != 200 || a.Body["correct"].(float64) != 4 || a.Body["total"].(float64) != 13 || a.Body["band"] != nil {
		t.Fatalf("part attempt: %v", a.Body)
	}
	answers := a.Body["answers"].(map[string]any)
	if _, kept := answers["30"]; kept {
		t.Fatal("answers outside the run must not be stored")
	}
	var q6 map[string]any
	for _, r := range a.Body["results"].([]any) {
		if r.(map[string]any)["n"].(float64) == 6 {
			q6 = r.(map[string]any)
		}
	}
	if q6["correct"] != false || q6["over_limit"] != true {
		t.Fatalf("q6 must be over the limit: %v", q6)
	}
	review := a.Body["test"].(map[string]any)["parts"].([]any)
	if len(review) != 1 || review[0].(map[string]any)["key"].(map[string]any)["1"] == nil {
		t.Fatalf("review must carry the key of the attempt's part only: %v", review)
	}

	// The bank now shows the best score of part 1.
	books = h.doList("/api/v1/exam/books?skill=reading", access)
	best := books[0]["tests"].([]any)[0].(map[string]any)["parts"].([]any)[0].(map[string]any)["best"].(map[string]any)
	if best["correct"].(float64) != 4 || best["total"].(float64) != 13 {
		t.Fatalf("best: %v", best)
	}

	// A full test gets a band: 23 right is band 6.0.
	full := map[string]any{}
	keyRows, err := testDB.Query(`SELECT k.key FROM exam_keys k JOIN exam_parts p ON p.id = k.part_id WHERE p.slug LIKE 's1-t1-r%'`)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]string{}
	for keyRows.Next() {
		var raw []byte
		_ = keyRows.Scan(&raw)
		for n, item := range decodeKey(t, raw) {
			keys[n] = item
		}
	}
	_ = keyRows.Close()
	correctWanted := 23
	given := 0
	for n := 1; n <= 40 && given < correctWanted; n++ {
		if v, ok := keys[fmt.Sprint(n)]; ok {
			full[fmt.Sprint(n)] = v
			given++
		}
	}
	fs := h.do("POST", "/api/v1/exam/attempts", map[string]any{"test": "s1-t1", "skill": "reading", "kind": "full", "answers": full,
		"mode": "exam", "time_limit_sec": 3600, "duration_sec": 3500}, access)
	fa := h.do("GET", "/api/v1/exam/attempts/"+fs.Body["id"].(string), nil, access)
	if fa.Body["total"].(float64) != 40 || fa.Body["correct"].(float64) != 23 || fa.Body["band"].(float64) != 6 {
		t.Fatalf("full attempt: %v %v", fa.Body["correct"], fa.Body["band"])
	}

	// Redo the wrong questions of the part attempt.
	redo := h.do("POST", "/api/v1/exam/attempts", map[string]any{"test": "s1-t1", "skill": "reading", "kind": "redo",
		"parent_attempt_id": partID, "questions": []int{6, 9}, "answers": map[string]any{"6": "Darwin", "9": "FALSE"}}, access)
	ra := h.do("GET", "/api/v1/exam/attempts/"+redo.Body["id"].(string), nil, access)
	if ra.Body["total"].(float64) != 2 || ra.Body["correct"].(float64) != 2 || ra.Body["kind"] != "redo" || ra.Body["parent_attempt_id"] != partID {
		t.Fatalf("redo: %v", ra.Body)
	}
	if r := h.do("POST", "/api/v1/exam/attempts", map[string]any{"test": "s1-t1", "skill": "reading", "kind": "redo",
		"parent_attempt_id": partID, "questions": []int{20}, "answers": map[string]any{}}, access); errCode(r) != "EXAM_INVALID_SELECTION" {
		t.Fatalf("a redo of a question outside the parent: %v", r.Body)
	}
	if r := h.do("POST", "/api/v1/exam/attempts", map[string]any{"test": "s1-t1", "skill": "reading", "kind": "part", "parts": []int{4},
		"answers": map[string]any{}}, access); r.Status != 400 {
		t.Fatalf("a part that does not exist: %d", r.Status)
	}

	// History (newest first) and stats (redo left out).
	hist := h.do("GET", "/api/v1/exam/attempts?skill=reading", nil, access)
	items := hist.Body["items"].([]any)
	if len(items) != 3 || items[0].(map[string]any)["kind"] != "redo" {
		t.Fatalf("history: %v", hist.Body)
	}
	stats := h.do("GET", "/api/v1/exam/stats?skill=reading", nil, access)
	if stats.Body["attempts"].(float64) != 2 || len(stats.Body["bands"].([]any)) != 1 || len(stats.Body["by_type"].([]any)) == 0 {
		t.Fatalf("stats: %v", stats.Body)
	}

	// Another learner cannot open this attempt.
	h.google.identity.Subject, h.google.identity.Email = "google-sub-2", "other@example.com"
	other, _ := h.signIn()
	if r := h.do("GET", "/api/v1/exam/attempts/"+partID, nil, other); r.Status != 404 || errCode(r) != "EXAM_ATTEMPT_NOT_FOUND" {
		t.Fatalf("another learner's attempt: %d", r.Status)
	}
}

// decodeKey returns each question's first accepted answer.
func decodeKey(t *testing.T, raw []byte) map[string]string {
	t.Helper()
	var key map[string]struct {
		Answer []string `json:"answer"`
	}
	if err := json.Unmarshal(raw, &key); err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for n, k := range key {
		out[n] = k.Answer[0]
	}
	return out
}
