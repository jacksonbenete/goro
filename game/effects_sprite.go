package game

import (
	"image/color"
	"log"
	"math"
	"strings"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

func (m *WorldMode) draw3DSpriteEffect(screen *render.Image, ctx client.Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, worldX, worldY, worldZ float64, size float64, alpha float64, starts time.Time, now time.Time) {
	view := m.effectSpriteView(ctx.Resources, component.spriteFile)
	if view == nil || len(view.act.Actions) == 0 {
		return
	}
	actionIndex := 0
	action := view.act.Actions[actionIndex]
	if len(action.Animations) == 0 {
		return
	}
	delayMS := float64(action.DelayMS)
	if component.spriteDelay > 0 {
		delayMS = float64(component.spriteDelay / time.Millisecond)
	}
	motion := 0
	if component.spriteRepeat {
		motion = spriteMotionIndexWithDelay(action, starts, now, true, delayMS)
	} else {
		motion = spriteMotionIndexWithDelay(action, starts, now, false, delayMS)
	}
	if motion < 0 || motion >= len(action.Animations) {
		return
	}
	ignoreLayerAngles := component.rotateToTarget
	key := singleSpriteBillboardKey{actionIndex: actionIndex, motion: motion, ignoreLayerAngles: ignoreLayerAngles}
	billboard, ok := view.billboards[key]
	if !ok {
		var baseOK bool
		billboard, baseOK = composeSingleSpriteBillboardWithOptions(view, action.Animations[motion], ignoreLayerAngles)
		if !baseOK {
			return
		}
		view.billboards[key] = billboard
	}
	tint := effectComponentTint(component, 1)
	if component.worldSizedSprite {
		scale := size / 100
		angle := -worldEffectSpriteAngle(component) * math.Pi / 180
		options := spriteBillboardTriangleDrawOptions()
		if component.blendAdditive {
			options.Blend = render.BlendLighter
		}
		drawSpriteBillboardTintAlphaWorld3DWithOptions(screen, projection, billboard, worldX, worldY, worldZ, scale, angle, alpha, 1, tint, options)
		return
	}
	scale := effect3DSpriteScale(size)
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	options := effect3DSpriteDrawOptions(component)
	if angle, ok := effectSpriteScreenRotation(ctx, projection, component, effect); ok {
		drawSpriteBillboardTintAlphaRotated3DWithOptions(screen, projection, billboard, worldX, worldY, worldZ, scale, angle, alpha, 1, tint, options)
		return
	}
	drawSpriteBillboardTintAlpha3DWithOptions(screen, projection, billboard, worldX, worldY, worldZ, scale, alpha, 1, tint, options)
}

func effect3DSpriteDrawOptions(component worldEffectComponent) *render.DrawTrianglesOptions {
	options := spriteBillboardTriangleDrawOptions()
	if component.blendAdditive {
		options.Blend = render.BlendLighter
	}
	return options
}

func effect3DSpriteScale(size float64) float64 {
	if size <= 0 || math.IsNaN(size) || math.IsInf(size, 0) {
		return 1
	}
	// reference client's SpriteRenderer applies size as (size / 175) * xSize, with
	// xSize defaulting to 5. effectTableSize already stores that scale.
	return size
}

func effectSpriteScreenRotation(ctx client.Context, projection sceneProjection, component worldEffectComponent, effect worldEffect) (float64, bool) {
	if component.rotateToTarget {
		startX, startY, startZ, endX, endY, endZ, ok := effectTrajectoryEndpoints(ctx, component, effect)
		if ok {
			start := projection.Project(startX, startY, startZ)
			end := projection.Project(endX, endY, endZ)
			dx := float64(end.x - start.x)
			dy := float64(end.y - start.y)
			if math.Hypot(dx, dy) > 0.001 {
				return math.Atan2(dy, dx) - math.Pi/2, true
			}
		}
	}
	angle := worldEffectBillboardAngle(component, projection, 0)
	if math.Abs(angle) > 0.0001 {
		return angle, true
	}
	return 0, false
}

func effectTrajectoryEndpoints(ctx client.Context, component worldEffectComponent, effect worldEffect) (float64, float64, float64, float64, float64, float64, bool) {
	if ctx.World == nil {
		return 0, 0, 0, 0, 0, 0, false
	}
	x, y, ok := effectAnchor(ctx, effect.actorID)
	if !ok {
		return 0, 0, 0, 0, 0, 0, false
	}
	worldX := cellCenter(float64(x))
	worldY := cellCenter(float64(y))
	worldZ := terrainHeightAt(ctx.World, float64(x), float64(y)) + 0.07
	startX := component.posX + component.posXStartMiddle
	startY := component.posY + component.posYStartMiddle
	startZ := component.posZ + component.posZStartMiddle
	endX := component.posXEnd + component.posXEndMiddle
	endY := component.posYEnd + component.posYEndMiddle
	endZ := component.posZEnd + component.posZEndMiddle
	if component.posXEnd == 0 && component.posXEndMiddle == 0 && component.posXStartMiddle == 0 {
		endX = startX
	}
	if component.posYEnd == 0 && component.posYEndMiddle == 0 && component.posYStartMiddle == 0 {
		endY = startY
	}
	if component.posZEnd == 0 && component.posZEndMiddle == 0 {
		endZ = startZ
	}
	if component.fromSrc || component.toSrc {
		otherX, otherY, otherZ, otherOK := effectOtherEndpoint(ctx, effect, worldX, worldY, worldZ)
		if otherOK {
			dx := otherX - worldX
			dy := otherY - worldY
			dz := otherZ - worldZ
			if component.fromSrc {
				endX += dx
				endY += dy
				endZ += dz
			}
			if component.toSrc {
				startX += dx
				startY += dy
				startZ += dz
			}
		}
	}
	return worldX + startX, worldY + startY, worldZ + startZ, worldX + endX, worldY + endY, worldZ + endZ, true
}

func worldEffectSpriteAngle(component worldEffectComponent) float64 {
	angle := component.angleStart
	if !component.rotateToTarget {
		return angle
	}
	startX, startY := component.posX, component.posY
	endX, endY := component.posXEnd, component.posYEnd
	if component.posXEnd == 0 && component.posXEndRand == 0 {
		endX = startX
	}
	if component.posYEnd == 0 && component.posYEndRand == 0 {
		endY = startY
	}
	return angle + 90 - math.Atan2(endY-startY, endX-startX)*180/math.Pi
}

func (m *WorldMode) drawSPREffect(screen *render.Image, ctx client.Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, worldX, worldY, worldZ float64, now time.Time) {
	view := m.effectSpriteView(ctx.Resources, component.spriteFile)
	if view == nil || len(view.act.Actions) == 0 {
		return
	}
	actionIndex := component.spriteFrame
	if component.spriteDirection {
		if actor, ok := ctx.World.Actors[effect.actorID]; ok {
			actionIndex = actor.RenderDirection(now) % len(view.act.Actions)
		} else if isLocalActor(ctx, effect.actorID) {
			actionIndex = ctx.World.Player.RenderDirection(now) % len(view.act.Actions)
		}
	}
	if actionIndex < 0 || actionIndex >= len(view.act.Actions) {
		actionIndex = 0
	}
	action := view.act.Actions[actionIndex]
	if len(action.Animations) == 0 {
		return
	}
	delayMS := float64(action.DelayMS)
	if component.spriteDelay > 0 {
		delayMS = float64(component.spriteDelay / time.Millisecond)
	}
	motion := 0
	if component.spriteRepeat {
		motion = spriteMotionIndexWithDelay(action, effect.starts, now, true, delayMS)
	} else {
		motion = spriteMotionIndexWithDelay(action, effect.starts, now, false, delayMS)
		if motion >= len(action.Animations)-1 && !component.spriteStopAtEnd && component.duration <= 0 {
			return
		}
	}
	if motion < 0 || motion >= len(action.Animations) {
		return
	}
	key := singleSpriteBillboardKey{
		actionIndex: actionIndex,
		motion:      motion,
		anchorX:     component.spriteXOffset,
		anchorY:     component.spriteYOffset,
	}
	billboard, ok := view.billboards[key]
	if !ok {
		base, baseOK := composeSingleSpriteBillboard(view, action.Animations[motion])
		if !baseOK {
			return
		}
		copy := *base
		copy.anchorX -= component.spriteXOffset
		copy.anchorY -= component.spriteYOffset
		billboard = &copy
		view.billboards[key] = billboard
	}
	z := worldZ + component.posZ
	if component.spriteHead {
		z += 2.0
	}
	if component.worldSizedSprite {
		drawSpriteBillboardTintAlphaWorld3D(screen, projection, billboard, worldX, worldY, z, effectPixelRatio, 0, 1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		return
	}
	drawSpriteBillboardTintAlpha3D(screen, projection, billboard, worldX, worldY, z, 1, 1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
}

const effectSpriteRoot = "data\\sprite\\\xC0\xCC\xC6\xD1\xC6\xAE\\"

func (m *WorldMode) effectSpriteView(manager *res.Manager, file string) *spriteView {
	file = strings.TrimSpace(file)
	if manager == nil || file == "" {
		return nil
	}
	if m.effectViews == nil {
		m.effectViews = make(map[string]*spriteView)
	}
	if m.effectViewMiss == nil {
		m.effectViewMiss = make(map[string]struct{})
	}
	key := strings.ReplaceAll(file, "/", "\\")
	if view, ok := m.effectViews[key]; ok {
		return view
	}
	if _, ok := m.effectViewMiss[key]; ok {
		return nil
	}
	actCandidates := effectSpriteResourceCandidates(file, "act")
	sprCandidates := effectSpriteResourceCandidates(file, "spr")
	view, status := loadSpriteView(manager, actCandidates, sprCandidates, nil, "effect sprite "+file)
	if view == nil {
		m.effectViewMiss[key] = struct{}{}
		log.Printf("effect sprite unavailable file=%q: %s", file, status)
		return nil
	}
	m.effectViews[key] = view
	log.Printf("effect sprite resources file=%q %s", file, status)
	return view
}

func effectSpriteResourceCandidates(file, ext string) []string {
	normalized := strings.TrimSpace(strings.ReplaceAll(file, "/", "\\"))
	normalized = strings.TrimSuffix(normalized, ".spr")
	normalized = strings.TrimSuffix(normalized, ".act")
	if normalized == "" {
		return nil
	}
	base := normalized
	if !strings.HasPrefix(strings.ToLower(base), "data\\sprite\\") {
		base = effectSpriteRoot + base
	}
	path := base + "." + ext
	return []string{path, strings.ReplaceAll(path, "\\", "/")}
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
