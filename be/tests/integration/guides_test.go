package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"go.uber.org/fx"

	guidesin "ielts-practice/app/modules/guides/ports/in"
	storagein "ielts-practice/app/modules/storage/ports/in"
	"ielts-practice/seeds"
)

func TestGuideVideos(t *testing.T) {
	store := &memStorage{objs: map[string]string{}}
	var seeder guidesin.IGuideSeeder
	h := newHarness(t,
		fx.Decorate(func(storagein.IObjectStorage) storagein.IObjectStorage { return store }),
		fx.Populate(&seeder),
	)
	if _, err := testDB.Exec("TRUNCATE guide_videos"); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// The committed seed file: both videos, idempotent.
	res, err := seeder.SeedGuides(ctx, seeds.Guides, "guides/videos.json")
	if err != nil || res.Inserted != 2 {
		t.Fatalf("first seed: %+v %v", res, err)
	}
	if res, _ = seeder.SeedGuides(ctx, seeds.Guides, "guides/videos.json"); res.Unchanged != 2 {
		t.Fatalf("second seed: %+v", res)
	}
	bad := fstest.MapFS{"v.json": {Data: []byte(`{"videos":[{"slug":"x-y","topic":"math","title":{"vi":"a","en":"b"},"video":"a.mp4","duration_seconds":1}]}`)}}
	if _, err := seeder.SeedGuides(ctx, bad, "v.json"); err == nil {
		t.Fatal("a guide with an unknown topic must be refused")
	}

	access, _ := h.signIn()
	if r := h.do("GET", "/api/v1/guides", nil, ""); r.Status != 401 {
		t.Fatalf("list without sign-in: %d", r.Status)
	}
	list := h.doList("/api/v1/guides?topic=writing", access)
	if len(list) != 1 || list[0]["slug"] != "writing-levels" {
		t.Fatalf("writing guides: %v", list)
	}
	g := list[0]
	video, poster := g["video_url"].(string), g["poster_url"].(string)
	if !strings.HasPrefix(video, "/api/v1/guides/media/") || !strings.HasSuffix(video, ".mp4") || !strings.HasSuffix(poster, ".jpg") ||
		g["title"].(map[string]any)["en"] == "" || g["lang"] != "vi" || g["duration_seconds"].(float64) <= 0 {
		t.Fatalf("guide: %v", g)
	}
	if all := h.doList("/api/v1/guides", access); len(all) != 2 {
		t.Fatalf("all guides: %d", len(all))
	}
	if r := h.do("GET", "/api/v1/guides?topic=math", nil, access); r.Status != 400 {
		t.Fatalf("unknown topic: %d", r.Status)
	}

	// Media: a presigned redirect for a seeded file only (public, no session needed).
	get := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.engine.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		return w
	}
	w := get(video)
	if w.Code != http.StatusFound || !strings.HasPrefix(w.Header().Get("Location"), "https://bucket.example/guides/videos/") {
		t.Fatalf("video redirect: %d %s", w.Code, w.Header().Get("Location"))
	}
	if w = get(poster); w.Code != http.StatusFound || !strings.Contains(w.Header().Get("Location"), "/guides/posters/") {
		t.Fatalf("poster redirect: %d", w.Code)
	}
	for _, path := range []string{"/api/v1/guides/media/" + strings.Repeat("a", 64) + ".mp4", "/api/v1/guides/media/../x.mp4", "/api/v1/guides/media/abc.mp4"} {
		if w = get(path); w.Code != http.StatusNotFound {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
}
