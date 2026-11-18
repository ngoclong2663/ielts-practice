package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"slices"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/guides/domain"
	"ielts-practice/app/modules/guides/domain/entity"
	"ielts-practice/app/modules/guides/domain/rules"
	guidesin "ielts-practice/app/modules/guides/ports/in"
)

var _ guidesin.IGuideSeeder = &GuideSeeder{}

// GuideSeeder loads seeds/guides/videos.json (Epic 7 §5.2). Idempotent: an unchanged entry is skipped, a
// changed one is updated in place, a new slug is inserted; guides are never deleted by the seed.
type GuideSeeder struct {
	repo   domain.IGuidesRepository
	logger zerolog.Logger
}

func NewGuideSeeder(repo domain.IGuidesRepository, logger zerolog.Logger) *GuideSeeder {
	return &GuideSeeder{repo: repo, logger: logger.With().Str("module", "guides").Logger()}
}

type seedFile struct {
	Videos []seedVideo `json:"videos"`
}

type seedVideo struct {
	Slug            string           `json:"slug"`
	Topic           string           `json:"topic"`
	Sort            int              `json:"sort"`
	Lang            string           `json:"lang"`
	Title           entity.Localized `json:"title"`
	Description     entity.Localized `json:"description"`
	DurationSeconds float64          `json:"duration_seconds"`
	Video           string           `json:"video"`  // <sha256>.mp4, uploaded with `api upload-media`
	Poster          string           `json:"poster"` // <sha256>.jpg, optional
}

var (
	seedSlug   = regexp.MustCompile(`^[a-z0-9-]{3,64}$`)
	seedVideoF = regexp.MustCompile(`^[0-9a-f]{64}\.mp4$`)
	seedPoster = regexp.MustCompile(`^[0-9a-f]{64}\.jpg$`)
)

func (v seedVideo) validate() error {
	switch {
	case !seedSlug.MatchString(v.Slug):
		return fmt.Errorf("guide %q: bad slug", v.Slug)
	case !slices.Contains(entity.Topics, v.Topic):
		return fmt.Errorf("guide %s: unknown topic %q", v.Slug, v.Topic)
	case v.Title.VI == "" || v.Title.EN == "":
		return fmt.Errorf("guide %s: title needs vi and en", v.Slug)
	case !seedVideoF.MatchString(v.Video):
		return fmt.Errorf("guide %s: video must be <sha256>.mp4", v.Slug)
	case v.Poster != "" && !seedPoster.MatchString(v.Poster):
		return fmt.Errorf("guide %s: poster must be <sha256>.jpg", v.Slug)
	case v.DurationSeconds <= 0:
		return fmt.Errorf("guide %s: duration_seconds is required", v.Slug)
	}
	return nil
}

func (s *GuideSeeder) SeedGuides(ctx context.Context, fsys fs.FS, file string) (guidesin.SeedResult, error) {
	var res guidesin.SeedResult
	raw, err := fs.ReadFile(fsys, file)
	if err != nil {
		return res, err
	}
	var f seedFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return res, fmt.Errorf("%s: %w", file, err)
	}
	seen := map[string]bool{}
	for _, v := range f.Videos {
		if err := v.validate(); err != nil {
			return res, err
		}
		if seen[v.Slug] {
			return res, fmt.Errorf("guide %s: duplicate slug", v.Slug)
		}
		seen[v.Slug] = true
		canonical, _ := json.Marshal(v)
		sum := sha256.Sum256(canonical)
		hash := hex.EncodeToString(sum[:])
		existing, err := s.repo.FindBySlug(ctx, v.Slug)
		if err != nil {
			return res, err
		}
		if existing != nil && existing.ContentHash == hash {
			res.Unchanged++
			continue
		}
		lang := v.Lang
		if lang == "" {
			lang = "vi"
		}
		g := &entity.GuideVideo{
			Slug: v.Slug, Topic: v.Topic, Sort: v.Sort, Lang: lang, Title: v.Title, Description: v.Description,
			VideoKey: rules.KeyForFile(v.Video), DurationSeconds: v.DurationSeconds, Status: entity.StatusPublished, ContentHash: hash,
		}
		if v.Poster != "" {
			g.PosterKey = rules.KeyForFile(v.Poster)
		}
		if err := s.repo.Upsert(ctx, g); err != nil {
			return res, fmt.Errorf("guide %s: %w", v.Slug, err)
		}
		if existing == nil {
			res.Inserted++
		} else {
			res.Updated++
		}
	}
	s.logger.Info().Int("inserted", res.Inserted).Int("updated", res.Updated).Int("unchanged", res.Unchanged).Msg("Guide videos seeded")
	return res, nil
}
