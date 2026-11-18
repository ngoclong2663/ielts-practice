// Package seeds holds reference data compiled into the API binary and applied by `api seed`.
package seeds

import "embed"

// Writing contains one folder per pack: pack.json plus the figures it references (Epic 5 §6).
//
//go:embed writing
var Writing embed.FS

// Guides lists the guide videos (Epic 7 §5): videos.json. The MP4s and posters are not in the repository;
// they are uploaded to S3 with `api upload-media` and referenced here by their sha256.
//
//go:embed guides
var Guides embed.FS
