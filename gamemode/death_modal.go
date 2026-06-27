package gamemode

import (
	"fmt"
	"image/color"

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
	deathModalTitleColor  = color.RGBA{R: 255, G: 214, B: 142, A: 255}
	deathModalTextColor   = color.RGBA{R: 238, G: 232, B: 218, A: 255}
	deathModalMutedColor  = color.RGBA{R: 148, G: 154, B: 164, A: 255}
	deathModalErrorColor  = color.RGBA{R: 255, G: 132, B: 132, A: 255}
	deathModalButtonColor = color.RGBA{R: 58, G: 64, B: 74, A: 235}
	deathModalHoverColor  = color.RGBA{R: 82, G: 92, B: 108, A: 245}
)

type deathModalState struct {
	open    bool
	pending deathModalAction
	status  string
}

type deathModalAction int

const (
	deathModalActionNone deathModalAction = iota
	deathModalActionSavePoint
	deathModalActionCharSelect
	deathModalActionExit
)

type deathModalButton struct {
	label  string
	action deathModalAction
}

var deathModalButtons = []deathModalButton{
	{label: "Return to Save Point", action: deathModalActionSavePoint},
	{label: "Character Select", action: deathModalActionCharSelect},
	{label: "Exit to Windows", action: deathModalActionExit},
}

func (m *deathModalState) openDeath() {
	m.open = true
	m.pending = deathModalActionNone
	m.status = ""
}

func (m *deathModalState) reset() {
	*m = deathModalState{}
}

func (m *deathModalState) clearIfAlive(ctx Context) {
	if !m.open || ctx.Session == nil {
		return
	}
	if ctx.Session.Vitals.HP > 0 || ctx.Session.Selected.HP > 0 {
		m.reset()
	}
}

func (m *deathModalState) applyRestartAck(ack network.RestartAck) {
	if !m.open || m.pending != deathModalActionCharSelect {
		return
	}
	if ack.Allowed {
		m.status = "Returning to character select..."
		return
	}
	m.pending = deathModalActionNone
	m.status = "Please wait before changing characters."
}

func (m *deathModalState) update(ctx Context) bool {
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
		if m.pending != deathModalActionNone && button.action != deathModalActionExit {
			return true
		}
		m.activate(ctx, button.action)
		return true
	}
	return true
}

func (m *deathModalState) activate(ctx Context, action deathModalAction) {
	switch action {
	case deathModalActionSavePoint:
		m.pending = action
		m.status = "Returning to save point..."
		if ctx.Network == nil {
			m.pending = deathModalActionNone
			m.status = "Respawn failed: not connected"
			return
		}
		if err := ctx.Network.SendRestart(network.RestartTypeRespawn); err != nil {
			m.pending = deathModalActionNone
			m.status = fmt.Sprintf("Respawn failed: %v", err)
		}
	case deathModalActionCharSelect:
		m.pending = action
		m.status = "Requesting character select..."
		if ctx.Network == nil {
			m.pending = deathModalActionNone
			m.status = "Character select failed: not connected"
			return
		}
		if err := ctx.Network.SendRestart(network.RestartTypeCharSelect); err != nil {
			m.pending = deathModalActionNone
			m.status = fmt.Sprintf("Character select failed: %v", err)
		}
	case deathModalActionExit:
		m.open = false
		if ctx.RequestQuit != nil {
			ctx.RequestQuit()
		}
	}
}

func (m *deathModalState) draw(screen *render.Image, ctx Context, width, height int) {
	if !m.open || screen == nil {
		return
	}
	drawUISurface(screen, 0, 0, width, height, color.RGBA{A: 112}, color.RGBA{})
	x, y, w, h := deathModalBounds(width, height)
	drawNPCWindowFrame(screen, x, y, w, h)
	render.DebugPrintAtColor(screen, "You have died", x+deathModalPad, y+10, deathModalTitleColor)
	render.DrawRect(screen, float64(x+8), float64(y+deathModalTitleH), float64(w-16), 1, color.RGBA{R: 210, G: 200, B: 170, A: 80})
	render.DebugPrintAtColor(screen, "Choose what to do next.", x+deathModalPad, y+deathModalTitleH+14, deathModalTextColor)

	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i, button := range deathModalButtons {
		bx, by, bw, bh := deathModalButtonBounds(x, y, w, i)
		fill := deathModalButtonColor
		textColor := deathModalTextColor
		enabled := m.pending == deathModalActionNone || button.action == deathModalActionExit
		if !enabled {
			fill = escapeMenuDisabledColor
			textColor = deathModalMutedColor
		} else if pointInRect(mx, my, bx, by, bw, bh) {
			fill = deathModalHoverColor
		}
		drawUIButtonSurface(screen, bx, by, bw, bh, fill)
		tx := bx + (bw-len([]rune(button.label))*7)/2
		render.DebugPrintAtColor(screen, button.label, tx, by+7, textColor)
	}

	if m.status != "" {
		statusColor := deathModalMutedColor
		if m.pending == deathModalActionNone {
			statusColor = deathModalErrorColor
		}
		render.DebugPrintAtColor(screen, trimRunes(m.status, 38), x+deathModalPad, y+h-22, statusColor)
	}
}

func (m *deathModalState) cursorAction(ctx Context) (int, bool) {
	if !m.open || ctx.Input == nil {
		return 0, false
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := deathModalBounds(width, height)
	if m.pending != deathModalActionNone {
		bx, by, bw, bh := deathModalButtonBounds(x, y, w, len(deathModalButtons)-1)
		if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, bx, by, bw, bh) {
			return cursorActionClick, true
		}
		return cursorActionDefault, true
	}
	for i := range deathModalButtons {
		bx, by, bw, bh := deathModalButtonBounds(x, y, w, i)
		if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, bx, by, bw, bh) {
			return cursorActionClick, true
		}
	}
	return cursorActionDefault, true
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
