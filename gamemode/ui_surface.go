package gamemode

import (
	"image/color"
	"math"

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
	radius float32
}

var uiSurfaceCache = map[uiSurfaceKey]*render.Image{}

var (
	uiWindowRadius      = float32(4)
	uiButtonRadius      = float32(4)
	uiWindowBodyColor   = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	uiWindowTitleTop    = color.RGBA{R: 214, G: 232, B: 250, A: 255}
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
	drawUIRoundedSurface(screen, x, y, w, h, uiWindowBodyColor, uiWindowBorderColor, uiWindowRadius)
}

func drawUITitledWindowFrame(screen *render.Image, x, y, w, h, titleH int) {
	drawUIWindowFrame(screen, x, y, w, h)
	drawUITitleBar(screen, x, y, w, titleH)
}

func drawUITitleBar(screen *render.Image, x, y, w, titleH int) {
	if titleH <= 1 {
		return
	}
	barH := titleH - 1
	for row := 0; row < barH; row++ {
		t := 0.0
		if barH > 1 {
			t = float64(row) / float64(barH-1)
		}
		inset := uiTitleBarRowInset(row)
		render.DrawRect(screen, float64(x+1+inset), float64(y+1+row), float64(w-2-2*inset), 1, lerpUIColor(uiWindowTitleTop, uiWindowTitleColor, t))
	}
	render.DrawRect(screen, float64(x+1), float64(y+titleH), float64(w-2), 1, uiSeparatorColor)
}

func uiTitleBarRowInset(row int) int {
	radius := int(math.Ceil(float64(uiWindowRadius)))
	if radius <= 0 || row >= radius {
		return 0
	}
	dy := float64(radius-row) - 0.5
	r := float64(radius)
	if dy <= 0 || dy >= r {
		return radius
	}
	return int(math.Ceil(r - math.Sqrt(r*r-dy*dy)))
}

func lerpUIColor(a, b color.RGBA, t float64) color.RGBA {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t + 0.5),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t + 0.5),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t + 0.5),
		A: uint8(float64(a.A) + (float64(b.A)-float64(a.A))*t + 0.5),
	}
}

func drawUIPanelSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	drawUISurface(screen, x, y, w, h, bg, uiWindowBorderColor)
}

func drawUIButtonSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	drawUIRoundedSurface(screen, x, y, w, h, bg, uiButtonBorderColor, uiButtonRadius)
	drawUIGradientInterior(screen, x, y, w, h, uiLighten(bg, 0.42), bg, int(math.Ceil(float64(uiButtonRadius)))-1)
}

func drawUIRowSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	drawUISurface(screen, x, y, w, h, bg, color.RGBA{})
}

func drawUISurface(screen *render.Image, x, y, w, h int, bg, border color.RGBA) {
	drawUIRoundedSurface(screen, x, y, w, h, bg, border, 0)
}

func drawUIRoundedSurface(screen *render.Image, x, y, w, h int, bg, border color.RGBA, radius float32) {
	if screen == nil || w <= 0 || h <= 0 {
		return
	}
	img := cachedUISurface(w, h, bg, border, radius)
	if img == nil {
		return
	}
	var opts render.DrawImageOptions
	opts.GeoM.Translate(float64(x), float64(y))
	opts.Filter = render.FilterNearest
	screen.DrawImage(img, &opts)
}

func drawUIGradientInterior(screen *render.Image, x, y, w, h int, top, bottom color.RGBA, radius int) {
	innerW := w - 2
	innerH := h - 2
	if screen == nil || innerW <= 0 || innerH <= 0 {
		return
	}
	for row := 0; row < innerH; row++ {
		t := 0.0
		if innerH > 1 {
			t = float64(row) / float64(innerH-1)
		}
		inset := uiRoundedRowInset(row, innerH, radius)
		if innerW-2*inset <= 0 {
			continue
		}
		render.DrawRect(screen, float64(x+1+inset), float64(y+1+row), float64(innerW-2*inset), 1, lerpUIColor(top, bottom, t))
	}
}

func uiRoundedRowInset(row, height, radius int) int {
	if radius <= 0 || height <= 0 {
		return 0
	}
	edge := row
	if bottom := height - 1 - row; bottom < edge {
		edge = bottom
	}
	if edge >= radius {
		return 0
	}
	dy := float64(radius-edge) - 0.5
	r := float64(radius)
	if dy <= 0 || dy >= r {
		return radius
	}
	return int(math.Ceil(r - math.Sqrt(r*r-dy*dy)))
}

func uiLighten(c color.RGBA, amount float64) color.RGBA {
	return lerpUIColor(c, color.RGBA{R: 255, G: 255, B: 255, A: c.A}, clampUnit(amount))
}

func cachedUISurface(w, h int, bg, border color.RGBA, radius float32) *render.Image {
	key := uiSurfaceKey{w: w, h: h, bg: bg, border: border, radius: radius}
	if img, ok := uiSurfaceCache[key]; ok {
		return img
	}
	root := primitives.Box().
		Width(float32(w)).
		Height(float32(h)).
		Background(uiColor(bg))
	if radius > 0 {
		root.Rounded(radius)
	}
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
