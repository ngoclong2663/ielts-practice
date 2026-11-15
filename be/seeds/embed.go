// Package seeds holds reference data compiled into the API binary and applied by `api seed`.
package seeds

import "embed"

// Writing contains one folder per pack: pack.json plus the figures it references (Epic 5 §6).
//
//go:embed writing
var Writing embed.FS
