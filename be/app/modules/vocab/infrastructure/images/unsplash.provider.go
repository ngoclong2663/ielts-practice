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
	unsplashAPI = "https://api.unsplash.com"
	// Unsplash guidelines: attribution links carry these UTM parameters.
	unsplashUTM = "utm_source=ielts_practice&utm_medium=referral"
)

// UnsplashProvider: https://unsplash.com/documentation — photos must be hotlinked (never re-hosted),
// credited to the photographer, and the download endpoint called when a photo is chosen.
type UnsplashProvider struct {
	key   string
	cache *ttlCache[[]entity.CardImage]
}

var _ vocabout.IImageProvider = (*UnsplashProvider)(nil)

func NewUnsplashProvider(key string) *UnsplashProvider {
	// Unsplash demo apps get 50 requests/hour, so repeated searches are cached.
	return &UnsplashProvider{key: key, cache: newTTLCache[[]entity.CardImage](6*time.Hour, 2000)}
}

func (p *UnsplashProvider) Name() string     { return "unsplash" }
func (p *UnsplashProvider) Configured() bool { return p.key != "" }
func (p *UnsplashProvider) Expiring() bool   { return false }

func (p *UnsplashProvider) headers() map[string]string {
	return map[string]string{"Authorization": "Client-ID " + p.key, "Accept-Version": "v1"}
}

type unsplashPhoto struct {
	ID             string `json:"id"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	AltDescription string `json:"alt_description"`
	Description    string `json:"description"`
	URLs           struct {
		Regular string `json:"regular"`
		Small   string `json:"small"`
	} `json:"urls"`
	Links struct {
		HTML             string `json:"html"`
		DownloadLocation string `json:"download_location"`
	} `json:"links"`
	User struct {
		Name  string `json:"name"`
		Links struct {
			HTML string `json:"html"`
		} `json:"links"`
	} `json:"user"`
}

func (p *UnsplashProvider) Search(ctx context.Context, query string, limit int) ([]entity.CardImage, error) {
	if !p.Configured() {
		return nil, errors.New("unsplash: missing access key")
	}
	query = strings.TrimSpace(query)
	cacheKey := strings.ToLower(query) + "|" + strconv.Itoa(limit)
	if hit, ok := p.cache.get(cacheKey); ok {
		return hit, nil
	}
	params := url.Values{
		"query":          {httpjson.Truncate(query, 100)},
		"per_page":       {strconv.Itoa(max(1, min(limit, 30)))},
		"content_filter": {"high"},
	}
	var res struct {
		Results []unsplashPhoto `json:"results"`
	}
	if err := httpjson.Get(ctx, unsplashAPI+"/search/photos?"+params.Encode(), p.headers(), &res); err != nil {
		return nil, fmt.Errorf("unsplash search: %w", err)
	}
	images := make([]entity.CardImage, 0, len(res.Results))
	for _, ph := range res.Results {
		images = append(images, unsplashImage(ph))
	}
	p.cache.set(cacheKey, images)
	return images, nil
}

func (p *UnsplashProvider) Lookup(ctx context.Context, id string) (*entity.CardImage, error) {
	var ph unsplashPhoto
	err := httpjson.Get(ctx, unsplashAPI+"/photos/"+url.PathEscape(id), p.headers(), &ph)
	var se *httpjson.StatusError
	if errors.As(err, &se) && se.Code == 404 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unsplash lookup: %w", err)
	}
	img := unsplashImage(ph)
	return &img, nil
}

func (p *UnsplashProvider) TrackDownload(ctx context.Context, img entity.CardImage) error {
	if img.DownloadLocation == "" || !p.Configured() {
		return nil
	}
	if !strings.HasPrefix(img.DownloadLocation, unsplashAPI+"/") {
		return fmt.Errorf("unsplash: unexpected download location")
	}
	return httpjson.Get(ctx, img.DownloadLocation, p.headers(), nil)
}

func withUTM(link string) string {
	if link == "" {
		return ""
	}
	if strings.Contains(link, "?") {
		return link + "&" + unsplashUTM
	}
	return link + "?" + unsplashUTM
}

func unsplashImage(ph unsplashPhoto) entity.CardImage {
	alt := ph.AltDescription
	if alt == "" {
		alt = ph.Description
	}
	width, height := ph.Width, ph.Height
	if width > 0 { // "regular" is 1080 px wide
		height = height * 1080 / width
		width = 1080
	}
	return entity.CardImage{
		Provider:         "unsplash",
		ProviderID:       ph.ID,
		URL:              ph.URLs.Regular,
		ThumbURL:         ph.URLs.Small,
		Width:            width,
		Height:           height,
		Alt:              alt,
		AuthorName:       ph.User.Name,
		AuthorURL:        withUTM(ph.User.Links.HTML),
		SourceURL:        withUTM(ph.Links.HTML),
		DownloadLocation: ph.Links.DownloadLocation,
	}
}
