package ports

import (
	"context"
	"io/fs"
)

// SeedResult counts what one seed run did.
type SeedResult struct {
	Packs           int
	Inserted        int
	Updated         int
	Unchanged       int
	FiguresUploaded int
}

// ITaskSeeder loads the writing seed packs (Epic 5 §6.4). Used by `api seed`.
type ITaskSeeder interface {
	// SeedTasks reads every <pack>/pack.json under root in fsys.
	SeedTasks(ctx context.Context, fsys fs.FS, root string) (SeedResult, error)
}
