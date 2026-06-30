package gamemode

import (
	"fmt"
	"image/color"
	"log"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

const (
	teleportSkillID      = 26
	teleportRandomMap    = "Random"
	teleportSavePointMap = "SavePoint"

	teleportModalWidth   = 260
	teleportModalHeight  = 178
	teleportModalPad     = 16
	teleportModalTitleH  = 32
	teleportModalButtonH = 28
	teleportModalGap     = 8
)

type teleportModalState struct {
	open     bool
	skill    session.Skill
	mapNames []string
	status   string
}

type teleportModalAction int

const (
	teleportModalActionRandom teleportModalAction = iota
	teleportModalActionSavePoint
	teleportModalActionCancel
)

type teleportModalButton struct {
	label  string
	action teleportModalAction
}

var teleportModalButtons = []teleportModalButton{
	{label: "Random", action: teleportModalActionRandom},
	{label: "Save Point", action: teleportModalActionSavePoint},
	{label: "Cancel", action: teleportModalActionCancel},
}

func (m *teleportModalState) openSkill(skill session.Skill) {
	m.open = true
	m.skill = skill
	m.mapNames = nil
	m.status = ""
}

func (m *teleportModalState) openWarpPointList(list network.WarpPointList, skill session.Skill) {
	m.open = true
	m.skill = skill
	m.mapNames = append(m.mapNames[:0], list.MapNames...)
	m.status = ""
}

func (m *WorldMode) applyWarpPointList(ctx Context, list network.WarpPointList) {
	if list.SkillID != teleportSkillID {
		log.Printf("warp point list ignored skill=%d maps=%v", list.SkillID, list.MapNames)
		return
	}
	skill, ok := skillByID(ctx.Session, list.SkillID)
	if !ok {
		skill = session.Skill{ID: list.SkillID, Level: 1}
		if len(list.MapNames) > 1 {
			skill.Level = 2
		}
	}
	if teleportWarpListBypassesModal(skill, list) {
		m.autoSelectTeleportRandom(ctx, list)
		return
	}
	m.teleportModal.openWarpPointList(list, skill)
	m.status = "choose teleport destination"
	log.Printf("teleport destination list skill=%d maps=%v", list.SkillID, list.MapNames)
}

func (m *WorldMode) autoSelectTeleportRandom(ctx Context, list network.WarpPointList) {
	mapName := teleportRandomMap
	for _, name := range list.MapNames {
		if name == teleportRandomMap {
			mapName = name
			break
		}
	}
	if ctx.Network == nil {
		m.status = "Teleport failed: not connected"
		return
	}
	if err := ctx.Network.SendSelectWarpPoint(list.SkillID, mapName); err != nil {
		m.status = fmt.Sprintf("Teleport failed: %v", err)
		return
	}
	m.teleportModal = teleportModalState{}
	m.status = "teleporting"
	log.Printf("teleport random selected automatically skill=%d maps=%v", list.SkillID, list.MapNames)
}

func teleportWarpListBypassesModal(skill session.Skill, list network.WarpPointList) bool {
	if list.SkillID != teleportSkillID {
		return false
	}
	if isLevelOneTeleportSkill(skill) {
		return true
	}
	for _, name := range list.MapNames {
		if name != "" && name != teleportRandomMap {
			return false
		}
	}
	return true
}

func (m *teleportModalState) update(ctx Context, mode *WorldMode) bool {
	if !m.open || ctx.Input == nil {
		return m.open
	}
	if ctx.Input.JustPressed(render.KeyEscape) || ctx.Input.MouseJustPressed(render.MouseButtonRight) {
		m.open = false
		return true
	}
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return true
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := teleportModalBounds(width, height)
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	for i, button := range teleportModalButtons {
		if !m.buttonEnabled(button) {
			continue
		}
		bx, by, bw, bh := teleportModalButtonBounds(x, y, w, i)
		if !pointInRect(mx, my, bx, by, bw, bh) {
			continue
		}
		m.activate(ctx, mode, button.action)
		return true
	}
	return true
}

func (m *teleportModalState) activate(ctx Context, mode *WorldMode, action teleportModalAction) {
	switch action {
	case teleportModalActionRandom:
		m.selectWarpPoint(ctx, mode, m.randomMapName())
	case teleportModalActionSavePoint:
		if !m.savePointEnabled() {
			return
		}
		m.selectWarpPoint(ctx, mode, m.savePointMapName())
	case teleportModalActionCancel:
		m.open = false
	}
}

func (m *teleportModalState) selectWarpPoint(ctx Context, mode *WorldMode, mapName string) {
	if ctx.Network == nil {
		m.status = "Teleport failed: not connected"
		return
	}
	skillID := uint16(teleportSkillID)
	if m.skill.ID != 0 {
		skillID = m.skill.ID
	}
	if err := ctx.Network.SendSelectWarpPoint(skillID, mapName); err != nil {
		m.status = fmt.Sprintf("Teleport failed: %v", err)
		return
	}
	if mode != nil {
		mode.addWorldEffect(ctx, effectTeleportation, localSkillTarget(ctx))
	}
	m.open = false
}

func (m teleportModalState) buttonEnabled(button teleportModalButton) bool {
	return button.action != teleportModalActionSavePoint || m.savePointEnabled()
}

func (m teleportModalState) savePointEnabled() bool {
	return m.savePointMapName() != ""
}

func (m teleportModalState) randomMapName() string {
	for _, name := range m.mapNames {
		if name == teleportRandomMap {
			return name
		}
	}
	return teleportRandomMap
}

func (m teleportModalState) savePointMapName() string {
	for _, name := range m.mapNames {
		if name != "" && name != teleportRandomMap {
			return name
		}
	}
	if len(m.mapNames) == 0 && m.skill.Level >= 2 {
		return teleportSavePointMap
	}
	return ""
}

func isLevelOneTeleportSkill(skill session.Skill) bool {
	return skill.ID == teleportSkillID && skill.Level <= 1
}

func (m *teleportModalState) draw(screen *render.Image, ctx Context, width, height int) {
	if !m.open || screen == nil {
		return
	}
	drawUISurface(screen, 0, 0, width, height, color.RGBA{A: 72}, color.RGBA{})
	x, y, w, h := teleportModalBounds(width, height)
	drawUITitledWindowFrame(screen, x, y, w, h, teleportModalTitleH)
	drawUIWindowTitle(screen, x, y, teleportModalTitleH, teleportModalPad, "Teleport", uiTitleTextColor)
	render.DebugPrintAtColor(screen, "Choose destination.", x+teleportModalPad, y+teleportModalTitleH+12, uiTextColor)

	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i, button := range teleportModalButtons {
		bx, by, bw, bh := teleportModalButtonBounds(x, y, w, i)
		fill := uiButtonColor
		textColor := uiTextColor
		if !m.buttonEnabled(button) {
			fill = uiDisabledColor
			textColor = uiMutedTextColor
		} else if pointInRect(mx, my, bx, by, bw, bh) {
			fill = uiButtonHoverColor
		}
		drawUIButtonLabel(screen, bx, by, bw, bh, button.label, fill, textColor)
	}
	if m.status != "" {
		render.DebugPrintAtColor(screen, trimRunes(m.status, 30), x+teleportModalPad, y+h-16, uiErrorTextColor)
	}
}

func (m *teleportModalState) cursorAction(ctx Context) (int, bool) {
	if !m.open || ctx.Input == nil {
		return 0, false
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := teleportModalBounds(width, height)
	for i, button := range teleportModalButtons {
		if !m.buttonEnabled(button) {
			continue
		}
		bx, by, bw, bh := teleportModalButtonBounds(x, y, w, i)
		if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, bx, by, bw, bh) {
			return cursorActionClick, true
		}
	}
	return cursorActionDefault, true
}

func teleportModalBounds(width, height int) (int, int, int, int) {
	w := minInt(teleportModalWidth, maxInt(220, width-40))
	h := minInt(teleportModalHeight, maxInt(160, height-40))
	x := (width - w) / 2
	y := (height - h) / 2
	if y < 16 {
		y = 16
	}
	return x, y, w, h
}

func teleportModalButtonBounds(x, y, w, index int) (int, int, int, int) {
	bx := x + teleportModalPad
	by := y + teleportModalTitleH + 34 + index*(teleportModalButtonH+teleportModalGap)
	bw := w - 2*teleportModalPad
	return bx, by, bw, teleportModalButtonH
}
