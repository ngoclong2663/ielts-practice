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
	application := app.NewApp()
	application.Run()
}
