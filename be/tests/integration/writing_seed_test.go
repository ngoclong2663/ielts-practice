package integration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"go.uber.org/fx"

	storagein "ielts-practice/app/modules/storage/ports/in"
	writingin "ielts-practice/app/modules/writing/ports/in"
	"ielts-practice/seeds"
)

// memStorage records uploads instead of calling S3.
type memStorage struct {
	mu   sync.Mutex
	objs map[string]string // key -> content type
	puts int
}

func (m *memStorage) Configured() bool                     { return true }
func (m *memStorage) Ping(context.Context) error           { return nil }
func (m *memStorage) Delete(context.Context, string) error { return nil }
func (m *memStorage) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://bucket.example/" + key, nil
}
func (m *memStorage) Put(_ context.Context, key string, body io.Reader, _ int64, contentType string) error {
	if _, err := io.ReadAll(body); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objs[key] = contentType
	m.puts++
	return nil
}

func newSeedHarness(t *testing.T) (writingin.ITaskSeeder, *memStorage) {
	seeder, store, _ := newSeedHarnessH(t)
	return seeder, store
}

func newSeedHarnessH(t *testing.T) (writingin.ITaskSeeder, *memStorage, *harness) {
	t.Helper()
	store := &memStorage{objs: map[string]string{}}
	var seeder writingin.ITaskSeeder
	h := newHarness(t,
		fx.Decorate(func(storagein.IObjectStorage) storagein.IObjectStorage { return store }),
		fx.Populate(&seeder),
	)
	if _, err := testDB.Exec("TRUNCATE writing_tasks CASCADE"); err != nil {
		t.Fatal(err)
	}
	return seeder, store, h
}

func TestWritingSeedCam21(t *testing.T) {
	seeder, store := newSeedHarness(t)
	ctx := context.Background()

	res, err := seeder.SeedTasks(ctx, seeds.Writing, "writing")
	if err != nil {
		t.Fatal(err)
	}
	if res.Inserted != 8 || res.Updated != 0 || res.Unchanged != 0 || res.FiguresUploaded != 4 {
		t.Fatalf("first run: %+v", res)
	}
	for key, ct := range store.objs {
		if !strings.HasPrefix(key, "writing/figures/") || !strings.HasSuffix(key, ".svg") || ct != "image/svg+xml" {
			t.Fatalf("unexpected object %s (%s)", key, ct)
		}
	}

	var (
		taskNo, minWords, minutes int
		taskType, status          string
		figures, facts            []byte
	)
	row := testDB.QueryRow(`SELECT task_no, task_type, min_words, minutes, status, figures, facts
		FROM writing_tasks WHERE slug = 'cam21-test1-task1'`)
	if err := row.Scan(&taskNo, &taskType, &minWords, &minutes, &status, &figures, &facts); err != nil {
		t.Fatal(err)
	}
	if taskNo != 1 || taskType != "line_graph" || minWords != 150 || minutes != 20 || status != "published" {
		t.Fatalf("task row: %d %s %d %d %s", taskNo, taskType, minWords, minutes, status)
	}
	var figs []map[string]any
	_ = json.Unmarshal(figures, &figs)
	if len(figs) != 1 || figs[0]["width"].(float64) == 0 || store.objs[figs[0]["key"].(string)] == "" {
		t.Fatalf("figure not stored: %s", figures)
	}
	if !strings.Contains(string(facts), `"Construction"`) {
		t.Fatalf("facts not stored: %s", facts)
	}

	// Second run: nothing changes, nothing is uploaded.
	putsBefore := store.puts
	res, err = seeder.SeedTasks(ctx, seeds.Writing, "writing")
	if err != nil {
		t.Fatal(err)
	}
	if res.Unchanged != 8 || res.Inserted+res.Updated != 0 || store.puts != putsBefore {
		t.Fatalf("second run should be a no-op: %+v puts=%d", res, store.puts-putsBefore)
	}
}

