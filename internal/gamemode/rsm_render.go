package gamemode

import (
	"bytes"
	"image/color"
	"log"
	"math"
	"os"
	"sort"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kivutar/goro/internal/res"
)

type modelPoint3 struct {
	x float64
	y float64
	z float64
}

type modelTriangle struct {
	points      [3]screenPoint
	uvs         [3]texturePoint
	depth       float64
	color       color.RGBA
	textureName string
}

type modelBounds struct {
	min modelPoint3
	max modelPoint3
}

type rsmBounds struct {
	model modelBounds
	main  modelBounds
	nodes map[string]modelBounds
}

type mat4 [16]float64

func (m *WorldMode) drawRSMModels(screen *ebiten.Image, manager *res.Manager, rsw *res.RSW, models map[string]*res.RSM, gnd *res.GND, projection sceneProjection) {
	triangles := m.collectRSMModelTriangles(screen, manager, rsw, models, gnd, projection)
	for _, tri := range triangles {
		m.drawModelTriangle(screen, manager, tri)
	}
}

func (m *WorldMode) collectRSMModelTriangles(screen *ebiten.Image, manager *res.Manager, rsw *res.RSW, models map[string]*res.RSM, gnd *res.GND, projection sceneProjection) []modelTriangle {
	if rsw == nil || gnd == nil || models == nil {
		return nil
	}
	if m.whitePixel == nil {
		m.whitePixel = ebiten.NewImage(1, 1)
		m.whitePixel.Fill(color.White)
	}

	width := screen.Bounds().Dx()
	height := screen.Bounds().Dy()
	radius := rsmRenderRadius()
	maxFaces := rsmMaxFaces()

	type visiblePlacement struct {
		model res.RSWModel
		baseX float64
		baseY float64
		dist2 float64
	}
	var visible []visiblePlacement
	for _, placement := range rsw.Models {
		if placement.Filename == "" {
			continue
		}
		baseX := float64(placement.Position.X) + float64(gnd.Width)
		baseY := float64(placement.Position.Z) + float64(gnd.Height)
		dx := baseX - projection.playerX
		dy := baseY - projection.playerY
		if math.Abs(dx) > radius*2 || math.Abs(dy) > radius*2 {
			continue
		}
		visible = append(visible, visiblePlacement{
			model: placement,
			baseX: baseX,
			baseY: baseY,
			dist2: dx*dx + dy*dy,
		})
	}
	sort.SliceStable(visible, func(i, j int) bool {
		return visible[i].dist2 < visible[j].dist2
	})

	var triangles []modelTriangle
	type boundsCacheKey struct {
		rsm  *res.RSM
		root string
	}
	boundsCache := make(map[boundsCacheKey]rsmBounds)
	nodeMatrixCache := make(map[*res.RSM]map[string]mat4)
	for _, visiblePlacement := range visible {
		placement := visiblePlacement.model
		if placement.Filename == "" {
			continue
		}
		baseX := visiblePlacement.baseX
		baseY := visiblePlacement.baseY
		rsm, ok := models[placement.Filename]
		if !ok {
			loaded, err := loadRSMModel(manager, placement.Filename)
			if err == nil {
				rsm = loaded
			}
			models[placement.Filename] = rsm
		}
		if rsm == nil {
			continue
		}

		nodeIndices := selectedRSMNodeIndices(rsm, placement.NodeName)
		boundsKey := boundsCacheKey{rsm: rsm, root: selectedRSMRootName(rsm, placement.NodeName)}
		bounds, ok := boundsCache[boundsKey]
		if !ok {
			bounds = calculateRSMBoundsForNodes(rsm, nodeIndices)
			boundsCache[boundsKey] = bounds
		}
		nodeMatrices, ok := nodeMatrixCache[rsm]
		if !ok {
			nodeMatrices = buildRSMNodeMatrices(rsm)
			nodeMatrixCache[rsm] = nodeMatrices
		}
		instance := modelInstance{
			placement: placement,
			bounds:    bounds.model,
			baseX:     baseX,
			baseY:     baseY,
			matrix:    buildRSMInstanceMatrix(rsm, placement, baseX, baseY, bounds.model),
		}
		m.logRSMTransformDebug(placement, instance)
		for _, nodeIndex := range nodeIndices {
			node := &rsm.Nodes[nodeIndex]
			for _, tri := range buildRSMNodeTriangles(rsm, node, nodeMatrices[node.Name], instance, projection, float64(width), float64(height)) {
				triangles = append(triangles, tri)
				if maxFaces > 0 && len(triangles) >= maxFaces {
					break
				}
			}
			if maxFaces > 0 && len(triangles) >= maxFaces {
				break
			}
		}
		if maxFaces > 0 && len(triangles) >= maxFaces {
			break
		}
	}

	sort.SliceStable(triangles, func(i, j int) bool {
		return triangles[i].depth > triangles[j].depth
	})
	return triangles
}

