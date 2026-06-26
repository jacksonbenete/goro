package gamemode

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"strings"
	"time"

	"github.com/kivutar/goro/internal/network"
	"github.com/kivutar/goro/internal/render"
	"github.com/kivutar/goro/internal/session"
)

const (
	skillWindowWidth  = 360
	skillWindowHeight = 382
	skillWindowTitleH = 28
	skillWindowPad    = 12
	skillRowH         = 28
	skillListTop      = 80
	skillListBottom   = 42
	skillButtonSize   = 17
)

var (
	skillWindowTitleColor  = color.RGBA{R: 255, G: 230, B: 150, A: 255}
	skillWindowTextColor   = color.RGBA{R: 236, G: 232, B: 220, A: 255}
	skillWindowMutedColor  = color.RGBA{R: 166, G: 174, B: 184, A: 255}
	skillWindowGoodColor   = color.RGBA{R: 144, G: 210, B: 142, A: 255}
	skillWindowErrorColor  = color.RGBA{R: 255, G: 116, B: 116, A: 255}
	skillWindowButtonColor = color.RGBA{R: 56, G: 62, B: 72, A: 235}
	skillWindowHoverColor  = color.RGBA{R: 82, G: 92, B: 108, A: 245}
	skillWindowDownColor   = color.RGBA{R: 98, G: 106, B: 122, A: 245}
	skillWindowDisabled    = color.RGBA{R: 42, G: 46, B: 54, A: 205}
	skillWindowPassive     = color.RGBA{R: 126, G: 206, B: 226, A: 255}
	skillWindowActive      = color.RGBA{R: 148, G: 170, B: 240, A: 255}
)

type skillWindowState struct {
	open       bool
	x          int
	y          int
	positioned bool
	dragging   bool
	dragDX     int
	dragDY     int
	scroll     int
	status     string
	statusGood bool
	statusAt   time.Time
}

func (w *skillWindowState) toggle(ctx Context) {
	if w.open {
		w.open = false
		w.dragging = false
		return
	}
	w.open = true
	w.ensurePosition(ctx)
	w.clampScroll(ctx.Session)
}

func (w *skillWindowState) update(ctx Context) bool {
	if !w.open || ctx.Input == nil {
		return false
	}
	w.ensurePosition(ctx)
	w.clampScroll(ctx.Session)
	width, height := ctx.ScreenSize()
	if w.dragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			w.x = clampSkillWindowInt(ctx.Input.MouseX-w.dragDX, 8, maxInt(8, width-skillWindowWidth-8))
			w.y = clampSkillWindowInt(ctx.Input.MouseY-w.dragDY, 8, maxInt(8, height-skillWindowHeight-8))
			return true
		}
		w.dragging = false
		return true
	}
	if ctx.Input.JustPressed(render.KeyEscape) {
		w.open = false
		return true
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	inside := pointInRect(mx, my, w.x, w.y, skillWindowWidth, skillWindowHeight)
	if inside && ctx.Input.WheelY != 0 {
		w.scrollBy(ctx.Input.WheelY, ctx.Session)
		return true
	}
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return inside
	}
	if !inside {
		return false
	}
	cx, cy, cw, ch := w.closeBounds()
	if pointInRect(mx, my, cx, cy, cw, ch) {
		w.open = false
		return true
	}
	if pointInRect(mx, my, w.x, w.y, skillWindowWidth, skillWindowTitleH) {
		w.dragging = true
		w.dragDX = mx - w.x
		w.dragDY = my - w.y
		return true
	}
	for row, skill := range visibleSkills(ctx.Session, w.scroll, visibleSkillRows()) {
		bx, by, bw, bh := w.levelButtonBounds(row)
		if !pointInRect(mx, my, bx, by, bw, bh) {
			continue
		}
		if !canIncreaseSkill(ctx.Session, skill) {
			w.setStatus("No skill points or skill is maxed", false)
			return true
		}
		if ctx.Network == nil {
			w.setStatus("Not connected", false)
			return true
		}
		if err := ctx.Network.SendSkillLevelUp(skill.ID); err != nil {
			w.setStatus(err.Error(), false)
			return true
		}
		w.setStatus(fmt.Sprintf("%s level-up requested", skillLabel(skill)), true)
		return true
	}
	return true
}

