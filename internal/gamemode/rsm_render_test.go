package gamemode

import (
	"math"
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

func TestRSMInstanceMatrixMirrorsVerticalBasisRotations(t *testing.T) {
	rsm := &res.RSM{}
	tests := []struct {
		name      string
		rotation  res.RSWVector3
		point     modelPoint3
		wantPoint modelPoint3
	}{
		{
			name:      "pitch",
			rotation:  res.RSWVector3{X: 90},
			point:     modelPoint3{y: 1},
			wantPoint: modelPoint3{z: -1},
		},
		{
			name:      "yaw",
			rotation:  res.RSWVector3{Y: 90},
			point:     modelPoint3{x: 1},
			wantPoint: modelPoint3{z: -1},
		},
		{
			name:      "roll",
			rotation:  res.RSWVector3{Z: 90},
			point:     modelPoint3{x: 1},
			wantPoint: modelPoint3{y: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matrix := buildRSMInstanceMatrix(rsm, res.RSWModel{
				Rotation: tt.rotation,
				Scale:    res.RSWVector3{X: 1, Y: 1, Z: 1},
			}, 0, 0, modelBounds{})
			got := mat4TransformPoint(matrix, tt.point)
			if math.Abs(got.x-tt.wantPoint.x) > 0.0001 || math.Abs(got.y-tt.wantPoint.y) > 0.0001 || math.Abs(got.z-tt.wantPoint.z) > 0.0001 {
				t.Fatalf("point = (%.3f, %.3f, %.3f), want (%.3f, %.3f, %.3f)", got.x, got.y, got.z, tt.wantPoint.x, tt.wantPoint.y, tt.wantPoint.z)
			}
		})
	}
}

func identityRSMMatrix() [9]float32 {
	return [9]float32{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}
