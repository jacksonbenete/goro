package gamemode

import (
	"image/color"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

const (
	retainedWorldMeshMaxVertices = 65535
	gndRetainedChunkSize         = 16
)

type retainedWorldMesh struct {
	mesh *render.WorldMesh
}

type gndRetainedMeshCache struct {
	gnd    *res.GND
	rsw    *res.RSW
	chunks map[gndRetainedChunkKey][]retainedWorldMesh
}

type gndRetainedChunkKey struct {
	x int
	y int
}

type retainedMeshKey struct {
	texture *render.Image
	options render.DrawTrianglesOptions
}

type retainedMeshBuilder struct {
	texture  *render.Image
	options  render.DrawTrianglesOptions
	vertices []render.Vertex3D
	indices  []uint16
	meshes   []retainedWorldMesh
}

func (b *retainedMeshBuilder) addTriangle(a, c, d render.Vertex3D) {
	if b.texture == nil {
		return
	}
	if len(b.vertices)+3 > retainedWorldMeshMaxVertices {
		b.flush()
	}
	base := uint16(len(b.vertices))
	b.vertices = append(b.vertices, a, c, d)
	b.indices = append(b.indices, base, base+1, base+2)
}

func (b *retainedMeshBuilder) addQuad(vertices [4]render.Vertex3D, indices []uint16) {
	if b.texture == nil {
		return
	}
	if len(b.vertices)+len(vertices) > retainedWorldMeshMaxVertices {
		b.flush()
	}
	base := uint16(len(b.vertices))
	b.vertices = append(b.vertices, vertices[:]...)
	for _, index := range indices {
		b.indices = append(b.indices, base+index)
	}
}

func (b *retainedMeshBuilder) flush() {
	if len(b.vertices) == 0 || len(b.indices) == 0 || b.texture == nil {
		b.vertices = b.vertices[:0]
		b.indices = b.indices[:0]
		return
	}
	mesh := render.NewWorldMesh(b.vertices, b.indices, b.texture, &b.options)
	b.meshes = append(b.meshes, retainedWorldMesh{mesh: mesh})
	b.vertices = nil
	b.indices = nil
}

func (m *WorldMode) drawGNDMeshes(screen *render.Image, manager *res.Manager, gnd *res.GND, rsw *res.RSW, projection sceneProjection) {
	if gnd == nil {
		return
	}
	cache := m.gndMeshCache
	if cache == nil || cache.gnd != gnd || cache.rsw != rsw {
		cache = &gndRetainedMeshCache{gnd: gnd, rsw: rsw, chunks: make(map[gndRetainedChunkKey][]retainedWorldMesh)}
		m.gndMeshCache = cache
	}
	width := screen.Bounds().Dx()
	height := screen.Bounds().Dy()
	startX, endX, startY, endY, ok := gndDrawBounds(gnd, projection, width, height)
	if !ok {
		return
	}
	startChunkX := startX / gndRetainedChunkSize
	endChunkX := endX / gndRetainedChunkSize
	startChunkY := startY / gndRetainedChunkSize
	endChunkY := endY / gndRetainedChunkSize
	for chunkY := startChunkY; chunkY <= endChunkY; chunkY++ {
		for chunkX := startChunkX; chunkX <= endChunkX; chunkX++ {
			key := gndRetainedChunkKey{x: chunkX, y: chunkY}
			meshes, ok := cache.chunks[key]
			if !ok {
				x0 := chunkX * gndRetainedChunkSize
				y0 := chunkY * gndRetainedChunkSize
				x1 := minInt(gnd.Width-1, x0+gndRetainedChunkSize-1)
				y1 := minInt(gnd.Height-1, y0+gndRetainedChunkSize-1)
				meshes = m.buildGNDMeshChunk(manager, gnd, rsw, x0, x1, y0, y1)
				cache.chunks[key] = meshes
			}
			for _, mesh := range meshes {
				screen.DrawWorldMesh(mesh.mesh)
			}
		}
	}
}

func (m *WorldMode) buildGNDMeshChunk(manager *res.Manager, gnd *res.GND, rsw *res.RSW, startX, endX, startY, endY int) []retainedWorldMesh {
	if gnd == nil {
		return nil
	}
	startX = maxInt(0, startX)
	startY = maxInt(0, startY)
	endX = minInt(gnd.Width-1, endX)
	endY = minInt(gnd.Height-1, endY)
	if startX > endX || startY > endY {
		return nil
	}
	if m.whitePixel == nil {
		m.whitePixel = render.NewImage(1, 1)
		m.whitePixel.Fill(color.White)
	}
	lighting := sceneLightingFromRSW(rsw)
	topNormals := m.smoothGNDTopNormals(gnd)
	builders := make(map[retainedMeshKey]*retainedMeshBuilder)
	builderFor := func(texture *render.Image, options *render.DrawTrianglesOptions) *retainedMeshBuilder {
		if texture == nil || options == nil {
			return nil
		}
		key := retainedMeshKey{texture: texture, options: *options}
		builder := builders[key]
		if builder == nil {
			builder = &retainedMeshBuilder{texture: texture, options: *options}
			builders[key] = builder
		}
		return builder
	}
	addTextured := func(texture *render.Image, verts [4]modelPoint3, uvs [4]texturePoint, indices []uint16, tints [4]color.RGBA, options *render.DrawTrianglesOptions) {
		builder := builderFor(texture, options)
		if builder == nil {
			return
		}
		bounds := texture.Bounds()
		w, h := float32(bounds.Dx()), float32(bounds.Dy())
		builder.addQuad([4]render.Vertex3D{
			texturedSurfaceVertex3D(verts[0], uvs[0], tints[0], w, h),
			texturedSurfaceVertex3D(verts[1], uvs[1], tints[1], w, h),
			texturedSurfaceVertex3D(verts[2], uvs[2], tints[2], w, h),
			texturedSurfaceVertex3D(verts[3], uvs[3], tints[3], w, h),
		}, indices)
	}
	addColored := func(verts [4]modelPoint3, indices []uint16, tints [4]color.RGBA, options *render.DrawTrianglesOptions) {
		builder := builderFor(m.whitePixel, options)
		if builder == nil {
			return
		}
		builder.addQuad([4]render.Vertex3D{
			coloredSurfaceVertex3D(verts[0], 0, 0, tints[0]),
			coloredSurfaceVertex3D(verts[1], 1, 0, tints[1]),
			coloredSurfaceVertex3D(verts[2], 1, 1, tints[2]),
			coloredSurfaceVertex3D(verts[3], 0, 1, tints[3]),
		}, indices)
	}
	addLightmapped := func(texture *render.Image, verts [4]modelPoint3, uvs [4]texturePoint, baseTints [4]color.RGBA, lightmap res.GNDLightmap, lightScales [4]modelPoint3) {
		const steps = 6
		if texture == nil {
			return
		}
		textureBounds := texture.Bounds()
		textureWidth := float32(textureBounds.Dx())
		textureHeight := float32(textureBounds.Dy())
		baseBuilder := builderFor(texture, groundTextureDrawOptions())
		lightOptions := triangleDrawOptions(render.FilterNearest, render.AddressUnsafe)
		lightOptions.Blend = render.BlendLighter
		lightBuilder := builderFor(m.whitePixel, lightOptions)
		if baseBuilder == nil || lightBuilder == nil {
			return
		}
		row := steps + 1
		localBase := make([]render.Vertex3D, 0, row*row)
		localLight := make([]render.Vertex3D, 0, row*row)
		for y := 0; y <= steps; y++ {
			t := float64(y) / steps
			for x := 0; x <= steps; x++ {
				s := float64(x) / steps
				alpha := float64(res.GNDLightmapSampleAlpha(lightmap, s, t)) / 255
				lm := posterizeGNDLightmapColor(res.GNDLightmapSampleColor(lightmap, s, t))
				lightScale := bilerpModelPoint(lightScales, s, t)
				base := bilerpColor(baseTints, s, t)
				tint := color.RGBA{
					R: clampColor(float64(base.R) * lightScale.x * alpha),
					G: clampColor(float64(base.G) * lightScale.y * alpha),
					B: clampColor(float64(base.B) * lightScale.z * alpha),
					A: 255,
				}
				world := bilerpModelPoint(verts, s, t)
				localBase = append(localBase, texturedSurfaceVertex3D(world, bilerpTexturePoint(uvs, s, t), tint, textureWidth, textureHeight))
				localLight = append(localLight, coloredSurfaceVertex3D(world, 0, 0, lm))
			}
		}
		for y := 0; y < steps; y++ {
			for x := 0; x < steps; x++ {
				topLeft := y*row + x
				topRight := y*row + x + 1
				bottomLeft := (y+1)*row + x
				bottomRight := (y+1)*row + x + 1
				baseBuilder.addTriangle(localBase[topLeft], localBase[topRight], localBase[bottomRight])
				baseBuilder.addTriangle(localBase[topLeft], localBase[bottomRight], localBase[bottomLeft])
				lightBuilder.addTriangle(localLight[topLeft], localLight[topRight], localLight[bottomRight])
				lightBuilder.addTriangle(localLight[topLeft], localLight[bottomRight], localLight[bottomLeft])
			}
		}
	}

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			cell, ok := gnd.Cell(x, y)
			if !ok {
				continue
			}
			if cell.Top >= 0 {
				if surface, ok := gnd.Surface(cell.Top); ok {
					vertexOrder := [4]int{0, 1, 3, 2}
					verts := [4]modelPoint3{
						{x: float64(x) * 2, y: float64(cell.Heights[0]), z: float64(y) * 2},
						{x: float64(x+1) * 2, y: float64(cell.Heights[1]), z: float64(y) * 2},
						{x: float64(x+1) * 2, y: float64(cell.Heights[3]), z: float64(y+1) * 2},
						{x: float64(x) * 2, y: float64(cell.Heights[2]), z: float64(y+1) * 2},
					}
					uvs := surfaceUVs(surface, vertexOrder)
					baseTints := topGNDSurfaceBaseTints(gnd, x, y, surface.Color)
					normals := gndTopNormalsAt(topNormals, gnd, x, y)
					if texture := m.groundTexture(manager, gndTextureName(gnd, surface.TextureID)); texture != nil {
						if lightmap, ok := gnd.Lightmap(surface.LightmapID); ok {
							addLightmapped(texture, verts, uvs, baseTints, lightmap, vertexLightScales(lighting, normals))
						} else {
							addTextured(texture, verts, uvs, quadIndices012023, surfaceVertexTints(gnd, surface, baseTints, vertexOrder, cell.Heights, normals, lighting), groundTextureDrawOptions())
						}
					} else {
						addColored(verts, quadIndices012023, groundSurfaceVertexColors(gndTextureName(gnd, surface.TextureID), surface.Color, cell.Heights, normals, lighting), worldOpaqueTriangleDrawOptions(render.FilterNearest, render.AddressUnsafe))
					}
				}
			}
			if cell.Front >= 0 && y+1 < gnd.Height {
				neighbor, neighborOK := gnd.Cell(x, y+1)
				surface, surfaceOK := gnd.Surface(cell.Front)
				if neighborOK && surfaceOK {
					vertexOrder := [4]int{0, 1, 3, 2}
					verts := [4]modelPoint3{
						{x: float64(x) * 2, y: float64(cell.Heights[2]), z: float64(y+1) * 2},
						{x: float64(x+1) * 2, y: float64(cell.Heights[3]), z: float64(y+1) * 2},
						{x: float64(x+1) * 2, y: float64(neighbor.Heights[1]), z: float64(y+1) * 2},
						{x: float64(x) * 2, y: float64(neighbor.Heights[0]), z: float64(y+1) * 2},
					}
					heights := [4]float32{cell.Heights[2], cell.Heights[3], neighbor.Heights[1], neighbor.Heights[0]}
					normals := uniformGNDNormals(modelPoint3{z: 1})
					addGNDRetainedSurface(m, manager, gnd, surface, vertexOrder, verts, heights, normals, lighting, addTextured, addColored)
				}
			}
			if cell.Right >= 0 && x+1 < gnd.Width {
				neighbor, neighborOK := gnd.Cell(x+1, y)
				surface, surfaceOK := gnd.Surface(cell.Right)
				if neighborOK && surfaceOK {
					vertexOrder := [4]int{0, 1, 3, 2}
					verts := [4]modelPoint3{
						{x: float64(x+1) * 2, y: float64(cell.Heights[3]), z: float64(y+1) * 2},
						{x: float64(x+1) * 2, y: float64(cell.Heights[1]), z: float64(y) * 2},
						{x: float64(x+1) * 2, y: float64(neighbor.Heights[0]), z: float64(y) * 2},
						{x: float64(x+1) * 2, y: float64(neighbor.Heights[2]), z: float64(y+1) * 2},
					}
					heights := [4]float32{cell.Heights[3], cell.Heights[1], neighbor.Heights[0], neighbor.Heights[2]}
					normals := uniformGNDNormals(modelPoint3{x: 1})
					addGNDRetainedSurface(m, manager, gnd, surface, vertexOrder, verts, heights, normals, lighting, addTextured, addColored)
				}
			}
		}
	}
	var meshes []retainedWorldMesh
	for _, builder := range builders {
		builder.flush()
		meshes = append(meshes, builder.meshes...)
	}
	return meshes
}

func addGNDRetainedSurface(m *WorldMode, manager *res.Manager, gnd *res.GND, surface res.GNDSurface, vertexOrder [4]int, verts [4]modelPoint3, heights [4]float32, normals [4]modelPoint3, lighting sceneLighting, addTextured func(*render.Image, [4]modelPoint3, [4]texturePoint, []uint16, [4]color.RGBA, *render.DrawTrianglesOptions), addColored func([4]modelPoint3, []uint16, [4]color.RGBA, *render.DrawTrianglesOptions)) {
	baseTints := uniformGNDSurfaceBaseTints(surface.Color)
	uvs := surfaceUVs(surface, vertexOrder)
	textureName := gndTextureName(gnd, surface.TextureID)
	if texture := m.groundTexture(manager, textureName); texture != nil {
		addTextured(texture, verts, uvs, quadIndices012023, surfaceVertexTints(gnd, surface, baseTints, vertexOrder, heights, normals, lighting), groundTextureDrawOptions())
		return
	}
	addColored(verts, quadIndices012023, groundSurfaceVertexColors(textureName, surface.Color, heights, normals, lighting), worldOpaqueTriangleDrawOptions(render.FilterNearest, render.AddressUnsafe))
}
