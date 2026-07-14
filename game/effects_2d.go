package game

import (
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
)

func (m *WorldMode) draw2DEffect(screen *render.Frame, ctx client.Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, componentIndex int, worldX, worldY, worldZ, progress float64, now time.Time) {
	texture := m.effectFileTexture(ctx.Resources, component.textureFile)
	if texture == nil {
		return
	}
	salt := componentIndex * 1009
	if component.durationRandMax > 0 {
		duration := deterministicDurationRange(effect, salt+19, component.durationRandMin, component.durationRandMax)
		progress = worldEffectComponentProgress(effect.starts.Add(component.delay), duration, now)
		if progress >= 1 {
			return
		}
	}
	alpha := effectBillboardAlpha(progress, component)
	if alpha <= 0 {
		return
	}
	offsetX, offsetY, offsetZ := m.effect3DOffset(ctx, component, effect, salt, 0, progress, worldX, worldY, worldZ)
	sizeX, sizeY := effect3DSize(component, effect, salt, progress, 0)
	if sizeX <= 0 || sizeY <= 0 {
		return
	}
	angle := worldEffectBillboardAngleForEffect(component, projection, effect, salt, progress)
	drawTexturedEffectBillboardRotatedXY(screen, projection, texture, worldX+offsetX, worldY+offsetY, worldZ+offsetZ, sizeX, sizeY, angle, effectComponentTint(component, alpha), true)
}
