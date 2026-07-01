package ui

import (
	"fmt"
	"image/color"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
)

const (
	deathModalWidth   = 312
	deathModalHeight  = 190
	deathModalPad     = 16
	deathModalTitleH  = 32
	deathModalButtonH = 28
	deathModalGap     = 8
)

var (
	deathModalTitleColor  = TitleTextColor
	deathModalTextColor   = TextColor
	deathModalMutedColor  = MutedTextColor
	deathModalErrorColor  = ErrorTextColor
	deathModalButtonColor = ButtonColor
	deathModalHoverColor  = ButtonHoverColor
)

type DeathModal struct {
	open    bool
	pending DeathModalAction
	status  string
}

type DeathModalAction int

const (
	DeathModalActionNone DeathModalAction = iota
	DeathModalActionSavePoint
	DeathModalActionCharSelect
	DeathModalActionExit
)

type deathModalButton struct {
	label  string
	action DeathModalAction
}

var deathModalButtons = []deathModalButton{
	{label: "Return to Save Point", action: DeathModalActionSavePoint},
	{label: "Character Select", action: DeathModalActionCharSelect},
	{label: "Exit to Windows", action: DeathModalActionExit},
}

func (m *DeathModal) OpenDeath() {
	m.open = true
	m.pending = DeathModalActionNone
	m.status = ""
}

func (m *DeathModal) Reset() {
	*m = DeathModal{}
}

func (m *DeathModal) ClearIfAlive(ctx client.Context) {
	if !m.open || ctx.Session == nil {
		return
	}
	if ctx.Session.Vitals.HP > 0 || ctx.Session.Selected.HP > 0 {
		m.Reset()
	}
}

func (m *DeathModal) ApplyRestartAck(ack network.RestartAck) bool {
	if !m.open || m.pending != DeathModalActionCharSelect {
		return false
	}
	if ack.Allowed {
		m.status = "Returning to character select..."
		return true
	}
	m.pending = DeathModalActionNone
	m.status = "Please wait before changing characters."
	return false
}

func (m *DeathModal) Update(ctx client.Context) bool {
	if !m.open || ctx.Input == nil {
		return m.open
	}
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return true
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := deathModalBounds(width, height)
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	for i, button := range deathModalButtons {
		bx, by, bw, bh := deathModalButtonBounds(x, y, w, i)
		if !pointInRect(mx, my, bx, by, bw, bh) {
			continue
		}
		if m.pending != DeathModalActionNone && button.action != DeathModalActionExit {
			return true
		}
		m.activate(ctx, button.action)
		return true
	}
	return true
}

func (m *DeathModal) activate(ctx client.Context, action DeathModalAction) {
	switch action {
	case DeathModalActionSavePoint:
		m.pending = action
		m.status = "Returning to save point..."
		if ctx.Network == nil {
			m.pending = DeathModalActionNone
			m.status = "Respawn failed: not connected"
			return
		}
		if err := ctx.Network.SendRestart(network.RestartTypeRespawn); err != nil {
			m.pending = DeathModalActionNone
			m.status = fmt.Sprintf("Respawn failed: %v", err)
		}
	case DeathModalActionCharSelect:
		m.pending = action
		m.status = "Requesting character select..."
		if ctx.Network == nil {
			m.pending = DeathModalActionNone
			m.status = "Character select failed: not connected"
			return
		}
		if err := ctx.Network.SendRestart(network.RestartTypeCharSelect); err != nil {
			m.pending = DeathModalActionNone
			m.status = fmt.Sprintf("Character select failed: %v", err)
		}
	case DeathModalActionExit:
		m.open = false
		if ctx.RequestQuit != nil {
			ctx.RequestQuit()
		}
	}
}

func (m *DeathModal) Draw(screen *render.Image, ctx client.Context, width, height int) {
	if !m.open || screen == nil {
		return
	}
	DrawSurface(screen, 0, 0, width, height, color.RGBA{A: 112}, color.RGBA{})
	x, y, w, h := deathModalBounds(width, height)
	DrawTitledWindowFrame(screen, x, y, w, h, deathModalTitleH)
	DrawWindowTitle(screen, x, y, deathModalTitleH, deathModalPad, "You have died", deathModalTitleColor)
	render.DebugPrintAtColor(screen, "Choose what to do next.", x+deathModalPad, y+deathModalTitleH+14, deathModalTextColor)

	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i, button := range deathModalButtons {
		bx, by, bw, bh := deathModalButtonBounds(x, y, w, i)
		fill := deathModalButtonColor
		textColor := deathModalTextColor
		enabled := m.pending == DeathModalActionNone || button.action == DeathModalActionExit
		if !enabled {
			fill = escapeMenuDisabledColor
			textColor = deathModalMutedColor
		} else if pointInRect(mx, my, bx, by, bw, bh) {
			fill = deathModalHoverColor
		}
		DrawButtonLabel(screen, bx, by, bw, bh, button.label, fill, textColor)
	}

	if m.status != "" {
		statusColor := deathModalMutedColor
		if m.pending == DeathModalActionNone {
			statusColor = deathModalErrorColor
		}
		render.DebugPrintAtColor(screen, trimRunes(m.status, 38), x+deathModalPad, y+h-22, statusColor)
	}
}

func (m *DeathModal) CursorAction(ctx client.Context) (int, bool) {
	if !m.open || ctx.Input == nil {
		return 0, false
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := deathModalBounds(width, height)
	if m.pending != DeathModalActionNone {
		bx, by, bw, bh := deathModalButtonBounds(x, y, w, len(deathModalButtons)-1)
		if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, bx, by, bw, bh) {
			return CursorActionClick, true
		}
		return CursorActionDefault, true
	}
	for i := range deathModalButtons {
		bx, by, bw, bh := deathModalButtonBounds(x, y, w, i)
		if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, bx, by, bw, bh) {
			return CursorActionClick, true
		}
	}
	return CursorActionDefault, true
}

func (m *DeathModal) IsOpen() bool {
	return m.open
}

func (m *DeathModal) PendingAction() DeathModalAction {
	return m.pending
}

func deathModalBounds(width, height int) (int, int, int, int) {
	w := minInt(deathModalWidth, maxInt(240, width-40))
	h := minInt(deathModalHeight, maxInt(168, height-40))
	x := (width - w) / 2
	y := (height - h) / 2
	if y < 16 {
		y = 16
	}
	return x, y, w, h
}

func deathModalButtonBounds(x, y, w, index int) (int, int, int, int) {
	bx := x + deathModalPad
	by := y + deathModalTitleH + 40 + index*(deathModalButtonH+deathModalGap)
	bw := w - 2*deathModalPad
	return bx, by, bw, deathModalButtonH
}
