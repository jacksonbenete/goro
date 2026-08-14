package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestTitleBarGradientEndsWithDarkerBottomBorder(t *testing.T) {
	bounds := geometry.NewRect(3, 5, 80, 4)
	canvas := &uitest.MockCanvas{}

	drawTitleBarGradient(canvas, bounds)

	if len(canvas.Rects) != 4 {
		t.Fatalf("title bar rows = %d, want 3 gradient rows and 1 border row", len(canvas.Rects))
	}
	uitest.AssertColorEqual(t, canvas.Rects[0].Color, rotheme.Default.Colors.WindowTitleTop)
	uitest.AssertColorEqual(t, canvas.Rects[2].Color, rotheme.Default.Colors.WindowTitle)
	border := canvas.Rects[3]
	if border.Bounds != geometry.NewRect(bounds.Min.X, bounds.Max.Y-1, bounds.Width(), 1) {
		t.Fatalf("title bar border bounds = %v, want bottom row of %v", border.Bounds, bounds)
	}
	uitest.AssertColorEqual(t, border.Color, rotheme.Default.Colors.WindowBorder)
}
