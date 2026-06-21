package gamemode

import (
	"math"
	"testing"

	"github.com/kivutar/goro/internal/res"
	worldstate "github.com/kivutar/goro/internal/world"
)

func TestBilinearHeight(t *testing.T) {
	heights := [4]float32{0, 10, 20, 30}
	if got := bilinearHeight(heights, 0.5, 0.5); got != 15 {
		t.Fatalf("height = %v, want 15", got)
	}
}

func TestTerrainHeightPrefersGAT(t *testing.T) {
	world := &worldstate.World{
		GAT: &res.GAT{
			Width:  1,
			Height: 1,
			Cells:  []res.GATCell{{Heights: [4]float32{10, 20, 30, 40}}},
		},
		GND: &res.GND{
			Width:  1,
			Height: 1,
			Cells:  []res.GNDCell{{Heights: [4]float32{1, 1, 1, 1}}},
		},
	}
	if got := terrainHeightAt(world, 0, 0); got != 25 {
		t.Fatalf("terrain height = %v, want GAT center height 25", got)
	}
}

func TestProjectionCentersPlayerCell(t *testing.T) {
	projection := sceneProjection{
		playerX:     10.5,
		playerY:     20.5,
		centerX:     400,
		centerY:     300,
		tileW:       sceneTileW,
		tileH:       sceneTileH,
		heightScale: 2,
	}
	point := projection.Project(10.5, 20.5, 5)
	if point.x != 400 || point.y != 290 {
		t.Fatalf("projected point = %.1f, %.1f, want 400, 290", point.x, point.y)
	}
}

func TestCameraProjectionCentersPlayerCell(t *testing.T) {
	projection := sceneProjection{
		playerX:        10.5,
		playerY:        20.5,
		playerZ:        5,
		centerX:        400,
		centerY:        300,
		screenW:        800,
		screenH:        600,
		camera:         true,
		viewProjection: sceneCameraMatrix(800, 600, 10.5, 20.5, 5),
	}
	point := projection.Project(10.5, 20.5, 5)
	if math.Abs(float64(point.x)-400) > 0.01 || math.Abs(float64(point.y)-300) > 0.01 {
		t.Fatalf("projected point = %.1f, %.1f, want 400, 300", point.x, point.y)
	}
}

func TestCameraProjectionMovesHigherWorldPointUpScreen(t *testing.T) {
	projection := sceneProjection{
		screenW:        800,
		screenH:        600,
		camera:         true,
		viewProjection: sceneCameraMatrix(800, 600, 10.5, 20.5, 5),
	}
	base := projection.Project(10.5, 20.5, 5)
	higher := projection.Project(10.5, 20.5, 15)
	if higher.y >= base.y {
		t.Fatalf("higher point y = %.1f, base y = %.1f, want higher point above base", higher.y, base.y)
	}
}

func TestCameraDepthOrdersFarBeforeNear(t *testing.T) {
	projection := sceneProjection{
		screenW:        800,
		screenH:        600,
		camera:         true,
		viewProjection: sceneCameraMatrix(800, 600, 10.5, 20.5, 5),
	}
	near := projection.Depth(10.5, 20.5, 5)
	far := projection.Depth(10.5, 30.5, 5)
	if far <= near {
		t.Fatalf("far depth %.2f must be greater than near depth %.2f for painter sorting", far, near)
	}
}
