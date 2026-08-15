package ui

import "testing"

func TestTooltipStaysHiddenDuringWindowDrag(t *testing.T) {
	app := &windowDragTestApp{}
	ctx := Context{UIApp: app}
	var tooltip tooltipState

	tooltip.Show(ctx, "Item details", 100, 120, 80)
	if !tooltip.Open() {
		t.Fatal("tooltip should open before dragging")
	}

	app.active = true
	if !TooltipsSuppressed(ctx) {
		t.Fatal("tooltips should be suppressed during a window drag")
	}
	tooltip.Show(ctx, "Item details", 100, 120, 80)
	if tooltip.Open() {
		t.Fatal("tooltip reopened during a window drag")
	}

	app.active = false
	tooltip.Show(ctx, "Item details", 100, 120, 80)
	if !tooltip.Open() {
		t.Fatal("tooltip should open after dragging")
	}
}
