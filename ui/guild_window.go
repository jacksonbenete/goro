package ui

import (
	"fmt"
	"image"
	"strings"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	guildWindowWidth     = 400
	guildWindowHeight    = 317
	guildWindowTabHeight = 23
	guildWindowTabWidth  = 64
	guildEmblemSize      = 24
)

type GuildWindow struct {
	Window
	tab         guildWindowTab
	snapshot    string
	EmblemImage func(Context) image.Image
}

type guildWindowTab int

const (
	guildWindowTabInfo guildWindowTab = iota
	guildWindowTabMembers
	guildWindowTabPositions
	guildWindowTabSkills
	guildWindowTabHistory
	guildWindowTabNotice
)

type guildWindowTabDef struct {
	tab   guildWindowTab
	label string
}

var guildWindowTabs = []guildWindowTabDef{
	{tab: guildWindowTabInfo, label: "Info"},
	{tab: guildWindowTabMembers, label: "Members"},
	{tab: guildWindowTabPositions, label: "Position"},
	{tab: guildWindowTabSkills, label: "Skill"},
	{tab: guildWindowTabHistory, label: "History"},
	{tab: guildWindowTabNotice, label: "Notice"},
}

func (w *GuildWindow) Toggle(ctx Context) {
	if w.IsOpen() {
		w.Close()
		return
	}
	w.OpenWindow(ctx)
}

