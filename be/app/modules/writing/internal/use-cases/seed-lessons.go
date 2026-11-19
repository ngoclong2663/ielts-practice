package usecases

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	decks "ielts-practice/app/modules/decks/domain/entity"
	libraryin "ielts-practice/app/modules/library/ports/in"
	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/app/modules/writing/domain/rules"
	writingin "ielts-practice/app/modules/writing/ports/in"
)

// Seeded lessons and vocabulary (Epic 8 §4), reviewed by hand so learners never wait for the AI to open a task:
//
//	<pack>/lessons/<task>.json             Task 1 (one lesson)
//	<pack>/lessons/<task>.<position>.json  Task 2 (one per position)
//	<pack>/vocab/<task>.json               the task's useful vocabulary, each term with its card
//
// A lesson file records the task hash it was written for: when the task's text changes, the stale lesson is
// skipped and the AI writes a new one on first use, as before.

type seedLessonFile struct {
	Task     string               `json:"task"`
	Position string               `json:"position"`
	TaskHash string               `json:"task_hash"`
	Lesson   entity.LessonContent `json:"lesson"`
}

type seedVocabFile struct {
	Task   string           `json:"task"`
	Groups []seedVocabGroup `json:"groups"`
}

type seedVocabGroup struct {
	Key   string          `json:"key"`
	Items []seedVocabItem `json:"items"`
}

type seedVocabItem struct {
	entity.VocabItem
	// The term's card: written here, or a library word reused by its headword (card_ref).
	Card         *decks.CardContent `json:"card,omitempty"`
	CardRef      string             `json:"card_ref,omitempty"`
	ImageQuery   string             `json:"image_query,omitempty"`
	Visualizable bool               `json:"visualizable,omitempty"`
}

func strictJSON(raw []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// seedLessons loads the lessons and vocabulary of one pack folder; folders without them are fine.
func (uc *TaskSeeder) seedLessons(ctx context.Context, fsys fs.FS, dir string, res *writingin.SeedResult) error {
	files, _ := fs.Glob(fsys, path.Join(dir, "lessons", "*.json"))
	for _, file := range files {
		raw, err := fs.ReadFile(fsys, file)
		if err != nil {
			return err
		}
		var f seedLessonFile
		if err := strictJSON(raw, &f); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		task, err := uc.repo.FindTaskBySlug(ctx, f.Task)
		if err != nil {
			return err
		}
		if err := ValidateSeedLesson(task, &f, path.Base(file)); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		if f.TaskHash != task.ContentHash {
			uc.logger.Warn().Str("file", file).Msg("Seeded lesson skipped: the task changed since it was written")
			continue
		}
		changed, err := uc.repo.SeedLesson(ctx, &entity.Lesson{TaskID: task.ID, Position: f.Position, TaskHash: task.ContentHash,
			Content: f.Lesson, Model: entity.ModelSeed})
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		if changed {
			res.Lessons++
		}
	}

	files, _ = fs.Glob(fsys, path.Join(dir, "vocab", "*.json"))
	for _, file := range files {
		raw, err := fs.ReadFile(fsys, file)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		hash := hex.EncodeToString(sum[:])
		var f seedVocabFile
		if err := strictJSON(raw, &f); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		task, err := uc.repo.FindTaskBySlug(ctx, f.Task)
		if err != nil {
			return err
		}
		if task == nil || path.Base(file) != f.Task+".json" {
			return fmt.Errorf("%s: unknown task %q or file name not <task>.json", file, f.Task)
		}
		if old, err := uc.repo.TaskVocabHash(ctx, task.ID); err != nil || old == hash {
			if err != nil {
				return err
			}
			continue
		}
		vocab, err := uc.vocabWithCards(ctx, &f)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		if err := uc.repo.SeedTaskVocab(ctx, &entity.StoredVocab{TaskID: task.ID, Content: vocab, ContentHash: hash}); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		res.Vocab++
	}
	return nil
}

// ValidateSeedLesson checks a lesson file against its task: name, position, the practice flow's lesson checks
// and the exam's minimum length.
func ValidateSeedLesson(task *entity.Task, f *seedLessonFile, name string) error {
	if task == nil {
		return fmt.Errorf("unknown task %q", f.Task)
	}
	want := f.Task + ".json"
	if f.Position != "" {
		want = f.Task + "." + f.Position + ".json"
	}
	if name != want {
		return fmt.Errorf("file should be named %s", want)
	}
	if !slices.Contains(rules.Positions(task), f.Position) {
		return fmt.Errorf("position %q is not one of %v", f.Position, rules.Positions(task))
	}
	if err := rules.CleanLesson(&f.Lesson, rules.ParagraphKeys(task.TaskNo)); err != nil {
		return err
	}
	if n := rules.LessonWords(f.Lesson); n < task.MinWords {
		return fmt.Errorf("model answer has %d words, the task needs %d", n, task.MinWords)
	}
	if len(f.Lesson.Questions) == 0 {
		return errors.New("no plan questions")
	}
	return nil
}

// vocabWithCards checks the vocabulary and links every term to a ready-made card.
func (uc *TaskSeeder) vocabWithCards(ctx context.Context, f *seedVocabFile) (entity.LessonVocab, error) {
	var vocab entity.LessonVocab
	var cards []libraryin.SeedCard
	var slots []*entity.VocabItem
	for _, g := range f.Groups {
		group := entity.VocabGroup{Key: g.Key}
		for _, it := range g.Items {
			group.Items = append(group.Items, it.VocabItem)
		}
		vocab.Groups = append(vocab.Groups, group)
	}
	if err := rules.CleanVocab(&vocab); err != nil {
		return vocab, err
	}
	byTerm := map[string]seedVocabItem{}
	for _, g := range f.Groups {
		for _, it := range g.Items {
			byTerm[strings.ToLower(strings.TrimSpace(it.Term))] = it
		}
	}
	for gi := range vocab.Groups {
		for ii := range vocab.Groups[gi].Items {
			item := &vocab.Groups[gi].Items[ii]
			src := byTerm[strings.ToLower(item.Term)]
			switch {
			case src.CardRef != "":
				id, err := uc.cards.FindCard(ctx, src.CardRef)
				if err != nil {
					return vocab, err
				}
				if id == "" {
					return vocab, fmt.Errorf("term %q: card_ref %q is not a library card", item.Term, src.CardRef)
				}
				item.LexemeID = id
			case src.Card != nil:
				cards = append(cards, libraryin.SeedCard{ImageQuery: src.ImageQuery, Visualizable: src.Visualizable, Card: *src.Card})
				slots = append(slots, item)
			default:
				return vocab, fmt.Errorf("term %q has neither card nor card_ref", item.Term)
			}
		}
	}
	ids, err := uc.cards.StoreCards(ctx, cards)
	if err != nil {
		return vocab, err
	}
	for i, id := range ids {
		slots[i].LexemeID = id
	}
	return vocab, nil
}
