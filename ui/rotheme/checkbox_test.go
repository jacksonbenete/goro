package rotheme

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestCheckboxMarkUsesShortLeftAndLongRightBranches(t *testing.T) {
	canvas := &tableViewHeaderCanvas{}
	drawCheckboxV(canvas, geometry.NewRect(10, 20, CheckboxSize, CheckboxSize), widget.ColorBlack)

	if len(canvas.lines) != 2 {
		t.Fatalf("checkbox mark lines = %d, want 2", len(canvas.lines))
	}
	if got := canvas.lines[0].from; got != geometry.Pt(14, 28) {
		t.Fatalf("checkbox left branch starts at %v, want 14,28", got)
	}
	joint := geometry.Pt(17, 31)
	if got := canvas.lines[0].to; got != joint {
		t.Fatalf("checkbox left branch ends at %v, want %v", got, joint)
	}
	if got := canvas.lines[1].from; got != joint {
		t.Fatalf("checkbox right branch starts at %v, want %v", got, joint)
	}
	if got := canvas.lines[1].to; got != geometry.Pt(24, 24) {
		t.Fatalf("checkbox right branch ends at %v, want 24,24", got)
	}
}