func (m *WorldMode) drawModelTriangle(screen *ebiten.Image, manager *res.Manager, tri modelTriangle) {
	if texture := m.groundTexture(manager, tri.textureName); texture != nil {
		drawTexturedTriangle(screen, texture, tri.points, tri.uvs, tri.color)
		return
	}
	drawColoredTriangle(screen, m.whitePixel, tri.points, tri.color)
}

func (m *WorldMode) logRSMTransformDebug(placement res.RSWModel, instance modelInstance) {
	if os.Getenv("GORO_DEBUG_RSM_TRANSFORMS") != "1" || !isRSMDebugBridgeName(placement.Filename) {
		return
	}
	if m.rsmDebugLog == nil {
		m.rsmDebugLog = make(map[string]struct{})
	}
	key := placement.Filename + "|" + placement.NodeName
	if _, ok := m.rsmDebugLog[key]; ok {
		return
	}
	m.rsmDebugLog[key] = struct{}{}
	refOffset := modelPoint3{
		x: -(instance.bounds.min.x + instance.bounds.max.x) * 0.5,
		y: -instance.bounds.max.y,
		z: -(instance.bounds.min.z + instance.bounds.max.z) * 0.5,
	}
	log.Printf("RSMDBG actor rawModel=%q nameBytes=%x node=%q pos=(%.3f,%.3f,%.3f) rot=(%.3f,%.3f,%.3f) scale=(%.3f,%.3f,%.3f) boundsMin=(%.3f,%.3f,%.3f) boundsMax=(%.3f,%.3f,%.3f) openMidgardOffset=(%.3f,%.3f,%.3f) localAnchor=(%.3f,%.3f,%.3f) worldT=(%.3f,%.3f,%.3f)",
		placement.Filename,
		[]byte(placement.Filename),
		placement.NodeName,
		placement.Position.X,
		placement.Position.Y,
		placement.Position.Z,
		placement.Rotation.X,
		placement.Rotation.Y,
		placement.Rotation.Z,
		placement.Scale.X,
		placement.Scale.Y,
		placement.Scale.Z,
		instance.bounds.min.x,
		instance.bounds.min.y,
		instance.bounds.min.z,
		instance.bounds.max.x,
		instance.bounds.max.y,
		instance.bounds.max.z,
		refOffset.x,
		refOffset.y,
		refOffset.z,
		refOffset.x,
		-refOffset.y,
		refOffset.z,
		instance.matrix[12],
		instance.matrix[13],
		instance.matrix[14])
}

func isRSMDebugBridgeName(name string) bool {
	data := []byte(name)
	return bytes.Contains(data, []byte{0xb9, 0xe8, 0xb4, 0xd9, 0xb8, 0xae}) ||
		bytes.Contains(data, []byte{0xb9, 0xe8})
}

func selectedRSMRootName(rsm *res.RSM, rootName string) string {
	if rsm == nil || rootName == "" {
		return ""
	}
	for i := range rsm.Nodes {
		if rsm.Nodes[i].Name == rootName {
			return rootName
		}
	}
	return ""
}

