package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/exam/domain"
	"ielts-practice/app/modules/exam/domain/entity"
	"ielts-practice/app/modules/exam/domain/rules"
	examin "ielts-practice/app/modules/exam/ports/in"
)

var _ examin.IExamSeeder = &ExamSeeder{}

// ExamSeeder loads the test bank (Epic 10 §8):
//
//	<root>/<book>/book.json                       {slug, source_ref, title {vi, en}, sort}
//	<root>/<book>/test<N>/reading-<P>.json        one file per part (SeedPart)
//	<root>/<book>/test<N>/listening-<P>.json
//
// Every file is checked before anything is written, so one broken file stops the whole seed. Idempotent: an
// unchanged part is skipped, a changed part is updated in place (attempts keep pointing at it). Nothing is
// deleted.
type ExamSeeder struct {
	repo   domain.IExamRepository
	tx     *database.TransactionStruct
	logger zerolog.Logger
}

func NewExamSeeder(repo domain.IExamRepository, tx *database.TransactionStruct, logger zerolog.Logger) *ExamSeeder {
	return &ExamSeeder{repo: repo, tx: tx, logger: logger.With().Str("module", "exam").Logger()}
}

type seedBook struct {
	Slug      string           `json:"slug"`
	SourceRef string           `json:"source_ref"`
	Title     entity.Localized `json:"title"`
	Sort      int              `json:"sort"`
	Status    string           `json:"status"`
}

// SeedPart is the file format of one part.
type SeedPart struct {
	SourceRef  string             `json:"source_ref"`
	Title      string             `json:"title"`
	Subtitle   string             `json:"subtitle"`
	Topic      string             `json:"topic"`
	QFrom      int                `json:"q_from"`
	QTo        int                `json:"q_to"`
	Paragraphs []entity.Paragraph `json:"paragraphs"`
	Lines      []entity.Line      `json:"lines"`
	Audio      *entity.Audio      `json:"audio"`
	Groups     []entity.Group     `json:"groups"`
	Key        entity.Key         `json:"key"`
	Status     string             `json:"status"`
	// Voices maps each transcript speaker to a TTS voice ("kokoro:bf_emma", "melo:EN-AU"). It is read by the
	// audio build (scripts/exam-audio) only and never stored.
	Voices map[string]string `json:"voices"`
}

// LoadedPart is a parsed and checked part, ready to store.
type LoadedPart struct {
	BookSlug string
	TestNo   int
	TestSlug string
	Part     entity.Part
	Key      entity.Key
}

// LoadedBook is a parsed book with its parts.
type LoadedBook struct {
	Book  entity.Book
	Parts []LoadedPart
}

var (
	bookSlugRe = regexp.MustCompile(`^[a-z0-9]{1,16}$`)
	testDirRe  = regexp.MustCompile(`^test([1-9][0-9]?)$`)
	partFileRe = regexp.MustCompile(`^(reading|listening)-([1-4])\.json$`)
	audioKeyRe = regexp.MustCompile(`^exam/media/[0-9a-f]{64}\.(mp3|m4a)$`)
)

// LoadExamSeeds parses and checks every book under root without touching the database. It is shared by
// `api seed` and TestCommittedExamSeeds.
func LoadExamSeeds(fsys fs.FS, root string) ([]LoadedBook, error) {
	dirs, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, err
	}
	var out []LoadedBook
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		lb, err := loadBook(fsys, path.Join(root, d.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, *lb)
	}
	return out, nil
}

func loadBook(fsys fs.FS, dir string) (*LoadedBook, error) {
	raw, err := fs.ReadFile(fsys, path.Join(dir, "book.json"))
	if err != nil {
		return nil, err
	}
	var sb seedBook
	if err := json.Unmarshal(raw, &sb); err != nil {
		return nil, fmt.Errorf("%s/book.json: %w", dir, err)
	}
	if !bookSlugRe.MatchString(sb.Slug) || sb.Slug != path.Base(dir) {
		return nil, fmt.Errorf("%s/book.json: slug must match the folder name", dir)
	}
	if sb.Title.VI == "" || sb.Title.EN == "" {
		return nil, fmt.Errorf("%s/book.json: title needs vi and en", dir)
	}
	status := sb.Status
	if status == "" {
		status = entity.StatusPublished
	}
	lb := &LoadedBook{Book: entity.Book{Slug: sb.Slug, SourceRef: sb.SourceRef, Title: sb.Title, Sort: sb.Sort, Status: status}}

	tests, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	for _, td := range tests {
		m := testDirRe.FindStringSubmatch(td.Name())
		if !td.IsDir() || m == nil {
			continue
		}
		testNo, _ := strconv.Atoi(m[1])
		testSlug := fmt.Sprintf("%s-t%d", sb.Slug, testNo)
		files, err := fs.ReadDir(fsys, path.Join(dir, td.Name()))
		if err != nil {
			return nil, err
		}
		bySkill := map[string][]*entity.Part{}
		for _, f := range files {
			fm := partFileRe.FindStringSubmatch(f.Name())
			if fm == nil {
				return nil, fmt.Errorf("%s/%s/%s: unexpected file (want reading-N.json or listening-N.json)", dir, td.Name(), f.Name())
			}
			lp, err := loadPart(fsys, dir, path.Join(dir, td.Name(), f.Name()), fm[1], fm[2], testSlug)
			if err != nil {
				return nil, err
			}
			lp.BookSlug, lp.TestNo, lp.TestSlug = sb.Slug, testNo, testSlug
			lb.Parts = append(lb.Parts, *lp)
		}
		for i := range lb.Parts {
			if lb.Parts[i].TestSlug == testSlug {
				bySkill[lb.Parts[i].Part.Skill] = append(bySkill[lb.Parts[i].Part.Skill], &lb.Parts[i].Part)
			}
		}
		for skill, parts := range bySkill {
			if err := rules.ValidateTest(testSlug, skill, parts); err != nil {
				return nil, err
			}
		}
	}
	return lb, nil
}

