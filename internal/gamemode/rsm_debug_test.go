package gamemode

import (
	"math"
	"os"
	"sort"
	"testing"

	"github.com/kivutar/goro/internal/res"
	worldstate "github.com/kivutar/goro/internal/world"
)

func TestDebugGeffenCenterModels(t *testing.T) {
	if os.Getenv("GORO_DEBUG_GEFFEN_RSM") != "1" {
		t.Skip("set GORO_DEBUG_GEFFEN_RSM=1")
	}
	manager, err := res.NewManager("/home/kivutar/Téléchargements/OldRO")
	if err != nil {
		t.Fatal(err)
	}
	gndData, err := manager.ReadFile("data\\geffen.gnd")
	if err != nil {
		t.Fatal(err)
	}
	gnd, err := res.ParseGND(gndData)
	if err != nil {
		t.Fatal(err)
	}
	gatData, err := manager.ReadFile("data\\geffen.gat")
	if err != nil {
		t.Fatal(err)
	}
	gat, err := res.ParseGAT(gatData)
	if err != nil {
		t.Fatal(err)
	}
	rswData, err := manager.ReadFile("data\\geffen.rsw")
	if err != nil {
		t.Fatal(err)
	}
	rsw, err := res.ParseRSW(rswData)
	if err != nil {
		t.Fatal(err)
	}
	world := &worldstate.World{GND: gnd, GAT: gat}
	t.Logf("terrain center gnd=%.2f gat=%.2f runtime=%.2f camera=%.2f", terrainHeightAt(&worldstate.World{GND: gnd}, 120, 120), terrainHeightAt(&worldstate.World{GAT: gat}, 120, 120), terrainHeightAt(world, 120, 120), cameraTargetHeightAt(world, 120, 120))
	for _, index := range []int{204, 625, 628} {
		placement := rsw.Models[index]
		rsm, err := loadRSMModel(manager, placement.Filename)
		if err != nil {
			t.Fatalf("%d %s: %v", index, placement.Filename, err)
		}
		baseX := float64(placement.Position.X) + float64(gnd.Width)
		baseY := float64(placement.Position.Z) + float64(gnd.Height)
		projection := sceneProjection{
			screenW:        1024,
			screenH:        768,
			camera:         true,
			viewProjection: sceneCameraMatrix(1024, 768, 120.5, 120.5, cameraTargetHeightAt(world, 120, 120)),
		}
		minX, minY := math.Inf(1), math.Inf(1)
		maxX, maxY := math.Inf(-1), math.Inf(-1)
		minWorldY, maxWorldY := math.Inf(1), math.Inf(-1)
		inFront := 0
		triangles := 0
		bounds := calculateRSMBounds(rsm)
		nodes := buildRSMNodeMatrices(rsm)
		instance := modelInstance{
			placement: placement,
			bounds:    bounds.model,
			baseX:     baseX,
			baseY:     baseY,
			matrix:    buildRSMInstanceMatrix(rsm, placement, baseX, baseY, bounds.model),
		}
		for nodeIndex := range rsm.Nodes {
			node := &rsm.Nodes[nodeIndex]
			modelMatrix := debugRSMModelMatrix(rsm, node, nodes[node.Name], instance)
			triangles += len(buildRSMNodeTriangles(rsm, node, nodes[node.Name], instance, projection, 1024, 768))
			for _, vertex := range node.Vertices {
				world := mat4TransformPoint(modelMatrix, vectorFromRSM(vertex))
				minWorldY = math.Min(minWorldY, world.y)
				maxWorldY = math.Max(maxWorldY, world.y)
				_, _, _, clipW := mat4TransformVec4(projection.viewProjection, world.x, world.y, world.z, 1)
				if clipW > 0.001 {
					inFront++
				}
				point := projection.Project(world.x, world.z, world.y)
				minX = math.Min(minX, float64(point.x))
				minY = math.Min(minY, float64(point.y))
				maxX = math.Max(maxX, float64(point.x))
				maxY = math.Max(maxY, float64(point.y))
			}
		}
		t.Logf("#%d %s base=(%.1f,%.1f,%.1f) terrain=%.1f worldY=(%.1f,%.1f) inFront=%d triangles=%d screen=(%.1f,%.1f)-(%.1f,%.1f)", index, placement.Filename, baseX, baseY, placement.Position.Y, terrainHeightAt(world, baseX, baseY), minWorldY, maxWorldY, inFront, triangles, minX, minY, maxX, maxY)
	}
}

