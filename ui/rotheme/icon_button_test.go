package rotheme

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestIconButtonGlyphKeepsIntegerXAndQuarterPixelY(t *testing.T) {
	canvas := &tableViewHeaderCanvas{}
	drawIconGlyph(canvas, geometry.NewRect(0, 0, IconButtonSize, IconButtonSize), IconButtonClose, widget.ColorBlack)

	if len(canvas.lines) != 2 {
		t.Fatalf("close icon lines = %d, want 2", len(canvas.lines))
	}
	if got := canvas.lines[0].from; got != geometry.Pt(5, 5.25) {
		t.Fatalf("close icon first point = %v, want 5,5.25", got)
	}
	if got := canvas.lines[0].to; got != geometry.Pt(11, 11.25) {
		t.Fatalf("close icon second point = %v, want 11,11.25", got)
	}
}

func TestIconButtonChevronKeepsIntegerXAndQuarterPixelY(t *testing.T) {
	canvas := &tableViewHeaderCanvas{}
	drawIconGlyph(canvas, geometry.NewRect(0, 0, IconButtonSize, IconButtonSize), IconButtonLeft, widget.ColorBlack)

	if len(canvas.lines) != 2 {
		t.Fatalf("left icon lines = %d, want 2", len(canvas.lines))
	}
	if got := canvas.lines[0].from; got != geometry.Pt(11, 4.25) {
		t.Fatalf("left icon first point = %v, want 11,4.25", got)
	}
	if got := canvas.lines[0].to; got != geometry.Pt(5, 8.25) {
		t.Fatalf("left icon second point = %v, want 5,8.25", got)
	}
}
