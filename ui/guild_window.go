package ui

import (
	"fmt"
	"image"
	"math"
	"sort"
	"strings"

	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/res"
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
	tab           guildWindowTab
	snapshot      string
	action        GuildWindowAction
	EmblemImage   func(Context) image.Image
	emblemOptions []GuildEmblemOption
	skillIcons    map[uint16]image.Image
	skillMiss     map[uint16]struct{}
}

type GuildWindowAction struct {
	RequestMenu        bool
	MenuTab            uint32
	SelectedEmblemPath string
}

type GuildEmblemOption struct {
	Label string
	Path  string
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

func (w *GuildWindow) PopAction() GuildWindowAction {
	action := w.action
	w.action = GuildWindowAction{}
	return action
}

func (w *GuildWindow) SetEmblemOptions(ctx Context, options []GuildEmblemOption) {
	w.emblemOptions = append(w.emblemOptions[:0], options...)
	if w.IsOpen() {
		w.refresh(ctx)
	}
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
					w.action = GuildWindowAction{RequestMenu: true, MenuTab: uint32(def.tab)}
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
		return w.membersTab(ctx)
	case guildWindowTabPositions:
		return w.positionsTab(ctx)
	case guildWindowTabSkills:
		return w.skillsTab(ctx)
	case guildWindowTabHistory:
		return w.historyTab(ctx)
	case guildWindowTabNotice:
		return w.noticeTab(ctx)
	default:
		return guildWindowPlaceholder("")
	}
}

func (w *GuildWindow) membersTab(ctx Context) widget.Widget {
	guild := guildSessionInfo(ctx.Session)
	members := guild.Members
	rows := make([]widget.Widget, 0, len(members)+1)
	rows = append(rows, guildMemberHeaderRow())
	totalExp := guildMembersTotalExp(members)
	if len(members) == 0 {
		rows = append(rows,
			primitives.Box(rotheme.Text("No guild members loaded.")).
				Height(24).
				PaddingXY(4, 4).
				Background(rotheme.Default.Colors.WindowBody),
		)
	}
	for i, member := range members {
		rows = append(rows, guildMemberRow(member, guild.Positions, totalExp, i%2 == 0))
	}
	return primitives.Box(
		scrollview.New(
			primitives.Box(rows...),
			scrollview.DirectionOpt(scrollview.Vertical),
			scrollview.ScrollbarOpt(scrollview.ScrollbarAuto),
			scrollview.ScrollStep(20),
		),
	).
		PaddingXY(7, 7).
		Background(rotheme.Default.Colors.WindowBody)
}

func guildMemberHeaderRow() widget.Widget {
	return primitives.HBox(
		guildMemberCell("Name", 78, true, false),
		guildMemberCell("Position", 62, true, false),
		guildMemberCell("Job", 52, true, false),
		guildMemberCell("Lv", 26, true, false),
		guildMemberCell("Note", 62, true, false),
		guildMemberCell("Dev.", 38, true, false),
		guildMemberCell("Tax", 42, true, false),
	).Height(20)
}

func guildMemberRow(member session.GuildMember, positions []session.GuildPosition, totalExp uint32, dark bool) widget.Widget {
	return primitives.HBox(
		guildMemberCell(guildText(member.CharName), 78, false, dark),
		guildMemberCell(guildPositionName(positions, member.PositionID), 62, false, dark),
		guildMemberCell(db.JobDisplayName(int(member.Job)), 52, false, dark),
		guildMemberCell(fmt.Sprintf("%d", member.Level), 26, false, dark),
		guildMemberCell(member.Memo, 62, false, dark),
		guildMemberCell(guildMemberDevotion(member.MemberExp, totalExp), 38, false, dark),
		guildMemberCell(fmt.Sprintf("%d", member.MemberExp), 42, false, dark),
	).Height(20)
}

func guildMemberCell(text string, width float32, header, dark bool) widget.Widget {
	text = trimRunes(strings.TrimSpace(text), int(width/7))
	bg := rotheme.Default.Colors.WindowBody
	if header {
		bg = rotheme.Default.Colors.WindowTitle
	} else if dark {
		bg = rotheme.Default.Colors.PanelBody
	}
	return primitives.HBox(rotheme.Text(text).Align(widget.TextAlignLeft)).
		Width(width).
		Height(20).
		PaddingLeft(4).
		CrossAlign(primitives.CrossAxisCenter).
		Background(bg)
}

func guildMembersTotalExp(members []session.GuildMember) uint32 {
	var total uint32
	for _, member := range members {
		total += member.MemberExp
	}
	return total
}

