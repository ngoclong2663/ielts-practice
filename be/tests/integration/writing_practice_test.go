package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ielts-practice/seeds"
)

// doList is h.do for endpoints that return a JSON array.
func (h *harness) doList(path, access string) []map[string]any {
	h.t.Helper()
	req := httptest.NewRequest("GET", path, &bytes.Buffer{})
	req.AddCookie(&http.Cookie{Name: "access_token", Value: access})
	w := httptest.NewRecorder()
	h.engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		h.t.Fatalf("GET %s: %d %s", path, w.Code, w.Body.String())
	}
	var out []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		h.t.Fatal(err)
	}
	return out
}

// tasksOnly hides the seeded lessons and vocabulary, so the fake coach writes them (the AI path).
type tasksOnly struct{ fs.FS }

func (f tasksOnly) Open(name string) (fs.File, error) {
	if strings.Contains(name, "/lessons") || strings.Contains(name, "/vocab") {
		return nil, fs.ErrNotExist
	}
	return f.FS.Open(name)
}

func (f tasksOnly) ReadDir(name string) ([]fs.DirEntry, error) {
	if strings.Contains(name, "/lessons") || strings.Contains(name, "/vocab") {
		return nil, fs.ErrNotExist
	}
	entries, err := fs.ReadDir(f.FS, name)
	out := entries[:0]
	for _, e := range entries {
		if e.Name() != "lessons" && e.Name() != "vocab" {
			out = append(out, e)
		}
	}
	return out, err
}

// newWritingHarness seeds the task packs without their lessons (fake coach, in-memory S3) and signs in.
func newWritingHarness(t *testing.T) (*harness, string, map[string]string) {
	t.Helper()
	seeder, _, h := newSeedHarnessH(t)
	if _, err := seeder.SeedTasks(context.Background(), tasksOnly{seeds.Writing}, "writing"); err != nil {
		t.Fatal(err)
	}
	access, _ := h.signIn()
	slugs := map[string]string{}
	rows, err := testDB.Query(`SELECT slug, id FROM writing_tasks`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var slug, id string
		_ = rows.Scan(&slug, &id)
		slugs[slug] = id
	}
	return h, access, slugs
}

func words(n int) string {
	return strings.TrimSpace(strings.Repeat("The number of workers rose steadily over the period. ", n/8+1))
}