func (w *skillWindowState) draw(screen *render.Image, ctx Context) {
	if !w.open || screen == nil {
		return
	}
	w.ensurePosition(ctx)
	w.clampScroll(ctx.Session)
	x, y := w.x, w.y
	drawNPCWindowFrame(screen, x, y, skillWindowWidth, skillWindowHeight)
	render.DebugPrintAtColor(screen, "Skill Tree", x+skillWindowPad, y+9, skillWindowTitleColor)
	cx, cy, cw, ch := w.closeBounds()
	drawUIButtonSurface(screen, cx, cy, cw, ch, skillWindowButtonColor)
	render.DebugPrintAtColor(screen, "x", cx+5, cy+2, skillWindowTextColor)
	render.DrawRect(screen, float64(x+8), float64(y+skillWindowTitleH), float64(skillWindowWidth-16), 1, color.RGBA{R: 210, G: 200, B: 170, A: 80})

	points := sessionSkillPoints(ctx.Session)
	render.DebugPrintAtColor(screen, fmt.Sprintf("Skill Points : %d", points), x+skillWindowPad, y+skillWindowTitleH+10, skillWindowTextColor)
	headerY := y + skillWindowTitleH + 32
	render.DebugPrintAtColor(screen, "Name", x+skillWindowPad, headerY, skillWindowMutedColor)
	render.DebugPrintAtColor(screen, "Lv", x+204, headerY, skillWindowMutedColor)
	render.DebugPrintAtColor(screen, "SP", x+244, headerY, skillWindowMutedColor)
	render.DebugPrintAtColor(screen, "Range", x+282, headerY, skillWindowMutedColor)

	mx, my := -1, -1
	mouseDown := false
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
		mouseDown = ctx.Input.MousePressed(render.MouseButtonLeft)
	}
	skills := sessionSkills(ctx.Session)
	if len(skills) == 0 {
		render.DebugPrintAtColor(screen, "No skills received from server yet.", x+skillWindowPad, y+skillListTop+18, skillWindowMutedColor)
	} else {
		for row, skill := range visibleSkills(ctx.Session, w.scroll, visibleSkillRows()) {
			ry := w.skillRowY(row)
			rowColor := color.RGBA{R: 32, G: 36, B: 44, A: 185}
			if row%2 == 1 {
				rowColor = color.RGBA{R: 38, G: 42, B: 50, A: 185}
			}
			drawUIRowSurface(screen, x+skillWindowPad, ry, skillWindowWidth-2*skillWindowPad, skillRowH-2, rowColor)
			typeColor := skillWindowPassive
			typeLabel := "P"
			if skill.Type != 0 {
				typeColor = skillWindowActive
				typeLabel = "A"
			}
			render.DebugPrintAtColor(screen, typeLabel, x+skillWindowPad+5, ry+7, typeColor)
			nameColor := skillWindowTextColor
			if skill.Level <= 0 {
				nameColor = skillWindowMutedColor
			}
			render.DebugPrintAtColor(screen, trimRunes(skillLabel(skill), 22), x+skillWindowPad+22, ry+7, nameColor)
			render.DebugPrintAtColor(screen, fmt.Sprintf("%d", skill.Level), x+204, ry+7, nameColor)
			render.DebugPrintAtColor(screen, fmt.Sprintf("%d", skill.SPCost), x+244, ry+7, skillWindowMutedColor)
			render.DebugPrintAtColor(screen, fmt.Sprintf("%d", skill.Range), x+292, ry+7, skillWindowMutedColor)
			bx, by, bw, bh := w.levelButtonBounds(row)
			fill := skillWindowButtonColor
			textColor := skillWindowGoodColor
			if !canIncreaseSkill(ctx.Session, skill) {
				fill = skillWindowDisabled
				textColor = skillWindowMutedColor
			} else if pointInRect(mx, my, bx, by, bw, bh) {
				if mouseDown {
					fill = skillWindowDownColor
				} else {
					fill = skillWindowHoverColor
				}
			}
			drawUIButtonSurface(screen, bx, by, bw, bh, fill)
			render.DebugPrintAtColor(screen, "+", bx+5, by+1, textColor)
		}
		w.drawScrollBar(screen, ctx.Session)
	}
	if w.status != "" && time.Since(w.statusAt) < 1800*time.Millisecond {
		statusColor := skillWindowErrorColor
		if w.statusGood {
			statusColor = skillWindowGoodColor
		}
		render.DebugPrintAtColor(screen, trimRunes(w.status, 44), x+skillWindowPad, y+skillWindowHeight-20, statusColor)
	}
}

