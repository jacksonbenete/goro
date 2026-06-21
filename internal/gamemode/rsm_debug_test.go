package gamemode

import (
	"math"
	"os"
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
			mainBox:   bounds.main,
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

func debugRSMModelMatrix(rsm *res.RSM, node *res.RSMNode, nodeMatrix mat4, instance modelInstance) mat4 {
	localMatrix := mat4Identity()
	localMatrix = mat4Translate(localMatrix, modelPoint3{
		x: -(instance.mainBox.min.x + instance.mainBox.max.x) * 0.5,
		y: instance.mainBox.max.y,
		z: -(instance.mainBox.min.z + instance.mainBox.max.z) * 0.5,
	})
	localMatrix = mat4Scale(localMatrix, modelPoint3{x: 1, y: -1, z: 1})
	localMatrix = mat4Multiply(localMatrix, nodeMatrix)
	if len(rsm.Nodes) != 1 {
		localMatrix = mat4Translate(localMatrix, vectorFromRSM(node.Offset))
	}
	localMatrix = mat4Multiply(localMatrix, mat4FromMat3(node.Matrix))
	return mat4Multiply(instance.matrix, localMatrix)
}
