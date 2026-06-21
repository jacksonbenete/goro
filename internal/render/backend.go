package render

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kivutar/goro/internal/core"
)

const BackendName = "ebitengine-opengl"

type Game interface {
	Update() error
	Draw(*ebiten.Image)
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}

func Run(game Game, cfg core.WindowConfig) error {
	ebiten.SetWindowTitle(cfg.Title)
	ebiten.SetWindowSize(cfg.Width, cfg.Height)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	return ebiten.RunGameWithOptions(game, &ebiten.RunGameOptions{
		GraphicsLibrary: ebiten.GraphicsLibraryOpenGL,
	})
}
