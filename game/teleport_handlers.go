package game

import (
	"fmt"
	"github.com/kivutar/goro/client"
	"log"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	gameui "github.com/kivutar/goro/ui"
)

func (m *WorldMode) applyWarpPointList(ctx client.Context, list network.WarpPointList) {
	if list.SkillID != gameui.TeleportSkillID && list.SkillID != gameui.WarpPortalSkillID {
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
	if gameui.TeleportWarpListBypassesModal(skill, list) {
		m.autoSelectTeleportRandom(ctx, list)
		return
	}
	m.teleportModal.OpenWarpPointList(list, skill)
	m.status = fmt.Sprintf("choose %s destination", warpPointSkillLabel(list.SkillID))
	log.Printf("warp point destination list skill=%d maps=%v", list.SkillID, list.MapNames)
}

func (m *WorldMode) applyRememberWarpPointAck(_ client.Context, ack network.RememberWarpPointAck) {
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

func (m *WorldMode) autoSelectTeleportRandom(ctx client.Context, list network.WarpPointList) {
	mapName := gameui.TeleportRandomMap
	for _, name := range list.MapNames {
		if name == gameui.TeleportRandomMap {
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
	m.teleportModal.Reset()
	m.status = "teleporting"
	log.Printf("teleport random selected automatically skill=%d maps=%v", list.SkillID, list.MapNames)
}

func warpPointSkillLabel(skillID uint16) string {
	if skillID == gameui.WarpPortalSkillID {
		return "warp portal"
	}
	return "teleport"
}
