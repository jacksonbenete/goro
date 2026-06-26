package gamemode

import (
	"image/color"

	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/primitives"
	uiwidget "github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/render"
)

type uiSurfaceKey struct {
	w      int
	h      int
	bg     color.RGBA
	border color.RGBA
}

var uiSurfaceCache = map[uiSurfaceKey]*render.Image{}

func drawUIWindowFrame(screen *render.Image, x, y, w, h int) {
	drawUISurface(screen, x+3, y+4, w, h, color.RGBA{A: 110}, color.RGBA{})
	drawUISurface(screen, x, y, w, h, color.RGBA{R: 24, G: 26, B: 31, A: 232}, color.RGBA{R: 232, G: 218, B: 172, A: 150})
}

func drawUIPanelSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	drawUISurface(screen, x+2, y+3, w, h, color.RGBA{A: 82}, color.RGBA{})
	drawUISurface(screen, x, y, w, h, bg, color.RGBA{R: 232, G: 218, B: 172, A: 125})
}

func drawUIButtonSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	drawUISurface(screen, x, y, w, h, bg, color.RGBA{R: 228, G: 218, B: 184, A: 105})
}

func drawUIRowSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	drawUISurface(screen, x, y, w, h, bg, color.RGBA{})
}

func drawUISurface(screen *render.Image, x, y, w, h int, bg, border color.RGBA) {
	if screen == nil || w <= 0 || h <= 0 {
		return
	}
	img := cachedUISurface(w, h, bg, border)
	if img == nil {
		return
	}
	var opts render.DrawImageOptions
	opts.GeoM.Translate(float64(x), float64(y))
	opts.Filter = render.FilterNearest
	screen.DrawImage(img, &opts)
}

func cachedUISurface(w, h int, bg, border color.RGBA) *render.Image {
	key := uiSurfaceKey{w: w, h: h, bg: bg, border: border}
	if img, ok := uiSurfaceCache[key]; ok {
		return img
	}
	root := primitives.Box().
		Width(float32(w)).
		Height(float32(h)).
		Background(uiColor(bg))
	if border.A != 0 {
		root.BorderStyle(1, uiColor(border))
	}
	r := offscreen.NewRenderer(w, h, offscreen.WithBackground(uiwidget.ColorTransparent))
	r.Render(root)
	src := r.Image()
	if src == nil {
		return nil
	}
	img := render.NewImageFromImage(src)
	if len(uiSurfaceCache) > 256 {
		uiSurfaceCache = map[uiSurfaceKey]*render.Image{}
	}
	uiSurfaceCache[key] = img
	return img
}

func uiColor(c color.RGBA) uiwidget.Color {
	return uiwidget.RGBA8(c.R, c.G, c.B, c.A)
}
