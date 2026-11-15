package images

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"ielts-practice/app/modules/vocab/domain/entity"
	vocabout "ielts-practice/app/modules/vocab/ports/out"
)

type FakeImageProvider struct{}

var _ vocabout.IImageProvider = (*FakeImageProvider)(nil)

func NewFakeImageProvider() *FakeImageProvider { return &FakeImageProvider{} }

func (FakeImageProvider) Name() string     { return "fake" }
func (FakeImageProvider) Configured() bool { return true }
func (FakeImageProvider) Expiring() bool   { return false }

var fakeColors = []string{"#10b981", "#0ea5e9", "#f59e0b", "#8b5cf6", "#ef4444", "#14b8a6"}

func (FakeImageProvider) Search(_ context.Context, query string, limit int) ([]entity.CardImage, error) {
	out := make([]entity.CardImage, 0, limit)
	for i := 0; i < limit && i < len(fakeColors); i++ {
		svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="640" height="400"><rect width="100%%" height="100%%" fill="%s"/><text x="50%%" y="50%%" font-size="40" text-anchor="middle" fill="white" font-family="sans-serif">%s</text></svg>`,
			fakeColors[i], escapeXML(query))
		uri := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
		out = append(out, entity.CardImage{
			Provider: "fake", ProviderID: fmt.Sprintf("%s-%d", strings.ReplaceAll(query, " ", "-"), i),
			URL: uri, ThumbURL: uri, Width: 640, Height: 400, Alt: query, AuthorName: "Test images",
		})
	}
	return out, nil
}

func (FakeImageProvider) Lookup(context.Context, string) (*entity.CardImage, error) { return nil, nil }
func (FakeImageProvider) TrackDownload(context.Context, entity.CardImage) error     { return nil }

func escapeXML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
