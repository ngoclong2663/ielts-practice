// Package in is what the exam module offers: loading the test bank seed files (used by `api seed`).
package in

import (
	"context"
	"io/fs"
)

type SeedResult struct {
	Inserted, Updated, Unchanged int
}

// IExamSeeder loads seeds/exam/<book>/ (Epic 10 §8). Audio must already be in S3.
type IExamSeeder interface {
	SeedExam(ctx context.Context, fsys fs.FS, root string) (SeedResult, error)
}
