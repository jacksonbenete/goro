package ui

import (
	"log"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	escapeMenuWidth  = 252
	escapeMenuHeight = 200
	escapeMenuPad    = 16
	escapeMenuGap    = 8
)

type EscapeMenu struct {
	Window
	action  EscapeMenuAction
	pending bool
	ctx     client.Context
}

type EscapeMenuAction int

const (
	EscapeMenuActionNone EscapeMenuAction = iota
	EscapeMenuActionCharacterSelect
	EscapeMenuActionSettings
	EscapeMenuActionCancel
	EscapeMenuActionExit
)

func (m *EscapeMenu) OpenMenu() {
	m.open = true
	m.action = EscapeMenuActionNone
	m.pending = false
	m.CloseOnEsc = true
}

func (m *EscapeMenu) Update(ctx client.Context) bool {
	m.ctx = ctx
	if m.action != EscapeMenuActionNone {
		return true
	}
	if ctx.Input == nil {
		return false
	}
	if !m.IsOpen() {
		if ctx.Input.JustPressed(render.KeyEscape) {
			m.OpenMenu()
			m.openWindow(ctx)
			return true
		}
		return false
	}
	m.openWindow(ctx)
	if m.Window.Update(ctx) {
		if !m.IsOpen() {
			m.Publish(ctx)
		}
		return true
	}
	return true
}

func (m *EscapeMenu) RequestCharacterSelect(ctx client.Context) {
	m.pending = true
	if ctx.Network == nil {
		m.pending = false
		m.refresh(ctx)
		return
	}
	if err := ctx.Network.SendRestart(network.RestartTypeCharSelect); err != nil {
		m.pending = false
		log.Printf("escape menu character select failed: %v", err)
		m.refresh(ctx)
		return
	}
	m.refresh(ctx)
}

func (m *EscapeMenu) ApplyRestartAck(ack network.RestartAck) bool {
	if !m.pending {
		return false
	}
	if ack.Allowed {
		m.refresh(m.ctx)
		return true
	}
	m.pending = false
	m.EnsureWindow(escapeMenuWidth, escapeMenuHeight)
	m.open = true
	m.refresh(m.ctx)
	return false
}

func (m *EscapeMenu) ConsumeAction() EscapeMenuAction {
	action := m.action
	m.action = EscapeMenuActionNone
	return action
}

func (m *EscapeMenu) Pending() bool {
	return m.pending
}

func (m *EscapeMenu) Action() EscapeMenuAction {
	return m.action
}

func (m *EscapeMenu) openWindow(ctx client.Context) {
	m.EnsureWindow(escapeMenuWidth, escapeMenuHeight)
	if !m.IsOpen() {
		m.Window.Open(ctx, m.widgetTree(ctx))
	}
	m.Publish(ctx)
}

func (m *EscapeMenu) refresh(ctx client.Context) {
	m.EnsureWindow(escapeMenuWidth, escapeMenuHeight)
	if !m.IsOpen() {
		return
	}
	m.SetContent(m.widgetTree(ctx))
	m.Publish(ctx)
}

func (m *EscapeMenu) widgetTree(ctx client.Context) widget.Widget {
	return Win(
		Title("Menu"),
		CloseButton(false),
		Size(escapeMenuWidth, escapeMenuHeight),
		Content(
			primitives.Box(
				rotheme.LargeButtonDisabled("Character Select", m.pending, func() {
					m.action = EscapeMenuActionCharacterSelect
					m.refresh(ctx)
				}),
				rotheme.LargeButtonDisabled("Settings", m.pending, func() {
					m.action = EscapeMenuActionSettings
					m.Window.Close()
				}),
				rotheme.LargeButtonDisabled("Cancel", m.pending, func() {
					m.Window.Close()
				}),
				rotheme.LargeButton("Exit to Windows", func() {
					m.Window.Close()
					if ctx.RequestQuit != nil {
						ctx.RequestQuit()
					}
				}),
			).
				Padding(escapeMenuPad).
				Gap(escapeMenuGap).
				CrossAlign(primitives.CrossAxisStretch),
		),
	)
}