func selectedRSMNodeIndices(rsm *res.RSM, rootName string) []int {
	if rsm == nil || len(rsm.Nodes) == 0 {
		return nil
	}
	rootName = selectedRSMRootName(rsm, rootName)
	if rootName == "" {
		indices := make([]int, len(rsm.Nodes))
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	selected := map[string]struct{}{rootName: {}}
	changed := true
	for changed {
		changed = false
		for i := range rsm.Nodes {
			node := &rsm.Nodes[i]
			if _, ok := selected[node.Name]; ok {
				continue
			}
			if _, ok := selected[node.ParentName]; !ok {
				continue
			}
			selected[node.Name] = struct{}{}
			changed = true
		}
	}

	indices := make([]int, 0, len(selected))
	for i := range rsm.Nodes {
		if _, ok := selected[rsm.Nodes[i].Name]; ok {
			indices = append(indices, i)
		}
	}
	return indices
}

type modelInstance struct {
	placement res.RSWModel
	bounds    modelBounds
	baseX     float64
	baseY     float64
	matrix    mat4
}

func buildRSMNodeTriangles(rsm *res.RSM, node *res.RSMNode, nodeMatrix mat4, instance modelInstance, projection sceneProjection, screenWidth, screenHeight float64) []modelTriangle {
	if len(node.Vertices) == 0 || len(node.Faces) == 0 {
		return nil
	}

	localMatrix := mat4Identity()
	localMatrix = mat4Translate(localMatrix, modelPoint3{
		x: -(instance.bounds.min.x + instance.bounds.max.x) * 0.5,
		y: instance.bounds.max.y,
		z: -(instance.bounds.min.z + instance.bounds.max.z) * 0.5,
	})
	localMatrix = mat4Scale(localMatrix, modelPoint3{x: 1, y: -1, z: 1})
	localMatrix = mat4Multiply(localMatrix, nodeMatrix)
	if len(rsm.Nodes) != 1 {
		localMatrix = mat4Translate(localMatrix, vectorFromRSM(node.Offset))
	}
	localMatrix = mat4Multiply(localMatrix, mat4FromMat3(node.Matrix))
	modelMatrix := mat4Multiply(instance.matrix, localMatrix)

	worldVerts := make([]modelPoint3, len(node.Vertices))
	screenVerts := make([]screenPoint, len(node.Vertices))
	for i, vertex := range node.Vertices {
		world := mat4TransformPoint(modelMatrix, vectorFromRSM(vertex))
		worldVerts[i] = world
		screenVerts[i] = projection.Project(world.x, world.z, world.y)
	}

	var triangles []modelTriangle
	for _, face := range node.Faces {
		if int(face.VertexIndices[0]) >= len(worldVerts) || int(face.VertexIndices[1]) >= len(worldVerts) || int(face.VertexIndices[2]) >= len(worldVerts) {
			continue
		}
		points := [3]screenPoint{
			screenVerts[face.VertexIndices[0]],
			screenVerts[face.VertexIndices[1]],
			screenVerts[face.VertexIndices[2]],
		}
		if triangleOutside(points, screenWidth, screenHeight) {
			continue
		}

		a := worldVerts[face.VertexIndices[0]]
		b := worldVerts[face.VertexIndices[1]]
		c := worldVerts[face.VertexIndices[2]]
		if !projection.VisibleForTriangle(a.x, a.z, a.y) ||
			!projection.VisibleForTriangle(b.x, b.z, b.y) ||
			!projection.VisibleForTriangle(c.x, c.z, c.y) {
			continue
		}
		depth := (projection.Depth(a.x, a.z, a.y) + projection.Depth(b.x, b.z, b.y) + projection.Depth(c.x, c.z, c.y)) / 3
		textureName, uvs := rsmFaceTexture(rsm, node, face)
		triangles = append(triangles, modelTriangle{
			points:      points,
			uvs:         uvs,
			depth:       depth,
			color:       rsmFaceColor(textureName, a, b, c),
			textureName: textureName,
		})
	}
	return triangles
}

func buildRSMInstanceMatrix(rsm *res.RSM, placement res.RSWModel, baseX, baseY float64, bounds modelBounds) mat4 {
	matrix := mat4Identity()
	matrix = mat4Translate(matrix, modelPoint3{x: baseX, y: float64(placement.Position.Y), z: baseY})
	matrix = mat4RotateZ(matrix, degreesToRadians(float64(placement.Rotation.Z)))
	matrix = mat4RotateX(matrix, degreesToRadians(float64(placement.Rotation.X)))
	matrix = mat4RotateY(matrix, degreesToRadians(float64(placement.Rotation.Y)))
	matrix = mat4Scale(matrix, vectorFromRSW(placement.Scale))

	if main := rsmMainNode(rsm); main != nil && rsmVersionAtLeast(rsm, 2, 2) {
		matrix = mat4Scale(matrix, modelPoint3{x: 1, y: -1, z: 1})
		matrix = mat4Translate(matrix, vectorFromRSM(main.Offset))
		matrix = mat4Translate(matrix, modelPoint3{y: (bounds.max.y - bounds.min.y) * 0.5})
		matrix = mat4Translate(matrix, modelPoint3{
			x: (bounds.max.x + bounds.min.x) * 0.5,
			y: (bounds.max.y + bounds.min.y) * 0.5,
			z: (bounds.max.z + bounds.min.z) * 0.5,
		})
	}
	return matrix
}

func buildRSMNodeMatrices(rsm *res.RSM) map[string]mat4 {
	nodes := make(map[string]*res.RSMNode, len(rsm.Nodes))
	for i := range rsm.Nodes {
		nodes[rsm.Nodes[i].Name] = &rsm.Nodes[i]
	}
	out := make(map[string]mat4, len(rsm.Nodes))
	visiting := make(map[string]bool, len(rsm.Nodes))
	var build func(*res.RSMNode) mat4
	build = func(node *res.RSMNode) mat4 {
		if matrix, ok := out[node.Name]; ok {
			return matrix
		}
		if visiting[node.Name] {
			return mat4Identity()
		}
		visiting[node.Name] = true
		matrix := mat4Identity()
		if parent := nodes[node.ParentName]; parent != nil && parent != node && node.ParentName != node.Name {
			matrix = build(parent)
		}
		matrix = mat4Translate(matrix, vectorFromRSM(node.Position))
		if len(node.RotationKeyframes) > 0 {
			matrix = mat4RotateQuat(matrix, node.RotationKeyframes[0].Quaternion)
		} else {
			matrix = mat4RotateAxis(matrix, vectorFromRSM(node.RotationAxis), float64(node.RotationAngle))
		}
		matrix = mat4Scale(matrix, vectorFromRSM(node.Scale))
		out[node.Name] = matrix
		visiting[node.Name] = false
		return matrix
	}
	for i := range rsm.Nodes {
		build(&rsm.Nodes[i])
	}
	return out
}

func calculateRSMBounds(rsm *res.RSM) rsmBounds {
	indices := make([]int, len(rsm.Nodes))
	for i := range indices {
		indices[i] = i
	}
	return calculateRSMBoundsForNodes(rsm, indices)
}

func calculateRSMBoundsForNodes(rsm *res.RSM, nodeIndices []int) rsmBounds {
	bounds := rsmBounds{
		model: emptyModelBounds(),
		main:  modelBounds{},
		nodes: make(map[string]modelBounds, len(rsm.Nodes)),
	}
	nodeMatrices := buildRSMNodeMatrices(rsm)
	for _, nodeIndex := range nodeIndices {
		if nodeIndex < 0 || nodeIndex >= len(rsm.Nodes) {
			continue
		}
		node := &rsm.Nodes[nodeIndex]
		matrix := nodeMatrices[node.Name]
		if len(rsm.Nodes) != 1 {
			matrix = mat4Translate(matrix, vectorFromRSM(node.Offset))
		}
		matrix = mat4Multiply(matrix, mat4FromMat3(node.Matrix))

		nodeBounds := transformedNodeBounds(node, matrix)
		bounds.nodes[node.Name] = nodeBounds
		if !nodeBounds.empty() {
			bounds.model.include(nodeBounds)
		}
	}
	if bounds.model.empty() {
		bounds.model = modelBounds{}
	}
	if len(nodeIndices) > 0 && nodeIndices[0] >= 0 && nodeIndices[0] < len(rsm.Nodes) {
		main := &rsm.Nodes[nodeIndices[0]]
		bounds.main = bounds.nodes[main.Name]
	}
	if bounds.main.empty() {
		bounds.main = bounds.model
	}
	return bounds
}

func transformedNodeBounds(node *res.RSMNode, matrix mat4) modelBounds {
	bounds := emptyModelBounds()
	for _, vertex := range node.Vertices {
		point := mat4TransformPoint(matrix, vectorFromRSM(vertex))
		bounds.min.x = math.Min(bounds.min.x, point.x)
		bounds.min.y = math.Min(bounds.min.y, point.y)
		bounds.min.z = math.Min(bounds.min.z, point.z)
		bounds.max.x = math.Max(bounds.max.x, point.x)
		bounds.max.y = math.Max(bounds.max.y, point.y)
		bounds.max.z = math.Max(bounds.max.z, point.z)
	}
	return bounds
}

func emptyModelBounds() modelBounds {
	return modelBounds{
		min: modelPoint3{x: math.Inf(1), y: math.Inf(1), z: math.Inf(1)},
		max: modelPoint3{x: math.Inf(-1), y: math.Inf(-1), z: math.Inf(-1)},
	}
}

func (b modelBounds) empty() bool {
	return math.IsInf(b.min.x, 0)
}

func (b *modelBounds) include(other modelBounds) {
	if other.empty() {
		return
	}
	if b.empty() {
		*b = other
		return
	}
	b.min.x = math.Min(b.min.x, other.min.x)
	b.min.y = math.Min(b.min.y, other.min.y)
	b.min.z = math.Min(b.min.z, other.min.z)
	b.max.x = math.Max(b.max.x, other.max.x)
	b.max.y = math.Max(b.max.y, other.max.y)
	b.max.z = math.Max(b.max.z, other.max.z)
}

func rsmMainNode(rsm *res.RSM) *res.RSMNode {
	if rsm == nil || len(rsm.Nodes) == 0 {
		return nil
	}
	if rsm.MainNodeName != "" {
		for i := range rsm.Nodes {
			if rsm.Nodes[i].Name == rsm.MainNodeName {
				return &rsm.Nodes[i]
			}
		}
	}
	return &rsm.Nodes[0]
}

func vectorFromRSM(v res.RSMVector3) modelPoint3 {
	return modelPoint3{x: float64(v.X), y: float64(v.Y), z: float64(v.Z)}
}

func vectorFromRSW(v res.RSWVector3) modelPoint3 {
	return modelPoint3{x: float64(v.X), y: float64(v.Y), z: float64(v.Z)}
}

func rsmVersionAtLeast(rsm *res.RSM, major, minor byte) bool {
	return rsm != nil && (rsm.VersionMajor > major || rsm.VersionMajor == major && rsm.VersionMinor >= minor)
}

func mat4Identity() mat4 {
	return mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func mat4Multiply(a, b mat4) mat4 {
	var out mat4
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			out[col*4+row] =
				a[0*4+row]*b[col*4+0] +
					a[1*4+row]*b[col*4+1] +
					a[2*4+row]*b[col*4+2] +
					a[3*4+row]*b[col*4+3]
		}
	}
	return out
}

