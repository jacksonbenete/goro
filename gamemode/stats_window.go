package gamemode

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

const (
	statsWindowWidth  = 286
	statsWindowHeight = 338
	statsWindowTitleH = 28
	statsWindowPad    = 12
	statsRowH         = 22
	statsButtonSize   = 17
)

var (
	statsWindowTitleColor  = color.RGBA{R: 255, G: 230, B: 150, A: 255}
	statsWindowTextColor   = color.RGBA{R: 236, G: 232, B: 220, A: 255}
	statsWindowMutedColor  = color.RGBA{R: 166, G: 174, B: 184, A: 255}
	statsWindowGoodColor   = color.RGBA{R: 144, G: 210, B: 142, A: 255}
	statsWindowErrorColor  = color.RGBA{R: 255, G: 116, B: 116, A: 255}
	statsWindowButtonColor = color.RGBA{R: 56, G: 62, B: 72, A: 235}
	statsWindowHoverColor  = color.RGBA{R: 82, G: 92, B: 108, A: 245}
	statsWindowDownColor   = color.RGBA{R: 98, G: 106, B: 122, A: 245}
	statsWindowDisabled    = color.RGBA{R: 42, G: 46, B: 54, A: 205}
)

type statsWindowState struct {
	open       bool
	x          int
	y          int
	positioned bool
	dragging   bool
	dragDX     int
	dragDY     int
	status     string
	statusGood bool
	statusAt   time.Time
}

type statRow struct {
	label    string
	statusID uint16
	value    int
	bonus    int
	cost     int
}

func (w *statsWindowState) toggle(ctx Context) {
	if w.open {
		w.open = false
		w.dragging = false
		return
	}
	w.open = true
	w.ensurePosition(ctx)
}

func (w *statsWindowState) update(ctx Context) bool {
	if !w.open || ctx.Input == nil {
		return false
	}
	w.ensurePosition(ctx)
	width, height := ctx.ScreenSize()
	if w.dragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			w.x = clampStatsWindowInt(ctx.Input.MouseX-w.dragDX, 8, maxInt(8, width-statsWindowWidth-8))
			w.y = clampStatsWindowInt(ctx.Input.MouseY-w.dragDY, 8, maxInt(8, height-statsWindowHeight-8))
			return true
		}
		w.dragging = false
		return true
	}
	if ctx.Input.JustPressed(render.KeyEscape) {
		w.open = false
		return true
	}
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, statsWindowWidth, statsWindowHeight)
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	if !pointInRect(mx, my, w.x, w.y, statsWindowWidth, statsWindowHeight) {
		return false
	}
	cx, cy, cw, ch := w.closeBounds()
	if pointInRect(mx, my, cx, cy, cw, ch) {
		w.open = false
		return true
	}
	if pointInRect(mx, my, w.x, w.y, statsWindowWidth, statsWindowTitleH) {
		w.dragging = true
		w.dragDX = mx - w.x
		w.dragDY = my - w.y
		return true
	}
	for _, row := range statsRows(ctx.Session) {
		bx, by, bw, bh := w.statButtonBounds(row.statusID)
		if !pointInRect(mx, my, bx, by, bw, bh) {
			continue
		}
		if !canIncreaseStat(ctx.Session, row) {
			w.setStatus("Not enough status points", false)
			return true
		}
		if ctx.Network == nil {
			w.setStatus("Not connected", false)
			return true
		}
		if err := ctx.Network.SendStatusIncrease(row.statusID); err != nil {
			w.setStatus(err.Error(), false)
			return true
		}
		w.setStatus(fmt.Sprintf("%s increase requested", row.label), true)
		return true
	}
	return true
}

