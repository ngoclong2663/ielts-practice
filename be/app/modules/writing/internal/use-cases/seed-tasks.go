package usecases

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // register decoders for figure sizes
	_ "image/png"
	"io/fs"
	"path"
	"regexp"
	"strconv"
	"strings"

	"ielts-practice/app/modules/writing/domain/rules"

	"github.com/rs/zerolog"

	storagein "ielts-practice/app/modules/storage/ports/in"
	"ielts-practice/app/modules/writing/domain"
	"ielts-practice/app/modules/writing/domain/entity"
	writingin "ielts-practice/app/modules/writing/ports/in"
	"ielts-practice/configs"
)

var _ writingin.ITaskSeeder = &TaskSeeder{}

// TaskSeeder loads the seed packs (Epic 5 §6.4). It needs no AI model, so `api seed` can run it alone.
type TaskSeeder struct {
	repo    domain.IWritingRepository
	storage storagein.IObjectStorage
	cfg     *configs.Config
	logger  zerolog.Logger
}

func NewTaskSeeder(repo domain.IWritingRepository, storage storagein.IObjectStorage, cfg *configs.Config, logger zerolog.Logger) *TaskSeeder {
	return &TaskSeeder{repo: repo, storage: storage, cfg: cfg, logger: logger.With().Str("module", "writing").Logger()}
}

// FigureKeyPrefix is where seeded figures live in the bucket (content-addressed, never overwritten).
const FigureKeyPrefix = "writing/figures/"

type seedPack struct {
	Pack  string     `json:"pack"`
	Tasks []seedTask `json:"tasks"`
}

type seedFigure struct {
	File string `json:"file"`
	Alt  string `json:"alt"`
}

type seedTask struct {
	Slug      string          `json:"slug"`
	SourceRef string          `json:"source_ref"`
	TestNo    int             `json:"test_no"`
	TaskNo    int             `json:"task_no"`
	Module    string          `json:"module,omitempty"`
	TaskType  string          `json:"task_type"`
	Topic     string          `json:"topic"`
	Title     string          `json:"title"`
	Prompt    string          `json:"prompt"`
	Figures   []seedFigure    `json:"figures"`
	Facts     json.RawMessage `json:"facts"`
	Positions []string        `json:"positions,omitempty"`
}

type loadedFigure struct {
	figure entity.Figure
	data   []byte
}

var (
	slugPattern  = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	viewBoxAttr  = regexp.MustCompile(`viewBox="\s*[-\d.]+[\s,]+[-\d.]+[\s,]+([\d.]+)[\s,]+([\d.]+)\s*"`)
	figureTypes  = map[string]string{".svg": "image/svg+xml", ".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".webp": "image/webp"}
	errSeedInput = errors.New("invalid seed pack")
)

// SeedTasks loads every <root>/<pack>/pack.json. Unchanged tasks are skipped; changed ones are updated in place.
func (uc *TaskSeeder) SeedTasks(ctx context.Context, fsys fs.FS, root string) (writingin.SeedResult, error) {
	var res writingin.SeedResult
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return res, fmt.Errorf("read seed folder: %w", err)
	}
	seen := map[string]string{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := path.Join(root, e.Name())
		if _, err := fs.Stat(fsys, path.Join(dir, "pack.json")); err != nil {
			continue
		}
		if err := uc.seedPack(ctx, fsys, dir, e.Name(), seen, &res); err != nil {
			return res, err
		}
		res.Packs++
	}
	uc.logger.Info().Int("packs", res.Packs).Int("inserted", res.Inserted).Int("updated", res.Updated).
		Int("unchanged", res.Unchanged).Int("figures_uploaded", res.FiguresUploaded).Msg("Writing tasks seeded")
	return res, nil
}

