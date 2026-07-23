package game

import (
	"math"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

type skillUnitModel struct {
	unitID          uint16
	x               int
	y               int
	modelPath       string
	triggerEffectID int
}

func skillUnitModelSpec(unitID uint16) (db.SkillUnitModelSpec, bool) {
	spec, ok := db.SkillUnitModels[unitID]
	return spec, ok
}

func skillUnitModelFromEntry(entry network.SkillUnitEntry, spec db.SkillUnitModelSpec) skillUnitModel {
	return skillUnitModel{
		unitID:          entry.UnitID,
		x:               int(entry.X),
		y:               int(entry.Y),
		modelPath:       spec.ModelPath,
		triggerEffectID: spec.TriggerEffectID,
	}
}

func (m *WorldMode) applySkillUnitModelEntry(entry network.SkillUnitEntry) bool {
	spec, ok := skillUnitModelSpec(entry.UnitID)
	if !ok {
		return false
	}
	unit := skillUnitModelFromEntry(entry, spec)
	if entry.Visible {
		if m.skillUnitModels == nil {
			m.skillUnitModels = make(map[uint32]skillUnitModel)
		}
		m.skillUnitModels[entry.ID] = unit
		if m.hiddenSkillUnits != nil {
			delete(m.hiddenSkillUnits, entry.ID)
		}
		glog.Debugf("skill unit model unit=%d id=%d creator=%d cell=%d,%d model=%s", entry.UnitID, entry.ID, entry.CreatorID, entry.X, entry.Y, unit.modelPath)
		return true
	}
	if m.hiddenSkillUnits == nil {
		m.hiddenSkillUnits = make(map[uint32]skillUnitModel)
	}
	m.hiddenSkillUnits[entry.ID] = unit
	if m.skillUnitModels != nil {
		delete(m.skillUnitModels, entry.ID)
	}
	glog.Debugf("hidden skill unit model unit=%d id=%d creator=%d cell=%d,%d model=%s", entry.UnitID, entry.ID, entry.CreatorID, entry.X, entry.Y, unit.modelPath)
	return true
}

func (m *WorldMode) applySkillUnitUpdate(update network.SkillUnitUpdate) {
	if update.ID == 0 || m.hiddenSkillUnits == nil {
		return
	}
	unit, ok := m.hiddenSkillUnits[update.ID]
	if !ok {
		return
	}
	delete(m.hiddenSkillUnits, update.ID)
	if m.skillUnitModels == nil {
		m.skillUnitModels = make(map[uint32]skillUnitModel)
	}
	m.skillUnitModels[update.ID] = unit
	glog.Debugf("hidden skill unit model revealed id=%d unit=%d cell=%d,%d model=%s", update.ID, unit.unitID, unit.x, unit.y, unit.modelPath)
}

func (m *WorldMode) applyUsedTrapLookChange(ctx client.Context, look network.ActorLookChange) bool {
	if look.Type != 0 || look.Value != uint32(db.SkillUnitUsedTraps) || look.ID == 0 {
		return false
	}
	unit, ok := m.removeSkillUnitModel(look.ID)
	if !ok {
		return false
	}
	if unit.triggerEffectID > 0 {
		now := time.Now()
		if m.addWorldEffectAtCellLifetime(ctx, unit.triggerEffectID, 0, unit.x, unit.y, now, 0, false) {
			glog.Debugf("skill unit trap triggered id=%d unit=%d cell=%d,%d effect=%d", look.ID, unit.unitID, unit.x, unit.y, unit.triggerEffectID)
		}
	}
	return true
}

func (m *WorldMode) removeSkillUnitModel(id uint32) (skillUnitModel, bool) {
	if m.skillUnitModels != nil {
		if unit, ok := m.skillUnitModels[id]; ok {
			delete(m.skillUnitModels, id)
			return unit, true
		}
	}
	if m.hiddenSkillUnits != nil {
		if unit, ok := m.hiddenSkillUnits[id]; ok {
			delete(m.hiddenSkillUnits, id)
			return unit, true
		}
	}
	return skillUnitModel{}, false
}

func (m *WorldMode) removeSkillUnitModelOnly(id uint32) bool {
	removed := false
	if m.skillUnitModels != nil {
		if _, ok := m.skillUnitModels[id]; ok {
			delete(m.skillUnitModels, id)
			removed = true
		}
	}
	if m.hiddenSkillUnits != nil {
		if _, ok := m.hiddenSkillUnits[id]; ok {
			delete(m.hiddenSkillUnits, id)
			removed = true
		}
	}
	return removed
}

func (m *WorldMode) skillUnitModelCell(id uint32) (int, int, bool) {
	if m.skillUnitModels != nil {
		if unit, ok := m.skillUnitModels[id]; ok {
			return unit.x, unit.y, true
		}
	}
	if m.hiddenSkillUnits != nil {
		if unit, ok := m.hiddenSkillUnits[id]; ok {
			return unit.x, unit.y, true
		}
	}
	return 0, 0, false
}

func (m *WorldMode) drawSkillUnitRSMModels(screen *render.Frame, ctx client.Context, projection sceneProjection, now time.Time) {
	if screen == nil || ctx.World == nil || ctx.World.GND == nil || ctx.Resources == nil || len(m.skillUnitModels) == 0 {
		return
	}
	radius := rsmRenderRadius() * 2
	for id, unit := range m.skillUnitModels {
		worldX := cellCenter(float64(unit.x))
		worldY := cellCenter(float64(unit.y))
		if math.Abs(worldX-projection.playerX) > radius || math.Abs(worldY-projection.playerY) > radius {
			continue
		}
		rsm := m.runtimeRSMModel(ctx.Resources, unit.modelPath)
		if rsm == nil {
			continue
		}
		placement := skillUnitRSMPlacement(ctx.World, id, unit)
		frame, _ := rsmAnimationFrame(rsm, placement.model, now)
		m.drawAnimatedRSMPlacement(screen, ctx.Resources, ctx.World.RSW, rsm, placement, frame)
	}
}

func (m *WorldMode) runtimeRSMModel(manager *res.Manager, modelPath string) *res.RSM {
	if manager == nil || modelPath == "" {
		return nil
	}
	if m.runtimeRSMModels == nil {
		m.runtimeRSMModels = make(map[string]*res.RSM)
	}
	if rsm, ok := m.runtimeRSMModels[modelPath]; ok {
		return rsm
	}
	rsm, err := loadRSMModel(manager, modelPath)
	if err != nil {
		glog.Warnf("runtime rsm model unavailable model=%s: %v", modelPath, err)
		m.runtimeRSMModels[modelPath] = nil
		return nil
	}
	m.runtimeRSMModels[modelPath] = rsm
	return rsm
}

func skillUnitRSMPlacement(world *worldstate.World, id uint32, unit skillUnitModel) visibleRSMPlacement {
	baseX := cellCenter(float64(unit.x))
	baseY := cellCenter(float64(unit.y))
	return visibleRSMPlacement{
		index: int(id),
		model: res.RSWModel{
			Name:     "skill_unit",
			Filename: unit.modelPath,
			Position: res.RSWVector3{
				Y: float32(terrainHeightAt(world, float64(unit.x), float64(unit.y))),
			},
			Scale: res.RSWVector3{X: 1, Y: 1, Z: 1},
		},
		baseX: baseX,
		baseY: baseY,
	}
}
