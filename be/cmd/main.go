package main

import (
	// Embed the IANA timezone database so timezone validation works in minimal containers.
	_ "time/tzdata"

	"ielts-practice/app"
)

func main() {
	application := app.NewApp()
	application.Run()
}
