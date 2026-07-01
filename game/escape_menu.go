package game

import (
	"fmt"
	"image/color"

	"github.com/kivutar/goro/network"
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
	escapeMenuTextColor     = uiTextColor
	escapeMenuMutedColor    = uiMutedTextColor
	escapeMenuTitleColor    = uiTitleTextColor
	escapeMenuButtonColor   = uiButtonColor
	escapeMenuDisabledColor = uiDisabledColor
	escapeMenuHoverColor    = uiButtonHoverColor
)

type escapeMenuState struct {
	open    bool
	action  escapeMenuAction
	pending bool
	status  string
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

func (m *escapeMenuState) openMenu() {
	m.open = true
	m.action = escapeMenuActionNone
	m.pending = false
	m.status = ""
}

var escapeMenuButtons = []escapeMenuButton{
	{label: "Character Select", action: escapeMenuActionCharacterSelect, enabled: true},
	{label: "Settings", action: escapeMenuActionSettings, enabled: true},
	{label: "Cancel", action: escapeMenuActionCancel, enabled: true},
	{label: "Exit to Windows", action: escapeMenuActionExit, enabled: true},
}

func (m *escapeMenuState) update(ctx Context) bool {
	if ctx.Input == nil {
		return false
	}
	if !m.open {
		if ctx.Input.JustPressed(render.KeyEscape) {
			m.openMenu()
			return true
		}
		return false
	}
	if ctx.Input.JustPressed(render.KeyEscape) && !m.pending {
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
		if !button.enabled || (m.pending && button.action != escapeMenuActionExit) {
			return true
		}
		switch button.action {
		case escapeMenuActionCharacterSelect:
			m.action = escapeMenuActionCharacterSelect
		case escapeMenuActionSettings:
			m.open = false
			m.action = escapeMenuActionSettings
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

func (m *escapeMenuState) requestCharacterSelect(ctx Context) {
	m.pending = true
	m.status = "Requesting character select..."
	if ctx.Network == nil {
		m.pending = false
		m.status = "Character select failed: not connected"
		return
	}
	if err := ctx.Network.SendRestart(network.RestartTypeCharSelect); err != nil {
		m.pending = false
		m.status = fmt.Sprintf("Character select failed: %v", err)
	}
}

func (m *escapeMenuState) applyRestartAck(ack network.RestartAck) bool {
	if !m.open || !m.pending {
		return false
	}
	if ack.Allowed {
		m.status = "Returning to character select..."
		return true
	}
	m.pending = false
	m.status = "Please wait before changing characters."
	return false
}

func (m *escapeMenuState) consumeAction() escapeMenuAction {
	action := m.action
	m.action = escapeMenuActionNone
	return action
}

func (m *escapeMenuState) draw(screen *render.Image, ctx Context, width, height int) {
	if !m.open || screen == nil {
		return
	}
	drawUISurface(screen, 0, 0, width, height, color.RGBA{A: 96}, color.RGBA{})
	x, y, w, h := escapeMenuBounds(width, height)
	drawUITitledWindowFrame(screen, x, y, w, h, escapeMenuTitleH)
	drawUIWindowTitle(screen, x, y, escapeMenuTitleH, escapeMenuPad, "Menu", escapeMenuTitleColor)

	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i, button := range escapeMenuButtons {
		bx, by, bw, bh := escapeMenuButtonBounds(x, y, w, i)
		fill := escapeMenuButtonColor
		textColor := escapeMenuTextColor
		enabled := button.enabled && (!m.pending || button.action == escapeMenuActionExit)
		if !enabled {
			fill = escapeMenuDisabledColor
			textColor = escapeMenuMutedColor
		} else if pointInRect(mx, my, bx, by, bw, bh) {
			fill = escapeMenuHoverColor
		}
		drawUIButtonLabel(screen, bx, by, bw, bh, button.label, fill, textColor)
	}
	if m.status != "" {
		statusColor := escapeMenuMutedColor
		if !m.pending {
			statusColor = uiErrorTextColor
		}
		render.DebugPrintAtColor(screen, trimRunes(m.status, 30), x+escapeMenuPad, y+h-18, statusColor)
	}
}

func (m *escapeMenuState) cursorAction(ctx Context) (int, bool) {
	if !m.open || ctx.Input == nil {
		return 0, false
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := escapeMenuBounds(width, height)
	for i, button := range escapeMenuButtons {
		if !button.enabled || (m.pending && button.action != escapeMenuActionExit) {
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
