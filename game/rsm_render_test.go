package game

import (
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/kivutar/goro/res"
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

func TestRSMFaceColorShadeNoneUsesDefaultModelNormal(t *testing.T) {
	lighting := sceneLighting{
		direction: modelPoint3{y: -1},
		diffuse:   modelPoint3{x: 1, y: 1, z: 1},
		env:       modelPoint3{x: 1, y: 1, z: 1},
	}
	triangleNormalWouldFaceAway := [3]modelPoint3{
		{},
		{x: 1},
		{z: -1},
	}

	got := rsmFaceColor(&res.RSM{ShadeType: 0}, "model.bmp",
		triangleNormalWouldFaceAway[0],
		triangleNormalWouldFaceAway[1],
		triangleNormalWouldFaceAway[2],
		lighting,
	)
	want := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if got != want {
		t.Fatalf("shade-none face color = %#v, want %#v", got, want)
	}

	flat := rsmFaceColor(&res.RSM{ShadeType: 1}, "model.bmp",
		triangleNormalWouldFaceAway[0],
		triangleNormalWouldFaceAway[1],
		triangleNormalWouldFaceAway[2],
		lighting,
	)
	if flat == want {
		t.Fatalf("flat-shaded face unexpectedly matched shade-none color %#v", flat)
	}
}

func TestRSMAnimationFrameLoopsOnMillisecondTick(t *testing.T) {
	rsm := &res.RSM{
		AnimLength: 12,
		Nodes: []res.RSMNode{
			{
				Name: "root",
				PositionKeyframes: []res.RSMPositionKeyframe{
					{Frame: 0},
					{Frame: 12, Pos: res.RSMVector3{X: 12}},
				},
			},
		},
	}

	frame, animated := rsmAnimationFrame(rsm, res.RSWModel{AnimType: 2}, time.UnixMilli(1015))
	if !animated {
		t.Fatal("model should be animated")
	}
	if frame != 7 {
		t.Fatalf("animation frame = %d, want 7", frame)
	}
}

func TestRSMAnimationFrameIgnoresRSWAnimationSpeedLikeRobr(t *testing.T) {
	rsm := &res.RSM{
		AnimLength: 6400,
		Nodes: []res.RSMNode{
			{
				Name: "root",
				RotationKeyframes: []res.RSMRotationKeyframe{
					{Frame: 0, Quaternion: [4]float32{0, 0, 0, 1}},
					{Frame: 160, Quaternion: [4]float32{0, 0, 1, 0}},
				},
			},
		},
	}

	frame, animated := rsmAnimationFrame(rsm, res.RSWModel{AnimType: 2, AnimSpeed: 0.7}, time.UnixMilli(160))
	if !animated {
		t.Fatal("model should be animated")
	}
	if frame != 160 {
		t.Fatalf("animation frame = %d, want 160", frame)
	}

	frame, animated = rsmAnimationFrame(rsm, res.RSWModel{AnimType: 2, AnimSpeed: 0.7}, time.UnixMilli(1000))
	if !animated {
		t.Fatal("model should be animated")
	}
	if frame != 1000 {
		t.Fatalf("animation frame = %d, want 1000", frame)
	}
}

func TestRSMAnimationFrameAnimatesKeyframedModelsWithoutRSWAnimType(t *testing.T) {
	rsm := &res.RSM{
		AnimLength: 12,
		Nodes: []res.RSMNode{
			{
				Name: "root",
				PositionKeyframes: []res.RSMPositionKeyframe{
					{Frame: 0},
					{Frame: 12, Pos: res.RSMVector3{X: 12}},
				},
			},
		},
	}

	frame, animated := rsmAnimationFrame(rsm, res.RSWModel{AnimType: 0}, time.UnixMilli(14))
	if !animated {
		t.Fatal("model should be animated")
	}
	if frame != 2 {
		t.Fatalf("animation frame = %d, want 2", frame)
	}
}

func TestRSMFaceColorUsesModelAlpha(t *testing.T) {
	got := rsmFaceColor(&res.RSM{ShadeType: 0, Alpha: 0.5}, "model.bmp",
		modelPoint3{x: 0, y: 0, z: 0},
		modelPoint3{x: 1, y: 0, z: 0},
		modelPoint3{x: 0, y: 0, z: 1},
		sceneLighting{})
	if got.A < 126 || got.A > 129 {
		t.Fatalf("alpha = %d, want about 128", got.A)
	}
}

func TestBuildRSMNodeMatricesSamplesAnimatedPositionAndScale(t *testing.T) {
	rsm := &res.RSM{Nodes: []res.RSMNode{
		{
			Name:   "root",
			Matrix: identityRSMMatrix(),
			PositionKeyframes: []res.RSMPositionKeyframe{
				{Frame: 0, Pos: res.RSMVector3{}},
				{Frame: 10, Pos: res.RSMVector3{X: 10}},
			},
			ScaleKeyframes: []res.RSMScaleKeyframe{
				{Frame: 0, Scale: res.RSMVector3{X: 1, Y: 1, Z: 1}},
				{Frame: 10, Scale: res.RSMVector3{X: 3, Y: 3, Z: 3}},
			},
		},
	}}

	matrix := buildRSMNodeMatrices(rsm, 5)["root"]
	got := mat4TransformPoint(matrix, modelPoint3{x: 1})
	want := modelPoint3{x: 7}
	if math.Abs(got.x-want.x) > 0.0001 || math.Abs(got.y-want.y) > 0.0001 || math.Abs(got.z-want.z) > 0.0001 {
		t.Fatalf("animated point = (%.3f, %.3f, %.3f), want (%.3f, %.3f, %.3f)", got.x, got.y, got.z, want.x, want.y, want.z)
	}
}

func TestBuildRSMNodeMatricesSlerpsAnimatedRotation(t *testing.T) {
	rsm := &res.RSM{Nodes: []res.RSMNode{
		{
			Name:   "root",
			Matrix: identityRSMMatrix(),
			Scale:  res.RSMVector3{X: 1, Y: 1, Z: 1},
			RotationKeyframes: []res.RSMRotationKeyframe{
				{Frame: 0, Quaternion: [4]float32{0, 0, 0, 1}},
				{Frame: 10, Quaternion: [4]float32{0, 0, 1, 0}},
			},
		},
	}}

	matrix := buildRSMNodeMatrices(rsm, 5)["root"]
	got := mat4TransformPoint(matrix, modelPoint3{x: 1})
	want := modelPoint3{y: 1}
	if math.Abs(got.x-want.x) > 0.0001 || math.Abs(got.y-want.y) > 0.0001 || math.Abs(got.z-want.z) > 0.0001 {
		t.Fatalf("animated rotation point = (%.3f, %.3f, %.3f), want (%.3f, %.3f, %.3f)", got.x, got.y, got.z, want.x, want.y, want.z)
	}
}

func identityRSMMatrix() [9]float32 {
	return [9]float32{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}