func TestWritingSeedUpdatesInPlace(t *testing.T) {
	seeder, _ := newSeedHarness(t)
	ctx := context.Background()
	pack := func(prompt string) fstest.MapFS {
		return fstest.MapFS{"writing/demo/pack.json": {Data: []byte(`{"pack":"demo","tasks":[{
			"slug":"demo-test1-task2","source_ref":"","test_no":1,"task_no":2,"task_type":"opinion",
			"topic":"work","title":"Demo","prompt":"` + prompt + `","figures":[],"facts":{}}]}`)}}
	}
	if _, err := seeder.SeedTasks(ctx, pack("First wording."), "writing"); err != nil {
		t.Fatal(err)
	}
	var id1 string
	_ = testDB.QueryRow(`SELECT id FROM writing_tasks WHERE slug = 'demo-test1-task2'`).Scan(&id1)
	// An archived task stays archived when its pack entry changes.
	_, _ = testDB.Exec(`UPDATE writing_tasks SET status = 'archived' WHERE id = $1`, id1)

	res, err := seeder.SeedTasks(ctx, pack("Second wording."), "writing")
	if err != nil {
		t.Fatal(err)
	}
	var id2, prompt, status string
	_ = testDB.QueryRow(`SELECT id, prompt, status FROM writing_tasks WHERE slug = 'demo-test1-task2'`).Scan(&id2, &prompt, &status)
	if res.Updated != 1 || id1 != id2 || prompt != "Second wording." || status != "archived" {
		t.Fatalf("update in place: %+v id %s→%s prompt %q status %s", res, id1, id2, prompt, status)
	}
}

func TestWritingSeedRejectsBadPacks(t *testing.T) {
	seeder, _ := newSeedHarness(t)
	cases := map[string]string{
		"wrong type":      `{"pack":"bad","tasks":[{"slug":"bad-1","task_no":2,"task_type":"line_graph","title":"x","prompt":"x","figures":[]}]}`,
		"task 1 no image": `{"pack":"bad","tasks":[{"slug":"bad-1","task_no":1,"task_type":"table","title":"x","prompt":"x","figures":[]}]}`,
		"missing file":    `{"pack":"bad","tasks":[{"slug":"bad-1","task_no":1,"task_type":"table","title":"x","prompt":"x","figures":[{"file":"figures/none.svg","alt":"x"}]}]}`,
		"folder mismatch": `{"pack":"other","tasks":[]}`,
		"unknown field":   `{"pack":"bad","tasks":[],"extra":1}`,
	}
	for name, body := range cases {
		fsys := fstest.MapFS{"writing/bad/pack.json": {Data: []byte(body)}}
		if _, err := seeder.SeedTasks(context.Background(), fsys, "writing"); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	var n int
	_ = testDB.QueryRow(`SELECT count(*) FROM writing_tasks`).Scan(&n)
	if n != 0 {
		t.Fatalf("bad packs must not insert rows, got %d", n)
	}
}

func TestWritingFigureRedirect(t *testing.T) {
	store := &memStorage{objs: map[string]string{}}
	h := newHarness(t, fx.Decorate(func(storagein.IObjectStorage) storagein.IObjectStorage { return store }))
	sha := strings.Repeat("a1", 32)

	r := httptest.NewRecorder()
	h.engine.ServeHTTP(r, httptest.NewRequest("GET", "/api/v1/writing/figures/"+sha+".svg", nil))
	if r.Code != 302 || r.Header().Get("Location") != "https://bucket.example/writing/figures/"+sha+".svg" {
		t.Fatalf("redirect: %d %q", r.Code, r.Header().Get("Location"))
	}
	if r.Header().Get("Cache-Control") != "public, max-age=3000" {
		t.Fatalf("cache header: %q", r.Header().Get("Cache-Control"))
	}
	for _, bad := range []string{"../secret.svg", "abc.svg", sha + ".exe", sha} {
		if got := h.do("GET", "/api/v1/writing/figures/"+bad, nil, ""); got.Status != 404 {
			t.Errorf("%q: expected 404, got %d", bad, got.Status)
		}
	}
}

func TestWritingFigureServedWithoutS3(t *testing.T) {
	h := newHarness(t) // S3_BUCKET_NAME is empty in tests
	raw, err := seeds.Writing.ReadFile("writing/cam21/figures/test1-task1.svg")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	r := httptest.NewRecorder()
	h.engine.ServeHTTP(r, httptest.NewRequest("GET", "/api/v1/writing/figures/"+hex.EncodeToString(sum[:])+".svg", nil))
	if r.Code != 200 || r.Header().Get("Content-Type") != "image/svg+xml" || r.Body.Len() != len(raw) {
		t.Fatalf("embedded figure: %d %q %d bytes", r.Code, r.Header().Get("Content-Type"), r.Body.Len())
	}
}
