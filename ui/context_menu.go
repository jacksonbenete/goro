package ui

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	contextMenuPadding = 2
	contextMenuRadius  = 5
	contextMenuRowH    = rotheme.ContextMenuItemHeight
)

func contextMenu(width int, rows ...widget.Widget) widget.Widget {
	return Win(
		TitleBar(false),
		Radius(contextMenuRadius),
		Size(float32(width), float32(contextMenuHeight(len(rows)))),
		Content(
			primitives.Box(rows...).
				CrossAlign(primitives.CrossAxisStretch).
				Padding(contextMenuPadding),
		),
	)
}

func contextMenuItem(label string, onClick func()) widget.Widget {
	return rotheme.ContextMenuItem(label, onClick)
}

func contextMenuHeight(rows int) int {
	return contextMenuRowH*rows + contextMenuPadding*2
}
