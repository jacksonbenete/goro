package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
)

func TestStatsWindowSectionsLayOutHorizontally(t *testing.T) {
	w := &StatsWindow{}
	body := w.statsBodyWidget(Context{Session: &session.Session{}})
	body.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(
		statsWindowWidth-2*statsWindowPad,
		statsWindowHeight-ROWindowTitleHeight-2*statsWindowPad,
	)))
	children := body.(interface{ Children() []widget.Widget }).Children()
	if len(children) != 2 {
		t.Fatalf("stats body children = %d, want primary and derived sections", len(children))
	}
	primaryChildren := children[0].(interface{ Children() []widget.Widget }).Children()
	if len(primaryChildren) != 2 {
		t.Fatalf("primary stats children = %d, want stat rows and status points without a header", len(primaryChildren))
	}
	primary := children[0].(interface{ Bounds() geometry.Rect }).Bounds()
	derived := children[1].(interface{ Bounds() geometry.Rect }).Bounds()
	if derived.Min.X < primary.Max.X {
		t.Fatalf("stats sections overlap horizontally: primary=%v derived=%v", primary, derived)
	}
	if derived.Min.Y != primary.Min.Y {
		t.Fatalf("stats sections top alignment differs: primary=%v derived=%v", primary, derived)
	}
}

func TestStatsPrimaryRowsUseTableRhythm(t *testing.T) {
	w := &StatsWindow{}
	rows := w.statRowsWidget(Context{Session: &session.Session{}})
	rows.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(
		statsPrimaryColumnWidth,
		6*statsRowH+5*statsRowGap,
	)))
	children := rows.(interface{ Children() []widget.Widget }).Children()
	if len(children) != 6 {
		t.Fatalf("primary stat rows = %d, want 6", len(children))
	}
	for i, child := range children {
		bounds := child.(interface{ Bounds() geometry.Rect }).Bounds()
		if bounds.Height() != statsRowH {
			t.Fatalf("primary row %d height = %.1f, want %d", i, bounds.Height(), statsRowH)
		}
		wantY := float32(i) * (statsRowH + statsRowGap)
		if bounds.Min.Y != wantY {
			t.Fatalf("primary row %d y = %.1f, want %.1f", i, bounds.Min.Y, wantY)
		}
	}
}

func TestStatPointCostPreRenewal(t *testing.T) {
	tests := []struct {
		current int
		want    int
	}{
		{current: 1, want: 2},
		{current: 10, want: 2},
		{current: 11, want: 3},
		{current: 20, want: 3},
		{current: 91, want: 11},
	}
	for _, test := range tests {
		if got := statPointCost(test.current); got != test.want {
			t.Fatalf("statPointCost(%d) = %d, want %d", test.current, got, test.want)
		}
	}
}

func TestStatCostPrefersServerValue(t *testing.T) {
	row := statRow{value: 31, cost: 8}
	if got := statCost(row); got != 8 {
		t.Fatalf("statCost() = %d, want 8", got)
	}
}
