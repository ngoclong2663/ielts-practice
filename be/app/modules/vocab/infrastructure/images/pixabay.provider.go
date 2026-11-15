package images

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ielts-practice/app/modules/vocab/domain/entity"
	vocabout "ielts-practice/app/modules/vocab/ports/out"
	"ielts-practice/pkg/httpjson"
)

const (
	pixabayAPI = "https://pixabay.com/api/"
	// Pixabay's direct links (webformatURL, previewURL) stop working after about 24 h.
	pixabayLinkTTL = 23 * time.Hour
)

// PixabayProvider: https://pixabay.com/api/docs/ — links are hotlinked through our refresh endpoint,
// never copied to our storage.
type PixabayProvider struct {
	key   string
	cache *ttlCache[pixabaySearch]
}

var _ vocabout.IImageProvider = (*PixabayProvider)(nil)

type pixabaySearch struct {
	images    []entity.CardImage
	fetchedAt time.Time
}

func NewPixabayProvider(key string) *PixabayProvider {
	// Pixabay asks API users to cache search results for 24 h.
	return &PixabayProvider{key: key, cache: newTTLCache[pixabaySearch](24*time.Hour, 2000)}
}

func (p *PixabayProvider) Name() string     { return "pixabay" }
func (p *PixabayProvider) Configured() bool { return p.key != "" }
func (p *PixabayProvider) Expiring() bool   { return true }

type pixabayHit struct {
	ID              int64  `json:"id"`
	PageURL         string `json:"pageURL"`
	Tags            string `json:"tags"`
	PreviewURL      string `json:"previewURL"`
	WebformatURL    string `json:"webformatURL"`
	WebformatWidth  int    `json:"webformatWidth"`
	WebformatHeight int    `json:"webformatHeight"`
	User            string `json:"user"`
	UserID          int64  `json:"user_id"`
}

type pixabayResponse struct {
	TotalHits int          `json:"totalHits"`
	Hits      []pixabayHit `json:"hits"`
}

func (p *PixabayProvider) Search(ctx context.Context, query string, limit int) ([]entity.CardImage, error) {
	if !p.Configured() {
		return nil, errors.New("pixabay: missing API key")
	}
	query = strings.TrimSpace(query)
	cacheKey := strings.ToLower(query) + "|" + strconv.Itoa(limit)
	if hit, ok := p.cache.get(cacheKey); ok {
		return withExpiry(hit.images, hit.fetchedAt), nil
	}
	params := url.Values{
		"key":        {p.key},
		"q":          {httpjson.Truncate(query, 100)},
		"image_type": {"all"},
		"safesearch": {"true"},
		"lang":       {"en"},
		"per_page":   {strconv.Itoa(max(3, min(limit, 200)))},
	}
	var res pixabayResponse
	if err := httpjson.Get(ctx, pixabayAPI+"?"+params.Encode(), nil, &res); err != nil {
		return nil, fmt.Errorf("pixabay search: %w", err)
	}
	images := make([]entity.CardImage, 0, len(res.Hits))
	for _, h := range res.Hits {
		images = append(images, pixabayImage(h))
		if len(images) == limit {
			break
		}
	}
	now := time.Now()
	p.cache.set(cacheKey, pixabaySearch{images: images, fetchedAt: now})
	return withExpiry(images, now), nil
}

func (p *PixabayProvider) Lookup(ctx context.Context, id string) (*entity.CardImage, error) {
	if !p.Configured() {
		return nil, errors.New("pixabay: missing API key")
	}
	var res pixabayResponse
	err := httpjson.Get(ctx, pixabayAPI+"?"+url.Values{"key": {p.key}, "id": {id}}.Encode(), nil, &res)
	var se *httpjson.StatusError
	// Pixabay answers 400 "[ERROR 400] Image #123 does not exist." for removed photos.
	if errors.As(err, &se) && se.Code == 400 && strings.Contains(strings.ToLower(se.Body), "does not exist") {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pixabay lookup: %w", err)
	}
	if len(res.Hits) == 0 {
		return nil, nil
	}
	img := pixabayImage(res.Hits[0])
	img.LinkExpiresAt = time.Now().Add(pixabayLinkTTL)
	return &img, nil
}

func (p *PixabayProvider) TrackDownload(context.Context, entity.CardImage) error { return nil }

func pixabayImage(h pixabayHit) entity.CardImage {
	return entity.CardImage{
		Provider:   "pixabay",
		ProviderID: strconv.FormatInt(h.ID, 10),
		URL:        h.WebformatURL,
		ThumbURL:   h.PreviewURL,
		Width:      h.WebformatWidth,
		Height:     h.WebformatHeight,
		Alt:        h.Tags,
		AuthorName: h.User,
		AuthorURL:  fmt.Sprintf("https://pixabay.com/users/%s-%d/", url.PathEscape(h.User), h.UserID),
		SourceURL:  h.PageURL,
	}
}

func withExpiry(images []entity.CardImage, fetchedAt time.Time) []entity.CardImage {
	out := make([]entity.CardImage, len(images))
	for i, img := range images {
		img.LinkExpiresAt = fetchedAt.Add(pixabayLinkTTL)
		out[i] = img
	}
	return out
}
