package rotheme

import (
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

const (
	RadioSize float32 = 13
	RadioGap  float32 = 7
)

func Radio(opts ...radio.GroupOption) *radio.Group {
	opts = append(opts, radio.GroupPainter(RadioPainter{}))
	return radio.NewGroup(opts...)
}

type RadioPainter struct{}

func (RadioPainter) PaintRadio(canvas widget.Canvas, state radio.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}
	center := geometry.Pt(state.Bounds.Min.X+RadioSize/2, state.Bounds.Min.Y+state.Bounds.Height()/2)
	border := Default.Colors.ButtonBorder
	text := Default.Colors.Text
	if state.Disabled {
		border = Default.Colors.Disabled
		text = Default.Colors.MutedText
	}
	canvas.DrawCircle(center, RadioSize/2, Default.Colors.Button)
	canvas.StrokeCircle(center, RadioSize/2, border, 1)
	if state.Selected {
		canvas.DrawCircle(center, 3, Default.Colors.TitleText)
	}
	if state.Focused && !state.Disabled {
		canvas.StrokeCircle(center, RadioSize/2+2, Default.Colors.InputFocus, 1)
	}
	if state.Label != "" {
		x := state.Bounds.Min.X + RadioSize + RadioGap
		DrawText(canvas, state.Label, geometry.NewRect(x, state.Bounds.Min.Y, state.Bounds.Width()-(x-state.Bounds.Min.X), state.Bounds.Height()), Default.Typography.TextSize, text, false, widget.TextAlignLeft)
	}
}

var _ radio.Painter = RadioPainter{}
