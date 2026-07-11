package ui

import (
	"fmt"
	"strings"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	friendsWindowWidth  = 286
	friendsWindowHeight = 256
	friendsRowHeight    = 24
	friendsTabWidth     = 72
	friendsTabHeight    = 24
	friendsListMax      = 40
	partyFooterH        = 42
)

type FriendsWindow struct {
	ContextWindowHandle
	snapshot string
	tab      friendsWindowTab
	action   FriendsWindowAction
}

type friendsWindowTab int

const (
	friendsWindowTabFriends friendsWindowTab = iota
	friendsWindowTabParty
)

type FriendsWindowAction int

const (
	FriendsWindowActionNone FriendsWindowAction = iota
	FriendsWindowActionPartySettings
	FriendsWindowActionPartyLeave
)

func (w *FriendsWindow) Toggle(ctx Context) {
	if w.IsOpen() {
		w.Close()
		return
	}
	w.OpenWindow(ctx)
}

func (w *FriendsWindow) OpenWindow(ctx Context) {
	w.ensureWindow()
	w.ctx = ctx
	w.snapshot = friendsWindowSnapshot(ctx.Session)
	w.tab = friendsWindowTabFriends
	w.window.Open(ctx, w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *FriendsWindow) Update(ctx Context) bool {
	w.ensureWindow()
	w.ctx = ctx
	if !w.window.IsOpen() {
		return false
	}
	nextSnapshot := friendsWindowSnapshot(ctx.Session)
	if nextSnapshot != w.snapshot {
		w.snapshot = nextSnapshot
		w.window.SetContent(w.widgetTree(ctx))
	}
	consumed := w.window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *FriendsWindow) Rebind(ctx Context) {
	w.ensureWindow()
	if !w.window.IsOpen() {
		return
	}
	w.ctx = ctx
	w.snapshot = friendsWindowSnapshot(ctx.Session)
	w.window.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *FriendsWindow) PopAction() FriendsWindowAction {
	action := w.action
	w.action = FriendsWindowActionNone
	return action
}

func (w *FriendsWindow) ensureWindow() {
	if w.window.width == 0 {
		w.window = NewWindowState(friendsWindowWidth, friendsWindowHeight)
	}
}

func (w *FriendsWindow) widgetTree(ctx Context) widget.Widget {
	friends := sessionFriends(ctx.Session)
	party := sessionParty(ctx.Session)
	title := fmt.Sprintf("Friends (%d/%d)", len(friends), friendsListMax)
	content := friendsList(friends)
	var footer widget.Widget
	var footerHeight float32
	if w.tab == friendsWindowTabParty {
		title = partyWindowTitle(party)
		content = partyList(party)
		footer = w.partyFooter(party)
		footerHeight = partyFooterH
	}
	return Window(
		Title(title),
		CloseButton(true),
		OnClose(w.Close),
		Size(friendsWindowWidth, friendsWindowHeight),
		Content(
			primitives.Box(
				w.friendsTabs(),
				content,
			).
				Gap(-1),
		),
		Footer(footer),
		FooterHeight(footerHeight),
	)
}

func (w *FriendsWindow) friendsTabs() widget.Widget {
	return primitives.HBox(
		newTabWidget(tabWidgetConfig{
			label:      "Friends",
			active:     w.tab == friendsWindowTabFriends,
			width:      friendsTabWidth,
			height:     friendsTabHeight,
			blendEdge:  tabBlendBottom,
			blendInset: 1,
			onClick: func() {
				w.tab = friendsWindowTabFriends
				w.refresh(w.ctx)
			},
		}),
		newTabWidget(tabWidgetConfig{
			label:      "Party",
			active:     w.tab == friendsWindowTabParty,
			width:      friendsTabWidth,
			height:     friendsTabHeight,
			blendEdge:  tabBlendBottom,
			blendInset: 1,
			onClick: func() {
				w.tab = friendsWindowTabParty
				w.refresh(w.ctx)
			},
		}),
		primitives.Expanded(primitives.Box()),
	).Gap(-1)
}

func (w *FriendsWindow) refresh(ctx Context) {
	w.snapshot = friendsWindowSnapshot(ctx.Session)
	w.window.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *FriendsWindow) partyFooter(party session.Party) widget.Widget {
	return primitives.HBox(
		rotheme.ButtonDisabled("Settings", !party.Active(), func() {
			w.action = FriendsWindowActionPartySettings
		}).Width(82),
		rotheme.ButtonDisabled("Leave", !party.Active(), func() {
			w.action = FriendsWindowActionPartyLeave
		}).Width(66),
	).
		Gap(8)
}

func friendsList(friends []session.Friend) widget.Widget {
	rows := make([]widget.Widget, 0, maxInt(1, len(friends)))
	if len(friends) == 0 {
		rows = append(rows,
			primitives.Box(
				rotheme.Text("No friends").
					Color(rotheme.Default.Colors.MutedText).
					Align(primitives.TextAlignCenter),
			).
				Height(friendsRowHeight).
				CrossAlign(primitives.CrossAxisStretch),
		)
	} else {
		for i, friend := range friends {
			rows = append(rows, friendRow(friend, i))
		}
	}
	return primitives.Box(rows...).
		BorderStyle(1, rotheme.Default.Colors.WindowBorder).
		CrossAlign(primitives.CrossAxisStretch)
}

func friendRow(friend session.Friend, index int) widget.Widget {
	state := "Offline"
	stateColor := rotheme.Default.Colors.MutedText
	if friend.Online() {
		state = "Online"
		stateColor = Color(GoodTextColor)
	}
	bg := rotheme.Default.Colors.WindowBody
	if index%2 == 0 {
		bg = Color(PanelAltColor)
	}
	name := strings.TrimSpace(friend.Name)
	if name == "" {
		name = "Unknown"
	}
	return primitives.HBox(
		rotheme.Text(trimRunes(name, 24)).
			Align(primitives.TextAlignStart),
		primitives.Expanded(primitives.Box()),
		rotheme.Text(state).
			Color(stateColor).
			Align(primitives.TextAlignEnd),
	).
		PaddingXY(8, 0).
		Height(friendsRowHeight).
		Background(bg).
		CrossAlign(primitives.CrossAxisCenter)
}

func partyWindowTitle(party session.Party) string {
	if !party.Active() {
		return "Party"
	}
	name := strings.TrimSpace(party.Name)
	if name == "" {
		name = "Party"
	}
	return name
}

func partyList(party session.Party) widget.Widget {
	members := party.Members
	rows := make([]widget.Widget, 0, maxInt(1, len(members)))
	if len(members) == 0 {
		rows = append(rows,
			primitives.Box(
				rotheme.Text("No party").
					Color(rotheme.Default.Colors.MutedText).
					Align(primitives.TextAlignCenter),
			).
				Height(friendsRowHeight).
				CrossAlign(primitives.CrossAxisStretch),
		)
	} else {
		for i, member := range members {
			rows = append(rows, partyRow(member, i))
		}
	}
	return primitives.Box(rows...).
		BorderStyle(1, rotheme.Default.Colors.WindowBorder).
		CrossAlign(primitives.CrossAxisStretch)
}

func partyRow(member session.PartyMember, index int) widget.Widget {
	bg := rotheme.Default.Colors.WindowBody
	if index%2 == 0 {
		bg = Color(PanelAltColor)
	}
	name := strings.TrimSpace(member.Name)
	if name == "" {
		name = "Player"
	}
	if member.Leader() {
		name += " *"
	}
	state := "Offline"
	stateColor := rotheme.Default.Colors.MutedText
	if member.Online() {
		state = trimRunes(member.MapName, 12)
		if state == "" {
			state = "Online"
		}
		stateColor = Color(GoodTextColor)
	}
	return primitives.Box(
		primitives.HBox(
			rotheme.Text(trimRunes(name, 20)).
				Align(primitives.TextAlignStart),
			primitives.Expanded(primitives.Box()),
			rotheme.Text(state).
				Color(stateColor).
				Align(primitives.TextAlignEnd),
		).
			Height(13).
			CrossAlign(primitives.CrossAxisCenter),
		newPartyHPBar(member.HP, member.MaxHP),
	).
		PaddingXY(8, 2).
		Height(friendsRowHeight).
		Background(bg).
		CrossAlign(primitives.CrossAxisStretch)
}

type partyHPBar struct {
	widget.WidgetBase
	hp    int
	maxHP int
}

func newPartyHPBar(hp, maxHP int) *partyHPBar {
	w := &partyHPBar{hp: hp, maxHP: maxHP}
	w.SetVisible(true)
	w.SetEnabled(false)
	return w
}

func (w *partyHPBar) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(geometry.Sz(120, 5))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *partyHPBar) Draw(_ widget.Context, canvas widget.Canvas) {
	bounds := w.Bounds()
	canvas.DrawRect(bounds, widget.RGBA8(224, 232, 242, 255))
	fillW := float32(0)
	if w.maxHP > 0 && w.hp > 0 {
		fillW = bounds.Width() * float32(w.hp) / float32(w.maxHP)
		if fillW > bounds.Width() {
			fillW = bounds.Width()
		}
	}
	if fillW > 0 {
		canvas.DrawRect(geometry.NewRect(bounds.Min.X, bounds.Min.Y, fillW, bounds.Height()), Color(PlayerHPBarColor))
	}
	canvas.StrokeRect(bounds, rotheme.Default.Colors.WindowBorder, 1)
}

func (w *partyHPBar) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (w *partyHPBar) Children() []widget.Widget {
	return nil
}

func sessionFriends(s *session.Session) []session.Friend {
	if s == nil {
		return nil
	}
	return s.Friends.List
}

func sessionParty(s *session.Session) session.Party {
	if s == nil {
		return session.Party{}
	}
	return s.Party
}

func friendsWindowSnapshot(s *session.Session) string {
	friends := sessionFriends(s)
	party := sessionParty(s)
	var b strings.Builder
	fmt.Fprintf(&b, "friends=%d", len(friends))
	for _, friend := range friends {
		fmt.Fprintf(&b, ";%d:%d:%s:%d", friend.AccountID, friend.CharID, friend.Name, friend.State)
	}
	fmt.Fprintf(&b, ";party=%s:%d:%d", party.Name, party.ExpShare, len(party.Members))
	for _, member := range party.Members {
		fmt.Fprintf(&b, ";%d:%s:%s:%d:%d:%d:%d:%d:%t", member.AccountID, member.Name, member.MapName, member.Role, member.State, member.HP, member.MaxHP, member.X, member.Dead)
	}
	return b.String()
}