func (w *GuildWindow) OpenWindow(ctx Context) {
	w.EnsureWindow(guildWindowWidth, guildWindowHeight)
	w.ctx = ctx
	if w.tab == 0 {
		w.tab = guildWindowTabInfo
	}
	w.snapshot = guildWindowSnapshot(ctx.Session)
	w.Open(ctx, w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *GuildWindow) Update(ctx Context) bool {
	w.EnsureWindow(guildWindowWidth, guildWindowHeight)
	w.ctx = ctx
	if !w.IsOpen() {
		return false
	}
	nextSnapshot := guildWindowSnapshot(ctx.Session)
	if nextSnapshot != w.snapshot {
		w.snapshot = nextSnapshot
		w.SetContent(w.widgetTree(ctx))
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *GuildWindow) Rebind(ctx Context) {
	w.EnsureWindow(guildWindowWidth, guildWindowHeight)
	if !w.IsOpen() {
		return
	}
	w.ctx = ctx
	w.snapshot = guildWindowSnapshot(ctx.Session)
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *GuildWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Guild"),
		CloseButton(true),
		OnClose(w.Close),
		Size(guildWindowWidth, guildWindowHeight),
		Content(
			primitives.Box(
				w.tabStrip(),
				primitives.Expanded(w.tabContent(ctx)),
			).
				CrossAlign(primitives.CrossAxisStretch),
		),
	)
}

func (w *GuildWindow) tabStrip() widget.Widget {
	tabs := make([]widget.Widget, 0, len(guildWindowTabs)+1)
	for _, def := range guildWindowTabs {
		def := def
		tabs = append(tabs,
			newTabWidget(tabWidgetConfig{
				label:      def.label,
				active:     w.tab == def.tab,
				width:      guildWindowTabWidth,
				height:     guildWindowTabHeight,
				blendEdge:  tabBlendBottom,
				blendInset: 1,
				onClick: func() {
					w.tab = def.tab
					w.refresh(w.ctx)
				},
			}),
		)
	}
	tabs = append(tabs, primitives.Expanded(primitives.Box()))
	return primitives.HBox(tabs...).
		Gap(-1).
		CrossAlign(primitives.CrossAxisStretch).
		Background(rotheme.Default.Colors.FooterLine)
}

func (w *GuildWindow) tabContent(ctx Context) widget.Widget {
	switch w.tab {
	case guildWindowTabInfo:
		return w.infoTab(ctx)
	case guildWindowTabMembers:
		return guildWindowPlaceholder("Guild member information is not loaded yet.")
	case guildWindowTabPositions:
		return guildWindowPlaceholder("Guild positions are not loaded yet.")
	case guildWindowTabSkills:
		return guildWindowPlaceholder("Guild skills are not loaded yet.")
	case guildWindowTabHistory:
		return guildWindowPlaceholder("Expel history is not loaded yet.")
	case guildWindowTabNotice:
		return guildWindowPlaceholder("Guild notice is not loaded yet.")
	default:
		return guildWindowPlaceholder("")
	}
}

func (w *GuildWindow) infoTab(ctx Context) widget.Widget {
	guild := guildSessionInfo(ctx.Session)
	leftRows := []widget.Widget{
		guildInfoRow("Guild Name", guild.Name),
		guildInfoRow("Guild lvl", guildNumber(guild.Level)),
		guildInfoRow("Guild Master", guildText(guild.MasterName)),
		guildInfoRow("Guildsmen", guildMembers(guild.UserNum, guild.MaxUserNum)),
		guildInfoRow("Avg.lvl of Guildsmen", guildNumber(guild.UserAverageLevel)),
		guildInfoRow("Territory", guildText(guild.ManageLand)),
		guildInfoSection("Tendency"),
		guildTendencyBox(guild.Honor, guild.Virtue),
	}
	rightRows := []widget.Widget{
		guildInfoRow("EXP", guildExp(guild.Exp, guild.MaxExp)),
		guildInfoSection("Emblem"),
		w.guildEmblemBox(ctx, guild.EmblemVersion),
		guildInfoRow("Tax Point", guildNumberAllowZero(guild.Point)),
		guildInfoSection("Alliance"),
		guildListBox(""),
		guildInfoSection("Antagonist"),
		guildListBox(""),
	}
	return primitives.HBox(
		primitives.Box(leftRows...).
			Gap(3).
			Width(188),
		primitives.Box(rightRows...).
			Gap(3).
			Width(174),
	).
		PaddingXY(9, 10).
		Gap(10).
		CrossAlign(primitives.CrossAxisStart)
}

func guildInfoRow(label, value string) widget.Widget {
	return primitives.Box(
		rotheme.Text(label + " : " + value),
	).
		Height(16)
}

func guildInfoSection(label string) widget.Widget {
	return primitives.Box(rotheme.Text(label + ":")).
		Height(16)
}

func guildTendencyBox(honor, virtue uint32) widget.Widget {
	value := ""
	if honor != 0 || virtue != 0 {
		value = fmt.Sprintf("H %d  V %d", honor, virtue)
	}
	return primitives.Box(
		primitives.Box(rotheme.Text("R")).Width(16),
		primitives.HBox(
			primitives.Box(
				primitives.Expanded(primitives.Box()),
				rotheme.Text("V"),
				primitives.Expanded(primitives.Box()),
			).
				PaddingLeft(3).
				Width(16).
				Height(90).
				CrossAlign(primitives.CrossAxisCenter),
			guildFramedBox(90, 90, value),
			primitives.Box(
				primitives.Expanded(primitives.Box()),
				rotheme.Text("F"),
				primitives.Expanded(primitives.Box()),
			).
				Width(16).
				Height(90).
				CrossAlign(primitives.CrossAxisCenter),
		).
			Gap(4).
			CrossAlign(primitives.CrossAxisCenter),
		primitives.Box(rotheme.Text("W")).
			PaddingTop(3).
			Width(16).
			Height(19),
	).
		CrossAlign(primitives.CrossAxisCenter).
		Gap(2)
}

func (w *GuildWindow) guildEmblemBox(ctx Context, version uint32) widget.Widget {
	var content widget.Widget
	if w.EmblemImage != nil {
		if img := w.EmblemImage(ctx); img != nil {
			content = newStaticImageWidget(img, guildEmblemSize, guildEmblemSize)
		}
	}
	if content != nil {
		return guildFramedWidget(guildEmblemSize, guildEmblemSize, 0, content)
	}
	text := "-"
	if version != 0 {
		text = fmt.Sprintf("v%d", version)
	}
	return guildFramedBox(guildEmblemSize, guildEmblemSize, text)
}

func guildListBox(text string) widget.Widget {
	return guildFramedBox(168, 42, text)
}

func guildFramedBox(width, height float32, text string) widget.Widget {
	content := widget.Widget(rotheme.Text(text))
	if strings.TrimSpace(text) == "" {
		content = primitives.Box()
	}
	return guildFramedWidget(width, height, 4, content)
}

func guildFramedWidget(width, height, padding float32, content widget.Widget) widget.Widget {
	return primitives.Box(content).
		Width(width).
		Height(height).
		Padding(padding).
		Background(rotheme.Default.Colors.WindowFooter).
		BorderStyle(1, rotheme.Default.Colors.FooterLine)
}

func guildWindowPlaceholder(text string) widget.Widget {
	if strings.TrimSpace(text) == "" {
		text = "Not loaded yet."
	}
	return primitives.Box(
		rotheme.Text(text),
	).
		Padding(10).
		Background(rotheme.Default.Colors.WindowBody)
}

func (w *GuildWindow) refresh(ctx Context) {
	w.snapshot = guildWindowSnapshot(ctx.Session)
	w.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *GuildWindow) Refresh(ctx Context) {
	if !w.IsOpen() {
		return
	}
	w.refresh(ctx)
}

func guildWindowSnapshot(s *session.Session) string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("%d|%d|%s|%d|%d|%d|%d|%d|%d|%d|%d|%d|%s|%s|%s|%d",
		s.GuildID,
		s.EmblemVersion,
		s.GuildName,
		s.Guild.Level,
		s.Guild.UserNum,
		s.Guild.MaxUserNum,
		s.Guild.UserAverageLevel,
		s.Guild.Exp,
		s.Guild.MaxExp,
		s.Guild.Point,
		s.Guild.Honor,
		s.Guild.Virtue,
		s.Guild.MasterName,
		s.Guild.ManageLand,
		s.Guild.Name,
		s.Guild.Zeny,
	)
}

func guildSessionInfo(s *session.Session) session.Guild {
	if s == nil {
		return session.Guild{Name: "-"}
	}
	guild := s.Guild
	if guild.ID == 0 {
		guild.ID = s.GuildID
	}
	if guild.EmblemVersion == 0 {
		guild.EmblemVersion = s.EmblemVersion
	}
	if strings.TrimSpace(guild.Name) == "" {
		guild.Name = strings.TrimSpace(s.GuildName)
	}
	if strings.TrimSpace(guild.Name) == "" {
		guild.Name = "-"
	}
	return guild
}

func guildText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "-"
	}
	return text
}

func guildNumber(value uint32) string {
	if value == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", value)
}

func guildNumberAllowZero(value uint32) string {
	return fmt.Sprintf("%d", value)
}

func guildMembers(current, max uint32) string {
	if current == 0 && max == 0 {
		return "- / -"
	}
	return fmt.Sprintf("%d / %d", current, max)
}

func guildExp(current, max uint32) string {
	if current == 0 && max == 0 {
		return "-"
	}
	if max == 0 {
		return fmt.Sprintf("%d", current)
	}
	return fmt.Sprintf("%d / %d", current, max)
}
