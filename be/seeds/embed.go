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

// Library holds the vocabulary library (Epic 8 §3): library/<collection>/collection.json and one
// <topic>.json per topic, each card reviewed by hand. Built with `api library-images` for the photos.
//
//go:embed library
var Library embed.FS

// Exam holds the Reading and Listening test bank (Epic 10 §8): exam/<book>/book.json and one file per part in
// exam/<book>/test<N>/. Audio is uploaded to S3 with `api exam-media` and referenced by its sha256.
//
//go:embed exam
var Exam embed.FS
