package gamemode

import (
	"time"

	gameaudio "github.com/kivutar/goro/audio"
	"github.com/kivutar/goro/core"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/world"
)

type Context struct {
	Config      core.Config
	Input       *input.State
	Resources   *res.Manager
	Session     *session.Session
	World       *world.World
	Network     *network.Client
	Audio       *gameaudio.BGM
	Started     time.Time
	ScreenW     int
	ScreenH     int
	RequestQuit func()
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
