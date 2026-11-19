package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/library/domain"
	"ielts-practice/app/modules/library/domain/entity"
	"ielts-practice/app/modules/library/domain/rules"
	libraryin "ielts-practice/app/modules/library/ports/in"
	vocabrules "ielts-practice/app/modules/vocab/domain/rules"
)

var (
	_ libraryin.ILibrarySeeder = &LibrarySeeder{}
	_ libraryin.ICardStore     = &LibrarySeeder{}
)

// LibrarySeeder loads seeds/library/<collection>/ (Epic 8 §3.4): collection.json, then one <topic>.json per
// topic it lists. A topic listed without a file yet is skipped, so a collection can grow topic by topic.
// Idempotent: an unchanged topic file is skipped; a changed one replaces the topic's words.
type LibrarySeeder struct {
	repo   domain.ILibraryRepository
	logger zerolog.Logger
}

func NewLibrarySeeder(repo domain.ILibraryRepository, logger zerolog.Logger) *LibrarySeeder {
	return &LibrarySeeder{repo: repo, logger: logger.With().Str("module", "library").Logger()}
}

// seedVersion is part of every topic hash: bump it when the loader changes how a file is stored.
const seedVersion = "1"

type collectionFile struct {
	Slug        string           `json:"slug"`
	Kind        string           `json:"kind"`
	Sort        int              `json:"sort"`
	Status      string           `json:"status"`
	Title       entity.Localized `json:"title"`
	Description entity.Localized `json:"description"`
	Topics      []string         `json:"topics"`
}

type topicFile struct {
	Topic string     `json:"topic"`
	Words []seedWord `json:"words"`
}

type seedWord struct {
	Source       string             `json:"source"`
	ImageQuery   string             `json:"image_query"`
	Visualizable bool               `json:"visualizable"`
	Card         entity.CardContent `json:"card"`
}

var (
	slugRe  = regexp.MustCompile(`^[a-z0-9-]{3,64}$`)
	topicRe = regexp.MustCompile(`^[a-z0-9_]{2,32}$`)
)

func (s *LibrarySeeder) SeedLibrary(ctx context.Context, fsys fs.FS, root string) (libraryin.SeedResult, error) {
	var res libraryin.SeedResult
	dirs, err := fs.ReadDir(fsys, root)
	if err != nil {
		return res, err
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		if err := s.seedCollection(ctx, fsys, path.Join(root, d.Name()), &res); err != nil {
			return res, err
		}
	}
	s.logger.Info().Int("topics", res.Topics).Int("unchanged", res.Unchanged).Int("words", res.Words).Msg("Library seeded")
	return res, nil
}

func (s *LibrarySeeder) seedCollection(ctx context.Context, fsys fs.FS, dir string, res *libraryin.SeedResult) error {
	raw, err := fs.ReadFile(fsys, path.Join(dir, "collection.json"))
	if err != nil {
		return err
	}
	var cf collectionFile
	if err := json.Unmarshal(raw, &cf); err != nil {
		return fmt.Errorf("%s/collection.json: %w", dir, err)
	}
	switch {
	case !slugRe.MatchString(cf.Slug):
		return fmt.Errorf("%s: bad slug %q", dir, cf.Slug)
	case !slices.Contains(entity.Kinds, cf.Kind):
		return fmt.Errorf("%s: unknown kind %q", dir, cf.Kind)
	case cf.Title.VI == "" || cf.Title.EN == "":
		return fmt.Errorf("%s: title needs vi and en", dir)
	}
	status := cf.Status
	if status == "" {
		status = entity.StatusPublished
	}
	sum := sha256.Sum256(raw)
	col := &entity.Collection{Slug: cf.Slug, Kind: cf.Kind, Sort: cf.Sort, Status: status, Title: cf.Title,
		Description: cf.Description, ContentHash: hex.EncodeToString(sum[:])}
	if err := s.repo.UpsertCollection(ctx, col); err != nil {
		return fmt.Errorf("collection %s: %w", cf.Slug, err)
	}

	for i, topic := range cf.Topics {
		if !topicRe.MatchString(topic) {
			return fmt.Errorf("collection %s: bad topic %q", cf.Slug, topic)
		}
		file := path.Join(dir, topic+".json")
		raw, err := fs.ReadFile(fsys, file)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		sum := sha256.Sum256(append([]byte(seedVersion+"|"), raw...))
		hash := hex.EncodeToString(sum[:])
		old, err := s.repo.TopicHash(ctx, col.ID, topic)
		if err != nil {
			return err
		}
		if old == hash {
			res.Unchanged++
			continue
		}
		words, err := ParseTopicFile(raw, topic)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		if err := s.repo.ReplaceTopic(ctx, col.ID, topic, i+1, hash, words); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		res.Topics++
		res.Words += len(words)
	}
	return nil
}

