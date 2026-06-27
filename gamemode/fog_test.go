package gamemode

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/kivutar/goro/core"
	"github.com/kivutar/goro/res"
)

func TestSceneFogFromMapUsesRObrowserScale(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "fogparametertable.txt"), []byte("prontera#0.5#1.5#ffffff#1.0#"), 0o644); err != nil {
		t.Fatal(err)
	}
	manager, err := res.NewManager(root)
	if err != nil {
		t.Fatal(err)
	}

	fog := sceneFogFromMap(manager, "prontera", core.FogConfig{Enabled: true})
	if !fog.enabled {
		t.Fatal("expected fog")
	}
	if fog.near != 120 || fog.far != 360 {
		t.Fatalf("unexpected scaled distances: near=%v far=%v", fog.near, fog.far)
	}
}

func TestSceneFogMixColorSmoothstepsToFogColor(t *testing.T) {
	fog := sceneFog{
		enabled: true,
		near:    10,
		far:     20,
		color:   color.RGBA{R: 200, G: 100, B: 50, A: 255},
		factor:  1,
	}
	base := color.RGBA{R: 100, G: 100, B: 100, A: 180}
	if got := fog.mixColor(base, 5); got != base {
		t.Fatalf("near color changed: %#v", got)
	}
	if got := fog.mixColor(base, 20); got != (color.RGBA{R: 200, G: 100, B: 50, A: 180}) {
		t.Fatalf("far color mismatch: %#v", got)
	}
	if got := fog.mixColor(base, 15); got != (color.RGBA{R: 150, G: 100, B: 75, A: 180}) {
		t.Fatalf("mid color mismatch: %#v", got)
	}
}

func TestSceneFogAttenuateColorSmoothstepsToBlack(t *testing.T) {
	fog := sceneFog{
		enabled: true,
		near:    10,
		far:     20,
		color:   color.RGBA{R: 200, G: 100, B: 50, A: 255},
	}
	base := color.RGBA{R: 100, G: 80, B: 60, A: 180}
	if got := fog.attenuateColor(base, 5); got != base {
		t.Fatalf("near color changed: %#v", got)
	}
	if got := fog.attenuateColor(base, 20); got != (color.RGBA{A: 180}) {
		t.Fatalf("far color mismatch: %#v", got)
	}
	if got := fog.attenuateColor(base, 15); got != (color.RGBA{R: 50, G: 40, B: 30, A: 180}) {
		t.Fatalf("mid color mismatch: %#v", got)
	}
}

func TestSceneFogVeilAlphaUsesCameraDepth(t *testing.T) {
	fog := sceneFog{
		enabled: true,
		near:    120,
		far:     360,
		color:   color.RGBA{R: 100, G: 160, B: 100, A: 255},
	}
	projection := sceneProjection{cameraZoom: 150}
	alpha := sceneFogVeilAlpha(fog, projection, core.FogConfig{Enabled: true, VeilStrength: 0.10, VeilDepthScale: 1.2})
	if alpha == 0 {
		t.Fatal("expected visible fog veil alpha")
	}
}

func TestSceneFogVeilAlphaIgnoresMapFactor(t *testing.T) {
	fog := sceneFog{
		enabled: true,
		near:    24,
		far:     216,
		color:   color.RGBA{R: 4, B: 154, A: 255},
		factor:  0.3,
	}
	projection := sceneProjection{cameraZoom: 150}
	got := sceneFogVeilAlpha(fog, projection, core.FogConfig{Enabled: true, VeilStrength: 0.10, VeilDepthScale: 1.2})
	withoutFactor := fog
	withoutFactor.factor = 0
	want := sceneFogVeilAlpha(withoutFactor, projection, core.FogConfig{Enabled: true, VeilStrength: 0.10, VeilDepthScale: 1.2})
	if got != want {
		t.Fatalf("veil alpha = %d, want factor-independent alpha %d", got, want)
	}
}