func (uc *TaskSeeder) seedPack(ctx context.Context, fsys fs.FS, dir, name string, seen map[string]string, res *writingin.SeedResult) error {
	raw, err := fs.ReadFile(fsys, path.Join(dir, "pack.json"))
	if err != nil {
		return err
	}
	var pack seedPack
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&pack); err != nil {
		return fmt.Errorf("%w %s: %v", errSeedInput, name, err)
	}
	if pack.Pack != name {
		return fmt.Errorf("%w %s: \"pack\" is %q, expected the folder name", errSeedInput, name, pack.Pack)
	}
	for i := range pack.Tasks {
		st := &pack.Tasks[i]
		if other, dup := seen[st.Slug]; dup {
			return fmt.Errorf("%w %s: slug %q is also used in %s", errSeedInput, name, st.Slug, other)
		}
		seen[st.Slug] = name
		if err := uc.seedTask(ctx, fsys, dir, pack.Pack, st, res); err != nil {
			return fmt.Errorf("%s/%s: %w", name, st.Slug, err)
		}
	}
	return nil
}

func (uc *TaskSeeder) seedTask(ctx context.Context, fsys fs.FS, dir, pack string, st *seedTask, res *writingin.SeedResult) error {
	if err := validateSeedTask(st); err != nil {
		return err
	}
	figures := make([]loadedFigure, 0, len(st.Figures))
	for _, f := range st.Figures {
		lf, err := loadFigure(fsys, dir, f)
		if err != nil {
			return err
		}
		figures = append(figures, lf)
	}

	hash, err := contentHash(pack, st, figures)
	if err != nil {
		return err
	}
	existing, err := uc.repo.FindTaskBySlug(ctx, st.Slug)
	if err != nil {
		return err
	}
	if existing != nil && existing.ContentHash == hash {
		res.Unchanged++
		return nil
	}

	uploaded, err := uc.uploadFigures(ctx, figures)
	if err != nil {
		return err
	}
	res.FiguresUploaded += uploaded

	minWords, minutes := rules.WordsAndMinutes(st.TaskNo)
	module := st.Module
	if module == "" {
		module = entity.ModuleAcademic
	}
	facts := st.Facts
	if len(facts) == 0 || string(facts) == "null" {
		facts = json.RawMessage("{}")
	}
	task := &entity.Task{
		Slug: st.Slug, Pack: pack, SourceRef: st.SourceRef, TestNo: st.TestNo, TaskNo: st.TaskNo,
		Module: module, TaskType: st.TaskType, Topic: st.Topic, Title: strings.TrimSpace(st.Title),
		Prompt: strings.TrimSpace(st.Prompt), Facts: facts, MinWords: minWords, Minutes: minutes,
		Status: entity.StatusPublished, ContentHash: hash, Figures: make([]entity.Figure, 0, len(figures)),
		Positions: st.Positions,
	}
	for _, f := range figures {
		task.Figures = append(task.Figures, f.figure)
	}
	if err := uc.repo.UpsertTask(ctx, task); err != nil {
		return err
	}
	if existing == nil {
		res.Inserted++
	} else {
		res.Updated++
	}
	uc.logger.Info().Str("slug", st.Slug).Bool("new", existing == nil).Msg("Writing task seeded")
	return nil
}

