package render

import "testing"

func TestApplyCameraFog3DMixesByCameraDepth(t *testing.T) {
	vertices := []Vertex3D{{
		X: 0, Y: 0, Z: 0,
		ColorR: 0.2, ColorG: 0.4, ColorB: 0.6, ColorA: 1,
	}}
	camera := Camera3D{
		Enabled: true,
		ViewProjection: [16]float32{
			0: 1, 5: 1, 10: 1, 15: 1,
		},
		Fog: Fog3D{
			Enabled: true,
			Near:    0,
			Far:     2,
			ColorR:  1,
			ColorG:  0,
			ColorB:  0,
			Factor:  1,
		},
	}

	applyCameraFog3D(vertices, camera)

	if vertices[0].ColorR <= 0.2 || vertices[0].ColorG >= 0.4 || vertices[0].ColorB >= 0.6 {
		t.Fatalf("fog did not mix toward red: %+v", vertices[0])
	}
	if vertices[0].ColorA != 1 {
		t.Fatalf("fog changed alpha: %+v", vertices[0])
	}
}
