package gamemode

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"strings"
	"time"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
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
	skillWindowTitleColor  = uiTitleTextColor
	skillWindowTextColor   = uiTextColor
	skillWindowMutedColor  = uiMutedTextColor
	skillWindowGoodColor   = uiGoodTextColor
	skillWindowErrorColor  = uiErrorTextColor
	skillWindowButtonColor = uiButtonColor
	skillWindowHoverColor  = uiButtonHoverColor
	skillWindowDownColor   = uiButtonDownColor
	skillWindowDisabled    = uiDisabledColor
	skillWindowPassive     = color.RGBA{R: 34, G: 142, B: 158, A: 255}
	skillWindowActive      = color.RGBA{R: 44, G: 92, B: 184, A: 255}
)

type skillWindowState struct {
	open        bool
	x           int
	y           int
	positioned  bool
	dragging    bool
	dragDX      int
	dragDY      int
	scroll      int
	status      string
	statusGood  bool
	statusAt    time.Time
	lastClick   uint16
	lastClickAt time.Time
	dragSkill   session.Skill
	dragActive  bool
	dragFrom    time.Time
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

func (w *skillWindowState) update(ctx Context, shortcuts *shortcutBarState, mode *WorldMode) bool {
	if !w.open || ctx.Input == nil {
		return false
	}
	w.ensurePosition(ctx)
	w.clampScroll(ctx.Session)
	width, height := ctx.ScreenSize()
	if w.dragActive {
		if ctx.Input.MouseJustReleased(render.MouseButtonLeft) || !ctx.Input.MousePressed(render.MouseButtonLeft) {
			skill := w.dragSkill
			w.dragActive = false
			w.dragSkill = session.Skill{}
			if shortcuts != nil && shortcuts.acceptSkillDrop(ctx, skill, ctx.Input.MouseX, ctx.Input.MouseY) {
				return true
			}
			return pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, skillWindowWidth, skillWindowHeight)
		}
	}
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
		if pointInRect(mx, my, bx, by, bw, bh) {
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
			w.setStatus(fmt.Sprintf("%s level-up requested", skillDisplayName(ctx.Resources, skill)), true)
			return true
		}
		rx, ry, rw, rh := w.skillRowBounds(row)
		if pointInRect(mx, my, rx, ry, rw, rh) {
			if skill.Level <= 0 {
				w.setStatus("Skill is not learned", false)
				return true
			}
			now := time.Now()
			if w.lastClick == skill.ID && now.Sub(w.lastClickAt) <= 360*time.Millisecond {
				w.lastClick = 0
				w.lastClickAt = time.Time{}
				if mode == nil {
					w.setStatus("No world mode", false)
					return true
				}
				if err := mode.useSkill(ctx, skill, "skill-window"); err != nil {
					w.setStatus(err.Error(), false)
					return true
				}
				w.setStatus(fmt.Sprintf("%s used", skillDisplayName(ctx.Resources, skill)), true)
				return true
			}
			w.lastClick = skill.ID
			w.lastClickAt = now
			w.dragSkill = skill
			w.dragActive = true
			w.dragFrom = now
			return true
		}
	}
	return true
}

