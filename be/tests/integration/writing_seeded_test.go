package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"ielts-practice/seeds"
)

// seededPack is the committed cam21 pack reduced to cam21-test1-task1, plus a lesson and a vocabulary file.
func seededPack(t *testing.T, lesson, vocab string) fstest.MapFS {
	t.Helper()
	raw, err := fs.ReadFile(seeds.Writing, "writing/cam21/pack.json")
	if err != nil {
		t.Fatal(err)
	}
	var pack map[string]any
	_ = json.Unmarshal(raw, &pack)
	var task map[string]any
	for _, x := range pack["tasks"].([]any) {
		if x.(map[string]any)["slug"] == "cam21-test1-task1" {
			task = x.(map[string]any)
		}
	}
	pack["tasks"] = []any{task}
	packJSON, _ := json.Marshal(pack)
	m := fstest.MapFS{"writing/cam21/pack.json": {Data: packJSON}}
	for _, f := range task["figures"].([]any) {
		name := f.(map[string]any)["file"].(string)
		data, err := fs.ReadFile(seeds.Writing, "writing/cam21/"+name)
		if err != nil {
			t.Fatal(err)
		}
		m["writing/cam21/"+name] = &fstest.MapFile{Data: data}
	}
	if lesson != "" {
		m["writing/cam21/lessons/cam21-test1-task1.json"] = &fstest.MapFile{Data: []byte(lesson)}
	}
	if vocab != "" {
		m["writing/cam21/vocab/cam21-test1-task1.json"] = &fstest.MapFile{Data: []byte(vocab)}
	}
	return m
}

func testLesson(taskHash, word string) string {
	sentence := func(n int) string {
		return fmt.Sprintf(`{"vi":"Câu %d.","note_vi":"ý chính %d","outline_en":"graph / jobs / %d","hints":[{"en":"rise steadily","vi":"tăng đều"}],"model_en":"The %s figures in sentence %d show a clear and steady change across the four industries over the whole period."}`, n, n, n, word, n)
	}
	para := func(key string, from, count int) string {
		var s []string
		for i := 0; i < count; i++ {
			s = append(s, sentence(from+i))
		}
		return fmt.Sprintf(`{"key":"%s","goal_vi":"Mục tiêu","sentences":[%s]}`, key, strings.Join(s, ","))
	}
	return fmt.Sprintf(`{"task":"cam21-test1-task1","position":"","task_hash":%q,"lesson":{"ideas_vi":["Ý 1","Ý 2","Ý 3"],
		"questions":[{"question_vi":"Biểu đồ cho thấy gì?","options_vi":["Việc làm","Dân số"],"answer":0,"explain_vi":"Đúng."}],
		"paragraphs":[%s,%s,%s,%s]}}`, taskHash, para("introduction", 1, 1), para("overview", 2, 2), para("body1", 4, 3), para("body2", 7, 3))
}

func testVocabCard(term string) string {
	return fmt.Sprintf(`"image_query":"factory workers","visualizable":false,"card":{"headword":%q,"lemma":%q,
		"ipa":{"uk":"/x/","us":"/x/"},"part_of_speech":"noun","cefr":"B2","meaning_vi":"nghĩa của %s","definition_en":"a test meaning",
		"word_forms":[{"form":%q,"pos":"noun","ipa":"/x/","meaning_vi":"nghĩa","example":{"en":"The %s changed a lot over the period.","vi":"Ví dụ."}}],
		"collocations":[{"phrase":"a %s","meaning_vi":"một"}],"extra_examples":[],"synonyms":[],"other_senses":[]}`, term, term, term, term, term, term)
}

func testVocab() string {
	item := func(term, card string) string {
		return fmt.Sprintf(`{"term":%q,"pos":"phrase","vi":"nghĩa %s","synonyms":["a b"],"example_en":"An example with %s in it here.","note_vi":"",%s}`, term, term, term, card)
	}
	var items []string
	for _, term := range []string{"labour market", "job sector", "service industry", "employment rate", "manufacturing jobs"} {
		items = append(items, item(term, testVocabCard(term)))
	}
	items = append(items, item("pollution", `"card_ref":"pollution"`))
	return `{"task":"cam21-test1-task1","groups":[{"key":"topic","items":[` + strings.Join(items, ",") + `]}]}`
}

func TestSeededLessonAndVocab(t *testing.T) {
	seeder, _, h := newSeedHarnessH(t)
	ctx := context.Background()
	if _, err := seeder.SeedTasks(ctx, seededPack(t, "", ""), "writing"); err != nil {
		t.Fatal(err)
	}
	var taskID, hash string
	if err := testDB.QueryRow(`SELECT id, content_hash FROM writing_tasks WHERE slug = 'cam21-test1-task1'`).Scan(&taskID, &hash); err != nil {
		t.Fatal(err)
	}

	// A broken lesson is refused; a stale one (another task text) is skipped.
	if _, err := seeder.SeedTasks(ctx, seededPack(t, `{"task":"cam21-test1-task1","position":"","task_hash":"x","lesson":{}}`, ""), "writing"); err == nil {
		t.Fatal("an empty lesson must be refused")
	}
	res, err := seeder.SeedTasks(ctx, seededPack(t, testLesson(strings.Repeat("0", 64), "stale"), ""), "writing")
	if err != nil || res.Lessons != 0 {
		t.Fatalf("stale lesson: %+v %v", res, err)
	}

	res, err = seeder.SeedTasks(ctx, seededPack(t, testLesson(hash, "employment"), testVocab()), "writing")
	if err != nil || res.Lessons != 1 || res.Vocab != 1 {
		t.Fatalf("seed: %+v %v", res, err)
	}
	if again, _ := seeder.SeedTasks(ctx, seededPack(t, testLesson(hash, "employment"), testVocab()), "writing"); again.Lessons != 0 || again.Vocab != 0 {
		t.Fatalf("second seed: %+v", again)
	}

	// The attempt uses the seeded lesson (no AI call) and the vocabulary has a card per term.
	access, _ := h.signIn()
	att := h.mustStatus(h.do("POST", "/api/v1/writing/tasks/"+taskID+"/attempts", map[string]any{"level": 1}, access), 201, "start")
	var model string
	_ = testDB.QueryRow(`SELECT l.model FROM writing_lessons l JOIN writing_attempts a ON a.lesson_id = l.id WHERE a.id = $1`, att.Body["id"]).Scan(&model)
	if model != "seed" {
		t.Fatalf("attempt lesson model %q, want seed", model)
	}
	vocab := h.mustStatus(h.do("GET", "/api/v1/writing/attempts/"+att.Body["id"].(string)+"/vocab", nil, access), 200, "vocab")
	var ids []string
	for _, g := range vocab.Body["groups"].([]any) {
		for _, it := range g.(map[string]any)["items"].([]any) {
			id, _ := it.(map[string]any)["lexeme_id"].(string)
			if id == "" {
				t.Fatalf("item without a card: %v", it)
			}
			ids = append(ids, id)
		}
	}
	if len(ids) != 6 {
		t.Fatalf("items: %d", len(ids))
	}
	var pollutionID string
	_ = testDB.QueryRow(`SELECT id FROM lexemes WHERE key = 'pollution||vi'`).Scan(&pollutionID)
	if ids[5] != pollutionID {
		t.Fatalf("card_ref must reuse the library card")
	}
	added := h.mustStatus(h.do("POST", "/api/v1/library/add", map[string]any{"lexeme_ids": ids, "new_deck_name": "Writing"}, access), 201, "add")
	if added.Body["created"].(float64) != 6 {
		t.Fatalf("add: %v", added.Body)
	}
}
