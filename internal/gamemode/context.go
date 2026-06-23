package gamemode

import (
	"time"

	gameaudio "github.com/kivutar/goro/internal/audio"
	"github.com/kivutar/goro/internal/core"
	"github.com/kivutar/goro/internal/input"
	"github.com/kivutar/goro/internal/network"
	"github.com/kivutar/goro/internal/res"
	"github.com/kivutar/goro/internal/session"
	"github.com/kivutar/goro/internal/world"
)

type Context struct {
	Config    core.Config
	Input     *input.State
	Resources *res.Manager
	Session   *session.Session
	World     *world.World
	Network   *network.Client
	Audio     *gameaudio.BGM
	Started   time.Time
	ScreenW   int
	ScreenH   int
}

func (c Context) ScreenSize() (int, int) {
	width, height := c.ScreenW, c.ScreenH
	if width <= 0 {
		width = c.Config.Window.Width
	}
	if height <= 0 {
		height = c.Config.Window.Height
	}
	return width, height
}
