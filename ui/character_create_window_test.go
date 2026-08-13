package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestCharacterSelectPage(t *testing.T) {
	if got := CharacterSelectPage(5); got != 1 {
		t.Fatalf("page = %d, want 1", got)
	}
}

func TestCharacterSelectFooterButtonStateForSelectedSlot(t *testing.T) {
	opts := CharacterSelectWindowOptions{
		SelectedSlot: 1,
		Characters: []session.Character{
			{ID: 10, Slot: 1},
		},
	}
	if characterSelectDeleteDisabled(opts) {
		t.Fatal("delete disabled for occupied slot")
	}
	if !characterSelectMakeDisabled(opts) {
		t.Fatal("make enabled for occupied slot")
	}

	opts.SelectedSlot = 2
	if !characterSelectDeleteDisabled(opts) {
		t.Fatal("delete enabled for empty slot")
	}
	if characterSelectMakeDisabled(opts) {
		t.Fatal("make disabled for empty slot")
	}
}

func TestCharacterSelectArrowHitboxesKeepFullWidth(t *testing.T) {
	tree := (&CharacterSelectWindow{opts: CharacterSelectWindowOptions{MaxSlots: 9}}).widgetTree()
	tree.Layout(
		widget.NewContext(),
		geometry.Tight(geometry.Sz(characterSelectWindowW, characterSelectWindowH)),
	)

	windowChildren := tree.Children()
	if len(windowChildren) < 2 || len(windowChildren[1].Children()) != 1 {
		t.Fatal("character-select window content tree is incomplete")
	}
	content := windowChildren[1].Children()[0]
	if len(content.Children()) == 0 {
		t.Fatal("character-select window has no content row")
	}
	rowChildren := content.Children()[0].Children()
	if len(rowChildren) != 3 {
		t.Fatalf("character-select row children = %d, want 3", len(rowChildren))
	}

	for i, arrow := range []widget.Widget{rowChildren[0], rowChildren[2]} {
		bounder, ok := arrow.(interface{ Bounds() geometry.Rect })
		if !ok {
			t.Fatalf("arrow %d does not expose bounds", i)
		}
		if got := bounder.Bounds().Width(); got != rotheme.IconButtonSize {
			t.Fatalf("arrow %d hitbox width = %.1f, want %.1f", i, got, rotheme.IconButtonSize)
		}
	}
}

func TestCharacterCreateHairControlsFrameHead(t *testing.T) {
	bounds := geometry.NewRect(13, 21, characterCreatePanelW, characterCreatePanelH)
	left := characterCreatePreviewButtonRect(bounds, 0)
	color := characterCreatePreviewButtonRect(bounds, 1)
	right := characterCreatePreviewButtonRect(bounds, 2)

	if left.Min.Y != right.Min.Y {
		t.Fatalf("hair style arrow heights differ: left %.1f, right %.1f", left.Min.Y, right.Min.Y)
	}
	if color.Min.Y < bounds.Min.Y || color.Max.Y >= left.Min.Y {
		t.Fatalf("hair color arrow bounds = %v, want above hair style arrows at %.1f", color, left.Min.Y)
	}
	if left.Min.Y >= bounds.Min.Y+bounds.Height()/3 {
		t.Fatalf("hair style arrows start at %.1f, want within preview's upper third", left.Min.Y)
	}
}

func TestCharacterCreateGraphDrawOrderIsValidHexagon(t *testing.T) {
	points := CharacterCreateGraphPoints(0, 0, 64)
	order := CharacterCreateGraphDrawOrder()
	seen := map[int]bool{}
	for _, stat := range order {
		if stat < 0 || stat >= CharacterCreateStatCount {
			t.Fatalf("stat index outside range in graph order: %d", stat)
		}
		if seen[stat] {
			t.Fatalf("duplicate stat index in graph order: %d", stat)
		}
		seen[stat] = true
	}

	if points[CharacterCreateStatDex][0] >= 0 || points[CharacterCreateStatDex][1] <= 0 {
		t.Fatalf("DEX graph point = %#v, want lower-left", points[CharacterCreateStatDex])
	}
	if points[CharacterCreateStatLuk][0] <= 0 || points[CharacterCreateStatLuk][1] <= 0 {
		t.Fatalf("LUK graph point = %#v, want lower-right", points[CharacterCreateStatLuk])
	}

	for i := 0; i < CharacterCreateStatCount; i++ {
		a1 := points[order[i]]
		a2 := points[order[(i+1)%CharacterCreateStatCount]]
		for j := i + 1; j < CharacterCreateStatCount; j++ {
			if graphEdgesAdjacent(i, j) {
				continue
			}
			b1 := points[order[j]]
			b2 := points[order[(j+1)%CharacterCreateStatCount]]
			if graphSegmentsCross(a1, a2, b1, b2) {
				t.Fatalf("graph edges %d and %d cross", i, j)
			}
		}
	}
}

func graphEdgesAdjacent(a, b int) bool {
	if a == b {
		return true
	}
	if a+1 == b || b+1 == a {
		return true
	}
	return (a == 0 && b == CharacterCreateStatCount-1) || (b == 0 && a == CharacterCreateStatCount-1)
}

func graphSegmentsCross(a1, a2, b1, b2 [2]float64) bool {
	o1 := graphOrientation(a1, a2, b1)
	o2 := graphOrientation(a1, a2, b2)
	o3 := graphOrientation(b1, b2, a1)
	o4 := graphOrientation(b1, b2, a2)
	return o1*o2 < 0 && o3*o4 < 0
}

func graphOrientation(a, b, c [2]float64) float64 {
	return (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
}
