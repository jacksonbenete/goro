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
	return primitives.Box(
		newMouseButton(label, func() bool { return disabled }, paddingY, ButtonPainter{}, onClick),
	).
		CrossAlign(primitives.CrossAxisStretch).
		Height(Default.Typography.TextSize + paddingY*2)
}

func buttonWithPaddingFn(label string, disabled func() bool, paddingY float32, onClick func()) *primitives.BoxWidget {
	return primitives.Box(
		newMouseButton(label, disabled, paddingY, ButtonPainter{}, onClick),
	).
		CrossAlign(primitives.CrossAxisStretch).
		Height(Default.Typography.TextSize + paddingY*2)
}

type mouseButtonWidget struct {
	widget.WidgetBase
	label     string
	disabled  func() bool
	onClick   func()
	painter   button.Painter
	paddingX  float32
	paddingY  float32
	minWidth  float32
	minHeight float32
	hovered   bool
	pressed   bool
}

func newMouseButton(label string, disabled func() bool, paddingY float32, painter button.Painter, onClick func()) *mouseButtonWidget {
	w := &mouseButtonWidget{
		label:    label,
		disabled: disabled,
		onClick:  onClick,
		painter:  painter,
		paddingX: ButtonPaddingX,
		paddingY: paddingY,
	}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *mouseButtonWidget) MinWidth(width float32) *mouseButtonWidget {
	w.minWidth = width
	return w
}

func (w *mouseButtonWidget) MinHeight(height float32) *mouseButtonWidget {
	w.minHeight = height
	return w
}

func (w *mouseButtonWidget) resolvedDisabled() bool {
	if w.disabled == nil {
		return false
	}
	return w.disabled()
}

func (w *mouseButtonWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	width := float32(len(w.label))*Default.Typography.TextSize*0.55 + w.paddingX*2
	if width < w.minWidth {
		width = w.minWidth
	}
	height := Default.Typography.TextSize + w.paddingY*2
	if height < w.minHeight {
		height = w.minHeight
	}
	return constraints.Constrain(geometry.Sz(width, height))
}

func (w *mouseButtonWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if w.painter == nil {
		return
	}
	w.painter.PaintButton(canvas, button.PaintState{
		Text:     w.label,
		Size:     button.Small,
		Hovered:  w.hovered,
		Pressed:  w.pressed,
		Disabled: w.resolvedDisabled(),
		Bounds:   w.Bounds(),
	})
}

func (w *mouseButtonWidget) Event(ctx widget.Context, e event.Event) bool {
	if w.resolvedDisabled() {
		return false
	}
	mouse, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	switch mouse.MouseType {
	case event.MouseEnter, event.MouseMove:
		w.hovered = true
		ctx.SetCursor(widget.CursorPointer)
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		return true
	case event.MouseLeave:
		w.hovered = false
		w.pressed = false
		ctx.SetCursor(widget.CursorDefault)
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		return true
	case event.MousePress:
		if mouse.Button != event.ButtonLeft {
			return false
		}
		w.pressed = true
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		return true
	case event.MouseRelease:
		if mouse.Button != event.ButtonLeft {
			return false
		}
		wasPressed := w.pressed
		inside := w.Bounds().Contains(mouse.Position)
		w.pressed = false
		w.hovered = inside
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		if wasPressed && inside && w.onClick != nil {
			w.onClick()
		}
		return true
	case event.MouseDrag:
		return w.pressed
	default:
		return false
	}
}

func (w *mouseButtonWidget) Children() []widget.Widget {
	return nil
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
