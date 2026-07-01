package game

import (
	"math"
	"time"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

func (m *WorldMode) drawRSWEffects(screen *render.Image, ctx Context, projection sceneProjection, now time.Time) {
	if screen == nil || ctx.World == nil || ctx.World.RSW == nil || ctx.World.GND == nil {
		return
	}
	radius := rsmRenderRadius()
	for index, effect := range ctx.World.RSW.Effects {
		effectID := int(effect.ID)
		spec, ok := worldEffectSpecForID(effectID)
		if !ok {
			continue
		}
		worldX, worldY, worldZ := rswEffectWorldPosition(ctx.World.GND, effect)
		if math.Abs(worldX-projection.playerX) > radius*2 || math.Abs(worldY-projection.playerY) > radius*2 {
			continue
		}
		m.drawRSWEffect(screen, ctx, projection, spec, effectID, index, effect, worldX, worldY, worldZ, now)
	}
}

func (m *WorldMode) drawRSWEffect(screen *render.Image, ctx Context, projection sceneProjection, spec worldEffectSpec, effectID int, index int, rswEffect res.RSWEffect, worldX, worldY, worldZ float64, now time.Time) {
	effect := worldEffect{
		effectID: effectID,
		x:        int(math.Round(worldX - 0.5)),
		y:        int(math.Round(worldY - 0.5)),
		expires:  now.Add(24 * time.Hour),
		duration: 24 * time.Hour,
	}
	for componentIndex, component := range spec.components {
		component = mapEffectComponent(component)
		duration := m.worldEffectResolvedComponentDuration(ctx.Resources, spec, component)
		if duration <= 0 {
			duration = 600 * time.Millisecond
		}
		starts := loopingRSWEffectStart(rswEffect, index, duration, now)
		effect.starts = starts
		progress := worldEffectComponentProgress(starts, duration, now)
		m.drawWorldEffectComponent(screen, ctx, projection, effect, component, componentIndex, worldX, worldY, worldZ, progress, now)
	}
}

func mapEffectComponent(component worldEffectComponent) worldEffectComponent {
	if component.kind == effectComponent3D && component.spriteFile != "" {
		component.worldSizedSprite = true
	}
	return component
}

func rswEffectWorldPosition(gnd *res.GND, effect res.RSWEffect) (float64, float64, float64) {
	if gnd == nil {
		return float64(effect.Position.X), float64(effect.Position.Z), float64(effect.Position.Y) + 1
	}
	return float64(effect.Position.X) + float64(gnd.Width), float64(effect.Position.Z) + float64(gnd.Height), float64(effect.Position.Y) + 1
}

func loopingRSWEffectStart(effect res.RSWEffect, index int, duration time.Duration, now time.Time) time.Time {
	if duration <= 0 {
		return now
	}
	phase := time.Duration(effect.Delay*float32(time.Millisecond)) + time.Duration(index%17)*17*time.Millisecond
	origin := time.Unix(0, 0).Add(phase)
	elapsed := now.Sub(origin)
	if elapsed < 0 {
		return origin
	}
	return now.Add(-(elapsed % duration))
}
