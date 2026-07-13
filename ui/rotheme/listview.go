package rotheme

import (
	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

const SelectListRowPadX float32 = 6

type SelectListPainter struct {
	EmptyText string
}

func SelectListRow(label string, enabled bool, height float32) widget.Widget {
	color := Default.Colors.Text
	if !enabled {
		color = Default.Colors.MutedText
	}
	return primitives.Box(
		Text(label).Color(color),
	).
		PaddingLeft(SelectListRowPadX).
		PaddingRight(SelectListRowPadX).
		Height(height).
		CrossAlign(primitives.CrossAxisCenter)
}

func (SelectListPainter) PaintDivider(widget.Canvas, listview.DividerState) {}

func (p SelectListPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	DrawText(canvas, p.EmptyText, bounds, Default.Typography.TextSize, Default.Colors.MutedText, false, widget.TextAlignCenter)
}

func (SelectListPainter) PaintItemBackground(canvas widget.Canvas, state listview.ItemPaintState) {
	fill := widget.RGBA8(246, 249, 253, 255)
	if state.Index%2 == 1 {
		fill = Default.Colors.PanelBody
	}
	if state.Hovered {
		fill = Default.Colors.ButtonHover
	}
	canvas.DrawRect(state.Bounds, fill)
}

func (SelectListPainter) PaintSelection(canvas widget.Canvas, state listview.ItemPaintState) {
	if state.Selected {
		canvas.DrawRect(state.Bounds, Default.Colors.ButtonDown)
	}
}

var _ listview.Painter = SelectListPainter{}
