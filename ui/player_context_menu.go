package ui

import (
	"github.com/kivutar/goro/input"
	"strings"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	playerContextMenuWidth = 118
	playerContextMenuRowH  = 28
)

type PlayerContextActionKind uint8

const (
	PlayerContextActionNone PlayerContextActionKind = iota
	PlayerContextActionAddFriend
	PlayerContextActionInviteParty
	PlayerContextActionInviteGuild
	PlayerContextActionTrade
	PlayerContextActionSeeEquipment
)

type PlayerContextAction struct {
	Kind    PlayerContextActionKind
	ActorID uint32
	Name    string
}

type PlayerContextMenu struct {
	Window
	actorID      uint32
	name         string
	canAddFriend bool
	canParty     bool
	canGuild     bool
	action       PlayerContextAction
}

func (m *PlayerContextMenu) Open(ctx Context, x, y int, actorID uint32, name string, canAddFriend bool, canParty bool, canGuild bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	m.EnsureWindow(playerContextMenuWidth, m.height())
	m.titleHeight = 0
	m.ctx = ctx
	m.actorID = actorID
	m.name = name
	m.canAddFriend = canAddFriend
	m.canParty = canParty
	m.canGuild = canGuild
	screenW, screenH := ctx.ScreenSize()
	height := m.height()
	m.SetSize(playerContextMenuWidth, height)
	x = clampWindowInt(x, 8, maxInt(8, screenW-playerContextMenuWidth-8))
	y = clampWindowInt(y, 8, maxInt(8, screenH-height-8))
	m.OpenAt(x, y, m.widgetTree(ctx))
	m.Publish(ctx)
}

func (m *PlayerContextMenu) Update(ctx Context) bool {
	m.EnsureWindow(playerContextMenuWidth, m.height())
	m.titleHeight = 0
	m.ctx = ctx
	if !m.IsOpen() {
		return false
	}
	if ctx.Input != nil {
		inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, m.x, m.y, playerContextMenuWidth, m.height())
		if ctx.Input.JustPressed(input.KeyEscape) || (!inside && (ctx.Input.MouseJustPressed(input.MouseButtonLeft) || ctx.Input.MouseJustPressed(input.MouseButtonRight))) {
			m.Close()
			return true
		}
	}
	consumed := m.Window.Update(ctx)
	m.Publish(ctx)
	return consumed
}

func (m *PlayerContextMenu) PopAction() PlayerContextAction {
	action := m.action
	m.action = PlayerContextAction{}
	return action
}

func (m *PlayerContextMenu) widgetTree(ctx Context) widget.Widget {
	rows := []widget.Widget{
		rotheme.Button("Trade", func() {
			m.action = PlayerContextAction{Kind: PlayerContextActionTrade, ActorID: m.actorID, Name: m.name}
			m.Close()
		}).
			Width(playerContextMenuWidth).
			Height(playerContextMenuRowH),
		rotheme.Button("See equipment", func() {
			m.action = PlayerContextAction{Kind: PlayerContextActionSeeEquipment, ActorID: m.actorID, Name: m.name}
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
	if m.canParty {
		rows = append(rows,
			rotheme.Button("Invite", func() {
				m.action = PlayerContextAction{Kind: PlayerContextActionInviteParty, ActorID: m.actorID, Name: m.name}
				m.Close()
			}).
				Width(playerContextMenuWidth).
				Height(playerContextMenuRowH),
		)
	}
	if m.canGuild {
		rows = append(rows,
			rotheme.Button("Invite Guild", func() {
				m.action = PlayerContextAction{Kind: PlayerContextActionInviteGuild, ActorID: m.actorID, Name: m.name}
				m.Close()
			}).
				Width(playerContextMenuWidth).
				Height(playerContextMenuRowH),
		)
	}
	return Win(
		TitleBar(false),
		Radius(0),
		Size(playerContextMenuWidth, float32(m.height())),
		Content(
			primitives.Box(rows...),
		),
	)
}

func (m *PlayerContextMenu) height() int {
	rows := 2
	if m.canAddFriend {
		rows++
	}
	if m.canParty {
		rows++
	}
	if m.canGuild {
		rows++
	}
	return playerContextMenuRowH * rows
}
