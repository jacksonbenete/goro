package app

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kivutar/goro/internal/core"
	"github.com/kivutar/goro/internal/gamemode"
	"github.com/kivutar/goro/internal/input"
	"github.com/kivutar/goro/internal/network"
	"github.com/kivutar/goro/internal/res"
	"github.com/kivutar/goro/internal/session"
	"github.com/kivutar/goro/internal/world"
)

type Game struct {
	cfg      core.Config
	input    *input.State
	resource *res.Manager
	session  *session.Session
	world    *world.World
	network  *network.Client
	modes    *gamemode.Manager
	started  time.Time
}

func New(cfg core.Config) (*Game, error) {
	resource, err := res.NewManager(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("resource manager: %w", err)
	}

	g := &Game{
		cfg:      cfg,
		input:    input.NewState(),
		resource: resource,
		session:  session.New(),
		world:    world.New(),
		network:  network.NewClient(cfg.Packet.ClientDate),
		started:  time.Now(),
	}

	ctx := g.modeContext()
	g.modes = gamemode.NewManager(ctx, gamemode.NewBootMode())
	return g, nil
}

func (g *Game) Update() error {
	g.input.Update()
	g.network.Pump()
	g.modes.UpdateContext(g.modeContext())
	return g.modes.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.modes.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth <= 0 || outsideHeight <= 0 {
		return g.cfg.Window.Width, g.cfg.Window.Height
	}
	return outsideWidth, outsideHeight
}

func (g *Game) modeContext() gamemode.Context {
	return gamemode.Context{
		Config:    g.cfg,
		Input:     g.input,
		Resources: g.resource,
		Session:   g.session,
		World:     g.world,
		Network:   g.network,
		Started:   g.started,
	}
}
