package main

import (
	"fmt"
	"github.com/stephanebruckert/purrfect-api/cmd/app"
	"os"
)

func main() {
	api, err := app.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := api.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
