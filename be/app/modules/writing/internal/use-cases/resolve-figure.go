package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"ielts-practice/pkg/exceptions"
	"ielts-practice/seeds"
)

// figureTTL is how long a presigned figure link works; browsers may cache the redirect a little less.
const figureTTL = time.Hour

// FigureCacheMaxAge is shorter than the presigned link's lifetime, so a cached redirect never points at an expired link.
const FigureCacheMaxAge = 50 * time.Minute

var figureFile = regexp.MustCompile(`^[0-9a-f]{64}\.(svg|png|jpg|jpeg|webp)$`)

// Figure is either a redirect to S3 or, without S3 outside production, the embedded seed file itself.
type Figure struct {
	RedirectURL string
	Data        []byte
	ContentType string
}

// ResolveFigure returns a short-lived S3 link for a seeded figure (Epic 5 §6.4). The bucket is private,
// so the database stores only the object key and the API hands out presigned links on demand.
func (uc *WritingUseCases) ResolveFigure(ctx context.Context, file string) (Figure, error) {
	notFound := exceptions.NewException(http.StatusNotFound, exceptions.CodeImageNotFound, "Figure not found", nil)
	if !figureFile.MatchString(file) {
		return Figure{}, notFound
	}
	if !uc.storage.Configured() {
		// Local development and e2e run without S3: serve the copy compiled into the binary.
		if f, ok := embeddedFigures()[file]; ok && !uc.cfg.IsProduction() {
			return f, nil
		}
		return Figure{}, notFound
	}
	url, err := uc.storage.PresignGet(ctx, FigureKeyPrefix+file, figureTTL)
	if err != nil {
		return Figure{}, exceptions.ErrInternalServerError("Failed to sign figure link", err)
	}
	return Figure{RedirectURL: url}, nil
}

// embeddedFigures indexes the seed figures by their object file name (<sha256>.<ext>).
var embeddedFigures = sync.OnceValue(func() map[string]Figure {
	out := map[string]Figure{}
	_ = fs.WalkDir(seeds.Writing, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := strings.ToLower(path.Ext(p))
		contentType, ok := figureTypes[ext]
		if !ok {
			return nil
		}
		data, err := fs.ReadFile(seeds.Writing, p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		out[hex.EncodeToString(sum[:])+ext] = Figure{Data: data, ContentType: contentType}
		return nil
	})
	return out
})
