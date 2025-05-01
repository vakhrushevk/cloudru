package main

import (
	"context"
	"log"

	"github.com/vakhrushevk/cloudru/internal/app"
)

func main() {
	app, err := app.NewApp(context.Background(), "configs/config.yaml")

	if err != nil {
		log.Fatal("error creating app:", err)
	}

	log.Fatal(app.Start())
}
