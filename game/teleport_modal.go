package game

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
	warpPortalSkillID    = 27
	teleportRandomMap    = "Random"
	teleportSavePointMap = "SavePoint"
	warpPointCancelMap   = "cancel"

	teleportModalWidth   = 260
	teleportModalMinH    = 160
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
	teleportModalActionCancel    teleportModalAction = -1
	teleportModalActionRandom    teleportModalAction = 0
	teleportModalActionSavePoint teleportModalAction = 1
)

type teleportModalButton struct {
	label   string
	action  teleportModalAction
	mapName string
	enabled bool
}

func (m *teleportModalState) openWarpPointList(list network.WarpPointList, skill session.Skill) {
	m.open = true
	m.skill = skill
	m.mapNames = append(m.mapNames[:0], list.MapNames...)
	m.status = ""
}

func (m *WorldMode) applyWarpPointList(ctx Context, list network.WarpPointList) {
	if list.SkillID != teleportSkillID && list.SkillID != warpPortalSkillID {
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
	m.status = fmt.Sprintf("choose %s destination", warpPointSkillLabel(list.SkillID))
	log.Printf("warp point destination list skill=%d maps=%v", list.SkillID, list.MapNames)
}

func (m *WorldMode) applyRememberWarpPointAck(_ Context, ack network.RememberWarpPointAck) {
	switch ack.Result {
	case 0:
		m.console.AddBlueMessage("Saved location as a Memo Point for Warp skill.")
	case 1:
		m.console.AddErrorMessage("Skill Level is not high enough.")
	case 2:
		m.console.AddErrorMessage("You haven't learned Warp.")
	default:
		m.console.AddErrorMessage("Memo failed.")
	}
	log.Printf("remember warp point ack result=%d", ack.Result)
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
		m.cancel(ctx)
		return true
	}
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return true
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := teleportModalBounds(width, height)
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	buttons := m.buttons()
	for i, button := range buttons {
		if !button.enabled {
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
	for _, button := range m.buttons() {
		if button.action != action || !button.enabled {
			continue
		}
		if button.action == teleportModalActionCancel {
			m.cancel(ctx)
			return
		}
		m.selectWarpPoint(ctx, mode, button.mapName)
		return
	}
}

func (m *teleportModalState) cancel(ctx Context) {
	if m.skill.ID == warpPortalSkillID && ctx.Network != nil {
		if err := ctx.Network.SendSelectWarpPoint(uint16(warpPortalSkillID), warpPointCancelMap); err != nil {
			m.status = fmt.Sprintf("Cancel failed: %v", err)
			return
		}
	}
	m.open = false
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
	if mode != nil && skillID == teleportSkillID {
		mode.addWorldEffect(ctx, effectTeleportation, localSkillTarget(ctx))
	}
	m.open = false
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
	if m.skill.ID == teleportSkillID && m.skill.Level >= 2 && m.hasSavePointChoice() {
		return teleportSavePointMap
	}
	return ""
}

func (m teleportModalState) hasSavePointChoice() bool {
	if m.skill.ID != teleportSkillID || m.skill.Level < 2 {
		return false
	}
	if len(m.mapNames) == 0 {
		return true
	}
	for _, name := range m.mapNames {
		if name != "" && name != teleportRandomMap {
			return true
		}
	}
	return false
}

func isLevelOneTeleportSkill(skill session.Skill) bool {
	return skill.ID == teleportSkillID && skill.Level <= 1
}

func (m teleportModalState) buttons() []teleportModalButton {
	if m.skill.ID == warpPortalSkillID {
		buttons := make([]teleportModalButton, 0, len(m.mapNames)+1)
		for i, name := range m.mapNames {
			if name == "" {
				continue
			}
			buttons = append(buttons, teleportModalButton{
				label:   warpPortalDestinationLabel(name, i),
				action:  teleportModalAction(i),
				mapName: name,
				enabled: true,
			})
		}
		buttons = append(buttons, teleportModalButton{label: "Cancel", action: teleportModalActionCancel, enabled: true})
		return buttons
	}
	return []teleportModalButton{
		{label: "Random", action: teleportModalActionRandom, mapName: m.randomMapName(), enabled: true},
		{label: "Save Point", action: teleportModalActionSavePoint, mapName: m.savePointMapName(), enabled: m.savePointEnabled()},
		{label: "Cancel", action: teleportModalActionCancel, enabled: true},
	}
}

func warpPortalDestinationLabel(mapName string, index int) string {
	if index == 0 {
		return fmt.Sprintf("Save Point: %s", mapName)
	}
	return mapName
}

func warpPointSkillLabel(skillID uint16) string {
	if skillID == warpPortalSkillID {
		return "warp portal"
	}
	return "teleport"
}

func (m *teleportModalState) draw(screen *render.Image, ctx Context, width, height int) {
	if !m.open || screen == nil {
		return
	}
	drawUISurface(screen, 0, 0, width, height, color.RGBA{A: 72}, color.RGBA{})
	x, y, w, h := teleportModalBounds(width, height)
	drawUITitledWindowFrame(screen, x, y, w, h, teleportModalTitleH)
	drawUIWindowTitle(screen, x, y, teleportModalTitleH, teleportModalPad, m.title(), uiTitleTextColor)
	render.DebugPrintAtColor(screen, "Choose destination.", x+teleportModalPad, y+teleportModalTitleH+12, uiTextColor)

	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i, button := range m.buttons() {
		bx, by, bw, bh := teleportModalButtonBounds(x, y, w, i)
		fill := uiButtonColor
		textColor := uiTextColor
		if !button.enabled {
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

func (m teleportModalState) title() string {
	if m.skill.ID == warpPortalSkillID {
		return "Warp Portal"
	}
	return "Teleport"
}

func (m *teleportModalState) cursorAction(ctx Context) (int, bool) {
	if !m.open || ctx.Input == nil {
		return 0, false
	}
	width, height := ctx.ScreenSize()
	x, y, w, _ := teleportModalBounds(width, height)
	for i, button := range m.buttons() {
		if !button.enabled {
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
	h := minInt(teleportModalHeightForButtons(5), maxInt(teleportModalMinH, height-40))
	x := (width - w) / 2
	y := (height - h) / 2
	if y < 16 {
		y = 16
	}
	return x, y, w, h
}

func teleportModalHeightForButtons(count int) int {
	if count < 3 {
		count = 3
	}
	return teleportModalTitleH + 34 + count*(teleportModalButtonH+teleportModalGap) + 18
}

func teleportModalButtonBounds(x, y, w, index int) (int, int, int, int) {
	bx := x + teleportModalPad
	by := y + teleportModalTitleH + 34 + index*(teleportModalButtonH+teleportModalGap)
	bw := w - 2*teleportModalPad
	return bx, by, bw, teleportModalButtonH
}
