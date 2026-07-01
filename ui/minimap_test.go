package ui

import "testing"

func TestNormalizeMinimapMapName(t *testing.T) {
	if got := normalizeMinimapMapName(`data\prontera.rsw`); got != "prontera" {
		t.Fatalf("normalized map = %q, want prontera", got)
	}
	if got := normalizeMinimapMapName("izlude.gat"); got != "izlude" {
		t.Fatalf("normalized gat = %q, want izlude", got)
	}
}

func TestMinimapCellToScreenInvertsY(t *testing.T) {
	rect := minimapRect{x: 10, y: 20, w: 100, h: 100}

	_, topY, ok := minimapCellToScreen(rect, 10, 10, 5, 9)
	if !ok {
		t.Fatal("top cell did not project")
	}
	_, bottomY, ok := minimapCellToScreen(rect, 10, 10, 5, 0)
	if !ok {
		t.Fatal("bottom cell did not project")
	}
	if topY >= bottomY {
		t.Fatalf("topY=%d bottomY=%d, want map Y inverted onto screen", topY, bottomY)
	}
}

func TestMinimapImageCandidatesIncludeROPaths(t *testing.T) {
	candidates := minimapImageCandidates("prontera")
	if len(candidates) == 0 {
		t.Fatal("no minimap candidates")
	}
	want := "data\\texture\\interface\\map\\prontera.bmp"
	for _, candidate := range candidates {
		if candidate == want {
			return
		}
	}
	t.Fatalf("candidate %q missing from %#v", want, candidates)
}
