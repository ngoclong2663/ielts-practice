package usecases

import (
	"os"
	"path"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"ielts-practice/seeds"
)

// TestCommittedExamSeeds runs the seed checks (Epic 10 §8.2) on every committed book, so a broken key,
// a quote missing from its passage or a numbering gap fails `make test` before it can reach `api seed`.
func TestCommittedExamSeeds(t *testing.T) {
	books, err := LoadExamSeeds(seeds.Exam, "exam")
	if err != nil {
		t.Fatal(err)
	}
	if len(books) == 0 {
		t.Fatal("no exam books committed")
	}
	slugs := map[string]bool{}
	for _, b := range books {
		for _, p := range b.Parts {
			if slugs[p.Part.Slug] {
				t.Errorf("duplicate part slug %s", p.Part.Slug)
			}
			slugs[p.Part.Slug] = true
			if len(p.Part.QuestionTypes) == 0 {
				t.Errorf("%s: no question types", p.Part.Slug)
			}
			for _, g := range p.Part.Content.Groups {
				if strings.Contains(strings.ToLower(g.Instructions), "cambridge") {
					t.Errorf("%s: instructions must not name the reference book", p.Part.Slug)
				}
			}
		}
	}
}

func TestLoadExamSeedsRefusesBrokenFiles(t *testing.T) {
	book := `{"slug":"zz","title":{"vi":"a","en":"b"}}`
	part := func(q string) string {
		return `{"title":"T","q_from":1,"q_to":1,"paragraphs":[{"id":"A","text":"Some text."}],
			"groups":[{"type":"tfng","from":1,"to":1,"instructions":"x","items":[{"n":1,"text":"s"}]}],
			"key":{"1":{"answer":["TRUE"],"evidence":{"block":"A","quote":"` + q + `"},"why":"w"}}}`
	}
	cases := map[string]fstest.MapFS{
		"quote not in passage": {"exam/zz/book.json": {Data: []byte(book)}, "exam/zz/test1/reading-1.json": {Data: []byte(part("Other text."))}},
		"incomplete test":      {"exam/zz/book.json": {Data: []byte(book)}, "exam/zz/test1/reading-1.json": {Data: []byte(part("Some text."))}},
		"unknown field": {"exam/zz/book.json": {Data: []byte(book)},
			"exam/zz/test1/reading-1.json": {Data: []byte(strings.Replace(part("Some text."), `"title"`, `"tittle":"x","title"`, 1))}},
		"stray file":               {"exam/zz/book.json": {Data: []byte(book)}, "exam/zz/test1/notes.txt": {Data: []byte("x")}},
		"slug differs from folder": {"exam/zz/book.json": {Data: []byte(`{"slug":"yy","title":{"vi":"a","en":"b"}}`)}},
	}
	for name, fsys := range cases {
		if _, err := LoadExamSeeds(fsys, "exam"); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

var seedFileRe = regexp.MustCompile(`^exam/([a-z0-9]+)/test([0-9]+)/(reading|listening)-([1-4])\.json$`)

// TestExamSeedFile checks one part file on its own, while the rest of its test is still being written:
//
//	EXAM_PART=exam/s1/test2/listening-1.json go test ./app/modules/exam/internal/use-cases/ -run TestExamSeedFile -v
//
// It reads the file from disk (not the embedded copy), so no rebuild is needed between edits.
func TestExamSeedFile(t *testing.T) {
	file := os.Getenv("EXAM_PART")
	if file == "" {
		t.Skip("set EXAM_PART=exam/<set>/test<N>/<skill>-<P>.json")
	}
	m := seedFileRe.FindStringSubmatch(file)
	if m == nil {
		t.Fatalf("EXAM_PART must look like exam/s1/test2/listening-1.json, got %q", file)
	}
	lp, err := loadPart(os.DirFS("../../../../../seeds"), path.Join("exam", m[1]), file, m[3], m[4], m[1]+"-t"+m[2])
	if err != nil {
		t.Fatal(err)
	}
	p := lp.Part
	t.Logf("%s ok: questions %d–%d, types %v", p.Slug, p.QFrom, p.QTo, p.QuestionTypes)
	if p.Audio != nil {
		t.Logf("audio %.0fs, %d transcript lines, %d reading pauses", p.Audio.Duration, len(p.Content.Lines), len(p.Audio.Cues))
	}
}
