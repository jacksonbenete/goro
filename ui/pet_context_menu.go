package ui

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	petContextMenuWidth = 140
	petContextMenuRowH  = 28
)

type PetContextActionKind uint8

const (
	PetContextActionNone PetContextActionKind = iota
	PetContextActionInfo
	PetContextActionFeed
	PetContextActionPerformance
	PetContextActionBackToEgg
	PetContextActionUnequipAccessory
)

type PetContextAction struct {
	Kind PetContextActionKind
}

type PetContextMenu struct {
	Window
	action PetContextAction
}

func (m *PetContextMenu) Open(ctx Context, x, y int) {
	m.EnsureWindow(petContextMenuWidth, m.height())
	m.titleHeight = 0
	m.ctx = ctx
	screenW, screenH := ctx.ScreenSize()
	x = clampWindowInt(x, 8, maxInt(8, screenW-petContextMenuWidth-8))
	y = clampWindowInt(y, 8, maxInt(8, screenH-m.height()-8))
	m.OpenAt(x, y, m.widgetTree())
	m.Publish(ctx)
}

func (m *PetContextMenu) Update(ctx Context) bool {
	m.EnsureWindow(petContextMenuWidth, m.height())
	m.titleHeight = 0
	m.ctx = ctx
	if !m.IsOpen() {
		return false
	}
	if ctx.Input != nil {
		inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, m.x, m.y, petContextMenuWidth, m.height())
		if ctx.Input.JustPressed(input.KeyEscape) || (!inside && (ctx.Input.MouseJustPressed(input.MouseButtonLeft) || ctx.Input.MouseJustPressed(input.MouseButtonRight))) {
			m.Close()
			return true
		}
	}
	consumed := m.Window.Update(ctx)
	m.Publish(ctx)
	return consumed
}

func (m *PetContextMenu) PopAction() PetContextAction {
	action := m.action
	m.action = PetContextAction{}
	return action
}

func (m *PetContextMenu) widgetTree() widget.Widget {
	return Win(
		TitleBar(false),
		Radius(0),
		Size(petContextMenuWidth, float32(m.height())),
		Content(
			primitives.Box(
				m.button("Check Pet Status", PetContextActionInfo),
				m.button("Feed", PetContextActionFeed),
				m.button("Performance", PetContextActionPerformance),
				m.button("Unequip Accessory", PetContextActionUnequipAccessory),
				m.button("Return to Egg", PetContextActionBackToEgg),
			),
		),
	)
}

func (m *PetContextMenu) button(label string, action PetContextActionKind) widget.Widget {
	return rotheme.Button(label, func() {
		m.action = PetContextAction{Kind: action}
		m.Close()
	}).
		Width(petContextMenuWidth).
		Height(petContextMenuRowH)
}

func (m *PetContextMenu) height() int {
	return petContextMenuRowH * 5
}