func (w *skillWindowState) draw(screen *render.Image, ctx Context, mode *WorldMode) {
	if !w.open || screen == nil {
		return
	}
	w.ensurePosition(ctx)
	w.clampScroll(ctx.Session)
	x, y := w.x, w.y
	drawUITitledWindowFrame(screen, x, y, skillWindowWidth, skillWindowHeight, skillWindowTitleH)
	drawUIWindowTitle(screen, x, y, skillWindowTitleH, skillWindowPad, "Skill Tree", skillWindowTitleColor)
	cx, cy, cw, ch := w.closeBounds()
	drawUICloseButton(screen, cx, cy, cw, ch, skillWindowButtonColor, skillWindowTextColor)

	points := sessionSkillPoints(ctx.Session)
	render.DebugPrintAtColor(screen, fmt.Sprintf("Skill Points : %d", points), x+skillWindowPad, y+skillWindowTitleH+10, skillWindowTextColor)
	headerY := y + skillWindowTitleH + 32
	render.DebugPrintAtColor(screen, "Name", x+skillWindowPad+34, headerY, skillWindowMutedColor)
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
			rowColor := uiPanelBodyColor
			if row%2 == 1 {
				rowColor = uiPanelAltColor
			}
			drawUIRowSurface(screen, x+skillWindowPad, ry, skillWindowWidth-2*skillWindowPad, skillRowH-2, rowColor)
			if mode != nil {
				mode.drawSkillIcon(screen, ctx.Resources, skill, x+skillWindowPad+3, ry+2, 22)
			}
			typeColor := skillWindowPassive
			typeLabel := "P"
			if skill.Type != 0 {
				typeColor = skillWindowActive
				typeLabel = "A"
			}
			nameColor := skillWindowTextColor
			if skill.Level <= 0 {
				nameColor = skillWindowMutedColor
			}
			render.DebugPrintAtColor(screen, typeLabel, x+skillWindowPad+28, ry+7, typeColor)
			render.DebugPrintAtColor(screen, trimRunes(skillDisplayName(ctx.Resources, skill), 18), x+skillWindowPad+44, ry+7, nameColor)
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
			drawUIButtonLabel(screen, bx, by, bw, bh, "+", fill, textColor)
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
	if w.dragActive && ctx.Input != nil && time.Since(w.dragFrom) > 80*time.Millisecond && mode != nil {
		mode.drawSkillIcon(screen, ctx.Resources, w.dragSkill, ctx.Input.MouseX-12, ctx.Input.MouseY-12, 24)
	}
	if !w.dragActive && ctx.Input != nil {
		if skill, ok := w.hoveredSkill(ctx); ok {
			drawSkillTooltip(screen, ctx, skill)
		}
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

func (w *skillWindowState) skillRowBounds(row int) (int, int, int, int) {
	return w.x + skillWindowPad, w.skillRowY(row), skillWindowWidth - 2*skillWindowPad, skillRowH - 2
}

func (w *skillWindowState) levelButtonBounds(row int) (int, int, int, int) {
	return w.x + skillWindowWidth - skillWindowPad - skillButtonSize, w.skillRowY(row) + 5, skillButtonSize, skillButtonSize
}

func (w *skillWindowState) hoveredSkill(ctx Context) (session.Skill, bool) {
	if !w.open || ctx.Input == nil {
		return session.Skill{}, false
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	for row, skill := range visibleSkills(ctx.Session, w.scroll, visibleSkillRows()) {
		x, y, width, height := w.skillRowBounds(row)
		if pointInRect(mx, my, x, y, width, height) {
			return skill, true
		}
	}
	return session.Skill{}, false
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

func skillDisplayName(manager *res.Manager, skill session.Skill) string {
	if manager != nil {
		if name, ok := manager.SkillDisplayName(int(skill.ID)); ok {
			return name
		}
	}
	return skillLabel(skill)
}

func drawSkillTooltip(screen *render.Image, ctx Context, skill session.Skill) {
	if screen == nil || ctx.Input == nil {
		return
	}
	const tooltipW = 292
	name := skillDisplayName(ctx.Resources, skill)
	lines := []string{
		fmt.Sprintf("Lv %d", skill.Level),
	}
	if skill.SPCost > 0 {
		lines = append(lines, fmt.Sprintf("SP Cost: %d", skill.SPCost))
	}
	if skill.Range > 0 {
		lines = append(lines, fmt.Sprintf("Range: %d", skill.Range))
	}
	hasDescription := false
	if ctx.Resources != nil {
		if desc, ok := ctx.Resources.SkillDescription(int(skill.ID)); ok {
			hasDescription = true
			lines = append(lines, "")
			for _, line := range desc {
				clean := strings.TrimSpace(stripItemInfoColorCodes(strings.ReplaceAll(line, "_", " ")))
				if clean == "" {
					lines = append(lines, "")
					continue
				}
				lines = append(lines, clean)
			}
		}
	}
	if !hasDescription {
		lines = append(lines, "", "No description available.")
	}
	wrapped := wrapItemInfoLines(lines, 38)
	tooltipH := 12 + itemInfoLineH*(len(wrapped)+1)
	x := ctx.Input.MouseX + 16
	y := ctx.Input.MouseY + 18
	screenW, screenH := ctx.ScreenSize()
	x = clampInventoryWindowInt(x, 8, maxInt(8, screenW-tooltipW-8))
	y = clampInventoryWindowInt(y, 8, maxInt(8, screenH-tooltipH-8))
	drawUISurface(screen, x, y, tooltipW, tooltipH, uiPanelBodyColor, uiWindowBorderColor)
	render.DebugPrintAtColor(screen, trimRunes(name, 38), x+7, y+6, skillWindowTextColor)
	lineY := y + 6 + itemInfoLineH
	for i, line := range wrapped {
		color := skillWindowMutedColor
		if i >= 4 {
			color = skillWindowTextColor
		}
		render.DebugPrintAtColor(screen, trimRunes(line, 38), x+7, lineY+i*itemInfoLineH, color)
	}
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