func TestDebugIzludeModelSizes(t *testing.T) {
	if os.Getenv("GORO_DEBUG_IZLUDE_RSM") != "1" {
		t.Skip("set GORO_DEBUG_IZLUDE_RSM=1")
	}
	manager, err := res.NewManager("/home/kivutar/Téléchargements/OldRO")
	if err != nil {
		t.Fatal(err)
	}
	gndData, err := manager.ReadFile("data\\izlude.gnd")
	if err != nil {
		t.Fatal(err)
	}
	gnd, err := res.ParseGND(gndData)
	if err != nil {
		t.Fatal(err)
	}
	rswData, err := manager.ReadFile("data\\izlude.rsw")
	if err != nil {
		t.Fatal(err)
	}
	rsw, err := res.ParseRSW(rswData)
	if err != nil {
		t.Fatal(err)
	}

	type modelSize struct {
		index         int
		filename      string
		nodeName      string
		version       string
		nodes         int
		selectedNodes int
		scale         res.RSWVector3
		width         float64
		height        float64
		depth         float64
		placement     res.RSWModel
	}
	var sizes []modelSize
	seen := make(map[string]struct{})
	for index, placement := range rsw.Models {
		if placement.Filename == "" {
			continue
		}
		key := placement.Filename
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		rsm, err := loadRSMModel(manager, placement.Filename)
		if err != nil {
			continue
		}
		nodeIndices := selectedRSMNodeIndices(rsm, placement.NodeName)
		bounds := calculateRSMBoundsForNodes(rsm, nodeIndices).model
		scale := vectorFromRSW(placement.Scale)
		if index == 587 || index == 599 {
			t.Logf("detail #%d rawBounds min=(%.1f,%.1f,%.1f) max=(%.1f,%.1f,%.1f) openMidgardOffset=(%.1f,%.1f,%.1f)",
				index,
				bounds.min.x,
				bounds.min.y,
				bounds.min.z,
				bounds.max.x,
				bounds.max.y,
				bounds.max.z,
				-(bounds.min.x+bounds.max.x)*0.5,
				-bounds.max.y,
				-(bounds.min.z+bounds.max.z)*0.5)
		}
		sizes = append(sizes, modelSize{
			index:         index,
			filename:      placement.Filename,
			nodeName:      placement.NodeName,
			version:       string([]byte{'0' + rsm.VersionMajor, '.', '0' + rsm.VersionMinor}),
			nodes:         len(rsm.Nodes),
			selectedNodes: len(nodeIndices),
			scale:         placement.Scale,
			width:         math.Abs(bounds.max.x-bounds.min.x) * math.Abs(scale.x),
			height:        math.Abs(bounds.max.y-bounds.min.y) * math.Abs(scale.y),
			depth:         math.Abs(bounds.max.z-bounds.min.z) * math.Abs(scale.z),
			placement:     placement,
		})
	}
	sort.SliceStable(sizes, func(i, j int) bool {
		return sizes[i].height > sizes[j].height
	})
	limit := minInt(24, len(sizes))
	for i := 0; i < limit; i++ {
		size := sizes[i]
		t.Logf("#%d %s nameBytes=%x node=%q v=%s nodes=%d selected=%d scale=(%.2f,%.2f,%.2f) rot=(%.1f,%.1f,%.1f) dims=(%.1f,%.1f,%.1f) pos=(%.1f,%.1f,%.1f)",
			size.index,
			size.filename,
			[]byte(size.filename),
			size.nodeName,
			size.version,
			size.nodes,
			size.selectedNodes,
			size.scale.X,
			size.scale.Y,
			size.scale.Z,
			size.placement.Rotation.X,
			size.placement.Rotation.Y,
			size.placement.Rotation.Z,
			size.width,
			size.height,
			size.depth,
			size.placement.Position.X+float32(gnd.Width),
			size.placement.Position.Y,
			size.placement.Position.Z+float32(gnd.Height),
		)
	}

	for _, index := range []int{63, 587} {
		placement := rsw.Models[index]
		rsm, err := loadRSMModel(manager, placement.Filename)
		if err != nil {
			t.Fatalf("%d %s: %v", index, placement.Filename, err)
		}
		nodeIndices := selectedRSMNodeIndices(rsm, placement.NodeName)
		bounds := calculateRSMBoundsForNodes(rsm, nodeIndices).model
		baseX := float64(placement.Position.X) + float64(gnd.Width)
		baseY := float64(placement.Position.Z) + float64(gnd.Height)
		terrain := terrainHeightAt(&worldstate.World{GND: gnd}, baseX, baseY)
		projection := newSceneProjectionForSize(1024, 768, int(math.Round(baseX)), int(math.Round(baseY)), terrain)
		nodes := buildRSMNodeMatrices(rsm)
		instance := modelInstance{
			placement: placement,
			bounds:    bounds,
			baseX:     baseX,
			baseY:     baseY,
			matrix:    buildRSMInstanceMatrix(rsm, placement, baseX, baseY, bounds),
		}
		minX, minY := math.Inf(1), math.Inf(1)
		maxX, maxY := math.Inf(-1), math.Inf(-1)
		minClipW, maxClipW := math.Inf(1), math.Inf(-1)
		visibleVertices := 0
		totalVertices := 0
		for _, nodeIndex := range nodeIndices {
			node := &rsm.Nodes[nodeIndex]
			modelMatrix := debugRSMModelMatrix(rsm, node, nodes[node.Name], instance)
			for _, vertex := range node.Vertices {
				world := mat4TransformPoint(modelMatrix, vectorFromRSM(vertex))
				totalVertices++
				_, _, _, clipW := mat4TransformVec4(projection.viewProjection, world.x, world.y, world.z, 1)
				minClipW = math.Min(minClipW, clipW)
				maxClipW = math.Max(maxClipW, clipW)
				if projection.VisibleForTriangle(world.x, world.z, world.y) {
					visibleVertices++
				}
				point := projection.Project(world.x, world.z, world.y)
				minX = math.Min(minX, float64(point.x))
				minY = math.Min(minY, float64(point.y))
				maxX = math.Max(maxX, float64(point.x))
				maxY = math.Max(maxY, float64(point.y))
			}
		}
		t.Logf("screen #%d base=(%.1f,%.1f,%.1f) terrain=%.1f deltaY=%.1f screen=(%.1f,%.1f)-(%.1f,%.1f) size=(%.1fx%.1f) clipW=(%.1f,%.1f) visibleVerts=%d/%d",
			index,
			baseX,
			baseY,
			placement.Position.Y,
			terrain,
			float64(placement.Position.Y)-terrain,
			minX,
			minY,
			maxX,
			maxY,
			maxX-minX,
			maxY-minY,
			minClipW,
			maxClipW,
			visibleVertices,
			totalVertices)
	}
}

func debugRSMModelMatrix(rsm *res.RSM, node *res.RSMNode, nodeMatrix mat4, instance modelInstance) mat4 {
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
	return mat4Multiply(instance.matrix, localMatrix)
}
