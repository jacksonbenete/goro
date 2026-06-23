package gamemode

import (
	"testing"

	"github.com/kivutar/goro/internal/res"
)

func TestCalculateRSMBoundsSeparatesMainNodeFromModel(t *testing.T) {
	rsm := &res.RSM{
		MainNodeName: "root",
		Nodes: []res.RSMNode{
			{
				Name:   "root",
				Matrix: identityRSMMatrix(),
				Scale:  res.RSMVector3{X: 1, Y: 1, Z: 1},
				Vertices: []res.RSMVector3{
					{X: -1, Y: 0, Z: -1},
					{X: 1, Y: 2, Z: 1},
				},
			},
			{
				Name:       "child",
				ParentName: "root",
				Matrix:     identityRSMMatrix(),
				Position:   res.RSMVector3{X: 100, Y: 0, Z: 0},
				Scale:      res.RSMVector3{X: 1, Y: 1, Z: 1},
				Vertices: []res.RSMVector3{
					{X: 0, Y: 0, Z: 0},
					{X: 2, Y: 3, Z: 2},
				},
			},
		},
	}

	bounds := calculateRSMBounds(rsm)
	if bounds.main.min.x != -1 || bounds.main.max.x != 1 {
		t.Fatalf("main bounds x = %.1f..%.1f, want -1..1", bounds.main.min.x, bounds.main.max.x)
	}
	if bounds.model.min.x != -1 || bounds.model.max.x != 102 {
		t.Fatalf("model bounds x = %.1f..%.1f, want -1..102", bounds.model.min.x, bounds.model.max.x)
	}
}

func TestRSMInstanceMatrixUsesParsedRSWModelY(t *testing.T) {
	rsm := &res.RSM{}
	placement := res.RSWModel{
		Position: res.RSWVector3{X: 10, Y: -7, Z: 20},
		Scale:    res.RSWVector3{X: 1, Y: 1, Z: 1},
	}

	matrix := buildRSMInstanceMatrix(rsm, placement, 110, 220, modelBounds{})
	point := mat4TransformPoint(matrix, modelPoint3{})
	if point.y != -7 {
		t.Fatalf("instance y = %.1f, want -7", point.y)
	}
}

func identityRSMMatrix() [9]float32 {
	return [9]float32{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}
