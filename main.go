package main

import (
	"flag"
	"log"

	"github.com/kivutar/goro/app"
	"github.com/kivutar/goro/core"
	"github.com/kivutar/goro/render"
)

func main() {
	cfg := core.LoadConfig()
	flag.BoolVar(&cfg.Window.Fullscreen, "fullscreen", cfg.Window.Fullscreen, "start in fullscreen mode")
	flag.Parse()

	game, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := render.Run(game, cfg.Window); err != nil {
		log.Fatal(err)
	}
}