func guildMemberDevotion(memberExp, totalExp uint32) string {
	if memberExp == 0 || totalExp == 0 {
		return "0 %"
	}
	return fmt.Sprintf("%d %%", int(math.Round(float64(memberExp)/float64(totalExp)*100)))
}

func (w *GuildWindow) positionsTab(ctx Context) widget.Widget {
	positions := guildSortedPositions(guildSessionInfo(ctx.Session).Positions)
	rows := make([]widget.Widget, 0, len(positions)+1)
	rows = append(rows, guildPositionHeaderRow())
	if len(positions) == 0 {
		rows = append(rows,
			primitives.Box(rotheme.Text("No guild positions loaded.")).
				Height(24).
				PaddingXY(4, 4).
				Background(rotheme.Default.Colors.WindowBody),
		)
	}
	for i, position := range positions {
		rows = append(rows, guildPositionRow(position, i%2 == 0))
	}
	return primitives.Box(
		scrollview.New(
			primitives.Box(rows...),
			scrollview.DirectionOpt(scrollview.Vertical),
			scrollview.ScrollbarOpt(scrollview.ScrollbarAuto),
			scrollview.ScrollStep(20),
		),
	).
		PaddingXY(7, 7).
		Background(rotheme.Default.Colors.WindowBody)
}

func guildPositionHeaderRow() widget.Widget {
	return primitives.HBox(
		guildMemberCell("Rank", 44, true, false),
		guildMemberCell("Position Title", 164, true, false),
		guildMemberCell("Invitation", 58, true, false),
		guildMemberCell("Punish", 48, true, false),
		guildMemberCell("Tax", 46, true, false),
	).Height(20)
}

func guildPositionRow(position session.GuildPosition, dark bool) widget.Widget {
	return primitives.HBox(
		guildMemberCell(fmt.Sprintf("%d", position.PositionID), 44, false, dark),
		guildMemberCell(guildPositionTitle(position), 164, false, dark),
		guildMemberCell(guildRightLabel(position.Right&0x01 != 0), 58, false, dark),
		guildMemberCell(guildRightLabel(position.Right&0x10 != 0), 48, false, dark),
		guildMemberCell(fmt.Sprintf("%d %%", position.PayRate), 46, false, dark),
	).Height(20)
}

func guildSortedPositions(positions []session.GuildPosition) []session.GuildPosition {
	out := append([]session.GuildPosition(nil), positions...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].PositionID < out[j].PositionID
	})
	return out
}

func guildPositionName(positions []session.GuildPosition, id uint32) string {
	for _, position := range positions {
		if position.PositionID == id {
			return guildPositionTitle(position)
		}
	}
	return fmt.Sprintf("Position %d", id)
}

func guildPositionTitle(position session.GuildPosition) string {
	if title := strings.TrimSpace(position.PosName); title != "" {
		return title
	}
	return fmt.Sprintf("Position %d", position.PositionID)
}

func guildRightLabel(enabled bool) string {
	if enabled {
		return "Yes"
	}
	return "-"
}

func (w *GuildWindow) skillsTab(ctx Context) widget.Widget {
	guild := guildSessionInfo(ctx.Session)
	rows := make([]widget.Widget, 0, len(guild.Skills))
	if len(guild.Skills) == 0 {
		rows = append(rows,
			primitives.Box(rotheme.Text("No guild skills loaded.").Color(rotheme.Default.Colors.MutedText)).
				Height(32).
				PaddingXY(6, 6).
				Background(rotheme.Default.Colors.WindowBody),
		)
	}
	for i, skill := range guild.Skills {
		rows = append(rows, w.guildSkillRow(ctx, skill, i%2 == 0))
	}
	return primitives.Box(
		primitives.Expanded(
			primitives.Box(
				scrollview.New(
					primitives.Box(rows...),
					scrollview.DirectionOpt(scrollview.Vertical),
					scrollview.ScrollbarOpt(scrollview.ScrollbarAuto),
					scrollview.ScrollStep(32),
				),
			).
				PaddingXY(7, 7),
		),
		primitives.HBox(
			rotheme.Text(fmt.Sprintf("Skill Points: %d", guild.SkillPoints)),
			primitives.Expanded(primitives.Box()),
		).
			Height(28).
			PaddingXY(9, 6).
			CrossAlign(primitives.CrossAxisCenter).
			Background(rotheme.Default.Colors.WindowFooter),
	).
		CrossAlign(primitives.CrossAxisStretch).
		Background(rotheme.Default.Colors.WindowBody)
}

