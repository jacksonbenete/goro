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

func teleportCylinderComponent(bottomSize, topSize, height float64) worldEffectComponent {
	return worldEffectComponent{
		kind:             effectComponentCylinder,
		color:            color.RGBA{R: 153, G: 153, B: 255, A: 255},
		textureName:      "ring_blue",
		duration:         1500 * time.Millisecond,
		alphaMax:         0.5,
		fade:             true,
		rotate:           true,
		animation:        5,
		bottomSize:       bottomSize,
		topSize:          topSize,
		height:           height,
		totalCircleSides: 32,
		circleSides:      32,
		blendAdditive:    true,
	}
}

func readyPortalCylinderComponent() worldEffectComponent {
	return worldEffectComponent{
		kind:             effectComponentCylinder,
		color:            color.RGBA{R: 153, G: 153, B: 255, A: 255},
		textureName:      "ring_blue",
		duration:         500 * time.Millisecond,
		repeat:           true,
		repeatDelay:      -300 * time.Millisecond,
		alphaMax:         0.4,
		fadeOut:          true,
		rotate:           true,
		animation:        4,
		bottomSize:       2.4,
		topSize:          3.9,
		height:           0.1,
		posZ:             0.1,
		totalCircleSides: 32,
		circleSides:      32,
		blendAdditive:    true,
	}
}

func portalCylinderComponent(bottomSize, topSize, height, posZ float64, textureName string, alphaMax float64) worldEffectComponent {
	return worldEffectComponent{
		kind:             effectComponentCylinder,
		color:            color.RGBA{R: 153, G: 153, B: 255, A: 255},
		textureName:      textureName,
		duration:         25000 * time.Millisecond,
		alphaMax:         alphaMax,
		fade:             true,
		rotate:           true,
		animation:        0,
		bottomSize:       bottomSize,
		topSize:          topSize,
		height:           height,
		posZ:             posZ,
		totalCircleSides: 32,
		circleSides:      32,
		blendAdditive:    true,
	}
}

func healCylinderComponent(bottomSize, topSize, height float64) worldEffectComponent {
	return worldEffectComponent{
		kind:             effectComponentCylinder,
		color:            color.RGBA{R: 178, G: 255, B: 178, A: 255},
		textureName:      "ring_white",
		duration:         1500 * time.Millisecond,
		alphaMax:         0.2,
		fade:             true,
		rotate:           true,
		animation:        1,
		bottomSize:       bottomSize,
		topSize:          topSize,
		height:           height,
		totalCircleSides: 32,
		circleSides:      32,
		blendAdditive:    true,
	}
}

func healOffensiveCylinderComponent(bottomSize, topSize, height float64) worldEffectComponent {
	component := healCylinderComponent(bottomSize, topSize, height)
	component.duration = time.Second
	component.color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	component.blendAdditive = true
	return component
}

func healParticleComponent(alpha float64, duration, delay, duplicateDelay time.Duration, duplicate int, posXRand, posYRand, posZStartRand, posZStartMiddle, posZEnd, posZEndRand, posZEndMiddle float64, sparkling bool) worldEffectComponent {
	component := worldEffectComponent{
		kind:            effectComponent3D,
		color:           color.RGBA{R: 255, G: 255, B: 255, A: 255},
		textureFile:     "effect/pok3.tga",
		duration:        duration,
		delay:           delay,
		duplicateDelay:  duplicateDelay,
		alphaMax:        alpha,
		sparkling:       sparkling,
		fadeIn:          true,
		fadeOut:         true,
		posXRand:        posXRand,
		posYRand:        posYRand,
		posZStartRand:   posZStartRand,
		posZStartMiddle: posZStartMiddle,
		posZEnd:         posZEnd,
		posZEndRand:     posZEndRand,
		posZEndMiddle:   posZEndMiddle,
		sizeStart:       9 * effectPixelRatio,
		sizeEnd:         9 * effectPixelRatio,
		sizeRand:        2 * effectPixelRatio,
		duplicate:       duplicate,
		blendAdditive:   true,
	}
	if sparkling {
		component.sparkNumber = 2
	}
	return component
}

func (m *WorldMode) drawFuncEffect(screen *render.Image, ctx client.Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, componentIndex int, worldX, worldY, worldZ, progress float64, now time.Time) {
	switch component.funcAdapter {
	case effectFuncGroundSample:
		m.drawGroundPlaneEffect(screen, ctx, component, effect, worldX, worldY, progress, now)
	case effectFuncCastRing:
		m.drawCastRingEffect(screen, ctx, component, effect, componentIndex, worldX, worldY, worldZ, progress)
	default:
	}
}