func mat4Translate(matrix mat4, v modelPoint3) mat4 {
	t := mat4Identity()
	t[12] = v.x
	t[13] = v.y
	t[14] = v.z
	return mat4Multiply(matrix, t)
}

func mat4Scale(matrix mat4, v modelPoint3) mat4 {
	s := mat4Identity()
	s[0] = v.x
	s[5] = v.y
	s[10] = v.z
	return mat4Multiply(matrix, s)
}

func mat4RotateX(matrix mat4, angle float64) mat4 {
	s, c := math.Sin(angle), math.Cos(angle)
	r := mat4Identity()
	r[5] = c
	r[6] = s
	r[9] = -s
	r[10] = c
	return mat4Multiply(matrix, r)
}

func mat4RotateY(matrix mat4, angle float64) mat4 {
	s, c := math.Sin(angle), math.Cos(angle)
	r := mat4Identity()
	r[0] = c
	r[2] = -s
	r[8] = s
	r[10] = c
	return mat4Multiply(matrix, r)
}

func mat4RotateZ(matrix mat4, angle float64) mat4 {
	s, c := math.Sin(angle), math.Cos(angle)
	r := mat4Identity()
	r[0] = c
	r[1] = s
	r[4] = -s
	r[5] = c
	return mat4Multiply(matrix, r)
}

