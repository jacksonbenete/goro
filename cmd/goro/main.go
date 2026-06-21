package main

import (
	"log"

	"github.com/kivutar/goro/internal/app"
	"github.com/kivutar/goro/internal/core"
	"github.com/kivutar/goro/internal/render"
)

func main() {
	cfg := core.LoadConfig()
	game, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := render.Run(game, cfg.Window); err != nil {
		log.Fatal(err)
	}
}
