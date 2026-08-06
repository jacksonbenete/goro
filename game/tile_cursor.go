package game

import (
	"image"
	"image/color"
	"math"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

const tileCursorDepthBias = 1.0 / 32768.0

func (m *WorldMode) drawTileCursor(screen *render.Frame, ctx client.Context, projection sceneProjection) {
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
	verts, ok := tileCursorCellVerts(ctx.World, x, y)
	if !ok {
		return
	}
	points := projectTileCursorVerts(projection, verts)
	if quadHasInvalidPoint(points) || quadOutside(points, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())) {
		return
	}
	drawTileCursorSurface3D(screen, m.tileCursorTexture(ctx.Resources), verts)
}

func tileCursorCellVerts(world *worldstate.World, x, y int) ([4]modelPoint3, bool) {
	if world == nil {
		return [4]modelPoint3{}, false
	}
	if world.GND != nil {
		return tileCursorGNDCellVerts(world.GND, x, y)
	}
	return tileCursorGATCellVerts(world.GAT, x, y)
}

func tileCursorGNDCellVerts(gnd *res.GND, x, y int) ([4]modelPoint3, bool) {
	if gnd == nil || x < 0 || y < 0 {
		return [4]modelPoint3{}, false
	}
	cell, ok := gnd.Cell(x/2, y/2)
	if !ok || cell.Top < 0 {
		return [4]modelPoint3{}, false
	}
	u0 := 0.0
	if x%2 != 0 {
		u0 = 0.5
	}
	v0 := 0.0
	if y%2 != 0 {
		v0 = 0.5
	}
	u1 := u0 + 0.5
	v1 := v0 + 0.5
	return [4]modelPoint3{
		{x: float64(x), y: bilinearHeight(cell.Heights, u0, v0), z: float64(y)},
		{x: float64(x + 1), y: bilinearHeight(cell.Heights, u1, v0), z: float64(y)},
		{x: float64(x), y: bilinearHeight(cell.Heights, u0, v1), z: float64(y + 1)},
		{x: float64(x + 1), y: bilinearHeight(cell.Heights, u1, v1), z: float64(y + 1)},
	}, true
}

func tileCursorGATCellVerts(gat *res.GAT, x, y int) ([4]modelPoint3, bool) {
	if gat == nil {
		return [4]modelPoint3{}, false
	}
	cell, ok := gat.Cell(x, y)
	if !ok {
		return [4]modelPoint3{}, false
	}
	return [4]modelPoint3{
		{x: float64(x), y: float64(cell.Heights[0]), z: float64(y)},
		{x: float64(x + 1), y: float64(cell.Heights[1]), z: float64(y)},
		{x: float64(x), y: float64(cell.Heights[2]), z: float64(y + 1)},
		{x: float64(x + 1), y: float64(cell.Heights[3]), z: float64(y + 1)},
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

func (m *WorldMode) tileCursorTexture(manager *res.Manager) *render.Image {
	if m.tileCursor != nil {
		return m.tileCursor
	}
	if manager != nil {
		if img, _, err := res.LoadImageExact(manager, []string{"data\\texture\\grid.tga", "data/texture/grid.tga"}); err == nil {
			m.tileCursor = render.NewImageFromImage(tintTileCursorImage(img))
			return m.tileCursor
		}
	}
	m.tileCursor = render.NewImageFromImage(fallbackTileCursorImage())
	return m.tileCursor
}

func tintTileCursorImage(src image.Image) *image.NRGBA {
	bounds := src.Bounds()
	out := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := src.At(x, y).RGBA()
			alpha := uint8((a >> 8) * 153 / 255)
			out.SetNRGBA(x, y, color.NRGBA{R: 50, G: 240, B: 160, A: alpha})
		}
	}
	return out
}

func fallbackTileCursorImage() image.Image {
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
				img.SetRGBA(x, y, color.RGBA{R: 50, G: 240, B: 160, A: alpha})
			}
		}
	}
	return img
}

func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func drawTileCursorSurface3D(screen *render.Frame, texture *render.Image, verts [4]modelPoint3) {
	if texture == nil {
		return
	}
	tints := [4]color.RGBA{
		{R: 255, G: 255, B: 255, A: 255},
		{R: 255, G: 255, B: 255, A: 255},
		{R: 255, G: 255, B: 255, A: 255},
		{R: 255, G: 255, B: 255, A: 255},
	}
	uvs := [4]texturePoint{
		{u: 0, v: 0},
		{u: 1, v: 0},
		{u: 0, v: 1},
		{u: 1, v: 1},
	}
	options := triangleDrawOptions(render.FilterLinear, render.AddressClampToZero)
	options.DepthBias = tileCursorDepthBias
	drawTexturedSurface3DWithOptions(screen, texture, verts, uvs, quadIndices012213, tints, options)
}