func (w *GuildWindow) guildSkillRow(ctx Context, skill session.Skill, dark bool) widget.Widget {
	bg := rotheme.Default.Colors.WindowBody
	if dark {
		bg = rotheme.Default.Colors.PanelBody
	}
	return primitives.HBox(
		primitives.Box(w.guildSkillIcon(ctx, skill)).
			Width(32).
			Height(32).
			CrossAlign(primitives.CrossAxisCenter),
		primitives.Box(
			rotheme.Text(trimRunes(skillDisplayName(ctx.Resources, skill), 28)).
				Align(widget.TextAlignLeft),
			rotheme.Text(guildSkillLevelText(skill)).
				Color(rotheme.Default.Colors.MutedText).
				Align(widget.TextAlignLeft),
		).
			Width(160).
			Height(32).
			CrossAlign(primitives.CrossAxisStretch),
		guildMemberCell(guildSkillKind(skill), 62, false, dark),
		guildMemberCell(guildSkillSP(skill), 54, false, dark),
		guildMemberCell(guildSkillRange(skill), 52, false, dark),
	).
		Height(32).
		CrossAlign(primitives.CrossAxisCenter).
		Background(bg)
}

func (w *GuildWindow) guildSkillIcon(ctx Context, skill session.Skill) widget.Widget {
	if img := w.guildSkillIconImage(ctx, skill); img != nil {
		return newStaticImageWidget(img, 24, 24)
	}
	return primitives.Box()
}

func (w *GuildWindow) guildSkillIconImage(ctx Context, skill session.Skill) image.Image {
	if ctx.Resources == nil || skill.ID == 0 {
		return nil
	}
	if w.skillIcons == nil {
		w.skillIcons = make(map[uint16]image.Image)
	}
	if img := w.skillIcons[skill.ID]; img != nil {
		return img
	}
	if w.skillMiss != nil {
		if _, ok := w.skillMiss[skill.ID]; ok {
			return nil
		}
	}
	resourceName, ok := ctx.Resources.SkillResourceName(int(skill.ID))
	if !ok {
		resourceName = strings.ToUpper(strings.ReplaceAll(skillLabel(skill), " ", "_"))
	}
	img, _, err := res.LoadImage(ctx.Resources, res.SkillIconTextureCandidates(resourceName, int(skill.ID)))
	if err != nil {
		if w.skillMiss == nil {
			w.skillMiss = make(map[uint16]struct{})
		}
		w.skillMiss[skill.ID] = struct{}{}
		return nil
	}
	w.skillIcons[skill.ID] = img
	return img
}

func guildSkillLevelText(skill session.Skill) string {
	if skill.MaxLevel > 0 {
		return fmt.Sprintf("Lv : %d / %d", skill.Level, skill.MaxLevel)
	}
	return fmt.Sprintf("Lv : %d", skill.Level)
}

func guildSkillKind(skill session.Skill) string {
	if skill.Type == 0 {
		return "Passive"
	}
	return "Active"
}

func guildSkillSP(skill session.Skill) string {
	if skill.Type == 0 {
		return ""
	}
	return fmt.Sprintf("SP : %d", skill.SPCost)
}

func guildSkillRange(skill session.Skill) string {
	if skill.Range <= 0 {
		return ""
	}
	return fmt.Sprintf("Range : %d", skill.Range)
}

func (w *GuildWindow) historyTab(ctx Context) widget.Widget {
	history := guildSessionInfo(ctx.Session).ExpelHistory
	rows := make([]widget.Widget, 0, len(history)+1)
	rows = append(rows, guildHistoryHeaderRow())
	if len(history) == 0 {
		rows = append(rows,
			primitives.Box(rotheme.Text("No expel history loaded.").Color(rotheme.Default.Colors.MutedText)).
				Height(24).
				PaddingXY(4, 4).
				Background(rotheme.Default.Colors.WindowBody),
		)
	}
	for i, entry := range history {
		rows = append(rows, guildHistoryRow(entry, i%2 == 0))
	}
	return primitives.Box(
		scrollview.New(
			primitives.Box(rows...),
			scrollview.DirectionOpt(scrollview.Vertical),
			scrollview.ScrollbarOpt(scrollview.ScrollbarAuto),
			scrollview.ScrollStep(20),
		),
	).
		PaddingXY(7, 7).
		Background(rotheme.Default.Colors.WindowBody)
}

func guildHistoryHeaderRow() widget.Widget {
	return primitives.HBox(
		guildMemberCell("Name", 116, true, false),
		guildMemberCell("The Reason of Expulsion", 244, true, false),
	).Height(20)
}

func guildHistoryRow(entry session.GuildExpelHistory, dark bool) widget.Widget {
	return primitives.HBox(
		guildMemberCell(guildText(entry.CharName), 116, false, dark),
		guildMemberCell(guildText(entry.Reason), 244, false, dark),
	).Height(20)
}