func mat4RotateAxis(matrix mat4, axis modelPoint3, angle float64) mat4 {
	axis = normalize3(axis)
	if axis == (modelPoint3{}) || angle == 0 {
		return matrix
	}
	x, y, z := axis.x, axis.y, axis.z
	s, c := math.Sin(angle), math.Cos(angle)
	t := 1 - c
	r := mat4Identity()
	r[0] = x*x*t + c
	r[1] = y*x*t + z*s
	r[2] = z*x*t - y*s
	r[4] = x*y*t - z*s
	r[5] = y*y*t + c
	r[6] = z*y*t + x*s
	r[8] = x*z*t + y*s
	r[9] = y*z*t - x*s
	r[10] = z*z*t + c
	return mat4Multiply(matrix, r)
}

func mat4RotateQuat(matrix mat4, q [4]float32) mat4 {
	x, y, z, w := float64(q[0]), float64(q[1]), float64(q[2]), float64(q[3])
	length := math.Sqrt(x*x + y*y + z*z + w*w)
	if length == 0 {
		return matrix
	}
	x /= length
	y /= length
	z /= length
	w /= length

	x2, y2, z2 := x+x, y+y, z+z
	xx, xy, xz := x*x2, x*y2, x*z2
	yy, yz, zz := y*y2, y*z2, z*z2
	wx, wy, wz := w*x2, w*y2, w*z2

	r := mat4Identity()
	r[0] = 1 - yy - zz
	r[1] = xy + wz
	r[2] = xz - wy
	r[4] = xy - wz
	r[5] = 1 - xx - zz
	r[6] = yz + wx
	r[8] = xz + wy
	r[9] = yz - wx
	r[10] = 1 - xx - yy
	return mat4Multiply(matrix, r)
}

