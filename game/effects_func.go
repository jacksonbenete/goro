package game

import (
	"image/color"
	"math"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
)

type effectFuncAdapter int

const (
	effectFuncUnknown effectFuncAdapter = iota
	effectFuncGroundSample
	effectFuncCastRing
)

func (m *WorldMode) drawFuncEffect(screen *render.Frame, ctx client.Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, componentIndex int, worldX, worldY, worldZ, progress float64, now time.Time) {
	switch component.funcAdapter {
	case effectFuncGroundSample:
		m.drawGroundPlaneEffect(screen, ctx, component, effect, worldX, worldY, progress, now)
	case effectFuncCastRing:
		m.drawCastRingEffect(screen, ctx, component, effect, componentIndex, worldX, worldY, worldZ, progress)
	default:
	}
}

func (m *WorldMode) drawCastRingEffect(screen *render.Frame, ctx client.Context, component worldEffectComponent, effect worldEffect, componentIndex int, x, y, z, progress float64) {
	alpha := effectComponentAlpha(progress, component)
	if alpha <= 0 {
		return
	}
	texture := m.effectTexture(ctx.Resources, component.textureName)
	tint := effectComponentTint(component, alpha)
	z += component.posZ
	angleOffset := 0.0
	if component.rotate {
		angleOffset = progress * 2 * math.Pi
		angleOffset += deterministicAngle(effect, componentIndex+17) * 0.08
	}
	drawWorldSoftRing(screen, m.whitePixel, x, y, z+0.015, component.bottomSize*1.15, 0.16, withAlpha(component.color, alpha*0.75), maxInt(component.circleSides, 20))
	if texture != nil {
		drawWorldCylinderBandRotated(screen, m.whitePixel, texture, x, y, z+0.035, component.bottomSize, component.topSize, component.height, tint, maxInt(component.circleSides, component.totalCircleSides), angleOffset)
		return
	}
	drawWorldCylinderBandRotated(screen, m.whitePixel, nil, x, y, z+0.035, component.bottomSize, component.topSize, component.height, tint, maxInt(component.circleSides, component.totalCircleSides), angleOffset)
}

func drawWorldCylinderBandRotated(screen *render.Frame, white, texture *render.Image, x, y, z, bottomRadius, topRadius, height float64, c color.RGBA, segments int, angleOffset float64) {
	if segments < 3 || bottomRadius <= 0.01 || topRadius <= 0.01 || height <= 0.01 || c.A == 0 {
		return
	}
	vertices := make([]render.Vertex3D, 0, (segments+1)*2)
	indices := make([]uint16, 0, segments*6)
	tint := c
	srcW, srcH := float32(1), float32(1)
	source := white
	if texture != nil {
		source = texture
		bounds := texture.Bounds()
		srcW = float32(bounds.Dx())
		srcH = float32(bounds.Dy())
	}
	for i := 0; i <= segments; i++ {
		u := float32(i) / float32(segments)
		angle := angleOffset + float64(i)*2*math.Pi/float64(segments)
		cosine := math.Cos(angle)
		sine := math.Sin(angle)
		topAngle := angleOffset + (float64(i)+0.5)*2*math.Pi/float64(segments)
		topCosine := math.Cos(topAngle)
		topSine := math.Sin(topAngle)
		vertices = append(vertices,
			warpEffectTexturedVertex3D(x+cosine*bottomRadius, y+sine*bottomRadius, z, u*srcW, srcH, tint),
			warpEffectTexturedVertex3D(x+topCosine*topRadius, y+topSine*topRadius, z+height, u*srcW, 0, tint),
		)
		if i == segments {
			continue
		}
		base := uint16(i * 2)
		indices = append(indices, base, base+1, base+3, base, base+3, base+2)
	}
	options := triangleDrawOptions(render.FilterLinear, render.AddressRepeat)
	options.Blend = render.BlendLighter
	screen.DrawTriangles3D(vertices, indices, source, options)
}

func (m *WorldMode) drawGroundPlaneEffect(screen *render.Frame, ctx client.Context, component worldEffectComponent, effect worldEffect, worldX, worldY, progress float64, now time.Time) {
	texture := m.effectFileTexture(ctx.Resources, component.textureFile)
	if texture == nil || ctx.World == nil {
		return
	}
	alpha := effectBillboardAlpha(progress, component)
	if alpha <= 0 {
		return
	}
	size := component.sizeStart
	if effect.size > 0 {
		size = effect.size
	}
	if size <= 0 {
		size = 1
	}
	half := size * 0.5
	angle := now.Sub(effect.starts).Seconds() * 40 * math.Pi / 180
	uv := func(u, v float64) texturePoint {
		sinA, cosA := math.Sin(angle), math.Cos(angle)
		x, y := u-0.5, v-0.5
		return texturePoint{
			u: float32(x*cosA - y*sinA + 0.5),
			v: float32(x*sinA + y*cosA + 0.5),
		}
	}
	point := func(dx, dy float64) modelPoint3 {
		x := worldX + dx
		y := worldY + dy
		return modelPoint3{x: x, y: terrainHeightAt(ctx.World, x-0.5, y-0.5) + component.posZ, z: y}
	}
	verts := [4]modelPoint3{
		point(-half, -half),
		point(half, -half),
		point(half, half),
		point(-half, half),
	}
	uvs := [4]texturePoint{
		uv(0, 0),
		uv(1, 0),
		uv(1, 1),
		uv(0, 1),
	}
	tint := effectComponentTint(component, alpha)
	drawTexturedSurface3DAlpha(screen, texture, verts, uvs, quadIndices012023, [4]color.RGBA{tint, tint, tint, tint})
}

