package main

import (
	"fmt"
	"os"

	// Embed the IANA timezone database so timezone validation works in minimal containers.
	_ "time/tzdata"

	"ielts-practice/app"
)

func main() {
	// `api seed` applies the embedded reference data (writing task packs) and exits.
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		if err := app.RunSeed(); err != nil {
			fmt.Fprintln(os.Stderr, "seed failed:", err)
			os.Exit(1)
		}
		return
	}
	// `api upload-media <file>...` puts guide videos and posters in S3 (Epic 7 §5.2) and exits.
	if len(os.Args) > 1 && os.Args[1] == "upload-media" {
		if err := app.RunUploadMedia(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "upload failed:", err)
			os.Exit(1)
		}
		return
	}
	// `api library-images <topic.json>...` picks photos for library seed files (Epic 8 §3.4) and exits.
	if len(os.Args) > 1 && os.Args[1] == "library-images" {
		if err := app.RunLibraryImages(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "library-images failed:", err)
			os.Exit(1)
		}
		return
	}
	// `api exam-media <file.mp3>...` uploads listening audio built by scripts/exam-audio (Epic 10 §8.3) and exits.
	if len(os.Args) > 1 && os.Args[1] == "exam-media" {
		if err := app.RunExamMedia(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "exam-media failed:", err)
			os.Exit(1)
		}
		return
	}
	application := app.NewApp()
	application.Run()
}
