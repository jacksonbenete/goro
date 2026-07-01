package main

import (
	"log"
	"os"

	"github.com/kivutar/goro/app"
	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/render"
)

func main() {
	cfg, err := config.LoadConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	game, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := render.Run(game, cfg.Window, cfg.Render); err != nil {
		log.Fatal(err)
	}
}
