package game

import (
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
)

var quadHornUVs = []texturePoint{
	{u: 0.0, v: 0}, {u: 0.0, v: 1}, {u: 0.2, v: 1},
	{u: 0.2, v: 0}, {u: 0.2, v: 1}, {u: 0.4, v: 1},
	{u: 0.4, v: 0}, {u: 0.4, v: 1}, {u: 0.6, v: 1},
	{u: 0.6, v: 0}, {u: 0.6, v: 1}, {u: 0.8, v: 1},
}

func (m *WorldMode) drawQuadHornEffect(screen *render.Frame, ctx client.Context, effect worldEffect, component worldEffectComponent, componentIndex int, worldX, worldY, worldZ, progress float64, componentDuration time.Duration) {
	if screen == nil || ctx.Resources == nil || component.textureFile == "" {
		return
	}
	texture := m.effectFileTexture(ctx.Resources, component.textureFile)
	if texture == nil {
		return
	}
	alpha := effectComponentAlpha(progress, component)
	if alpha <= 0 {
		return
	}

	salt := componentIndex * 1009
	height := quadHornRange(effect, salt+1, component.quadHornHeightMin, component.quadHornHeightMax)
	bottomSize := quadHornRange(effect, salt+2, component.quadHornBottomMin, component.quadHornBottomMax)
	if height <= 0 || bottomSize <= 0 {
		return
	}
	offsetX := quadHornRange(effect, salt+3, component.quadHornOffsetXMin, component.quadHornOffsetXMax)
	offsetY := quadHornRange(effect, salt+4, component.quadHornOffsetYMin, component.quadHornOffsetYMax)
	offsetZ := component.quadHornOffsetZ
	offsetX = quadHornDefaultOffset(offsetX)
	offsetY = quadHornDefaultOffset(offsetY)
	offsetZ = quadHornDefaultOffset(offsetZ)
	height, offsetZ = quadHornAnimation(component, progress, componentDuration, height, offsetZ)

	rotateX := quadHornRange(effect, salt+5, component.quadHornRotateXMin, component.quadHornRotateXMax)
	rotateY := quadHornRange(effect, salt+6, component.quadHornRotateYMin, component.quadHornRotateYMax)
	rotateZ := quadHornRange(effect, salt+7, component.quadHornRotateZMin, component.quadHornRotateZMax)
	x := worldX + component.posX
	y := worldY + component.posY
	z := worldZ + component.posZ
	if ctx.World != nil && (component.posX != 0 || component.posY != 0) {
		z = terrainHeightAtRenderPoint(ctx.World, x, y) + 0.07 + component.posZ
	}
	origin := modelPoint3{x: x + offsetX, y: z + height*0.9 + offsetZ, z: y + offsetY}

	locals := []modelPoint3{
		{x: 0, y: height, z: 0},
		{x: -bottomSize, y: -height, z: bottomSize},
		{x: bottomSize, y: -height, z: bottomSize},

		{x: 0, y: height, z: 0},
		{x: bottomSize, y: -height, z: bottomSize},
		{x: bottomSize, y: -height, z: -bottomSize},

		{x: 0, y: height, z: 0},
		{x: bottomSize, y: -height, z: -bottomSize},
		{x: -bottomSize, y: -height, z: -bottomSize},

		{x: 0, y: height, z: 0},
		{x: -bottomSize, y: -height, z: -bottomSize},
		{x: -bottomSize, y: -height, z: bottomSize},
	}

	bounds := texture.Bounds()
	textureWidth := float32(bounds.Dx())
	textureHeight := float32(bounds.Dy())
	tint := effectComponentTint(component, alpha)
	vertices := make([]render.Vertex3D, 0, len(locals))
	indices := make([]uint16, 0, len(locals))
	for i, local := range locals {
		point := add3(origin, rotateEffectCylinderVector(local, rotateX, rotateY, rotateZ))
		vertices = append(vertices, texturedSurfaceVertex3D(point, quadHornUVs[i], tint, textureWidth, textureHeight))
		indices = append(indices, uint16(i))
	}

	options := triangleDrawOptions(render.FilterLinear, render.AddressRepeat)
	if component.blendAdditive || component.blendMode == 2 {
		options.Blend = render.BlendLighter
	}
	screen.DrawTriangles3DOwned(vertices, indices, texture, options)
}

func quadHornRange(effect worldEffect, salt int, min, max float64) float64 {
	if max < min {
		return max
	}
	return deterministicFloatRange(effect, salt, min, max)
}

func quadHornDefaultOffset(value float64) float64 {
	if value == 0 {
		return 0.5
	}
	return value
}

func quadHornAnimation(component worldEffectComponent, progress float64, componentDuration time.Duration, height, offsetZ float64) (float64, float64) {
	speed := component.quadHornAnimSpeed
	if speed <= 0 {
		speed = 100 * time.Millisecond
	}
	elapsed := time.Duration(progress * float64(componentDuration))
	units := float64(elapsed) / float64(speed)

	switch component.animation {
	case 1:
		if units < height {
			height = units
		}
	case 2:
		if units <= offsetZ {
			offsetZ = units
		}
	case 3:
		if units < height/2 {
			offsetZ = units
		}
	}

	if component.quadHornAnimOut && componentDuration > speed {
		outStarts := componentDuration - speed
		if elapsed > outStarts {
			outUnits := float64(elapsed-outStarts) / float64(speed)
			offsetZ = -(outUnits * height)
		}
	}
	return height, offsetZ
}
