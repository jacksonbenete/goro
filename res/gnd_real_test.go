package res

import (
	"os"
	"testing"
)

func TestGNDRealFileWhenConfigured(t *testing.T) {
	data := readRealDataFile(t, "data\\geffen_in.gnd")
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
	grf, name := realDataArchiveFile(t, "geffen_in.gnd")
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
	manager := realDataManager(t)
	name := "geffen_in.gnd"
	data, err := manager.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	gnd, err := ParseGND(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}

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
	manager := realDataManager(t)
	name := "prontera.gnd"
	data, err := manager.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	gnd, err := ParseGND(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}

	t.Logf("%s: size=%dx%d textures=%d lightmaps=%d surfaces=%d", name, gnd.Width, gnd.Height, len(gnd.Textures), len(gnd.Lightmaps), len(gnd.Surfaces))
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
			lightmap := GNDLightmap{}
			if surface.LightmapID >= 0 && surface.LightmapID < len(gnd.Lightmaps) {
				lightmap = gnd.Lightmaps[surface.LightmapID]
			}
			t.Logf("cell %d,%d top=%d tex=%d %q lm=%d alpha=%v u=%v v=%v h=%v", x, y, cell.Top, surface.TextureID, texture, surface.LightmapID, lightmap.Alpha, surface.U, surface.V, cell.Heights)
		}
	}
}
