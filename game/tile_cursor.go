package game

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

func (m *WorldMode) drawTileCursor(screen *render.Image, ctx client.Context, projection sceneProjection, now time.Time) {
	if ctx.Input == nil || ctx.World == nil || ctx.World.GAT == nil {
		return
	}
	if uiPointerBlocked(ctx) {
		return
	}
	x, y, ok := m.hoveredWalkCell(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY)
	if !ok {
		return
	}
	verts, ok := tileCursorCellVerts(ctx.World.GAT, x, y, now)
	if !ok {
		return
	}
	points := projectTileCursorVerts(projection, verts)
	if quadHasInvalidPoint(points) || quadOutside(points, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())) {
		return
	}
	drawTileCursorSurface3D(screen, m.tileCursorTexture(), verts)
}

func tileCursorCellVerts(gat *res.GAT, x, y int, now time.Time) ([4]modelPoint3, bool) {
	if gat == nil {
		return [4]modelPoint3{}, false
	}
	cell, ok := gat.Cell(x, y)
	if !ok {
		return [4]modelPoint3{}, false
	}
	lift := tileCursorLift(now)
	return [4]modelPoint3{
		{x: float64(x), y: float64(cell.Heights[0]) + lift, z: float64(y)},
		{x: float64(x + 1), y: float64(cell.Heights[1]) + lift, z: float64(y)},
		{x: float64(x), y: float64(cell.Heights[2]) + lift, z: float64(y + 1)},
		{x: float64(x + 1), y: float64(cell.Heights[3]) + lift, z: float64(y + 1)},
	}, true
}

func projectTileCursorVerts(projection sceneProjection, verts [4]modelPoint3) [4]screenPoint {
	return [4]screenPoint{
		projection.Project(verts[0].x, verts[0].z, verts[0].y),
		projection.Project(verts[1].x, verts[1].z, verts[1].y),
		projection.Project(verts[2].x, verts[2].z, verts[2].y),
		projection.Project(verts[3].x, verts[3].z, verts[3].y),
	}
}

func quadOutside(points [4]screenPoint, width, height float64) bool {
	minX, minY := float64(points[0].x), float64(points[0].y)
	maxX, maxY := minX, minY
	for _, point := range points[1:] {
		minX = math.Min(minX, float64(point.x))
		minY = math.Min(minY, float64(point.y))
		maxX = math.Max(maxX, float64(point.x))
		maxY = math.Max(maxY, float64(point.y))
	}
	return maxX < -32 || maxY < -32 || minX > width+32 || minY > height+32
}

func quadHasInvalidPoint(points [4]screenPoint) bool {
	for _, point := range points {
		if !isFinite(float64(point.x)) || !isFinite(float64(point.y)) {
			return true
		}
		if point.x <= -1<<19 && point.y <= -1<<19 {
			return true
		}
	}
	return false
}

func tileCursorLift(now time.Time) float64 {
	seconds := float64(now.UnixNano()) / float64(time.Second)
	return 0.018 + 0.006*math.Sin(seconds*math.Pi*2/1.2)
}

func (m *WorldMode) tileCursorTexture() *render.Image {
	if m.tileCursor != nil {
		return m.tileCursor
	}
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dist := minInt(minInt(x, y), minInt(size-1-x, size-1-y))
			alpha := uint8(0)
			switch {
			case dist < 3:
				alpha = 190
			case dist < 6:
				alpha = 100
			case dist < 11:
				alpha = 32
			}
			if x == y || x == size-1-y {
				alpha = maxUint8(alpha, 34)
			}
			if alpha > 0 {
				img.SetRGBA(x, y, color.RGBA{R: 180, G: 230, B: 255, A: alpha})
			}
		}
	}
	m.tileCursor = render.NewImageFromImage(img)
	return m.tileCursor
}

func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func drawTileCursorSurface3D(screen, texture *render.Image, verts [4]modelPoint3) {
	if texture == nil {
		return
	}
	tints := [4]color.RGBA{
		{R: 255, G: 255, B: 255, A: 210},
		{R: 255, G: 255, B: 255, A: 210},
		{R: 255, G: 255, B: 255, A: 210},
		{R: 255, G: 255, B: 255, A: 210},
	}
	uvs := [4]texturePoint{
		{u: 0, v: 0},
		{u: 1, v: 0},
		{u: 0, v: 1},
		{u: 1, v: 1},
	}
	drawTexturedSurface3DWithOptions(screen, texture, verts, uvs, quadIndices012213, tints, triangleDrawOptions(render.FilterLinear, render.AddressClampToZero))
}