// ParseTopicFile validates a topic seed file: every card must be complete and pass the same checks as an
// AI card (no links, markup or instruction-like text; the card is about its headword).
func ParseTopicFile(raw []byte, topic string) ([]entity.SeedWord, error) {
	var tf topicFile
	if err := json.Unmarshal(raw, &tf); err != nil {
		return nil, err
	}
	if tf.Topic != topic {
		return nil, fmt.Errorf("topic is %q, want %q", tf.Topic, topic)
	}
	if len(tf.Words) == 0 {
		return nil, errors.New("no words")
	}
	seen := map[string]bool{}
	out := make([]entity.SeedWord, 0, len(tf.Words))
	for i, w := range tf.Words {
		if w.Source == "" {
			return nil, fmt.Errorf("word %d (%s): source is required", i+1, w.Card.Headword)
		}
		sw, err := toSeedWord(w, i+1)
		if err != nil {
			return nil, fmt.Errorf("word %d (%s): %w", i+1, w.Card.Headword, err)
		}
		if seen[sw.Key] {
			return nil, fmt.Errorf("word %d: %q appears twice", i+1, w.Card.Headword)
		}
		seen[sw.Key] = true
		out = append(out, sw)
	}
	return out, nil
}

func toSeedWord(w seedWord, sort int) (entity.SeedWord, error) {
	if w.Source != "" && !slices.Contains(entity.Sources, w.Source) {
		return entity.SeedWord{}, fmt.Errorf("source %q", w.Source)
	}
	c, imageQuery, pixabay, err := rules.CheckCard(w.Card, w.ImageQuery, w.Visualizable)
	if err != nil {
		return entity.SeedWord{}, err
	}
	return entity.SeedWord{
		Key:          vocabrules.LexemeKey(c.Headword, c.SenseHint), // a sense hint gives a second sense its own card
		Card:         c,
		CEFR:         c.CEFR,
		Source:       w.Source,
		ImageQuery:   imageQuery,
		SearchText:   rules.SearchText(&c),
		Sort:         sort,
		PixabayImage: pixabay,
	}, nil
}

// StoreCards stores reviewed cards that are not library words (seeded writing vocabulary): same checks as a
// library card. A headword that is already a library word keeps the library's card.
func (s *LibrarySeeder) StoreCards(ctx context.Context, in []libraryin.SeedCard) ([]string, error) {
	ids := make([]string, 0, len(in))
	for i, c := range in {
		sw, err := toSeedWord(seedWord{ImageQuery: c.ImageQuery, Visualizable: c.Visualizable, Card: c.Card}, i+1)
		if err != nil {
			return nil, fmt.Errorf("card %d (%s): %w", i+1, c.Card.Headword, err)
		}
		if id, err := s.repo.FindLibraryLexeme(ctx, sw.Key, true); err != nil || id != "" {
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
			continue
		}
		id, err := s.repo.UpsertLibraryLexeme(ctx, sw)
		if err != nil {
			return nil, fmt.Errorf("card %s: %w", c.Card.Headword, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *LibrarySeeder) FindCard(ctx context.Context, headword string) (string, error) {
	return s.repo.FindLibraryLexeme(ctx, vocabrules.LexemeKey(strings.TrimSpace(headword), ""), false)
}