func (w *skillWindowState) cursorAction(ctx Context) (int, bool) {
	if !w.open || ctx.Input == nil {
		return 0, false
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	cx, cy, cw, ch := w.closeBounds()
	if pointInRect(mx, my, cx, cy, cw, ch) {
		return cursorActionClick, true
	}
	if pointInRect(mx, my, w.x, w.y, skillWindowWidth, skillWindowTitleH) {
		return cursorActionClick, true
	}
	for row, skill := range visibleSkills(ctx.Session, w.scroll, visibleSkillRows()) {
		if !canIncreaseSkill(ctx.Session, skill) {
			continue
		}
		bx, by, bw, bh := w.levelButtonBounds(row)
		if pointInRect(mx, my, bx, by, bw, bh) {
			return cursorActionClick, true
		}
	}
	if pointInRect(mx, my, w.x, w.y, skillWindowWidth, skillWindowHeight) {
		return cursorActionDefault, true
	}
	return 0, false
}

func (w *skillWindowState) ensurePosition(ctx Context) {
	if w.positioned {
		return
	}
	width, height := ctx.ScreenSize()
	w.x = minInt(characterWindowX+characterWindowWidth+12, maxInt(8, width-skillWindowWidth-8))
	w.y = minInt(characterWindowY, maxInt(8, height-skillWindowHeight-8))
	if w.x < 8 {
		w.x = 8
	}
	if w.y < 8 {
		w.y = 8
	}
	w.positioned = true
}

func (w *skillWindowState) closeBounds() (int, int, int, int) {
	return w.x + skillWindowWidth - 24, w.y + 6, 16, 16
}

func (w *skillWindowState) skillRowY(row int) int {
	return w.y + skillListTop + row*skillRowH
}

func (w *skillWindowState) levelButtonBounds(row int) (int, int, int, int) {
	return w.x + skillWindowWidth - skillWindowPad - skillButtonSize, w.skillRowY(row) + 5, skillButtonSize, skillButtonSize
}

func (w *skillWindowState) setStatus(status string, good bool) {
	w.status = status
	w.statusGood = good
	w.statusAt = time.Now()
}

func (w *skillWindowState) scrollBy(wheelY float64, s *session.Session) {
	step := int(math.Ceil(math.Abs(wheelY))) * 3
	if step < 1 {
		step = 1
	}
	if wheelY > 0 {
		w.scroll += step
	} else {
		w.scroll -= step
	}
	w.clampScroll(s)
}

func (w *skillWindowState) clampScroll(s *session.Session) {
	maxScroll := maxInt(0, len(sessionSkills(s))-visibleSkillRows())
	if w.scroll < 0 {
		w.scroll = 0
	}
	if w.scroll > maxScroll {
		w.scroll = maxScroll
	}
}

func (w *skillWindowState) drawScrollBar(screen *render.Image, s *session.Session) {
	total := len(sessionSkills(s))
	visible := visibleSkillRows()
	if total <= visible {
		return
	}
	trackX := float64(w.x + skillWindowWidth - 7)
	trackY := float64(w.y + skillListTop)
	trackH := float64(visible*skillRowH - 2)
	thumbH := math.Max(18, trackH*float64(visible)/float64(total))
	maxScroll := maxInt(1, total-visible)
	thumbTravel := math.Max(1, trackH-thumbH)
	thumbY := trackY + thumbTravel*float64(w.scroll)/float64(maxScroll)
	render.DrawRect(screen, trackX, trackY, 3, trackH, color.RGBA{R: 110, G: 124, B: 142, A: 90})
	render.DrawRect(screen, trackX, thumbY, 3, thumbH, color.RGBA{R: 205, G: 218, B: 232, A: 150})
}

func visibleSkillRows() int {
	return (skillWindowHeight - skillListTop - skillListBottom) / skillRowH
}

func visibleSkills(s *session.Session, scroll, count int) []session.Skill {
	skills := sessionSkills(s)
	if scroll > len(skills) {
		scroll = len(skills)
	}
	end := scroll + count
	if end > len(skills) {
		end = len(skills)
	}
	return skills[scroll:end]
}

func sessionSkills(s *session.Session) []session.Skill {
	if s == nil {
		return nil
	}
	return s.Skills.List
}

func sessionSkillPoints(s *session.Session) int {
	if s == nil {
		return 0
	}
	return s.Skills.Points
}

func skillLabel(skill session.Skill) string {
	if strings.TrimSpace(skill.Name) != "" {
		return skill.Name
	}
	return fmt.Sprintf("Skill %d", skill.ID)
}

func canIncreaseSkill(s *session.Session, skill session.Skill) bool {
	return s != nil && s.Skills.Points > 0 && skill.Upgradable
}

func applySkillInfoList(ctx Context, list network.SkillInfoList) {
	if ctx.Session == nil {
		return
	}
	ctx.Session.Skills.List = ctx.Session.Skills.List[:0]
	for _, skill := range list.Skills {
		ctx.Session.Skills.List = append(ctx.Session.Skills.List, sessionSkillFromNetwork(skill))
	}
	log.Printf("skill list received count=%d points=%d", len(ctx.Session.Skills.List), ctx.Session.Skills.Points)
}

func applySkillInfoUpdate(ctx Context, update network.SkillInfoUpdate) {
	if ctx.Session == nil {
		return
	}
	upsertSessionSkill(ctx.Session, sessionSkillFromNetwork(update.Skill))
	log.Printf("skill update id=%d level=%d sp=%d range=%d upgradable=%t", update.Skill.ID, update.Skill.Level, update.Skill.SPCost, update.Skill.Range, update.Skill.Upgradable)
}

func sessionSkillFromNetwork(skill network.SkillInfo) session.Skill {
	return session.Skill{
		ID:         skill.ID,
		Type:       skill.Type,
		Level:      skill.Level,
		SPCost:     skill.SPCost,
		Range:      skill.Range,
		Name:       skill.Name,
		Upgradable: skill.Upgradable,
	}
}

func upsertSessionSkill(s *session.Session, skill session.Skill) {
	for i := range s.Skills.List {
		if s.Skills.List[i].ID != skill.ID {
			continue
		}
		if skill.Type == 0 {
			skill.Type = s.Skills.List[i].Type
		}
		if strings.TrimSpace(skill.Name) == "" {
			skill.Name = s.Skills.List[i].Name
		}
		s.Skills.List[i] = skill
		return
	}
	s.Skills.List = append(s.Skills.List, skill)
}

func clampSkillWindowInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
