package app

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	vocabentity "ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/infrastructure/images"
	vocabout "ielts-practice/app/modules/vocab/ports/out"
	"ielts-practice/configs"
)

// RunLibraryImages picks a photo for every picturable word of library seed files that has none
// (`api library-images [--force] [--provider unsplash|pixabay] <topic.json>...`, Epic 8 §3.4). It only searches
// the photo APIs (no AI), takes the first result for the word's image_query, and writes the file back for review.
// Unsplash comes first (stable links); when it fails (for example its hourly limit) the rest use Pixabay, whose
// links are served through /api/v1/images so they can be refreshed.
func RunLibraryImages(args []string) error {
	fs := flag.NewFlagSet("library-images", flag.ContinueOnError)
	force := fs.Bool("force", false, "replace images that are already chosen")
	only := fs.String("provider", "", "use only this provider (unsplash or pixabay)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("usage: api library-images [--force] [--provider unsplash|pixabay] <topic.json>...")
	}
	cfg, err := configs.LoadConfig()
	if err != nil {
		return err
	}
	cfg.FakeProviders = false
	cfg.Vocab.ImageProviders = []string{"unsplash", "pixabay"}
	var providers []vocabout.IImageProvider
	for _, p := range images.NewImageProviders(cfg) {
		if p.Configured() && (*only == "" || p.Name() == *only) {
			providers = append(providers, p)
		}
	}
	if len(providers) == 0 {
		return fmt.Errorf("no image provider configured (UNSPLASH_ACCESS_KEY, PIXABAY_API_KEY)")
	}
	for _, file := range fs.Args() {
		if err := fillImages(file, providers, *force); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
	}
	return nil
}

type imageSeedFile struct {
	Topic string          `json:"topic"`
	Words []imageSeedWord `json:"words"`
}

// Seeded writing vocabulary (Epic 8 §4): the same photo search for each item that has its own card.
type vocabSeedFile struct {
	Task   string           `json:"task"`
	Groups []vocabSeedGroup `json:"groups"`
}

type vocabSeedGroup struct {
	Key   string          `json:"key"`
	Items []vocabSeedItem `json:"items"`
}

type vocabSeedItem struct {
	Term         string                   `json:"term"`
	POS          string                   `json:"pos"`
	VI           string                   `json:"vi"`
	Synonyms     []string                 `json:"synonyms"`
	ExampleEN    string                   `json:"example_en"`
	NoteVI       string                   `json:"note_vi"`
	CardRef      string                   `json:"card_ref,omitempty"`
	ImageQuery   string                   `json:"image_query,omitempty"`
	Visualizable bool                     `json:"visualizable,omitempty"`
	Card         *vocabentity.CardContent `json:"card,omitempty"`
}

type imageSeedWord struct {
	Source       string                  `json:"source"`
	ImageQuery   string                  `json:"image_query"`
	Visualizable bool                    `json:"visualizable"`
	Card         vocabentity.CardContent `json:"card"`
}

func fillImages(file string, providers []vocabout.IImageProvider, force bool) error {
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if bytes.Contains(raw, []byte(`"groups"`)) {
		return fillVocabImages(file, raw, providers, force)
	}
	var f imageSeedFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	added := 0
	for i := range f.Words {
		w := &f.Words[i]
		if !w.Visualizable || (w.Card.Image != nil && !force) {
			continue
		}
		query := strings.TrimSpace(w.ImageQuery)
		if query == "" {
			query = w.Card.Headword
		}
		img, err := searchFirst(ctx, &providers, query, w.Card.Headword)
		if err != nil {
			return err
		}
		if img == nil {
			fmt.Printf("  %-22s no matching photo for %q\n", w.Card.Headword, query)
			w.Card.Image = nil
			continue
		}
		w.Card.Image = img
		added++
		fmt.Printf("  %-22s %-8s %s\n", w.Card.Headword, img.Provider, img.Alt)
		// Save after every photo, so an interrupted run keeps what it found.
		if err := writeSeedFile(file, &f); err != nil {
			return err
		}
	}
	fmt.Printf("%s: %d photos added\n", file, added)
	return writeSeedFile(file, &f)
}

