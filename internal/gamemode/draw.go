package gamemode

import (
	"fmt"
	"image/color"

	"github.com/kivutar/goro/internal/render"
)

func clear(screen *render.Image) {
	screen.Fill(render.ColorBackground)
}

func drawPanel(screen *render.Image, x, y, w, h float64) {
	render.DrawRect(screen, x, y, w, h, render.ColorPanel)
	render.DrawRect(screen, x, y, w, 2, render.ColorAccent)
}

func debugText(screen *render.Image, x, y int, format string, args ...any) {
	render.DebugPrintAt(screen, fmt.Sprintf(format, args...), x, y)
}

func drawLine(screen *render.Image, x1, y1, x2, y2 float64, c color.Color) {
	render.DrawLine(screen, x1, y1, x2, y2, c)
}
