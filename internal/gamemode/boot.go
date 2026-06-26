package gamemode

import (
	"strings"
	"time"

	"github.com/kivutar/goro/internal/render"
)

type BootMode struct {
	entered time.Time
}

func NewBootMode() *BootMode {
	return &BootMode{}
}

func (m *BootMode) Name() string {
	return "boot"
}

func (m *BootMode) Enter(Context) {
	m.entered = time.Now()
}

func (m *BootMode) Update(ctx Context) (Mode, error) {
	if time.Since(m.entered) > 700*time.Millisecond || ctx.Input.JustPressed(render.KeyEnter) {
		return NewLoginMode(), nil
	}
	return nil, nil
}

func (m *BootMode) Draw(ctx Context, screen *render.Image) {
	clear(screen)
	drawPanel(screen, 32, 32, 520, 178)
	debugText(screen, 52, 52, "goro")
	debugText(screen, 52, 76, "renderer: %s", render.BackendName)
	debugText(screen, 52, 96, "runtime root: %s", ctx.Resources.Root)
	debugText(screen, 52, 116, "clientinfo servers: %d", len(ctx.Resources.ClientInfo.Connections))
	debugText(screen, 52, 136, "grf archives: %d", len(ctx.Resources.Archives))
	if len(ctx.Resources.FoundFiles) > 0 {
		debugText(screen, 52, 156, "found: %s", strings.Join(shortPaths(ctx.Resources.FoundFiles), ", "))
	} else {
		debugText(screen, 52, 156, "found: loose clientinfo fallback only")
	}
	debugText(screen, 52, 188, "packet profile: %d / %d", ctx.Config.Packet.ClientDate, ctx.Config.Packet.Profile)
}

func shortPaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if len(path) > 42 {
			path = "..." + path[len(path)-39:]
		}
		out = append(out, path)
	}
	return out
}
