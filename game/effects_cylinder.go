package game

import (
	"image/color"
	"math"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
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
