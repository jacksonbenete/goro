package ui

import (
	"image/color"
	"math"

	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/primitives"
	uiwidget "github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/render"
)

type surfaceKey struct {
	w      int
	h      int
	bg     color.RGBA
	border color.RGBA
	radius float32
}

var surfaceCache = map[surfaceKey]*render.Image{}

var (
	WindowRadius      = float32(0)
	ButtonRadius      = float32(0)
	ButtonPaddingX    = 14
	WindowBodyColor   = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	WindowTitleTop    = color.RGBA{R: 214, G: 232, B: 250, A: 255}
	WindowTitleColor  = color.RGBA{R: 184, G: 214, B: 242, A: 255}
	WindowBorderColor = color.RGBA{R: 118, G: 160, B: 206, A: 255}
	WindowFooterColor = color.RGBA{R: 244, G: 246, B: 248, A: 255}
	PanelBodyColor    = color.RGBA{R: 250, G: 252, B: 255, A: 255}
	PanelAltColor     = color.RGBA{R: 236, G: 244, B: 252, A: 255}
	PanelHoverColor   = color.RGBA{R: 222, G: 236, B: 250, A: 255}
	DisabledColor     = color.RGBA{R: 226, G: 230, B: 235, A: 255}
	TextColor         = color.RGBA{R: 38, G: 48, B: 58, A: 255}
	MutedTextColor    = color.RGBA{R: 98, G: 112, B: 126, A: 255}
	TitleTextColor    = color.RGBA{R: 22, G: 54, B: 88, A: 255}
	GoodTextColor     = color.RGBA{R: 34, G: 142, B: 72, A: 255}
	ErrorTextColor    = color.RGBA{R: 204, G: 48, B: 48, A: 255}
	ButtonColor       = color.RGBA{R: 236, G: 244, B: 252, A: 255}
	ButtonHoverColor  = color.RGBA{R: 218, G: 235, B: 250, A: 255}
	ButtonDownColor   = color.RGBA{R: 198, G: 222, B: 245, A: 255}
	ButtonBorderColor = color.RGBA{R: 138, G: 174, B: 214, A: 255}
	SeparatorColor    = color.RGBA{R: 160, G: 190, B: 222, A: 190}
	FooterLineColor   = color.RGBA{R: 174, G: 180, B: 188, A: 255}
	SelectionColor    = color.RGBA{R: 206, G: 226, B: 248, A: 255}
	SelectionBorder   = color.RGBA{R: 82, G: 138, B: 200, A: 255}
)

func DrawWindowFrame(screen *render.Image, x, y, w, h int) {
	DrawRoundedSurface(screen, x, y, w, h, WindowBodyColor, WindowBorderColor, WindowRadius)
}

func DrawTitledWindowFrame(screen *render.Image, x, y, w, h, titleH int) {
	DrawWindowFrame(screen, x, y, w, h)
	DrawTitleBar(screen, x, y, w, titleH)
}

func DrawWindowTitle(screen *render.Image, x, y, titleH, pad int, title string, text color.RGBA) {
	DrawTitleTextAt(screen, x+pad, y, titleH, title, text)
}

func DrawTitleTextAt(screen *render.Image, x, y, titleH int, title string, text color.RGBA) {
	_, textH := render.DebugTextSize(title)
	ty := y + (titleH-textH)/2
	render.DebugPrintAtColor(screen, title, x, ty, text)
}

func DrawTitleBar(screen *render.Image, x, y, w, titleH int) {
	if titleH <= 1 {
		return
	}
	barH := titleH - 1
	for row := 0; row < barH; row++ {
		t := 0.0
		if barH > 1 {
			t = float64(row) / float64(barH-1)
		}
		inset := titleBarRowInset(row)
		render.DrawRect(screen, float64(x+1+inset), float64(y+1+row), float64(w-2-2*inset), 1, LerpColor(WindowTitleTop, WindowTitleColor, t))
	}
	render.DrawRect(screen, float64(x+1), float64(y+titleH), float64(w-2), 1, SeparatorColor)
}

func DrawWindowFooter(screen *render.Image, x, y, w, h, footerH int) {
	if screen == nil || w <= 2 || h <= 2 || footerH <= 0 {
		return
	}
	footerY := y + h - footerH
	if footerY <= y {
		footerY = y + 1
	}
	bottom := y + h - 1
	if footerY >= bottom {
		return
	}
	render.DrawRect(screen, float64(x+1), float64(footerY), float64(w-2), float64(bottom-footerY), WindowFooterColor)
	render.DrawRect(screen, float64(x+1), float64(footerY), float64(w-2), 1, FooterLineColor)
}

