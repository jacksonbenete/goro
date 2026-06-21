package res

import (
	"errors"
	"os"
	"testing"
)

func TestGNDRealFileWhenConfigured(t *testing.T) {
	path := os.Getenv("GORO_TEST_GND")
	if path == "" {
		t.Skip("set GORO_TEST_GND to run against a real GND file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	gnd, err := ParseGND(data)
	if err != nil {
		t.Fatal(err)
	}
	if gnd.Width <= 0 || gnd.Height <= 0 || len(gnd.Cells) != gnd.Width*gnd.Height {
		t.Fatalf("invalid parsed gnd: %dx%d cells=%d", gnd.Width, gnd.Height, len(gnd.Cells))
	}
	if len(gnd.Surfaces) == 0 {
		t.Fatal("real gnd has no surfaces")
	}
}

func TestGNDRealArchiveWhenConfigured(t *testing.T) {
	path := os.Getenv("GORO_TEST_GRF")
	if path == "" {
		t.Skip("set GORO_TEST_GRF to run against a real archive")
	}
	name := os.Getenv("GORO_TEST_GND_FILE")
	if name == "" {
		name = "geffen_in.gnd"
	}

	grf, err := OpenGRF(path)
	if err != nil {
		if errors.Is(err, ErrGRFUnsupportedVersion) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer grf.Close()

	if !grf.Has(name) {
		matches := grf.NamesWithSuffix(name)
		if len(matches) == 0 {
			t.Skipf("%s not present in %s", name, path)
		}
		name = matches[0]
	}
	data, err := grf.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	gnd, err := ParseGND(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	if gnd.Width <= 0 || gnd.Height <= 0 || len(gnd.Surfaces) == 0 {
		t.Fatalf("invalid parsed gnd %s: %dx%d surfaces=%d", name, gnd.Width, gnd.Height, len(gnd.Surfaces))
	}
}

func TestGNDRealArchiveTexturesWhenConfigured(t *testing.T) {
	path := os.Getenv("GORO_TEST_GRF")
	if path == "" {
		t.Skip("set GORO_TEST_GRF to run against a real archive")
	}
	name := os.Getenv("GORO_TEST_GND_FILE")
	if name == "" {
		name = "geffen_in.gnd"
	}

	grf, err := OpenGRF(path)
	if err != nil {
		if errors.Is(err, ErrGRFUnsupportedVersion) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer grf.Close()

	if !grf.Has(name) {
		matches := grf.NamesWithSuffix(name)
		if len(matches) == 0 {
			t.Skipf("%s not present in %s", name, path)
		}
		name = matches[0]
	}
	data, err := grf.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	gnd, err := ParseGND(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}

	manager := &Manager{Archives: []*GRF{grf}}
	found := 0
	for _, texture := range gnd.Textures {
		if texture == "" {
			continue
		}
		img, source, err := LoadImage(manager, GroundTextureCandidates(texture))
		if err != nil {
			continue
		}
		if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
			t.Fatalf("invalid texture %s from %s: %v", texture, source, img.Bounds())
		}
		found++
		if found >= 3 {
			break
		}
	}
	if found == 0 {
		t.Fatalf("decoded no textures from %s (%d names)", name, len(gnd.Textures))
	}
	t.Logf("decoded %d textures from %s", found, name)
}

func TestGNDRealArchiveTextureDebugWhenConfigured(t *testing.T) {
	if os.Getenv("GORO_DEBUG_GND_TEXTURES") == "" {
		t.Skip("set GORO_DEBUG_GND_TEXTURES=1 to inspect real GND texture usage")
	}
	path := os.Getenv("GORO_TEST_GRF")
	if path == "" {
		t.Skip("set GORO_TEST_GRF to run against a real archive")
	}
	name := os.Getenv("GORO_TEST_GND_FILE")
	if name == "" {
		name = "prontera.gnd"
	}

	grf, err := OpenGRF(path)
	if err != nil {
		if errors.Is(err, ErrGRFUnsupportedVersion) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer grf.Close()

	if !grf.Has(name) {
		matches := grf.NamesWithSuffix(name)
		if len(matches) == 0 {
			t.Skipf("%s not present in %s", name, path)
		}
		name = matches[0]
	}
	data, err := grf.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	gnd, err := ParseGND(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}

	t.Logf("%s: size=%dx%d textures=%d surfaces=%d", name, gnd.Width, gnd.Height, len(gnd.Textures), len(gnd.Surfaces))
	manager := &Manager{Archives: []*GRF{grf}}
	for i, texture := range gnd.Textures {
		if i >= 24 {
			break
		}
		_, source, err := LoadImage(manager, GroundTextureCandidates(texture))
		if err != nil {
			t.Logf("texture[%02d] %q missing: %v", i, texture, err)
			continue
		}
		t.Logf("texture[%02d] %q -> %s", i, texture, source)
	}

	minU, maxU := float32(1e9), float32(-1e9)
	minV, maxV := float32(1e9), float32(-1e9)
	for _, surface := range gnd.Surfaces {
		for i := range surface.U {
			if surface.U[i] < minU {
				minU = surface.U[i]
			}
			if surface.U[i] > maxU {
				maxU = surface.U[i]
			}
			if surface.V[i] < minV {
				minV = surface.V[i]
			}
			if surface.V[i] > maxV {
				maxV = surface.V[i]
			}
		}
	}
	t.Logf("uv range: u=%.3f..%.3f v=%.3f..%.3f", minU, maxU, minV, maxV)

	centerX, centerY := gnd.Width/2, gnd.Height/2
	for y := centerY - 2; y <= centerY+2; y++ {
		for x := centerX - 2; x <= centerX+2; x++ {
			cell, ok := gnd.Cell(x, y)
			if !ok || cell.Top < 0 {
				continue
			}
			surface, ok := gnd.Surface(cell.Top)
			if !ok {
				continue
			}
			texture := ""
			if surface.TextureID >= 0 && surface.TextureID < len(gnd.Textures) {
				texture = gnd.Textures[surface.TextureID]
			}
			t.Logf("cell %d,%d top=%d tex=%d %q u=%v v=%v h=%v", x, y, cell.Top, surface.TextureID, texture, surface.U, surface.V, cell.Heights)
		}
	}
}
