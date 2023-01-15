package main

import (
	"github.com/stephanebruckert/purrfect-api/cmd/app"
	"log"
	"os"
)

func main() {
	api, err := app.New()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if err := api.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