func TestWritingPracticeFlow(t *testing.T) {
	h, access, ids := newWritingHarness(t)
	task1 := ids["cam21-test1-task1"]

	// Task bank
	list := h.doList("/api/v1/writing/tasks?task_no=1", access)
	if len(list) != 12 { // 4 per pack: cam21, cam20, cam19 (newest book first)
		t.Fatalf("task 1 list: %d", len(list))
	}
	first := list[0]
	fig := first["figure"].(map[string]any)
	if !strings.HasPrefix(fig["url"].(string), "/api/v1/writing/figures/") {
		t.Fatalf("summary: %v", first)
	}
	// The reference book is internal: neither source_ref nor a "source" label is sent.
	for _, leaked := range []string{"source_ref", "source"} {
		if _, ok := first[leaked]; ok {
			t.Fatalf("%s must not be sent", leaked)
		}
	}
	detail := h.mustStatus(h.do("GET", "/api/v1/writing/tasks/"+ids["cam21-test4-task2"], nil, access), 200, "task detail")
	if got := detail.Body["positions"].([]any); len(got) != 3 || got[0] != "agree" {
		t.Fatalf("positions: %v", got)
	}
	status := h.mustStatus(h.do("GET", "/api/v1/writing/status", nil, access), 200, "status")
	if !status.Body["configured"].(bool) || status.Body["checks"].(map[string]any)["remaining"].(float64) != 30 {
		t.Fatalf("status: %v", status.Body)
	}

	// Positions are validated for Task 2.
	bad := h.do("POST", "/api/v1/writing/tasks/"+ids["cam21-test1-task2"]+"/attempts", map[string]any{"level": 1, "position": "favour_first"}, access)
	if bad.Status != 400 || errCode(bad) != "WRITING_INVALID_POSITION" {
		t.Fatalf("invalid position: %d %s", bad.Status, errCode(bad))
	}

	// Level 1 attempt: Vietnamese prompts, hints, no answers or models yet.
	att := h.mustStatus(h.do("POST", "/api/v1/writing/tasks/"+task1+"/attempts", map[string]any{"level": 1}, access), 201, "start")
	id := att.Body["id"].(string)
	paras := att.Body["paragraphs"].([]any)
	if len(paras) != 4 || paras[0].(map[string]any)["key"] != "introduction" || paras[1].(map[string]any)["key"] != "overview" {
		t.Fatalf("paragraphs: %v", paras)
	}
	s0 := paras[0].(map[string]any)["sentences"].([]any)[0].(map[string]any)
	if !strings.HasPrefix(s0["prompt"].(string), "Câu 1") || len(s0["hints"].([]any)) == 0 || s0["model_en"] != nil {
		t.Fatalf("level 1 sentence: %v", s0)
	}
	q0 := att.Body["questions"].([]any)[0].(map[string]any)
	if q0["answer"] != nil || len(att.Body["ideas_vi"].([]any)) == 0 || att.Body["model_answer"] != nil {
		t.Fatalf("hidden fields leaked: %v", att.Body)
	}

	// Useful vocabulary (Epic 7): written once per lesson, in group order.
	if att.Body["vocab_available"] != true {
		t.Fatalf("level 1 vocab should be available: %v", att.Body["vocab_available"])
	}
	vocab := h.mustStatus(h.do("GET", "/api/v1/writing/attempts/"+id+"/vocab", nil, access), 200, "vocab")
	groups := vocab.Body["groups"].([]any)
	if len(groups) != 3 || groups[0].(map[string]any)["key"] != "topic" || groups[2].(map[string]any)["key"] != "task" {
		t.Fatalf("vocab groups: %v", groups)
	}
	item := groups[1].(map[string]any)["items"].([]any)[0].(map[string]any)
	if item["term"] != "show" || len(item["synonyms"].([]any)) != 2 || item["example_en"] == "" {
		t.Fatalf("vocab item: %v", item)
	}
	h.mustStatus(h.do("GET", "/api/v1/writing/attempts/"+id+"/vocab", nil, access), 200, "vocab again")
	var vocabRows, vocabCalls int
	_ = testDB.QueryRow(`SELECT count(*) FROM writing_task_vocab`).Scan(&vocabRows)
	_ = testDB.QueryRow(`SELECT count(*) FROM writing_ai_requests WHERE kind = 'vocab'`).Scan(&vocabCalls)
	if vocabRows != 1 || vocabCalls != 1 {
		t.Fatalf("vocab rows %d, vocab calls %d", vocabRows, vocabCalls)
	}

	// Step 0
	if r := h.do("POST", "/api/v1/writing/attempts/"+id+"/plan", map[string]any{"answers": []int{0}}, access); r.Status != 400 {
		t.Fatalf("plan with too few answers: %d", r.Status)
	}
	plan := h.mustStatus(h.do("POST", "/api/v1/writing/attempts/"+id+"/plan", map[string]any{"answers": []int{0, 2}}, access), 200, "plan")
	if plan.Body["questions"].([]any)[1].(map[string]any)["answer"].(float64) != 1 {
		t.Fatalf("plan answers not revealed: %v", plan.Body["questions"])
	}

	// Paragraph check: literal translation and an empty sentence; the model sentences appear.
	chk := h.mustStatus(h.do("POST", "/api/v1/writing/attempts/"+id+"/paragraphs/introduction/check",
		map[string]any{"answers": []string{"Old people  drink medicine every day.", ""}}, access), 200, "check")
	intro := chk.Body["paragraphs"].([]any)[0].(map[string]any)
	fb := intro["feedback"].([]any)
	if fb[0].(map[string]any)["verdict"] != "understandable" || fb[1].(map[string]any)["verdict"] != "empty" {
		t.Fatalf("feedback: %v", fb)
	}
	if intro["answers"].([]any)[0] != "Old people drink medicine every day." || intro["checked_count"].(float64) != 1 {
		t.Fatalf("answers not normalised/stored: %v", intro)
	}
	if intro["sentences"].([]any)[0].(map[string]any)["model_en"] == nil ||
		chk.Body["paragraphs"].([]any)[1].(map[string]any)["sentences"].([]any)[0].(map[string]any)["model_en"] != nil {
		t.Fatal("model sentences must appear only for checked paragraphs")
	}
	if r := h.do("POST", "/api/v1/writing/attempts/"+id+"/paragraphs/conclusion/check", map[string]any{"answers": []string{"x", "y"}}, access); r.Status != 400 {
		t.Fatalf("Task 1 has no conclusion: %d", r.Status)
	}

	// Submit
	short := h.do("POST", "/api/v1/writing/attempts/"+id+"/submit", map[string]any{"text": "Too short."}, access)
	if short.Status != 400 || errCode(short) != "WRITING_TEXT_TOO_SHORT" {
		t.Fatalf("short answer: %d %s", short.Status, errCode(short))
	}
	sub := h.mustStatus(h.do("POST", "/api/v1/writing/attempts/"+id+"/submit", map[string]any{"text": words(160), "seconds_spent": 900}, access), 200, "submit")
	s := sub.Body["submission"].(map[string]any)
	if sub.Body["status"] != "submitted" || s["band_low"].(float64) != 5.5 || s["band_high"].(float64) != 6 || sub.Body["model_answer"] == "" {
		t.Fatalf("submission: %v", sub.Body)
	}
	if len(s["feedback"].(map[string]any)["criteria"].([]any)) != 4 {
		t.Fatalf("criteria: %v", s["feedback"])
	}
	if r := h.do("POST", "/api/v1/writing/attempts/"+id+"/submit", map[string]any{"text": words(160)}, access); r.Status != 409 {
		t.Fatalf("second submit: %d", r.Status)
	}
	if r := h.do("POST", "/api/v1/writing/attempts/"+id+"/paragraphs/body1/check", map[string]any{"answers": []string{"a b c d", "e f g h"}}, access); r.Status != 409 {
		t.Fatalf("check after submit: %d", r.Status)
	}

	// A second attempt reuses the lesson (one lesson AI call in total).
	lvl3 := h.mustStatus(h.do("POST", "/api/v1/writing/tasks/"+task1+"/attempts", map[string]any{"level": 3}, access), 201, "level 3")
	p3 := lvl3.Body["paragraphs"].([]any)[0].(map[string]any)["sentences"].([]any)[0].(map[string]any)
	if p3["prompt"] != "introduction / sentence 1" || len(p3["hints"].([]any)) != 0 {
		t.Fatalf("level 3 sentence: %v", p3)
	}
	lvl4 := h.mustStatus(h.do("POST", "/api/v1/writing/tasks/"+task1+"/attempts", map[string]any{"level": 4}, access), 201, "level 4")
	if len(lvl4.Body["paragraphs"].([]any)) != 0 || len(lvl4.Body["ideas_vi"].([]any)) != 0 || lvl4.Body["vocab_available"] != false {
		t.Fatalf("level 4 should be free writing: %v", lvl4.Body)
	}
	if r := h.do("GET", "/api/v1/writing/attempts/"+lvl4.Body["id"].(string)+"/vocab", nil, access); r.Status != 403 || errCode(r) != "WRITING_VOCAB_LOCKED" {
		t.Fatalf("level 4 vocab before submit: %d", r.Status)
	}
	var lessons, lessonCalls int
	_ = testDB.QueryRow(`SELECT count(*) FROM writing_lessons`).Scan(&lessons)
	_ = testDB.QueryRow(`SELECT count(*) FROM writing_ai_requests WHERE kind = 'lesson'`).Scan(&lessonCalls)
	if lessons != 1 || lessonCalls != 1 {
		t.Fatalf("lessons %d, lesson calls %d", lessons, lessonCalls)
	}

	// History and progress
	hist := h.doList("/api/v1/writing/attempts", access)
	if len(hist) != 3 || hist[2]["band_high"].(float64) != 6 {
		t.Fatalf("history: %v", hist)
	}
	again := h.doList("/api/v1/writing/tasks?task_no=1", access)
	for _, task := range again {
		if task["id"] == task1 && (task["attempts"].(float64) != 3 || task["best_band"].(float64) != 6) {
			t.Fatalf("progress: %v", task)
		}
	}
	st := h.mustStatus(h.do("GET", "/api/v1/writing/status", nil, access), 200, "status")
	if st.Body["checks"].(map[string]any)["used"].(float64) != 1 || st.Body["submissions"].(map[string]any)["used"].(float64) != 1 {
		t.Fatalf("usage: %v", st.Body)
	}

	// Another learner cannot open the attempt.
	h.google.identity.Subject, h.google.identity.Email = "google-sub-2", "second@example.com"
	other, _ := h.signIn()
	if r := h.do("GET", "/api/v1/writing/attempts/"+id, nil, other); r.Status != 404 {
		t.Fatalf("other user's attempt: %d", r.Status)
	}
}