func (m *WorldMode) drawCastRingEffect(screen *render.Image, ctx client.Context, component worldEffectComponent, effect worldEffect, componentIndex int, x, y, z, progress float64) {
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

func drawWorldCylinderBandRotated(screen, white, texture *render.Image, x, y, z, bottomRadius, topRadius, height float64, c color.RGBA, segments int, angleOffset float64) {
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

func (m *WorldMode) drawCylinderEffect(screen *render.Image, ctx client.Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, componentIndex int, x, y, z, progress float64) {
	texture := m.effectTexture(ctx.Resources, component.textureName)
	if texture == nil {
		return
	}
	alpha := effectComponentAlpha(progress, component)
	if alpha <= 0 {
		return
	}
	topSize := component.topSize
	if component.animation == 2 {
		topSize *= progress
	}
	bottomSize := component.bottomSize
	height := component.height
	if component.animation == 4 {
		bottomSize *= progress
		topSize *= progress
	}
	if !component.fixedPerspective {
		drawWorldCylinderBand(screen, m.whitePixel, texture, x, y, z+component.posZ, bottomSize, topSize, height, effectComponentTint(component, alpha), maxInt(component.circleSides, component.totalCircleSides))
		return
	}
	duplicates := maxInt(component.duplicate, 1)
	for i := 0; i < duplicates; i++ {
		angle := 0.0
		if component.rotate {
			angle += progress * 2 * math.Pi
			angle += deterministicAngle(effect, componentIndex*101+i+31) * 0.08
		}
		if component.angleZRandom != 0 {
			angle += deterministicAngle(effect, componentIndex*101+i) * component.angleZRandom / 360
		}
		drawTexturedEffectCylinder(screen, projection, texture, x, y, z+component.posZ, effectCylinderDraw{
			bottomSize:       bottomSize,
			topSize:          topSize,
			totalCircleSides: component.totalCircleSides,
			circleSides:      component.circleSides,
			alpha:            alpha,
			angle:            angle,
		})
	}
}

func (m *WorldMode) drawGroundPlaneEffect(screen *render.Image, ctx client.Context, component worldEffectComponent, effect worldEffect, worldX, worldY, progress float64, now time.Time) {
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

type effectCylinderDraw struct {
	bottomSize       float64
	topSize          float64
	totalCircleSides int
	circleSides      int
	alpha            float64
	angle            float64
}

func drawTexturedEffectCylinder(screen *render.Image, projection sceneProjection, texture *render.Image, worldX, worldY, worldZ float64, draw effectCylinderDraw) {
	if screen == nil || texture == nil || draw.alpha <= 0 || draw.topSize <= 0 || draw.totalCircleSides <= 0 || draw.circleSides <= 0 {
		return
	}
	right, up, _, ok := projection.BillboardBasis(worldX, worldY, worldZ)
	if !ok {
		return
	}
	bounds := texture.Bounds()
	w, h := float32(bounds.Dx()), float32(bounds.Dy())
	center := modelPoint3{x: worldX, y: worldZ, z: worldY}
	tint := color.RGBA{R: 255, G: 255, B: 255, A: uint8(clampFloat(draw.alpha, 0, 1) * 255)}
	vertices := make([]render.Vertex3D, 0, (draw.circleSides+1)*2)
	indices := make([]uint16, 0, draw.circleSides*6)
	point := func(radius, angle float64) modelPoint3 {
		return add3(add3(center, mul3(right, math.Sin(angle)*radius)), mul3(up, math.Cos(angle)*radius))
	}
	for i := 0; i <= draw.circleSides; i++ {
		a := float64(i) / float64(draw.totalCircleSides)
		angle := draw.angle + a*2*math.Pi
		u := float32(a * float64(draw.totalCircleSides) / float64(draw.circleSides))
		vertices = append(vertices,
			texturedSurfaceVertex3D(point(draw.bottomSize, angle), texturePoint{u: u, v: 1}, tint, w, h),
			texturedSurfaceVertex3D(point(draw.topSize, angle), texturePoint{u: u, v: 0}, tint, w, h),
		)
		if i < draw.circleSides {
			base := uint16(i * 2)
			indices = append(indices, base, base+1, base+2, base+1, base+3, base+2)
		}
	}
	screen.DrawTriangles3DOwned(vertices, indices, texture, triangleDrawOptions(render.FilterLinear, render.AddressRepeat))
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

func drawWarpZoneEffect(screen, white, ringTexture *render.Image, x, y, z float64, now time.Time) {
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

func drawWorldRadialGradient(screen, white *render.Image, x, y, z, innerRadius, outerRadius float64, c color.RGBA, segments int) {
	drawWorldRingBand(screen, white, x, y, z, innerRadius, outerRadius, c.A, 0, c, segments)
}

func drawWorldSoftRing(screen, white *render.Image, x, y, z, radius, width float64, c color.RGBA, segments int) {
	inner := math.Max(0, radius-width*0.5)
	mid := math.Max(inner+0.01, radius)
	outer := math.Max(mid+0.01, radius+width*0.5)
	drawWorldRingBand(screen, white, x, y, z, inner, mid, 0, c.A, c, segments)
	drawWorldRingBand(screen, white, x, y, z, mid, outer, c.A, 0, c, segments)
}

func drawWorldCylinderBand(screen, white, texture *render.Image, x, y, z, bottomRadius, topRadius, height float64, c color.RGBA, segments int) {
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
		angle := float64(i) * 2 * math.Pi / float64(segments)
		cosine := math.Cos(angle)
		sine := math.Sin(angle)
		vertices = append(vertices,
			warpEffectTexturedVertex3D(x+cosine*bottomRadius, y+sine*bottomRadius, z, u*srcW, srcH, tint),
			warpEffectTexturedVertex3D(x+cosine*topRadius, y+sine*topRadius, z+height, u*srcW, 0, tint),
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

func drawWorldRingBand(screen, white *render.Image, x, y, z, innerRadius, outerRadius float64, innerAlpha, outerAlpha uint8, c color.RGBA, segments int) {
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
