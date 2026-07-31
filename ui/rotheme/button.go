package rotheme

import (
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

const (
	ButtonRadius        float32 = 6
	ButtonPaddingX      float32 = 8
	ButtonPaddingY      float32 = 5.5
	LargeButtonPaddingY float32 = 8.5
)

func Button(label string, onClick func()) *primitives.BoxWidget {
	return ButtonDisabled(label, false, onClick)
}

func ButtonDisabled(label string, disabled bool, onClick func()) *primitives.BoxWidget {
	return buttonWithPadding(label, disabled, ButtonPaddingY, onClick)
}

func ButtonDisabledFn(label string, disabled func() bool, onClick func()) *primitives.BoxWidget {
	return buttonWithPaddingFn(label, disabled, ButtonPaddingY, onClick)
}

func LargeButton(label string, onClick func()) *primitives.BoxWidget {
	return LargeButtonDisabled(label, false, onClick)
}

func LargeButtonDisabled(label string, disabled bool, onClick func()) *primitives.BoxWidget {
	return buttonWithPadding(label, disabled, LargeButtonPaddingY, onClick)
}

func buttonWithPadding(label string, disabled bool, paddingY float32, onClick func()) *primitives.BoxWidget {
	opts := []button.Option{
		button.TextOpt(label),
		button.SizeOpt(button.Small),
		button.PainterOpt(ButtonPainter{}),
		button.RoundedOpt(ButtonRadius),
		button.Disabled(disabled),
	}
	if onClick != nil {
		opts = append(opts, button.OnClick(onClick))
	}
	return primitives.Box(
		newMouseOnlyButton(button.New(opts...).PaddingXY(ButtonPaddingX, paddingY)),
	).
		CrossAlign(primitives.CrossAxisStretch).
		Height(Default.Typography.TextSize + paddingY*2)
}

func buttonWithPaddingFn(label string, disabled func() bool, paddingY float32, onClick func()) *primitives.BoxWidget {
	opts := []button.Option{
		button.TextOpt(label),
		button.SizeOpt(button.Small),
		button.PainterOpt(ButtonPainter{}),
		button.RoundedOpt(ButtonRadius),
		button.DisabledFn(disabled),
	}
	if onClick != nil {
		opts = append(opts, button.OnClick(onClick))
	}
	return primitives.Box(
		newMouseOnlyButton(button.New(opts...).PaddingXY(ButtonPaddingX, paddingY)),
	).
		CrossAlign(primitives.CrossAxisStretch).
		Height(Default.Typography.TextSize + paddingY*2)
}

type mouseOnlyButtonWidget struct {
	widget.WidgetBase
	button *button.Widget
}

func newMouseOnlyButton(btn *button.Widget) *mouseOnlyButtonWidget {
	w := &mouseOnlyButtonWidget{button: btn}
	w.SetVisible(true)
	w.SetEnabled(true)
	btn.SetParent(w)
	return w
}

func (w *mouseOnlyButtonWidget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	return w.button.Layout(ctx, constraints)
}

func (w *mouseOnlyButtonWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	w.button.Draw(ctx, canvas)
}

func (w *mouseOnlyButtonWidget) Event(ctx widget.Context, e event.Event) bool {
	if _, ok := e.(*event.KeyEvent); ok {
		return false
	}
	consumed := w.button.Event(ctx, e)
	if mouse, ok := e.(*event.MouseEvent); ok && mouse.Button == event.ButtonLeft {
		ctx.ReleaseFocus(w.button)
	}
	return consumed
}

func (w *mouseOnlyButtonWidget) Children() []widget.Widget {
	return []widget.Widget{w.button}
}

func (w *mouseOnlyButtonWidget) SetBounds(bounds geometry.Rect) {
	w.WidgetBase.SetBounds(bounds)
	w.button.SetBounds(bounds)
}

type ButtonPainter struct{}

func (ButtonPainter) PaintButton(canvas widget.Canvas, state button.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}
	bg := Default.Colors.Button
	if state.Background != nil {
		bg = *state.Background
	}
	if state.Hovered {
		bg = Default.Colors.ButtonHover
	}
	if state.Pressed {
		bg = Default.Colors.ButtonDown
	}
	if state.Disabled {
		bg = Default.Colors.Disabled
	}
	border := Default.Colors.ButtonBorder
	if state.Disabled {
		border = Default.Colors.FooterLine
	}
	radius := ButtonRadius
	if state.Radius != nil {
		radius = *state.Radius
	}
	drawButtonGradient(canvas, state.Bounds, bg, radius)
	canvas.StrokeRoundRect(state.Bounds, border, radius, 1)

	text := Default.Colors.Text
	if state.Disabled {
		text = Default.Colors.MutedText
	}
	DrawText(canvas, state.Text, state.Bounds, Default.Typography.TextSize, text, false, widget.TextAlignCenter)
}

func drawButtonGradient(canvas widget.Canvas, bounds geometry.Rect, bottom widget.Color, radius float32) {
	top := bottom.Lerp(widget.RGBA(1, 1, 1, bottom.A), 0.42)
	height := int(bounds.Height())
	if height <= 1 {
		canvas.DrawRoundRect(bounds, bottom, radius)
		return
	}
	if radius > 0 {
		canvas.PushClipRoundRect(bounds, radius)
		defer canvas.PopClip()
	}
	for i := 0; i < height; i++ {
		t := float32(i) / float32(height-1)
		y := bounds.Min.Y + float32(i)
		canvas.DrawRect(geometry.NewRect(bounds.Min.X, y, bounds.Width(), 1), top.Lerp(bottom, t))
	}
}
