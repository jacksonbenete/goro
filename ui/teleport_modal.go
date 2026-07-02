package ui

import (
	"fmt"
	"image/color"

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

const (
	TeleportSkillID      = teleportSkillID
	WarpPortalSkillID    = warpPortalSkillID
	TeleportRandomMap    = teleportRandomMap
	TeleportSavePointMap = teleportSavePointMap
	WarpPointCancelMap   = warpPointCancelMap
)

type TeleportModal struct {
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

func (m *TeleportModal) OpenWarpPointList(list network.WarpPointList, skill session.Skill) {
	m.open = true
	m.skill = skill
	m.mapNames = append(m.mapNames[:0], list.MapNames...)
	m.status = ""
}

func (m *TeleportModal) Reset() {
	*m = TeleportModal{}
}

func TeleportWarpListBypassesModal(skill session.Skill, list network.WarpPointList) bool {
	if list.SkillID != teleportSkillID {
		return false
	}
	if IsLevelOneTeleportSkill(skill) {
		return true
	}
	for _, name := range list.MapNames {
		if name != "" && name != teleportRandomMap {
			return false
		}
	}
	return true
}

func (m *TeleportModal) Update(ctx Context, actions GameActions) bool {
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
		m.activate(ctx, actions, button.action)
		return true
	}
	return true
}

func (m *TeleportModal) activate(ctx Context, actions GameActions, action teleportModalAction) {
	for _, button := range m.buttons() {
		if button.action != action || !button.enabled {
			continue
		}
		if button.action == teleportModalActionCancel {
			m.cancel(ctx)
			return
		}
		m.selectWarpPoint(ctx, actions, button.mapName)
		return
	}
}

func (m *TeleportModal) cancel(ctx Context) {
	if m.skill.ID == warpPortalSkillID && ctx.Network != nil {
		if err := ctx.Network.SendSelectWarpPoint(uint16(warpPortalSkillID), warpPointCancelMap); err != nil {
			m.status = fmt.Sprintf("Cancel failed: %v", err)
			return
		}
	}
	m.open = false
}

func (m *TeleportModal) selectWarpPoint(ctx Context, actions GameActions, mapName string) {
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
	if actions != nil && skillID == teleportSkillID {
		actions.AddTeleportEffect(ctx)
	}
	m.open = false
}

func (m TeleportModal) savePointEnabled() bool {
	return m.savePointMapName() != ""
}

func (m TeleportModal) randomMapName() string {
	for _, name := range m.mapNames {
		if name == teleportRandomMap {
			return name
		}
	}
	return teleportRandomMap
}

func (m TeleportModal) savePointMapName() string {
	if m.skill.ID == teleportSkillID && m.skill.Level >= 2 && m.hasSavePointChoice() {
		return teleportSavePointMap
	}
	return ""
}

func (m TeleportModal) hasSavePointChoice() bool {
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

func IsLevelOneTeleportSkill(skill session.Skill) bool {
	return skill.ID == teleportSkillID && skill.Level <= 1
}

func (m TeleportModal) buttons() []teleportModalButton {
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

func (m *TeleportModal) Draw(screen *render.Image, ctx Context, width, height int) {
	if !m.open || screen == nil {
		return
	}
	DrawSurface(screen, 0, 0, width, height, color.RGBA{A: 72}, color.RGBA{})
	x, y, w, h := teleportModalBounds(width, height)
	DrawTitledWindowFrame(screen, x, y, w, h, teleportModalTitleH)
	DrawWindowTitle(screen, x, y, teleportModalTitleH, teleportModalPad, m.Title(), TitleTextColor)
	render.DebugPrintAtColor(screen, "Choose destination.", x+teleportModalPad, y+teleportModalTitleH+12, TextColor)

	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i, button := range m.buttons() {
		bx, by, bw, bh := teleportModalButtonBounds(x, y, w, i)
		fill := ButtonColor
		textColor := TextColor
		if !button.enabled {
			fill = DisabledColor
			textColor = MutedTextColor
		} else if pointInRect(mx, my, bx, by, bw, bh) {
			fill = ButtonHoverColor
		}
		DrawButtonLabel(screen, bx, by, bw, bh, button.label, fill, textColor)
	}
	if m.status != "" {
		render.DebugPrintAtColor(screen, trimRunes(m.status, 30), x+teleportModalPad, y+h-16, ErrorTextColor)
	}
}

func (m TeleportModal) Title() string {
	if m.skill.ID == warpPortalSkillID {
		return "Warp Portal"
	}
	return "Teleport"
}

func (m *TeleportModal) CursorAction(ctx Context) (int, bool) {
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
			return CursorActionClick, true
		}
	}
	return CursorActionDefault, true
}

func (m TeleportModal) IsOpen() bool {
	return m.open
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