func titleBarRowInset(row int) int {
	radius := int(math.Ceil(float64(WindowRadius)))
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

func LerpColor(a, b color.RGBA, t float64) color.RGBA {
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

func DrawPanelSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	DrawSurface(screen, x, y, w, h, bg, WindowBorderColor)
}

func DrawButtonSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	DrawRoundedSurface(screen, x, y, w, h, bg, ButtonBorderColor, ButtonRadius)
	DrawGradientInterior(screen, x, y, w, h, Lighten(bg, 0.42), bg, int(math.Ceil(float64(ButtonRadius)))-1)
}

func DrawButtonLabel(screen *render.Image, x, y, w, h int, label string, bg, text color.RGBA) {
	DrawButtonSurface(screen, x, y, w, h, bg)
	DrawCenteredText(screen, x, y, w, h, label, text)
}

func ButtonLabelWidth(label string) int {
	textW, _ := render.DebugTextSize(label)
	return textW + ButtonPaddingX*2
}

func DrawCloseButton(screen *render.Image, x, y, w, h int, bg, line color.RGBA) {
	DrawButtonSurface(screen, x, y, w, h, bg)
	icon := minInt(w, h) / 2
	if icon < 6 {
		icon = minInt(w, h) - 6
	}
	if icon < 2 {
		return
	}
	left := x + (w-icon)/2
	top := y + (h-icon)/2
	right := left + icon - 1
	bottom := top + icon - 1
	render.DrawLine(screen, float64(left), float64(top), float64(right), float64(bottom), line)
	render.DrawLine(screen, float64(right), float64(top), float64(left), float64(bottom), line)
}

func DrawCenteredText(screen *render.Image, x, y, w, h int, label string, text color.RGBA) {
	textW, textH := render.DebugTextSize(label)
	tx := x + (w-textW)/2
	ty := y + (h-textH)/2
	render.DebugPrintAtColor(screen, label, tx, ty, text)
}

func DrawRowSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	DrawSurface(screen, x, y, w, h, bg, color.RGBA{})
}

func DrawSurface(screen *render.Image, x, y, w, h int, bg, border color.RGBA) {
	DrawRoundedSurface(screen, x, y, w, h, bg, border, 0)
}

func DrawRoundedSurface(screen *render.Image, x, y, w, h int, bg, border color.RGBA, radius float32) {
	if screen == nil || w <= 0 || h <= 0 {
		return
	}
	img := cachedSurface(w, h, bg, border, radius)
	if img == nil {
		return
	}
	var opts render.DrawImageOptions
	opts.GeoM.Translate(float64(x), float64(y))
	opts.Filter = render.FilterNearest
	screen.DrawImage(img, &opts)
}

func DrawGradientInterior(screen *render.Image, x, y, w, h int, top, bottom color.RGBA, radius int) {
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
		inset := roundedRowInset(row, innerH, radius)
		if innerW-2*inset <= 0 {
			continue
		}
		render.DrawRect(screen, float64(x+1+inset), float64(y+1+row), float64(innerW-2*inset), 1, LerpColor(top, bottom, t))
	}
}

func roundedRowInset(row, height, radius int) int {
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

func Lighten(c color.RGBA, amount float64) color.RGBA {
	return LerpColor(c, color.RGBA{R: 255, G: 255, B: 255, A: c.A}, clampUnit(amount))
}

func cachedSurface(w, h int, bg, border color.RGBA, radius float32) *render.Image {
	key := surfaceKey{w: w, h: h, bg: bg, border: border, radius: radius}
	if img, ok := surfaceCache[key]; ok {
		return img
	}
	root := primitives.Box().
		Width(float32(w)).
		Height(float32(h)).
		Background(Color(bg))
	if radius > 0 {
		root.Rounded(radius)
	}
	if border.A != 0 {
		root.BorderStyle(1, Color(border))
	}
	r := offscreen.NewRenderer(w, h, offscreen.WithBackground(uiwidget.ColorTransparent))
	r.Render(root)
	src := r.Image()
	if src == nil {
		return nil
	}
	img := render.NewImageFromImage(src)
	if len(surfaceCache) > 256 {
		surfaceCache = map[surfaceKey]*render.Image{}
	}
	surfaceCache[key] = img
	return img
}

func Color(c color.RGBA) uiwidget.Color {
	return uiwidget.RGBA8(c.R, c.G, c.B, c.A)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampUnit(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
