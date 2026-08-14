package rotheme

import (
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

const (
	ContextMenuItemHeight               = 18
	contextMenuItemTextPaddingX float32 = 4
)

var contextMenuItemHover = widget.RGBA8(115, 156, 239, 255)

// ContextMenuItem keeps the standard button interaction behavior while using
// the flat, compact appearance of an action in a context menu.
func ContextMenuItem(label string, onClick func()) *primitives.BoxWidget {
	opts := []button.Option{
		button.TextOpt(label),
		button.SizeOpt(button.Small),
		button.PainterOpt(ContextMenuItemPainter{}),
		button.RoundedOpt(0),
	}
	if onClick != nil {
		opts = append(opts, button.OnClick(onClick))
	}
	return primitives.Box(
		newMouseOnlyButton(button.New(opts...).PaddingXY(contextMenuItemTextPaddingX, 0)),
	).
		CrossAlign(primitives.CrossAxisStretch).
		Height(ContextMenuItemHeight)
}

type ContextMenuItemPainter struct{}

func (ContextMenuItemPainter) PaintButton(canvas widget.Canvas, state button.PaintState) {
	if state.Bounds.IsEmpty() {
		return
	}

	background := Default.Colors.WindowBody
	if state.Hovered || state.Pressed {
		background = contextMenuItemHover
	}
	canvas.DrawRect(state.Bounds, background)

	textColor := Default.Colors.Text
	if state.Disabled {
		textColor = Default.Colors.MutedText
	}
	textWidth := max(float32(0), state.Bounds.Width()-2*contextMenuItemTextPaddingX)
	textBounds := geometry.NewRect(
		state.Bounds.Min.X+contextMenuItemTextPaddingX,
		state.Bounds.Min.Y,
		textWidth,
		state.Bounds.Height(),
	)
	DrawText(canvas, state.Text, textBounds, Default.Typography.TextSize, textColor, false, widget.TextAlignLeft)
}

var _ button.Painter = ContextMenuItemPainter{}