// searchFirst returns the photo that best matches the query among the first results of the first provider that
// has a match; a provider that fails is dropped for the rest of the run. A photo whose description shares no word
// with the query is not used: no photo is better than a wrong one on a flashcard.
func searchFirst(ctx context.Context, providers *[]vocabout.IImageProvider, query, headword string) (*vocabentity.CardImage, error) {
	if len(*providers) == 0 {
		return nil, fmt.Errorf("every image provider failed")
	}
	for i := 0; i < len(*providers); i++ {
		p := (*providers)[i]
		found, err := p.Search(ctx, query, 10)
		if err != nil {
			fmt.Printf("  (%s failed: %v; dropping it)\n", p.Name(), err)
			*providers = append((*providers)[:i:i], (*providers)[i+1:]...)
			i--
			continue
		}
		best, bestScore := -1, 0
		for k, img := range found {
			if sc := matchScore(img.Alt, query, headword); sc > bestScore {
				best, bestScore = k, sc
			}
		}
		if best < 0 {
			continue
		}
		img := found[best]
		img.LinkExpiresAt = time.Time{}
		if p.Expiring() {
			// Pixabay: the card points at our refresh endpoint; the seed registers the link.
			img.URL = "/api/v1/images/" + img.Provider + "/" + img.ProviderID
			img.ThumbURL = img.URL + "?size=thumb"
		} else if err := p.TrackDownload(ctx, img); err != nil { // Unsplash guideline: the photo is used
			fmt.Printf("  (track download failed: %v)\n", err)
		}
		return &img, nil
	}
	if len(*providers) == 0 {
		return nil, fmt.Errorf("every image provider failed")
	}
	return nil, nil
}

// matchScore counts the query's content words (and the headword, twice) found in a photo's description. A photo
// matching fewer than half the query's words (and not the headword) scores 0, as does an AI-generated picture.
func matchScore(alt, query, headword string) int {
	lower := strings.ToLower(alt)
	if strings.Contains(lower, "ai generated") || strings.Contains(lower, "ai-generated") {
		return 0
	}
	desc := " " + strings.Join(strings.FieldsFunc(lower, notLetter), " ") + " "
	words, hits := 0, 0
	for _, w := range strings.FieldsFunc(strings.ToLower(query), notLetter) {
		if len(w) < 3 || stopWords[w] {
			continue
		}
		words++
		if strings.Contains(desc, " "+stem(w)) {
			hits++
		}
	}
	head := 0
	for _, w := range strings.FieldsFunc(strings.ToLower(headword), notLetter) {
		if len(w) > 3 && strings.Contains(desc, " "+stem(w)) {
			head += 2
		}
	}
	if head == 0 && hits*2 < words {
		return 0
	}
	return hits + head
}

var stopWords = map[string]bool{"the": true, "and": true, "with": true, "for": true, "from": true, "into": true, "near": true, "during": true}

func notLetter(r rune) bool { return r < 'a' || r > 'z' }

// stem drops a plural or -ing ending so "panels" matches "panel" and "planting" matches "plant".
func stem(w string) string {
	for _, suf := range []string{"ing", "es", "s"} {
		if strings.HasSuffix(w, suf) && len(w)-len(suf) >= 4 {
			return strings.TrimSuffix(w, suf)
		}
	}
	return w
}

func fillVocabImages(file string, raw []byte, providers []vocabout.IImageProvider, force bool) error {
	var f vocabSeedFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	added := 0
	for gi := range f.Groups {
		for ii := range f.Groups[gi].Items {
			it := &f.Groups[gi].Items[ii]
			if it.Card == nil || !it.Visualizable || (it.Card.Image != nil && !force) {
				continue
			}
			query := strings.TrimSpace(it.ImageQuery)
			if query == "" {
				query = it.Card.Headword
			}
			img, err := searchFirst(ctx, &providers, query, it.Card.Headword)
			if err != nil {
				return err
			}
			it.Card.Image = img
			if img == nil {
				fmt.Printf("  %-22s no matching photo for %q\n", it.Term, query)
				continue
			}
			added++
			fmt.Printf("  %-22s %-8s %s\n", it.Term, img.Provider, img.Alt)
		}
	}
	fmt.Printf("%s: %d photos added\n", file, added)
	return writeSeedFile(file, &f)
}

func writeSeedFile(file string, f any) error {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(f); err != nil {
		return err
	}
	return os.WriteFile(file, b.Bytes(), 0o644)
}