func (w *statsWindowState) draw(screen *render.Image, ctx Context) {
	if !w.open || screen == nil {
		return
	}
	w.ensurePosition(ctx)
	x, y := w.x, w.y
	drawNPCWindowFrame(screen, x, y, statsWindowWidth, statsWindowHeight)
	render.DebugPrintAtColor(screen, "Status", x+statsWindowPad, y+9, statsWindowTitleColor)
	cx, cy, cw, ch := w.closeBounds()
	drawUIButtonSurface(screen, cx, cy, cw, ch, statsWindowButtonColor)
	render.DebugPrintAtColor(screen, "x", cx+5, cy+(ch-13)/2-1, statsWindowTextColor)
	render.DrawRect(screen, float64(x+8), float64(y+statsWindowTitleH), float64(statsWindowWidth-16), 1, color.RGBA{R: 210, G: 200, B: 170, A: 80})

	stats := sessionStats(ctx.Session)
	render.DebugPrintAtColor(screen, fmt.Sprintf("Status Point : %d", stats.Points), x+statsWindowPad, y+statsWindowTitleH+10, statsWindowTextColor)
	render.DebugPrintAtColor(screen, "Stat", x+statsWindowPad, y+statsWindowTitleH+32, statsWindowMutedColor)
	render.DebugPrintAtColor(screen, "Value", x+72, y+statsWindowTitleH+32, statsWindowMutedColor)
	render.DebugPrintAtColor(screen, "Need", x+132, y+statsWindowTitleH+32, statsWindowMutedColor)

	mx, my := -1, -1
	mouseDown := false
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
		mouseDown = ctx.Input.MousePressed(render.MouseButtonLeft)
	}
	for i, row := range statsRows(ctx.Session) {
		ry := w.statRowY(i)
		render.DebugPrintAtColor(screen, row.label, x+statsWindowPad, ry+4, statsWindowTextColor)
		render.DebugPrintAtColor(screen, formatStatValue(row.value, row.bonus), x+72, ry+4, statsWindowTextColor)
		render.DebugPrintAtColor(screen, fmt.Sprintf("%d", statCost(row)), x+132, ry+4, statsWindowMutedColor)
		bx, by, bw, bh := w.statButtonBounds(row.statusID)
		fill := statsWindowButtonColor
		textColor := statsWindowGoodColor
		if !canIncreaseStat(ctx.Session, row) {
			fill = statsWindowDisabled
			textColor = statsWindowMutedColor
		} else if pointInRect(mx, my, bx, by, bw, bh) {
			if mouseDown {
				fill = statsWindowDownColor
			} else {
				fill = statsWindowHoverColor
			}
		}
		drawUIButtonSurface(screen, bx, by, bw, bh, fill)
		render.DebugPrintAtColor(screen, "+", bx+5, by+1, textColor)
	}

	leftX := x + statsWindowPad
	rightX := x + 150
	derivedY := y + 220
	drawStatsDerived(screen, leftX, derivedY, "ATK", fmt.Sprintf("%d + %d", stats.Attack, stats.AttackBonus))
	drawStatsDerived(screen, leftX, derivedY+18, "MATK", fmt.Sprintf("%d - %d", stats.MatkMin, stats.MatkMax))
	drawStatsDerived(screen, leftX, derivedY+36, "HIT", fmt.Sprintf("%d", stats.Hit))
	drawStatsDerived(screen, leftX, derivedY+54, "CRIT", fmt.Sprintf("%d", stats.Critical))
	drawStatsDerived(screen, rightX, derivedY, "DEF", fmt.Sprintf("%d + %d", stats.Defense, stats.DefenseBonus))
	drawStatsDerived(screen, rightX, derivedY+18, "MDEF", fmt.Sprintf("%d + %d", stats.MDefense, stats.MDefenseBonus))
	drawStatsDerived(screen, rightX, derivedY+36, "FLEE", fmt.Sprintf("%d + %d", stats.Flee, stats.FleeBonus))
	drawStatsDerived(screen, rightX, derivedY+54, "ASPD", fmt.Sprintf("%d", displayASPD(stats.ASPD+stats.ASPDBonus)))

	if w.status != "" && time.Since(w.statusAt) < 1800*time.Millisecond {
		statusColor := statsWindowErrorColor
		if w.statusGood {
			statusColor = statsWindowGoodColor
		}
		render.DebugPrintAtColor(screen, trimRunes(w.status, 36), x+statsWindowPad, y+statsWindowHeight-20, statusColor)
	}
}