func warpEffectTexturedVertex3D(x, y, z float64, srcX, srcY float32, c color.RGBA) render.Vertex3D {
	point := modelPoint3{x: x, y: z, z: y}
	return render.Vertex3D{
		X:      float32(point.x),
		Y:      float32(point.y),
		Z:      float32(point.z),
		SrcX:   srcX,
		SrcY:   srcY,
		ColorR: float32(c.R) / 255,
		ColorG: float32(c.G) / 255,
		ColorB: float32(c.B) / 255,
		ColorA: float32(c.A) / 255,
		DepthX: float32(point.x),
		DepthY: float32(point.y),
		DepthZ: float32(point.z),
	}
}

func warpEffectVertex3D(x, y, z float64, c color.RGBA) render.Vertex3D {
	return warpEffectTexturedVertex3D(x, y, z, 0, 0, c)
}

func drawWarpZoneEffect(screen *render.Frame, white, ringTexture *render.Image, x, y, z float64, now time.Time) {
	const (
		segments       = 64
		ringCount      = 4
		baseRadius     = 0.25
		radiusRange    = 1.18
		bandWidth      = 0.34
		cycleSeconds   = 4.0
		bottomBaseSize = 0.95
		topBaseSize    = 1.58
		heightBase     = 1.10
		groundLift     = 0.04
	)
	z += groundLift
	seconds := float64(now.UnixNano()) / float64(time.Second)

	for i := 0; i < ringCount; i++ {
		phase := math.Mod(seconds+float64(i), cycleSeconds) / cycleSeconds
		sizeFactor := 1 - phase
		heightFactor := phase * 2
		if phase > 0.5 {
			heightFactor = (1 - phase) * 2
		}
		alpha := uint8(102 * warpCycleFade(phase))
		drawWorldCylinderBand(screen, white, ringTexture, x, y, z, bottomBaseSize*sizeFactor, topBaseSize*sizeFactor, heightBase*heightFactor, color.RGBA{R: 155, G: 205, B: 255, A: alpha}, segments)
	}
	drawWorldRadialGradient(screen, white, x, y, z, 0.18, 0.85, color.RGBA{R: 170, G: 210, B: 255, A: 54}, segments)
	for i := 0; i < ringCount; i++ {
		phase := math.Mod(seconds*0.55+float64(i)/ringCount, 1)
		radius := baseRadius + phase*radiusRange
		alpha := uint8(155 * (1 - phase))
		if alpha < 28 {
			alpha = 28
		}
		drawWorldSoftRing(screen, white, x, y, z, radius, bandWidth, color.RGBA{R: 185, G: 215, B: 255, A: alpha}, segments)
	}
	pulse := 0.5 + 0.5*math.Sin(seconds*2.4)
	drawWorldSoftRing(screen, white, x, y, z, 0.35+pulse*0.06, 0.26, color.RGBA{R: 235, G: 245, B: 255, A: 150}, segments)
}

func warpCycleFade(phase float64) float64 {
	switch {
	case phase < 0.25:
		return phase / 0.25
	case phase > 0.75:
		return (1 - phase) / 0.25
	default:
		return 1
	}
}

func drawWorldRadialGradient(screen *render.Frame, white *render.Image, x, y, z, innerRadius, outerRadius float64, c color.RGBA, segments int) {
	drawWorldRingBand(screen, white, x, y, z, innerRadius, outerRadius, c.A, 0, c, segments)
}

func drawWorldSoftRing(screen *render.Frame, white *render.Image, x, y, z, radius, width float64, c color.RGBA, segments int) {
	inner := math.Max(0, radius-width*0.5)
	mid := math.Max(inner+0.01, radius)
	outer := math.Max(mid+0.01, radius+width*0.5)
	drawWorldRingBand(screen, white, x, y, z, inner, mid, 0, c.A, c, segments)
	drawWorldRingBand(screen, white, x, y, z, mid, outer, c.A, 0, c, segments)
}

func drawWorldRingBand(screen *render.Frame, white *render.Image, x, y, z, innerRadius, outerRadius float64, innerAlpha, outerAlpha uint8, c color.RGBA, segments int) {
	if segments < 3 || outerRadius <= innerRadius {
		return
	}
	vertices := make([]render.Vertex3D, 0, (segments+1)*2)
	indices := make([]uint16, 0, segments*6)
	innerColor := c
	outerColor := c
	innerColor.A = innerAlpha
	outerColor.A = outerAlpha
	for i := 0; i <= segments; i++ {
		angle := float64(i) * 2 * math.Pi / float64(segments)
		cosine := math.Cos(angle)
		sine := math.Sin(angle)
		vertices = append(vertices,
			warpEffectVertex3D(x+cosine*innerRadius, y+sine*innerRadius, z, innerColor),
			warpEffectVertex3D(x+cosine*outerRadius, y+sine*outerRadius, z, outerColor),
		)
		if i == segments {
			continue
		}
		base := uint16(i * 2)
		indices = append(indices, base, base+1, base+3, base, base+3, base+2)
	}
	options := triangleDrawOptions(render.FilterNearest, render.AddressUnsafe)
	options.Blend = render.BlendLighter
	screen.DrawTriangles3D(vertices, indices, white, options)
}
