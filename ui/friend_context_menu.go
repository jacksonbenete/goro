package ui

import (
	"github.com/kivutar/goro/input"
	"strings"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	friendContextMenuWidth = 124
	friendContextMenuRowH  = 28
)

type FriendContextMenu struct {
	Window
	friend session.Friend
	action FriendsWindowAction
}

func (m *FriendContextMenu) Open(ctx Context, x, y int, friend session.Friend) {
	name := strings.TrimSpace(friend.Name)
	if name == "" {
		return
	}
	m.EnsureWindow(friendContextMenuWidth, m.height())
	m.titleHeight = 0
	m.ctx = ctx
	m.friend = friend
	screenW, screenH := ctx.ScreenSize()
	height := m.height()
	m.SetSize(friendContextMenuWidth, height)
	x = clampWindowInt(x, windowScreenMargin, maxInt(windowScreenMargin, screenW-friendContextMenuWidth-windowScreenMargin))
	y = clampWindowInt(y, windowScreenMargin, maxInt(windowScreenMargin, screenH-height-windowScreenMargin))
	m.OpenAt(x, y, m.widgetTree())
	m.Publish(ctx)
}

func (m *FriendContextMenu) Update(ctx Context) bool {
	m.EnsureWindow(friendContextMenuWidth, m.height())
	m.titleHeight = 0
	m.ctx = ctx
	if !m.IsOpen() {
		return false
	}
	if ctx.Input != nil {
		inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, m.x, m.y, friendContextMenuWidth, m.height())
		if ctx.Input.JustPressed(input.KeyEscape) || (!inside && (ctx.Input.MouseJustPressed(input.MouseButtonLeft) || ctx.Input.MouseJustPressed(input.MouseButtonRight))) {
			m.Close()
			return true
		}
	}
	consumed := m.Window.Update(ctx)
	m.Publish(ctx)
	return consumed
}

func (m *FriendContextMenu) Rebind(ctx Context) {
	if !m.IsOpen() {
		return
	}
	m.ctx = ctx
	m.SetContent(m.widgetTree())
	m.Publish(ctx)
}

func (m *FriendContextMenu) PopAction() FriendsWindowAction {
	action := m.action
	m.action = FriendsWindowAction{}
	return action
}

func (m *FriendContextMenu) widgetTree() widget.Widget {
	return Win(
		TitleBar(false),
		Radius(0),
		Size(friendContextMenuWidth, float32(m.height())),
		Content(
			primitives.Box(
				rotheme.Button("1:1 Chat", func() {
					m.action = FriendsWindowAction{Kind: FriendsWindowActionFriendWhisper, Friend: m.friend}
					m.Close()
				}).
					Width(friendContextMenuWidth).
					Height(friendContextMenuRowH),
				rotheme.Button("Delete Friend", func() {
					m.action = FriendsWindowAction{Kind: FriendsWindowActionFriendDelete, Friend: m.friend}
					m.Close()
				}).
					Width(friendContextMenuWidth).
					Height(friendContextMenuRowH),
				rotheme.Button("Block Whisper", func() {
					m.action = FriendsWindowAction{Kind: FriendsWindowActionFriendBlockWhisper, Friend: m.friend}
					m.Close()
				}).
					Width(friendContextMenuWidth).
					Height(friendContextMenuRowH),
			),
		),
	)
}

func (m *FriendContextMenu) height() int {
	return friendContextMenuRowH * 3
}