func (w *GuildWindow) noticeTab(ctx Context) widget.Widget {
	guild := guildSessionInfo(ctx.Session)
	return primitives.Box(
		rotheme.Text("Title"),
		guildNoticeBox(guildText(guild.NoticeSubject), 28, 1),
		rotheme.Text("Contents"),
		guildNoticeBox(guildText(guild.Notice), 140, 8),
	).
		PaddingXY(9, 10).
		Gap(5).
		CrossAlign(primitives.CrossAxisStretch).
		Background(rotheme.Default.Colors.WindowBody)
}

func guildNoticeBox(text string, height float32, maxLines int) widget.Widget {
	return primitives.Box(
		rotheme.Text(text).
			Align(widget.TextAlignLeft).
			MaxLines(maxLines).
			LineHeight(16/rotheme.Default.Typography.TextSize),
	).
		Height(height).
		PaddingXY(5, 4).
		CrossAlign(primitives.CrossAxisStretch).
		Background(rotheme.Default.Colors.WindowFooter).
		BorderStyle(1, rotheme.Default.Colors.FooterLine)
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
		w.guildEmblemEditor(ctx, guild.EmblemVersion),
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

func (w *GuildWindow) guildEmblemEditor(ctx Context, version uint32) widget.Widget {
	return primitives.HBox(
		w.guildEmblemBox(ctx, version),
		primitives.Box(w.guildEmblemDropdown()).
			Width(132).
			Height(22).
			CrossAlign(primitives.CrossAxisStretch),
	).
		Gap(6).
		CrossAlign(primitives.CrossAxisCenter)
}

func (w *GuildWindow) guildEmblemDropdown() widget.Widget {
	if len(w.emblemOptions) == 0 {
		return dropdown.New(
			dropdown.ItemDefs([]dropdown.ItemDef{{Value: "", Label: "No emblems found", Disabled: true}}),
			dropdown.Selected(0),
			dropdown.Disabled(true),
			dropdown.MaxVisibleItems(5),
			dropdown.PainterOpt(rotheme.DropdownPainter{}),
		)
	}
	items := make([]dropdown.ItemDef, 0, len(w.emblemOptions))
	for _, option := range w.emblemOptions {
		label := strings.TrimSpace(option.Label)
		if label == "" {
			label = "Emblem"
		}
		items = append(items, dropdown.ItemDef{
			Value: option.Path,
			Label: trimRunes(label, 20),
		})
	}
	return dropdown.New(
		dropdown.ItemDefs(items),
		dropdown.Placeholder("Edit"),
		dropdown.MaxVisibleItems(5),
		dropdown.PainterOpt(rotheme.DropdownPainter{}),
		dropdown.OnChange(func(_ int, value string) {
			w.action.SelectedEmblemPath = value
		}),
	)
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
	memberSnapshot := strings.Builder{}
	for _, member := range s.Guild.Members {
		fmt.Fprintf(&memberSnapshot, "|%d:%d:%d:%d:%d:%d:%d:%s:%s",
			member.AccountID,
			member.CharID,
			member.Job,
			member.Level,
			member.MemberExp,
			member.CurrentState,
			member.PositionID,
			member.CharName,
			member.Memo,
		)
	}
	positionSnapshot := strings.Builder{}
	for _, position := range s.Guild.Positions {
		fmt.Fprintf(&positionSnapshot, "|%d:%d:%d:%d:%s",
			position.PositionID,
			position.Right,
			position.Ranking,
			position.PayRate,
			position.PosName,
		)
	}
	skillSnapshot := strings.Builder{}
	for _, skill := range s.Guild.Skills {
		fmt.Fprintf(&skillSnapshot, "|%d:%d:%d:%d:%d:%s",
			skill.ID,
			skill.Type,
			skill.Level,
			skill.SPCost,
			skill.Range,
			skill.Name,
		)
	}
	historySnapshot := strings.Builder{}
	for _, entry := range s.Guild.ExpelHistory {
		fmt.Fprintf(&historySnapshot, "|%s:%s:%s",
			entry.CharName,
			entry.Account,
			entry.Reason,
		)
	}
	return fmt.Sprintf("%d|%d|%s|%d|%d|%d|%d|%d|%d|%d|%d|%d|%s|%s|%s|%d|%d|%s|%s%s%s%s%s",
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
		s.Guild.SkillPoints,
		s.Guild.NoticeSubject,
		s.Guild.Notice,
		memberSnapshot.String(),
		positionSnapshot.String(),
		skillSnapshot.String(),
		historySnapshot.String(),
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
