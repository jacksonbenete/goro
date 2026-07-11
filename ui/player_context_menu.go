package ui

import (
	"strings"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	playerContextMenuWidth  = 118
	playerContextMenuRowH   = 28
	playerContextMenuHeight = playerContextMenuRowH * 2
)

type PlayerContextActionKind uint8

const (
	PlayerContextActionNone PlayerContextActionKind = iota
	PlayerContextActionAddFriend
	PlayerContextActionTrade
)

type PlayerContextAction struct {
	Kind    PlayerContextActionKind
	ActorID uint32
	Name    string
}

type PlayerContextMenu struct {
	window       WindowState
	ctx          Context
	actorID      uint32
	name         string
	canAddFriend bool
	action       PlayerContextAction
}

func (m *PlayerContextMenu) Open(ctx Context, x, y int, actorID uint32, name string, canAddFriend bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	m.ensureWindow()
	m.ctx = ctx
	m.actorID = actorID
	m.name = name
	m.canAddFriend = canAddFriend
	screenW, screenH := ctx.ScreenSize()
	height := m.height()
	m.window.SetSize(playerContextMenuWidth, height)
	x = clampWindowInt(x, 8, maxInt(8, screenW-playerContextMenuWidth-8))
	y = clampWindowInt(y, 8, maxInt(8, screenH-height-8))
	m.window.OpenAt(x, y, m.widgetTree(ctx))
	m.Publish(ctx)
}

func (m *PlayerContextMenu) Update(ctx Context) bool {
	m.ensureWindow()
	m.ctx = ctx
	if !m.window.IsOpen() {
		return false
	}
	if ctx.Input != nil {
		inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, m.window.x, m.window.y, playerContextMenuWidth, m.height())
		if ctx.Input.JustPressed(render.KeyEscape) || (!inside && (ctx.Input.MouseJustPressed(render.MouseButtonLeft) || ctx.Input.MouseJustPressed(render.MouseButtonRight))) {
			m.Close()
			return true
		}
	}
	consumed := m.window.Update(ctx)
	m.Publish(ctx)
	return consumed
}

func (m *PlayerContextMenu) Close() {
	m.ensureWindow()
	m.window.Close()
	m.Publish(m.ctx)
}

func (m *PlayerContextMenu) Publish(ctx Context) {
	m.ensureWindow()
	m.window.Publish(ctx)
}

func (m *PlayerContextMenu) PopAction() PlayerContextAction {
	action := m.action
	m.action = PlayerContextAction{}
	return action
}

func (m *PlayerContextMenu) ensureWindow() {
	if m.window.width != 0 {
		return
	}
	m.window = NewWindowState(playerContextMenuWidth, playerContextMenuHeight)
	m.window.titleHeight = 0
}

func (m *PlayerContextMenu) widgetTree(ctx Context) widget.Widget {
	rows := []widget.Widget{
		rotheme.Button("Trade", func() {
			m.action = PlayerContextAction{Kind: PlayerContextActionTrade, ActorID: m.actorID, Name: m.name}
			m.Close()
		}).
			Width(playerContextMenuWidth).
			Height(playerContextMenuRowH),
	}
	if m.canAddFriend {
		rows = append(rows,
			rotheme.Button("Add Friend", func() {
				m.action = PlayerContextAction{Kind: PlayerContextActionAddFriend, ActorID: m.actorID, Name: m.name}
				m.Close()
			}).
				Width(playerContextMenuWidth).
				Height(playerContextMenuRowH),
		)
	}
	return Window(
		TitleBar(false),
		Radius(0),
		Size(playerContextMenuWidth, float32(m.height())),
		Content(
			primitives.Box(rows...),
		),
	)
}

func (m *PlayerContextMenu) height() int {
	if m.canAddFriend {
		return playerContextMenuHeight
	}
	return playerContextMenuRowH
}