func loadPart(fsys fs.FS, bookDir, file, skill, partNo, testSlug string) (*LoadedPart, error) {
	raw, err := fs.ReadFile(fsys, file)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	var sp SeedPart
	if err := dec.Decode(&sp); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	no, _ := strconv.Atoi(partNo)
	status := sp.Status
	if status == "" {
		status = entity.StatusPublished
	}
	p := entity.Part{
		Skill: skill, PartNo: no, Slug: fmt.Sprintf("%s-%c%d", testSlug, skill[0], no), SourceRef: sp.SourceRef,
		Title: sp.Title, Topic: sp.Topic, QFrom: sp.QFrom, QTo: sp.QTo, Audio: sp.Audio, Status: status,
		Content: entity.Content{Subtitle: sp.Subtitle, Paragraphs: sp.Paragraphs, Lines: sp.Lines, Groups: sp.Groups},
	}
	if p.Audio != nil && !audioKeyRe.MatchString(p.Audio.Key) {
		return nil, fmt.Errorf("%s: audio key must be exam/media/<sha256>.mp3 (upload it with `api exam-media`)", file)
	}
	if err := rules.ValidatePart(&p, sp.Key); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if p.Skill == entity.SkillListening {
		for _, l := range p.Content.Lines {
			if l.Speaker != "" && sp.Voices[l.Speaker] == "" {
				return nil, fmt.Errorf("%s: speaker %q has no voice in \"voices\"", file, l.Speaker)
			}
		}
	}
	// Figures are named by their content, so an edited drawing is a new file and changes the part's hash.
	h := sha256.New()
	h.Write(raw)
	for i := range p.Content.Groups {
		g := &p.Content.Groups[i]
		if g.Figure == "" {
			continue
		}
		name, err := figureName(fsys, bookDir, g.Figure)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		g.Figure = name
		h.Write([]byte(name))
	}
	p.QuestionTypes = rules.QuestionTypes(p.Content)
	p.ContentHash = hex.EncodeToString(h.Sum(nil))
	return &LoadedPart{Part: p, Key: sp.Key}, nil
}

func (s *ExamSeeder) SeedExam(ctx context.Context, fsys fs.FS, root string) (examin.SeedResult, error) {
	var res examin.SeedResult
	if _, err := fs.Stat(fsys, root); err != nil {
		return res, nil // no exam seeds yet
	}
	books, err := LoadExamSeeds(fsys, root)
	if err != nil {
		return res, err
	}
	err = s.tx.Execute(ctx, func(ctx context.Context) error {
		for _, lb := range books {
			book := lb.Book
			if err := s.repo.UpsertBook(ctx, &book); err != nil {
				return fmt.Errorf("book %s: %w", book.Slug, err)
			}
			testIDs := map[string]string{}
			for _, lp := range lb.Parts {
				id, ok := testIDs[lp.TestSlug]
				if !ok {
					t := &entity.Test{BookID: book.ID, TestNo: lp.TestNo, Slug: lp.TestSlug}
					if err := s.repo.UpsertTest(ctx, t); err != nil {
						return fmt.Errorf("test %s: %w", lp.TestSlug, err)
					}
					id, testIDs[lp.TestSlug] = t.ID, t.ID
				}
				part := lp.Part
				part.TestID = id
				existing, err := s.repo.FindPartBySlug(ctx, part.Slug)
				if err != nil {
					return err
				}
				if existing != nil && existing.ContentHash == part.ContentHash {
					res.Unchanged++
					continue
				}
				if err := s.repo.UpsertPart(ctx, &part, lp.Key); err != nil {
					return fmt.Errorf("part %s: %w", part.Slug, err)
				}
				if existing == nil {
					res.Inserted++
				} else {
					res.Updated++
				}
			}
		}
		return nil
	})
	if err != nil {
		return examin.SeedResult{}, err
	}
	s.logger.Info().Int("inserted", res.Inserted).Int("updated", res.Updated).Int("unchanged", res.Unchanged).Msg("Exam parts seeded")
	return res, nil
}

// maxFigureBytes keeps drawings small; they are compiled into the binary.
const maxFigureBytes = 300 << 10

var figurePathRe = regexp.MustCompile(`^figures/[a-z0-9-]+\.svg$`)

// figureName checks a figure file of a set and returns its content name (<sha256>.svg).
func figureName(fsys fs.FS, bookDir, rel string) (string, error) {
	if !figurePathRe.MatchString(rel) {
		return "", fmt.Errorf("figure %q must be figures/<name>.svg", rel)
	}
	b, err := fs.ReadFile(fsys, path.Join(bookDir, rel))
	if err != nil {
		return "", fmt.Errorf("figure %s: %w", rel, err)
	}
	if len(b) > maxFigureBytes || !strings.Contains(string(b[:min(len(b), 512)]), "<svg") {
		return "", fmt.Errorf("figure %s must be an SVG under 300 KB", rel)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]) + ".svg", nil
}
