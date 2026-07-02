package ui

import (
	"fmt"
	"image/color"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
)

const (
	escapeMenuWidth   = 252
	escapeMenuHeight  = 200
	escapeMenuPad     = 16
	escapeMenuTitleH  = 32
	escapeMenuButtonH = 28
	escapeMenuGap     = 8
)

var (
	escapeMenuTextColor     = TextColor
	escapeMenuMutedColor    = MutedTextColor
	escapeMenuTitleColor    = TitleTextColor
	escapeMenuButtonColor   = ButtonColor
	escapeMenuDisabledColor = DisabledColor
	escapeMenuHoverColor    = ButtonHoverColor
)

type EscapeMenu struct {
	open    bool
	action  EscapeMenuAction
	pending bool
	status  string
}

type EscapeMenuAction int

const (
	EscapeMenuActionNone EscapeMenuAction = iota
	EscapeMenuActionCharacterSelect
	EscapeMenuActionSettings
	EscapeMenuActionCancel
	EscapeMenuActionExit
)

type escapeMenuButton struct {
	label   string
	action  EscapeMenuAction
	enabled bool
}

func (m *EscapeMenu) OpenMenu() {
	m.open = true
	m.action = EscapeMenuActionNone
	m.pending = false
	m.status = ""
}

var escapeMenuButtons = []escapeMenuButton{
	{label: "Character Select", action: EscapeMenuActionCharacterSelect, enabled: true},
	{label: "Settings", action: EscapeMenuActionSettings, enabled: true},
	{label: "Cancel", action: EscapeMenuActionCancel, enabled: true},
	{label: "Exit to Windows", action: EscapeMenuActionExit, enabled: true},
}

func (m *EscapeMenu) Update(ctx client.Context) bool {
	if ctx.Input == nil {
		return false
	}
	if !m.open {
		if ctx.Input.JustPressed(render.KeyEscape) {
			m.OpenMenu()
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
		if !button.enabled || (m.pending && button.action != EscapeMenuActionExit) {
			return true
		}
		switch button.action {
		case EscapeMenuActionCharacterSelect:
			m.action = EscapeMenuActionCharacterSelect
		case EscapeMenuActionSettings:
			m.open = false
			m.action = EscapeMenuActionSettings
		case EscapeMenuActionCancel:
			m.open = false
		case EscapeMenuActionExit:
			m.open = false
			if ctx.RequestQuit != nil {
				ctx.RequestQuit()
			}
		}
		return true
	}
	return true
}

func (m *EscapeMenu) RequestCharacterSelect(ctx client.Context) {
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

func (m *EscapeMenu) ApplyRestartAck(ack network.RestartAck) bool {
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

func (m *EscapeMenu) ConsumeAction() EscapeMenuAction {
	action := m.action
	m.action = EscapeMenuActionNone
	return action
}

func (m *EscapeMenu) Draw(screen *render.Image, ctx client.Context, width, height int) {
	if !m.open || screen == nil {
		return
	}
	DrawSurface(screen, 0, 0, width, height, color.RGBA{A: 96}, color.RGBA{})
	x, y, w, h := escapeMenuBounds(width, height)
	DrawTitledWindowFrame(screen, x, y, w, h, escapeMenuTitleH)
	DrawWindowTitle(screen, x, y, escapeMenuTitleH, escapeMenuPad, "Menu", escapeMenuTitleColor)

	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i, button := range escapeMenuButtons {
		bx, by, bw, bh := escapeMenuButtonBounds(x, y, w, i)
		fill := escapeMenuButtonColor
		textColor := escapeMenuTextColor
		enabled := button.enabled && (!m.pending || button.action == EscapeMenuActionExit)
		if !enabled {
			fill = escapeMenuDisabledColor
			textColor = escapeMenuMutedColor
		} else if pointInRect(mx, my, bx, by, bw, bh) {
			fill = escapeMenuHoverColor
		}
		DrawButtonLabel(screen, bx, by, bw, bh, button.label, fill, textColor)
	}
}

func (m *EscapeMenu) CursorAction(ctx client.Context) (int, bool) {
	if !m.open || ctx.Input == nil {
		return 0, false
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := escapeMenuBounds(width, height)
	for i, button := range escapeMenuButtons {
		if !button.enabled || (m.pending && button.action != EscapeMenuActionExit) {
			continue
		}
		bx, by, bw, bh := escapeMenuButtonBounds(x, y, w, i)
		if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, bx, by, bw, bh) {
			return CursorActionClick, true
		}
	}
	return CursorActionDefault, true
}

func (m *EscapeMenu) IsOpen() bool {
	return m.open
}

func (m *EscapeMenu) Pending() bool {
	return m.pending
}

func (m *EscapeMenu) Status() string {
	return m.status
}

func (m *EscapeMenu) Action() EscapeMenuAction {
	return m.action
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
