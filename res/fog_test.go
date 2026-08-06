package res

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestFogParameterParsesTable(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	table := "prontera.gat#0.25#0.75#80C0FF#1.0#izlude#0.1#0.2#0x112233#0.5#"
	if err := os.WriteFile(filepath.Join(dataDir, "fogparametertable.txt"), []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}

	manager, err := NewManager(root)
	if err != nil {
		t.Fatal(err)
	}

	parameter, ok := manager.FogParameter("prontera")
	if !ok {
		t.Fatal("expected prontera fog")
	}
	if parameter.Near != 0.25 || parameter.Far != 0.75 || parameter.Factor != 1 {
		t.Fatalf("unexpected prontera fog distances: %+v", parameter)
	}
	if parameter.Color != (color.RGBA{R: 0x80, G: 0xc0, B: 0xff, A: 255}) {
		t.Fatalf("unexpected prontera fog color: %#v", parameter.Color)
	}

	parameter, ok = manager.FogParameter("izlude.rsw")
	if !ok {
		t.Fatal("expected izlude fog")
	}
	if parameter.Color != (color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 255}) {
		t.Fatalf("unexpected izlude fog color: %#v", parameter.Color)
	}
}

func TestPayonDungeonFogRealWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	for _, mapName := range []string{"pay_dun00.rsw", "pay_dun01.rsw", "pay_dun02.rsw", "pay_dun03.rsw"} {
		parameter, ok := manager.FogParameter(mapName)
		if !ok {
			t.Fatalf("%s fog missing", mapName)
		}
		t.Logf("%s fog near=%.3f far=%.3f color=%02x%02x%02x factor=%.3f", mapName, parameter.Near, parameter.Far, parameter.Color.R, parameter.Color.G, parameter.Color.B, parameter.Factor)
	}
}

func TestEinbrochFogRealWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	parameter, ok := manager.FogParameter("einbroch.rsw")
	if !ok {
		t.Fatal("einbroch.rsw fog missing")
	}
	if parameter.Factor != 0.5 {
		t.Fatalf("einbroch.rsw fog factor = %.3f, want 0.500", parameter.Factor)
	}
	t.Logf("einbroch.rsw fog near=%.3f far=%.3f color=%02x%02x%02x factor=%.3f", parameter.Near, parameter.Far, parameter.Color.R, parameter.Color.G, parameter.Color.B, parameter.Factor)
}
