package usecases

import (
	"encoding/json"
	"io/fs"
	"path"
	"strings"
	"testing"

	librules "ielts-practice/app/modules/library/domain/rules"
	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/app/modules/writing/domain/rules"
	"ielts-practice/seeds"
)

// seedTasks reads the committed packs as `api seed` would, with each task's content hash.
func seedTasks(t *testing.T) (map[string]*entity.Task, map[string]string) {
	t.Helper()
	tasks, packOf := map[string]*entity.Task{}, map[string]string{}
	dirs, _ := fs.ReadDir(seeds.Writing, "writing")
	for _, d := range dirs {
		dir := path.Join("writing", d.Name())
		raw, err := fs.ReadFile(seeds.Writing, path.Join(dir, "pack.json"))
		if err != nil {
			continue
		}
		var pack seedPack
		if err := json.Unmarshal(raw, &pack); err != nil {
			t.Fatal(err)
		}
		for i := range pack.Tasks {
			st := &pack.Tasks[i]
			figures := make([]loadedFigure, 0, len(st.Figures))
			for _, f := range st.Figures {
				lf, err := loadFigure(seeds.Writing, dir, f)
				if err != nil {
					t.Fatal(err)
				}
				figures = append(figures, lf)
			}
			hash, err := contentHash(pack.Pack, st, figures)
			if err != nil {
				t.Fatal(err)
			}
			minWords, _ := rules.WordsAndMinutes(st.TaskNo)
			tasks[st.Slug] = &entity.Task{Slug: st.Slug, TaskNo: st.TaskNo, TaskType: st.TaskType, Positions: st.Positions,
				MinWords: minWords, ContentHash: hash}
			packOf[st.Slug] = d.Name()
		}
	}
	return tasks, packOf
}

// libraryHeadwords lists the library words a vocabulary item may reuse with card_ref.
func libraryHeadwords(t *testing.T) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	files, _ := fs.Glob(seeds.Library, "library/*/*.json")
	for _, f := range files {
		raw, _ := fs.ReadFile(seeds.Library, f)
		var tf struct {
			Words []struct {
				Card struct {
					Headword string `json:"headword"`
				} `json:"card"`
			} `json:"words"`
		}
		_ = json.Unmarshal(raw, &tf)
		for _, w := range tf.Words {
			out[strings.ToLower(w.Card.Headword)] = true
		}
	}
	return out
}

// Every committed lesson and vocabulary file must load (Epic 8 §4). Run it after editing a seed file:
//
//	go test ./app/modules/writing/internal/use-cases/ -run TestCommittedWritingSeeds -v
func TestCommittedWritingSeeds(t *testing.T) {
	tasks, packOf := seedTasks(t)
	lessons, _ := fs.Glob(seeds.Writing, "writing/*/lessons/*.json")
	for _, file := range lessons {
		raw, _ := fs.ReadFile(seeds.Writing, file)
		var f seedLessonFile
		if err := strictJSON(raw, &f); err != nil {
			t.Errorf("%s: %v", file, err)
			continue
		}
		task := tasks[f.Task]
		if err := ValidateSeedLesson(task, &f, path.Base(file)); err != nil {
			t.Errorf("%s: %v", file, err)
			continue
		}
		if packOf[f.Task] != strings.Split(file, "/")[1] {
			t.Errorf("%s: task %s is in pack %s", file, f.Task, packOf[f.Task])
		}
		if f.TaskHash != task.ContentHash {
			t.Errorf("%s: task_hash must be %q", file, task.ContentHash)
		}
		t.Logf("%s: %d words", file, rules.LessonWords(f.Lesson))
	}

	library := libraryHeadwords(t)
	// A card_ref may also point at a card written in an earlier vocabulary file (packs and files load in order),
	// so a phrase such as "account for" has one card for every task.
	vocab, _ := fs.Glob(seeds.Writing, "writing/*/vocab/*.json")
	earlier := map[string]bool{}
	for _, file := range vocab {
		raw, _ := fs.ReadFile(seeds.Writing, file)
		var f seedVocabFile
		if err := strictJSON(raw, &f); err != nil {
			t.Errorf("%s: %v", file, err)
			continue
		}
		if tasks[f.Task] == nil || path.Base(file) != f.Task+".json" {
			t.Errorf("%s: unknown task %q or file not named <task>.json", file, f.Task)
			continue
		}
		var v entity.LessonVocab
		items := 0
		for _, g := range f.Groups {
			group := entity.VocabGroup{Key: g.Key}
			for _, it := range g.Items {
				items++
				group.Items = append(group.Items, it.VocabItem)
				switch {
				case it.CardRef != "" && it.Card != nil:
					t.Errorf("%s: %q has both card and card_ref", file, it.Term)
				case it.CardRef != "":
					if !library[strings.ToLower(it.CardRef)] && !earlier[strings.ToLower(it.CardRef)] {
						t.Errorf("%s: card_ref %q is neither a library word nor a card of an earlier vocabulary file", file, it.CardRef)
					}
				case it.Card != nil:
					if !strings.EqualFold(strings.TrimSpace(it.Card.Headword), strings.TrimSpace(it.Term)) {
						t.Errorf("%s: card headword %q differs from term %q", file, it.Card.Headword, it.Term)
					}
					if earlier[strings.ToLower(it.Card.Headword)] && it.Card.SenseHint == "" {
						t.Errorf("%s: %q already has a card in an earlier vocabulary file: use \"card_ref\"", file, it.Term)
					}
					if library[strings.ToLower(it.Card.Headword)] && it.Card.SenseHint == "" {
						t.Errorf("%s: %q is a library word: use \"card_ref\", or give the card a sense_hint when its sense differs", file, it.Term)
					}
					if _, _, _, err := librules.CheckCard(*it.Card, it.ImageQuery, it.Visualizable); err != nil {
						t.Errorf("%s: %q: %v", file, it.Term, err)
					}
				default:
					t.Errorf("%s: %q has neither card nor card_ref", file, it.Term)
				}
			}
			v.Groups = append(v.Groups, group)
		}
		if err := rules.CleanVocab(&v); err != nil {
			t.Errorf("%s: %v", file, err)
			continue
		}
		kept := 0
		for _, g := range v.Groups {
			kept += len(g.Items)
		}
		if kept != items {
			t.Errorf("%s: %d items, %d kept (duplicate term, empty field or more than 10 in a group)", file, items, kept)
		}
		for _, g := range f.Groups {
			for _, it := range g.Items {
				if it.Card != nil && it.Card.SenseHint == "" {
					earlier[strings.ToLower(strings.TrimSpace(it.Card.Headword))] = true
				}
			}
		}
		t.Logf("%s: %d items", file, items)
	}
}
