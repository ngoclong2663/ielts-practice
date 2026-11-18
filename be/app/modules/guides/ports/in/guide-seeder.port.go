// Package in is what the guides module offers: loading the guide video seed file (used by `api seed`).
package in

import (
	"context"
	"io/fs"
)

type SeedResult struct {
	Inserted, Updated, Unchanged int
}

// IGuideSeeder loads seeds/guides/videos.json. The media must already be in S3 (`api upload-media`).
type IGuideSeeder interface {
	SeedGuides(ctx context.Context, fsys fs.FS, file string) (SeedResult, error)
}
