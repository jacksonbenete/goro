package gamemode

import "testing"

func TestBasicMenuButtonAtFindsGridButtons(t *testing.T) {
	x, y, _, _ := basicMenuButtonBounds(3)
	index, ok := basicMenuButtonAt(x+basicMenuButtonW/2, y+basicMenuButtonH/2)
	if !ok {
		t.Fatal("expected menu button hit")
	}
	if index != 3 {
		t.Fatalf("button index = %d, want 3", index)
	}
}

func TestBasicMenuButtonAtRejectsPanelGaps(t *testing.T) {
	x := basicMenuX + basicMenuPad + basicMenuButtonW + basicMenuGapX/2
	y := basicMenuY + basicMenuPad + basicMenuButtonH/2
	if index, ok := basicMenuButtonAt(x, y); ok {
		t.Fatalf("gap hit button %d", index)
	}
}