func (w *statsWindowState) cursorAction(ctx Context) (int, bool) {
	if !w.open || ctx.Input == nil {
		return 0, false
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	cx, cy, cw, ch := w.closeBounds()
	if pointInRect(mx, my, cx, cy, cw, ch) {
		return cursorActionClick, true
	}
	if pointInRect(mx, my, w.x, w.y, statsWindowWidth, statsWindowTitleH) {
		return cursorActionClick, true
	}
	for _, row := range statsRows(ctx.Session) {
		if !canIncreaseStat(ctx.Session, row) {
			continue
		}
		bx, by, bw, bh := w.statButtonBounds(row.statusID)
		if pointInRect(mx, my, bx, by, bw, bh) {
			return cursorActionClick, true
		}
	}
	if pointInRect(mx, my, w.x, w.y, statsWindowWidth, statsWindowHeight) {
		return cursorActionDefault, true
	}
	return 0, false
}

func (w *statsWindowState) ensurePosition(ctx Context) {
	if w.positioned {
		return
	}
	width, height := ctx.ScreenSize()
	w.x = minInt(characterWindowX+characterWindowWidth+12, maxInt(8, width-statsWindowWidth-8))
	w.y = minInt(characterWindowY, maxInt(8, height-statsWindowHeight-8))
	if w.x < 8 {
		w.x = 8
	}
	if w.y < 8 {
		w.y = 8
	}
	w.positioned = true
}

func (w *statsWindowState) closeBounds() (int, int, int, int) {
	return w.x + statsWindowWidth - 24, w.y + 6, 16, 16
}

func (w *statsWindowState) statRowY(index int) int {
	return w.y + statsWindowTitleH + 50 + index*statsRowH
}

func (w *statsWindowState) statButtonBounds(statusID uint16) (int, int, int, int) {
	rows := statsRows(nil)
	for i, row := range rows {
		if row.statusID == statusID {
			return w.x + statsWindowWidth - statsWindowPad - statsButtonSize, w.statRowY(i) + 2, statsButtonSize, statsButtonSize
		}
	}
	return 0, 0, 0, 0
}

func (w *statsWindowState) setStatus(status string, good bool) {
	w.status = status
	w.statusGood = good
	w.statusAt = time.Now()
}

func statsRows(s *session.Session) []statRow {
	stats := sessionStats(s)
	return []statRow{
		{label: "STR", statusID: network.StatusStr, value: stats.Str, bonus: stats.StrBonus, cost: stats.StrCost},
		{label: "AGI", statusID: network.StatusAgi, value: stats.Agi, bonus: stats.AgiBonus, cost: stats.AgiCost},
		{label: "VIT", statusID: network.StatusVit, value: stats.Vit, bonus: stats.VitBonus, cost: stats.VitCost},
		{label: "INT", statusID: network.StatusInt, value: stats.Int, bonus: stats.IntBonus, cost: stats.IntCost},
		{label: "DEX", statusID: network.StatusDex, value: stats.Dex, bonus: stats.DexBonus, cost: stats.DexCost},
		{label: "LUK", statusID: network.StatusLuk, value: stats.Luk, bonus: stats.LukBonus, cost: stats.LukCost},
	}
}

func sessionStats(s *session.Session) session.Stats {
	if s == nil {
		return session.Stats{}
	}
	stats := s.Stats
	character := selectedCharacter(s)
	if stats.Str == 0 {
		stats.Str = int(character.Str)
	}
	if stats.Agi == 0 {
		stats.Agi = int(character.Agi)
	}
	if stats.Vit == 0 {
		stats.Vit = int(character.Vit)
	}
	if stats.Int == 0 {
		stats.Int = int(character.Int)
	}
	if stats.Dex == 0 {
		stats.Dex = int(character.Dex)
	}
	if stats.Luk == 0 {
		stats.Luk = int(character.Luk)
	}
	return stats
}

func canIncreaseStat(s *session.Session, row statRow) bool {
	if s == nil || row.value <= 0 || row.value >= 99 {
		return false
	}
	return sessionStats(s).Points >= statCost(row)
}

func statCost(row statRow) int {
	if row.cost > 0 {
		return row.cost
	}
	return statPointCost(row.value)
}

func statPointCost(current int) int {
	if current < 1 {
		current = 1
	}
	return 1 + (current+9)/10
}

func formatStatValue(value, bonus int) string {
	if bonus == 0 {
		return fmt.Sprintf("%d", value)
	}
	if bonus > 0 {
		return fmt.Sprintf("%d + %d", value, bonus)
	}
	return fmt.Sprintf("%d - %d", value, -bonus)
}

func drawStatsDerived(screen *render.Image, x, y int, label, value string) {
	render.DebugPrintAtColor(screen, label, x, y, statsWindowMutedColor)
	render.DebugPrintAtColor(screen, value, x+46, y, statsWindowTextColor)
}

func displayASPD(raw int) int {
	if raw <= 0 {
		return 0
	}
	return raw / 4
}

func clampStatsWindowInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func applyStatusSnapshot(ctx Context, snapshot network.StatusSnapshot) {
	if ctx.Session == nil {
		return
	}
	ctx.Session.Stats.Points = snapshot.Points
	setSessionStat(ctx.Session, network.StatusStr, snapshot.Str)
	setSessionStat(ctx.Session, network.StatusAgi, snapshot.Agi)
	setSessionStat(ctx.Session, network.StatusVit, snapshot.Vit)
	setSessionStat(ctx.Session, network.StatusInt, snapshot.Int)
	setSessionStat(ctx.Session, network.StatusDex, snapshot.Dex)
	setSessionStat(ctx.Session, network.StatusLuk, snapshot.Luk)
	ctx.Session.Stats.StrBonus = snapshot.StrBonus
	ctx.Session.Stats.AgiBonus = snapshot.AgiBonus
	ctx.Session.Stats.VitBonus = snapshot.VitBonus
	ctx.Session.Stats.IntBonus = snapshot.IntBonus
	ctx.Session.Stats.DexBonus = snapshot.DexBonus
	ctx.Session.Stats.LukBonus = snapshot.LukBonus
	ctx.Session.Stats.StrCost = snapshot.StrCost
	ctx.Session.Stats.AgiCost = snapshot.AgiCost
	ctx.Session.Stats.VitCost = snapshot.VitCost
	ctx.Session.Stats.IntCost = snapshot.IntCost
	ctx.Session.Stats.DexCost = snapshot.DexCost
	ctx.Session.Stats.LukCost = snapshot.LukCost
	ctx.Session.Stats.Attack = snapshot.Attack
	ctx.Session.Stats.AttackBonus = snapshot.AttackBonus
	ctx.Session.Stats.MatkMin = snapshot.MatkMin
	ctx.Session.Stats.MatkMax = snapshot.MatkMax
	ctx.Session.Stats.Defense = snapshot.Defense
	ctx.Session.Stats.DefenseBonus = snapshot.DefenseBonus
	ctx.Session.Stats.MDefense = snapshot.MDefense
	ctx.Session.Stats.MDefenseBonus = snapshot.MDefenseBonus
	ctx.Session.Stats.Hit = snapshot.Hit
	ctx.Session.Stats.Flee = snapshot.Flee
	ctx.Session.Stats.FleeBonus = snapshot.FleeBonus
	ctx.Session.Stats.Critical = snapshot.Critical
	ctx.Session.Stats.ASPD = snapshot.ASPD
	ctx.Session.Stats.ASPDBonus = snapshot.ASPDBonus
	log.Printf("status snapshot points=%d str=%d agi=%d vit=%d int=%d dex=%d luk=%d", snapshot.Points, snapshot.Str, snapshot.Agi, snapshot.Vit, snapshot.Int, snapshot.Dex, snapshot.Luk)
}

func (w *statsWindowState) applyStatusChangeAck(ctx Context, ack network.StatusChangeAck) {
	if ctx.Session == nil {
		return
	}
	label := statLabel(ack.StatusID)
	if !ack.Success {
		w.setStatus(fmt.Sprintf("%s increase failed", label), false)
		log.Printf("status increase ack status=%d success=false value=%d", ack.StatusID, ack.Value)
		return
	}
	setSessionStat(ctx.Session, ack.StatusID, ack.Value)
	if ctx.Session.Stats.Points > 0 {
		ctx.Session.Stats.Points--
	}
	w.setStatus(fmt.Sprintf("%s increased to %d", label, ack.Value), true)
	log.Printf("status increase ack status=%d success=true value=%d", ack.StatusID, ack.Value)
}

func setSessionStat(s *session.Session, statusID uint16, value int) {
	if value < 0 {
		value = 0
	}
	if value > 255 {
		value = 255
	}
	switch statusID {
	case network.StatusStr:
		s.Stats.Str = value
		s.Selected.Str = uint8(value)
	case network.StatusAgi:
		s.Stats.Agi = value
		s.Selected.Agi = uint8(value)
	case network.StatusVit:
		s.Stats.Vit = value
		s.Selected.Vit = uint8(value)
	case network.StatusInt:
		s.Stats.Int = value
		s.Selected.Int = uint8(value)
	case network.StatusDex:
		s.Stats.Dex = value
		s.Selected.Dex = uint8(value)
	case network.StatusLuk:
		s.Stats.Luk = value
		s.Selected.Luk = uint8(value)
	}
}

func setSessionStatCost(s *session.Session, statusID uint16, value int) {
	switch statusID {
	case network.StatusUStr:
		s.Stats.StrCost = value
	case network.StatusUAgi:
		s.Stats.AgiCost = value
	case network.StatusUVit:
		s.Stats.VitCost = value
	case network.StatusUInt:
		s.Stats.IntCost = value
	case network.StatusUDex:
		s.Stats.DexCost = value
	case network.StatusULuk:
		s.Stats.LukCost = value
	}
}

func statLabel(statusID uint16) string {
	switch statusID {
	case network.StatusStr:
		return "STR"
	case network.StatusAgi:
		return "AGI"
	case network.StatusVit:
		return "VIT"
	case network.StatusInt:
		return "INT"
	case network.StatusDex:
		return "DEX"
	case network.StatusLuk:
		return "LUK"
	default:
		return fmt.Sprintf("Stat %d", statusID)
	}
}
