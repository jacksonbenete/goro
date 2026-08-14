package rotheme

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

func TestContextMenuItemUsesFlatFillAndLeftAlignedText(t *testing.T) {
	item := ContextMenuItem("Trade", nil)
	bounds := geometry.NewRect(0, 0, 100, 18)
	item.Layout(widget.NewContext(), geometry.Tight(bounds.Size()))

	canvas := &uitest.MockCanvas{}
	item.Draw(widget.NewContext(), canvas)

	if len(canvas.Rects) != 1 {
		t.Fatalf("normal context menu item rectangles = %d, want one flat fill", len(canvas.Rects))
	}
	if len(canvas.Images) != 0 || len(canvas.RoundRects) != 0 || len(canvas.StrokeRects) != 0 || len(canvas.StrokeRoundRects) != 0 {
		t.Fatalf("normal context menu item drew gradient or bordered shapes: images=%d round=%d stroke=%d round-stroke=%d", len(canvas.Images), len(canvas.RoundRects), len(canvas.StrokeRects), len(canvas.StrokeRoundRects))
	}
	uitest.AssertColorEqual(t, canvas.Rects[0].Color, Default.Colors.WindowBody)
	if len(canvas.StyledTexts) != 1 {
		t.Fatalf("context menu item text draws = %d, want 1", len(canvas.StyledTexts))
	}
	text := canvas.StyledTexts[0]
	if text.Bounds != geometry.NewRect(contextMenuItemTextPaddingX, 0, 100-2*contextMenuItemTextPaddingX, 18) {
		t.Fatalf("context menu text bounds = %v, want horizontally inset bounds", text.Bounds)
	}
	if text.Style.Align != widget.TextAlignLeft {
		t.Fatalf("context menu text alignment = %v, want left", text.Style.Align)
	}
}

func TestContextMenuItemHoverUsesFlatBlueFill(t *testing.T) {
	item := ContextMenuItem("Trade", nil)
	bounds := geometry.NewRect(0, 0, 100, 18)
	item.Layout(widget.NewContext(), geometry.Tight(bounds.Size()))
	wrapper := item.Children()[0].(*mouseOnlyButtonWidget)
	center := bounds.Center()
	ctx := widget.NewContext()
	wrapper.Event(ctx, event.NewMouseEvent(
		event.MouseEnter,
		event.ButtonNone,
		0,
		center,
		center,
		event.ModNone,
	))

	canvas := &uitest.MockCanvas{}
	item.Draw(ctx, canvas)

	if len(canvas.Rects) != 1 {
		t.Fatalf("hovered context menu item rectangles = %d, want one flat fill", len(canvas.Rects))
	}
	uitest.AssertColorEqual(t, canvas.Rects[0].Color, contextMenuItemHover)
	if len(canvas.Images) != 0 || len(canvas.RoundRects) != 0 || len(canvas.StrokeRects) != 0 || len(canvas.StrokeRoundRects) != 0 {
		t.Fatalf("hovered context menu item drew gradient or bordered shapes: images=%d round=%d stroke=%d round-stroke=%d", len(canvas.Images), len(canvas.RoundRects), len(canvas.StrokeRects), len(canvas.StrokeRoundRects))
	}
}
