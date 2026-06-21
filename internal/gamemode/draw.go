package gamemode

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kivutar/goro/internal/render"
)

func clear(screen *ebiten.Image) {
	screen.Fill(render.ColorBackground)
}

func drawPanel(screen *ebiten.Image, x, y, w, h float64) {
	ebitenutil.DrawRect(screen, x, y, w, h, render.ColorPanel)
	ebitenutil.DrawRect(screen, x, y, w, 2, render.ColorAccent)
}

func debugText(screen *ebiten.Image, x, y int, format string, args ...any) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(format, args...), x, y)
}

func drawLine(screen *ebiten.Image, x1, y1, x2, y2 float64, c color.Color) {
	ebitenutil.DrawLine(screen, x1, y1, x2, y2, c)
}
