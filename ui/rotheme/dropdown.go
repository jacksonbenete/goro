package rotheme

import (
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

type DropdownPainter struct{}

func (DropdownPainter) PaintTrigger(canvas widget.Canvas, state *dropdown.TriggerPaintState) {
	if state.Bounds.IsEmpty() {
		return
	}
	bg := Default.Colors.Button
	if state.Hovered || state.Open {
		bg = Default.Colors.ButtonHover
	}
	if state.Disabled {
		bg = Default.Colors.Disabled
	}
	border := Default.Colors.ButtonBorder
	if state.Focused || state.Open {
		border = Default.Colors.InputFocus
	}
	drawButtonGradient(canvas, state.Bounds, bg, ButtonRadius)
	canvas.StrokeRoundRect(state.Bounds, border, ButtonRadius, 1)

	textColor := Default.Colors.Text
	if state.IsPlaceholder || state.Disabled {
		textColor = Default.Colors.MutedText
	}
	DrawText(canvas, state.SelectedText, geometry.Rect{
		Min: geometry.Pt(state.Bounds.Min.X+6, state.Bounds.Min.Y),
		Max: geometry.Pt(state.Bounds.Max.X-18, state.Bounds.Max.Y),
	}, Default.Typography.TextSize, textColor, false, widget.TextAlignLeft)
	drawDropdownArrow(canvas, state.Bounds, state.Open, textColor)
}

func (DropdownPainter) PaintMenu(canvas widget.Canvas, state *dropdown.MenuPaintState) {
	if state.Bounds.IsEmpty() {
		return
	}
	canvas.DrawRect(state.Bounds, Default.Colors.WindowBody)
	canvas.StrokeRect(state.Bounds, Default.Colors.FooterLine, 1)
	canvas.PushClip(state.Bounds)
	defer canvas.PopClip()

	end := state.ScrollOffset + state.VisibleCount
	if end > len(state.Items) {
		end = len(state.Items)
	}
	for i := state.ScrollOffset; i < end; i++ {
		row := i - state.ScrollOffset
		itemBounds := geometry.Rect{
			Min: geometry.Pt(state.Bounds.Min.X, state.Bounds.Min.Y+float32(row)*state.ItemHeight),
			Max: geometry.Pt(state.Bounds.Max.X, state.Bounds.Min.Y+float32(row+1)*state.ItemHeight),
		}
		switch {
		case i == state.HighlightedIndex && i == state.SelectedIndex:
			canvas.DrawRect(itemBounds, Default.Colors.ButtonDown)
		case i == state.HighlightedIndex:
			canvas.DrawRect(itemBounds, Default.Colors.ButtonHover)
		case i == state.SelectedIndex:
			canvas.DrawRect(itemBounds, Default.Colors.WindowTitle)
		}
		textColor := Default.Colors.Text
		if state.Items[i].Disabled {
			textColor = Default.Colors.MutedText
		}
		DrawText(canvas, state.Items[i].DisplayText(), geometry.Rect{
			Min: geometry.Pt(itemBounds.Min.X+6, itemBounds.Min.Y),
			Max: geometry.Pt(itemBounds.Max.X-6, itemBounds.Max.Y),
		}, Default.Typography.TextSize, textColor, false, widget.TextAlignLeft)
	}
}

func drawDropdownArrow(canvas widget.Canvas, bounds geometry.Rect, open bool, color widget.Color) {
	cx := bounds.Max.X - 10
	cy := bounds.Center().Y
	if open {
		canvas.DrawLine(geometry.Pt(cx-4, cy+2), geometry.Pt(cx, cy-2), color, 1)
		canvas.DrawLine(geometry.Pt(cx, cy-2), geometry.Pt(cx+4, cy+2), color, 1)
		return
	}
	canvas.DrawLine(geometry.Pt(cx-4, cy-2), geometry.Pt(cx, cy+2), color, 1)
	canvas.DrawLine(geometry.Pt(cx, cy+2), geometry.Pt(cx+4, cy-2), color, 1)
}
