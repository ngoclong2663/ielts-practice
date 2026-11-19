package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"go.uber.org/fx"

	libraryin "ielts-practice/app/modules/library/ports/in"
	notifyEntity "ielts-practice/app/modules/notify/domain/entity"
	notifyin "ielts-practice/app/modules/notify/ports/in"
	"ielts-practice/seeds"
)

// keepWord signs in learner n, pastes the word and saves its card to a new deck (the fake generator writes the
// card for the first learner; the others reuse it from the cache).
func (ah *adminHarness) keepWord(n int, word string) {
	ah.t.Helper()
	ah.google.identity.Subject, ah.google.identity.Email = fmt.Sprintf("google-sub-%d", n), fmt.Sprintf("learner%d@example.com", n)
	access, _ := ah.signIn()
	job := ah.mustStatus(ah.do("POST", "/api/v1/vocab/jobs", entriesBody(word), access), http.StatusCreated, "job")
	done := ah.waitJob(access, job.Body["id"].(string))
	item := jobItems(done)[word]
	ah.mustStatus(ah.do("POST", "/api/v1/vocab/jobs/"+job.Body["id"].(string)+"/commit",
		map[string]any{"new_deck_name": "Words", "item_ids": []string{item["id"].(string)}}, access), http.StatusCreated, "commit")
}

func TestLibraryCandidates(t *testing.T) {
	t.Setenv("LIBRARY_CANDIDATE_MIN_USERS", "3")
	var library libraryin.ILibrarySeeder
	ah := &adminHarness{notifier: &fakeNotifier{calls: make(chan notifyEntity.Notification, 50)}}
	ah.harness = newHarness(t,
		fx.Decorate(func(notifyin.IAdminNotifier) notifyin.IAdminNotifier { return ah.notifier }),
		fx.Populate(&ah.seeder, &library),
	)
	if _, err := testDB.Exec("TRUNCATE admin_accounts, library_collections CASCADE"); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := library.SeedLibrary(ctx, seeds.Library, "library"); err != nil {
		t.Fatal(err)
	}
	token := ah.adminToken()
	list := func() []any {
		t.Helper()
		r := ah.admin("GET", "/api/v1/admin/library/candidates", nil, token)
		if r.Status != http.StatusOK {
			t.Fatalf("candidates: %d %v", r.Status, r.Body)
		}
		return r.Body["items"].([]any)
	}

	// Admin only.
	learner, _ := ah.signIn()
	if r := ah.do("GET", "/api/v1/admin/library/candidates", nil, learner); r.Status != http.StatusUnauthorized && r.Status != http.StatusForbidden {
		t.Fatalf("a learner must not see candidates: %d", r.Status)
	}

	// Two learners keep "exacerbate": below the threshold. A third makes it a candidate.
	ah.keepWord(11, "exacerbate")
	ah.keepWord(12, "exacerbate")
	if items := list(); len(items) != 0 {
		t.Fatalf("2 learners < 3: %v", items)
	}
	ah.keepWord(13, "exacerbate")
	ah.keepWord(14, "mitigate")
	ah.keepWord(15, "mitigate")
	ah.keepWord(16, "mitigate")
	items := list()
	if len(items) != 2 {
		t.Fatalf("candidates: %v", items)
	}
	first := items[0].(map[string]any)
	if first["users"].(float64) != 3 || first["card"].(map[string]any)["headword"] == "" || first["suggested_topic"] != "environment" {
		t.Fatalf("candidate: %v", first)
	}
	ids := map[string]string{}
	for _, it := range items {
		m := it.(map[string]any)
		ids[m["card"].(map[string]any)["headword"].(string)] = m["lexeme_id"].(string)
	}

	// Unknown topic refused; approve "exacerbate" into a new topic, reject "mitigate".
	if r := ah.admin("POST", "/api/v1/admin/library/candidates/approve", map[string]any{"items": []any{map[string]any{"lexeme_id": ids["exacerbate"], "topic_id": "astrology"}}}, token); r.Status != http.StatusBadRequest {
		t.Fatalf("unknown topic: %d", r.Status)
	}
	ok := ah.admin("POST", "/api/v1/admin/library/candidates/approve", map[string]any{"items": []any{map[string]any{"lexeme_id": ids["exacerbate"], "topic_id": "other"}}}, token)
	if ok.Status != http.StatusOK || ok.Body["count"].(float64) != 1 {
		t.Fatalf("approve: %d %v", ok.Status, ok.Body)
	}
	rej := ah.admin("POST", "/api/v1/admin/library/candidates/reject", map[string]any{"lexeme_ids": []string{ids["mitigate"]}}, token)
	if rej.Status != http.StatusOK || rej.Body["count"].(float64) != 1 {
		t.Fatalf("reject: %d %v", rej.Status, rej.Body)
	}
	if items := list(); len(items) != 0 {
		t.Fatalf("decided words are not offered again: %v", items)
	}

	// The approved word is in the library, its card locked; approving it again does nothing.
	var origin string
	_ = testDB.QueryRow(`SELECT origin FROM lexemes WHERE id = $1`, ids["exacerbate"]).Scan(&origin)
	if origin != "library" {
		t.Fatalf("approved card origin %q", origin)
	}
	words := ah.mustStatus(ah.do("GET", "/api/v1/library/collections/ielts-academic/words?topic=other", nil, learner), http.StatusOK, "words")
	got := words.Body["words"].([]any)
	if len(got) != 1 || got[0].(map[string]any)["source"] != "learner" {
		t.Fatalf("other topic: %v", got)
	}
	again := ah.admin("POST", "/api/v1/admin/library/candidates/approve", map[string]any{"items": []any{map[string]any{"lexeme_id": ids["exacerbate"], "topic_id": "health"}}}, token)
	if again.Body["count"].(float64) != 0 {
		t.Fatalf("approved twice: %v", again.Body)
	}

	// A re-seed of the collection keeps the words learners added.
	if _, err := testDB.Exec(`UPDATE library_topics SET content_hash = ''`); err != nil {
		t.Fatal(err)
	}
	if _, err := testDB.Exec(`UPDATE library_words SET library_topic_id = (SELECT id FROM library_topics WHERE topic_id = 'environment') WHERE lexeme_id = $1`, ids["exacerbate"]); err != nil {
		t.Fatal(err)
	}
	if _, err := library.SeedLibrary(ctx, seeds.Library, "library"); err != nil {
		t.Fatal(err)
	}
	var n int
	_ = testDB.QueryRow(`SELECT count(*) FROM library_words WHERE lexeme_id = $1`, ids["exacerbate"]).Scan(&n)
	if n != 1 {
		t.Fatalf("re-seed removed the learner word")
	}
}