func mat4FromMat3(matrix [9]float32) mat4 {
	out := mat4Identity()
	out[0] = float64(matrix[0])
	out[1] = float64(matrix[1])
	out[2] = float64(matrix[2])
	out[4] = float64(matrix[3])
	out[5] = float64(matrix[4])
	out[6] = float64(matrix[5])
	out[8] = float64(matrix[6])
	out[9] = float64(matrix[7])
	out[10] = float64(matrix[8])
	return out
}

func mat4TransformPoint(matrix mat4, point modelPoint3) modelPoint3 {
	return modelPoint3{
		x: matrix[0]*point.x + matrix[4]*point.y + matrix[8]*point.z + matrix[12],
		y: matrix[1]*point.x + matrix[5]*point.y + matrix[9]*point.z + matrix[13],
		z: matrix[2]*point.x + matrix[6]*point.y + matrix[10]*point.z + matrix[14],
	}
}

func rsmFaceTexture(rsm *res.RSM, node *res.RSMNode, face res.RSMFace) (string, [3]texturePoint) {
	textureName := ""
	if int(face.TextureID) < len(node.TextureRefs) {
		textureIndex := node.TextureRefs[face.TextureID].Index
		if textureIndex >= 0 && int(textureIndex) < len(rsm.Textures) {
			textureName = rsm.Textures[textureIndex]
		}
	}
	var uvs [3]texturePoint
	for i, tvertIndex := range face.TextureVertexIndices {
		if int(tvertIndex) >= len(node.TextureVertices) {
			continue
		}
		tvert := node.TextureVertices[tvertIndex]
		uvs[i] = texturePoint{u: tvert.U, v: tvert.V}
	}
	return textureName, uvs
}