func TestWritingDailyLimits(t *testing.T) {
	t.Setenv("WRITING_DAILY_CHECKS", "1")
	t.Setenv("WRITING_DAILY_SUBMISSIONS", "1")
	h, access, ids := newWritingHarness(t)
	start := func() string {
		r := h.mustStatus(h.do("POST", "/api/v1/writing/tasks/"+ids["cam21-test1-task2"]+"/attempts",
			map[string]any{"level": 2, "position": "agree"}, access), 201, "start")
		return r.Body["id"].(string)
	}
	a, b := start(), start()
	check := map[string]any{"answers": []string{"Cities are growing quickly today.", "Tall buildings save a lot of land."}}
	h.mustStatus(h.do("POST", "/api/v1/writing/attempts/"+a+"/paragraphs/introduction/check", check, access), 200, "check")
	if r := h.do("POST", "/api/v1/writing/attempts/"+a+"/paragraphs/body1/check", map[string]any{"answers": []string{"a b c d", "e f g h"}}, access); r.Status != 429 || errCode(r) != "WRITING_CHECK_LIMIT" {
		t.Fatalf("check limit: %d %s", r.Status, errCode(r))
	}
	h.mustStatus(h.do("POST", "/api/v1/writing/attempts/"+a+"/submit", map[string]any{"text": words(260)}, access), 200, "submit")
	if r := h.do("POST", "/api/v1/writing/attempts/"+b+"/submit", map[string]any{"text": words(260)}, access); r.Status != 429 || errCode(r) != "WRITING_SUBMIT_LIMIT" {
		t.Fatalf("submit limit: %d %s", r.Status, errCode(r))
	}
}
