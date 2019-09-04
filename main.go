package main

import (
	"log"

	"github.com/wybiral/timeline/pkg/app"
)

func main() {
	a, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	a.Run()
}
