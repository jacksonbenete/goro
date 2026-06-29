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

var (
	uiWindowBodyColor   = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	uiWindowTitleColor  = color.RGBA{R: 184, G: 214, B: 242, A: 255}
	uiWindowBorderColor = color.RGBA{R: 118, G: 160, B: 206, A: 255}
	uiPanelBodyColor    = color.RGBA{R: 250, G: 252, B: 255, A: 255}
	uiPanelAltColor     = color.RGBA{R: 236, G: 244, B: 252, A: 255}
	uiPanelHoverColor   = color.RGBA{R: 222, G: 236, B: 250, A: 255}
	uiPanelDownColor    = color.RGBA{R: 204, G: 224, B: 246, A: 255}
	uiDisabledColor     = color.RGBA{R: 226, G: 230, B: 235, A: 255}
	uiTextColor         = color.RGBA{R: 38, G: 48, B: 58, A: 255}
	uiMutedTextColor    = color.RGBA{R: 98, G: 112, B: 126, A: 255}
	uiTitleTextColor    = color.RGBA{R: 22, G: 54, B: 88, A: 255}
	uiGoodTextColor     = color.RGBA{R: 34, G: 142, B: 72, A: 255}
	uiErrorTextColor    = color.RGBA{R: 204, G: 48, B: 48, A: 255}
	uiButtonColor       = color.RGBA{R: 236, G: 244, B: 252, A: 255}
	uiButtonHoverColor  = color.RGBA{R: 218, G: 235, B: 250, A: 255}
	uiButtonDownColor   = color.RGBA{R: 198, G: 222, B: 245, A: 255}
	uiButtonBorderColor = color.RGBA{R: 138, G: 174, B: 214, A: 255}
	uiSeparatorColor    = color.RGBA{R: 160, G: 190, B: 222, A: 190}
	uiSelectionColor    = color.RGBA{R: 206, G: 226, B: 248, A: 255}
	uiSelectionBorder   = color.RGBA{R: 82, G: 138, B: 200, A: 255}
)

func drawUIWindowFrame(screen *render.Image, x, y, w, h int) {
	drawUISurface(screen, x, y, w, h, uiWindowBodyColor, uiWindowBorderColor)
}

func drawUIPanelSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	drawUISurface(screen, x, y, w, h, bg, uiWindowBorderColor)
}

func drawUIButtonSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	drawUISurface(screen, x, y, w, h, bg, uiButtonBorderColor)
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
