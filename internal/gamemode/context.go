package gamemode

import (
	"time"

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
	Started   time.Time
}
