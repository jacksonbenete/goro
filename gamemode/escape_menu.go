package gamemode

import (
	"image/color"

	"github.com/kivutar/goro/render"
)

const (
	escapeMenuWidth   = 252
	escapeMenuHeight  = 214
	escapeMenuPad     = 16
	escapeMenuTitleH  = 32
	escapeMenuButtonH = 28
	escapeMenuGap     = 8
)

var (
	escapeMenuTextColor     = color.RGBA{R: 238, G: 232, B: 218, A: 255}
	escapeMenuMutedColor    = color.RGBA{R: 128, G: 134, B: 142, A: 255}
	escapeMenuTitleColor    = color.RGBA{R: 255, G: 230, B: 150, A: 255}
	escapeMenuButtonColor   = color.RGBA{R: 58, G: 64, B: 74, A: 235}
	escapeMenuDisabledColor = color.RGBA{R: 42, G: 46, B: 54, A: 210}
	escapeMenuHoverColor    = color.RGBA{R: 82, G: 92, B: 108, A: 245}
)

type escapeMenuState struct {
	open bool
}

type escapeMenuAction int

const (
	escapeMenuActionNone escapeMenuAction = iota
	escapeMenuActionCharacterSelect
	escapeMenuActionSettings
	escapeMenuActionCancel
	escapeMenuActionExit
)

type escapeMenuButton struct {
	label   string
	action  escapeMenuAction
	enabled bool
}

var escapeMenuButtons = []escapeMenuButton{
	{label: "Character Select", action: escapeMenuActionCharacterSelect},
	{label: "Settings", action: escapeMenuActionSettings},
	{label: "Cancel", action: escapeMenuActionCancel, enabled: true},
	{label: "Exit to Windows", action: escapeMenuActionExit, enabled: true},
}

func (m *escapeMenuState) update(ctx Context) bool {
	if ctx.Input == nil {
		return false
	}
	if !m.open {
		if ctx.Input.JustPressed(render.KeyEscape) {
			m.open = true
			return true
		}
		return false
	}
	if ctx.Input.JustPressed(render.KeyEscape) {
		m.open = false
		return true
	}
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return true
	}
	width, height := ctx.ScreenSize()
	x, y, w, h := escapeMenuBounds(width, height)
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	if !pointInRect(mx, my, x, y, w, h) {
		return true
	}
	for i, button := range escapeMenuButtons {
		bx, by, bw, bh := escapeMenuButtonBounds(x, y, w, i)
		if !pointInRect(mx, my, bx, by, bw, bh) {
			continue
		}
		if !button.enabled {
			return true
		}
		switch button.action {
		case escapeMenuActionCancel:
			m.open = false
		case escapeMenuActionExit:
			m.open = false
			if ctx.RequestQuit != nil {
				ctx.RequestQuit()
			}
		}
		return true
	}
	return true
}

func (m *escapeMenuState) draw(screen *render.Image, ctx Context, width, height int) {
	if !m.open || screen == nil {
		return
	}
	drawUISurface(screen, 0, 0, width, height, color.RGBA{A: 96}, color.RGBA{})
	x, y, w, h := escapeMenuBounds(width, height)
	drawNPCWindowFrame(screen, x, y, w, h)
	render.DebugPrintAtColor(screen, "Menu", x+escapeMenuPad, y+10, escapeMenuTitleColor)
	render.DrawRect(screen, float64(x+8), float64(y+escapeMenuTitleH), float64(w-16), 1, color.RGBA{R: 210, G: 200, B: 170, A: 80})

	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i, button := range escapeMenuButtons {
		bx, by, bw, bh := escapeMenuButtonBounds(x, y, w, i)
		fill := escapeMenuButtonColor
		textColor := escapeMenuTextColor
		if !button.enabled {
			fill = escapeMenuDisabledColor
			textColor = escapeMenuMutedColor
		} else if pointInRect(mx, my, bx, by, bw, bh) {
			fill = escapeMenuHoverColor
		}
		drawUIButtonSurface(screen, bx, by, bw, bh, fill)
		tx := bx + (bw-len([]rune(button.label))*7)/2
		render.DebugPrintAtColor(screen, button.label, tx, by+7, textColor)
	}
}

func (m *escapeMenuState) cursorAction(ctx Context) (int, bool) {
	if !m.open || ctx.Input == nil {
		return 0, false
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := escapeMenuBounds(width, height)
	for i, button := range escapeMenuButtons {
		if !button.enabled {
			continue
		}
		bx, by, bw, bh := escapeMenuButtonBounds(x, y, w, i)
		if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, bx, by, bw, bh) {
			return cursorActionClick, true
		}
	}
	return cursorActionDefault, true
}

func escapeMenuBounds(width, height int) (int, int, int, int) {
	w := minInt(escapeMenuWidth, maxInt(220, width-40))
	h := minInt(escapeMenuHeight, maxInt(168, height-40))
	x := (width - w) / 2
	y := (height - h) / 2
	if y < 16 {
		y = 16
	}
	return x, y, w, h
}

func escapeMenuButtonBounds(x, y, w, index int) (int, int, int, int) {
	bx := x + escapeMenuPad
	by := y + escapeMenuTitleH + 16 + index*(escapeMenuButtonH+escapeMenuGap)
	bw := w - 2*escapeMenuPad
	return bx, by, bw, escapeMenuButtonH
}