func rsmFaceColor(textureName string, a, b, c modelPoint3) color.RGBA {
	base := textureColor(textureName)
	if textureName != "" {
		base = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	normal := normalize3(cross3(sub3(b, a), sub3(c, a)))
	light := normalize3(modelPoint3{x: -0.45, y: -0.78, z: -0.42})
	diffuse := math.Abs(normal.x*light.x + normal.y*light.y + normal.z*light.z)
	top := math.Max(0, normal.y)
	side := math.Min(1, math.Abs(normal.x)+math.Abs(normal.z))
	shade := 0.42 + diffuse*0.38 + top*0.16 + side*0.08
	return color.RGBA{
		R: clampColor(float64(base.R) * shade),
		G: clampColor(float64(base.G) * shade),
		B: clampColor(float64(base.B) * shade),
		A: 255,
	}
}

func drawTexturedTriangle(screen, texture *ebiten.Image, points [3]screenPoint, uvs [3]texturePoint, tint color.RGBA) {
	bounds := texture.Bounds()
	w := float32(bounds.Dx())
	h := float32(bounds.Dy())
	r := float32(tint.R) / 255
	g := float32(tint.G) / 255
	b := float32(tint.B) / 255
	a := float32(tint.A) / 255
	vertices := []ebiten.Vertex{
		{DstX: points[0].x, DstY: points[0].y, SrcX: uvs[0].u * w, SrcY: uvs[0].v * h, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: points[1].x, DstY: points[1].y, SrcX: uvs[1].u * w, SrcY: uvs[1].v * h, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: points[2].x, DstY: points[2].y, SrcX: uvs[2].u * w, SrcY: uvs[2].v * h, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
	}
	op := &ebiten.DrawTrianglesOptions{
		Filter:  ebiten.FilterLinear,
		Address: ebiten.AddressRepeat,
	}
	screen.DrawTriangles(vertices, []uint16{0, 1, 2}, texture, op)
}

func drawColoredTriangle(screen, white *ebiten.Image, points [3]screenPoint, c color.RGBA) {
	r := float32(c.R) / 255
	g := float32(c.G) / 255
	b := float32(c.B) / 255
	a := float32(c.A) / 255
	vertices := []ebiten.Vertex{
		{DstX: points[0].x, DstY: points[0].y, SrcX: 0, SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: points[1].x, DstY: points[1].y, SrcX: 1, SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: points[2].x, DstY: points[2].y, SrcX: 1, SrcY: 1, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
	}
	screen.DrawTriangles(vertices, []uint16{0, 1, 2}, white, nil)
}

func triangleOutside(points [3]screenPoint, width, height float64) bool {
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

func signedScreenArea(points [3]screenPoint) float32 {
	return (points[1].x-points[0].x)*(points[2].y-points[0].y) - (points[2].x-points[0].x)*(points[1].y-points[0].y)
}

func applyMatrix3(point modelPoint3, matrix [9]float32) modelPoint3 {
	return modelPoint3{
		x: float64(matrix[0])*point.x + float64(matrix[3])*point.y + float64(matrix[6])*point.z,
		y: float64(matrix[1])*point.x + float64(matrix[4])*point.y + float64(matrix[7])*point.z,
		z: float64(matrix[2])*point.x + float64(matrix[5])*point.y + float64(matrix[8])*point.z,
	}
}

func rotateX(point modelPoint3, angle float64) modelPoint3 {
	s, c := math.Sin(angle), math.Cos(angle)
	return modelPoint3{x: point.x, y: point.y*c - point.z*s, z: point.y*s + point.z*c}
}

func rotateY(point modelPoint3, angle float64) modelPoint3 {
	s, c := math.Sin(angle), math.Cos(angle)
	return modelPoint3{x: point.x*c + point.z*s, y: point.y, z: -point.x*s + point.z*c}
}

func rotateZ(point modelPoint3, angle float64) modelPoint3 {
	s, c := math.Sin(angle), math.Cos(angle)
	return modelPoint3{x: point.x*c - point.y*s, y: point.x*s + point.y*c, z: point.z}
}

func rotateAxisAngle(point, axis modelPoint3, angle float64) modelPoint3 {
	axis = normalize3(axis)
	if axis == (modelPoint3{}) || angle == 0 {
		return point
	}
	s, c := math.Sin(angle), math.Cos(angle)
	dot := point.x*axis.x + point.y*axis.y + point.z*axis.z
	cross := cross3(axis, point)
	return modelPoint3{
		x: point.x*c + cross.x*s + axis.x*dot*(1-c),
		y: point.y*c + cross.y*s + axis.y*dot*(1-c),
		z: point.z*c + cross.z*s + axis.z*dot*(1-c),
	}
}

func sub3(a, b modelPoint3) modelPoint3 {
	return modelPoint3{x: a.x - b.x, y: a.y - b.y, z: a.z - b.z}
}

func cross3(a, b modelPoint3) modelPoint3 {
	return modelPoint3{
		x: a.y*b.z - a.z*b.y,
		y: a.z*b.x - a.x*b.z,
		z: a.x*b.y - a.y*b.x,
	}
}

func normalize3(v modelPoint3) modelPoint3 {
	length := math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
	if length == 0 {
		return modelPoint3{}
	}
	return modelPoint3{x: v.x / length, y: v.y / length, z: v.z / length}
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func rsmRenderRadius() float64 {
	raw := os.Getenv("GORO_RSM_RENDER_RADIUS")
	if raw == "" {
		return 42
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value <= 0 {
		return 42
	}
	return value
}

func rsmMaxFaces() int {
	raw := os.Getenv("GORO_RSM_MAX_FACES")
	if raw == "" {
		return 5000
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 5000
	}
	return value
}

func rsmHeightScale() float64 {
	raw := os.Getenv("GORO_RSM_HEIGHT_SCALE")
	if raw == "" {
		return 2.8
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value <= 0 {
		return 2.8
	}
	return value
}