func validateSeedTask(st *seedTask) error {
	switch {
	case !slugPattern.MatchString(st.Slug) || len(st.Slug) > 64:
		return fmt.Errorf("%w: slug must be lowercase words joined by '-' (max 64)", errSeedInput)
	case st.TaskNo != 1 && st.TaskNo != 2:
		return fmt.Errorf("%w: task_no must be 1 or 2", errSeedInput)
	case !rules.ValidTaskType(st.TaskNo, st.TaskType):
		return fmt.Errorf("%w: task_type %q is not valid for task %d", errSeedInput, st.TaskType, st.TaskNo)
	case st.Module != "" && st.Module != entity.ModuleAcademic && st.Module != "general":
		return fmt.Errorf("%w: module must be academic or general", errSeedInput)
	case strings.TrimSpace(st.Title) == "" || len([]rune(st.Title)) > 160:
		return fmt.Errorf("%w: title is required (max 160 characters)", errSeedInput)
	case strings.TrimSpace(st.Prompt) == "":
		return fmt.Errorf("%w: prompt is required", errSeedInput)
	case len(st.Topic) > 32 || len(st.SourceRef) > 64:
		return fmt.Errorf("%w: topic (max 32) or source_ref (max 64) too long", errSeedInput)
	case st.TaskNo == 1 && len(st.Figures) == 0:
		return fmt.Errorf("%w: Task 1 needs at least one figure", errSeedInput)
	}
	if len(st.Positions) > 0 && st.TaskNo != 2 {
		return fmt.Errorf("%w: positions are only for Task 2", errSeedInput)
	}
	for _, p := range st.Positions {
		if !rules.ValidPosition(p) {
			return fmt.Errorf("%w: unknown position %q", errSeedInput, p)
		}
	}
	if len(st.Facts) > 0 && !json.Valid(st.Facts) {
		return fmt.Errorf("%w: facts is not valid JSON", errSeedInput)
	}
	return nil
}

func loadFigure(fsys fs.FS, dir string, f seedFigure) (loadedFigure, error) {
	if !fs.ValidPath(f.File) || strings.TrimSpace(f.Alt) == "" {
		return loadedFigure{}, fmt.Errorf("%w: figure %q needs a relative file path and alt text", errSeedInput, f.File)
	}
	ext := strings.ToLower(path.Ext(f.File))
	contentType, ok := figureTypes[ext]
	if !ok {
		return loadedFigure{}, fmt.Errorf("%w: figure %q must be svg, png, jpg or webp", errSeedInput, f.File)
	}
	data, err := fs.ReadFile(fsys, path.Join(dir, f.File))
	if err != nil {
		return loadedFigure{}, fmt.Errorf("%w: %v", errSeedInput, err)
	}
	sum := sha256.Sum256(data)
	sha := hex.EncodeToString(sum[:])
	w, h := figureSize(ext, data)
	return loadedFigure{
		figure: entity.Figure{
			Key: FigureKeyPrefix + sha + ext, SHA256: sha, ContentType: contentType,
			Width: w, Height: h, Alt: strings.TrimSpace(f.Alt),
		},
		data: data,
	}, nil
}

func figureSize(ext string, data []byte) (int, int) {
	if ext == ".svg" {
		m := viewBoxAttr.FindSubmatch(data)
		if m == nil {
			return 0, 0
		}
		w, _ := strconv.ParseFloat(string(m[1]), 64)
		h, _ := strconv.ParseFloat(string(m[2]), 64)
		return int(w), int(h)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

// contentHash covers the pack entry and the figure bytes, so editing either re-seeds the task.
func contentHash(pack string, st *seedTask, figures []loadedFigure) (string, error) {
	shas := make([]string, len(figures))
	for i, f := range figures {
		shas[i] = f.figure.SHA256
	}
	b, err := json.Marshal(struct {
		Pack    string    `json:"pack"`
		Task    *seedTask `json:"task"`
		Figures []string  `json:"figures"`
	}{pack, st, shas})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// uploadFigures puts each figure in S3. Keys are content-addressed, so re-uploading is harmless.
func (uc *TaskSeeder) uploadFigures(ctx context.Context, figures []loadedFigure) (int, error) {
	if len(figures) == 0 {
		return 0, nil
	}
	if !uc.storage.Configured() {
		if uc.cfg.IsProduction() {
			return 0, fmt.Errorf("S3 is not configured: cannot upload writing figures")
		}
		uc.logger.Warn().Int("figures", len(figures)).Msg("S3 is not configured: writing figures not uploaded")
		return 0, nil
	}
	for _, f := range figures {
		if err := uc.storage.Put(ctx, f.figure.Key, bytes.NewReader(f.data), int64(len(f.data)), f.figure.ContentType); err != nil {
			return 0, fmt.Errorf("upload %s: %w", f.figure.Key, err)
		}
	}
	return len(figures), nil
}
