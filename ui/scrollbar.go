package ui

import "github.com/gogpu/ui/geometry"

const ROScrollbarGutter = 12

func scrollbarSafeWidth(width float32) float32 {
	if width <= ROScrollbarGutter {
		return 0
	}
	return width - ROScrollbarGutter
}

func scrollbarSafeIntWidth(width int) int {
	if width <= ROScrollbarGutter {
		return 0
	}
	return width - ROScrollbarGutter
}

func scrollbarSafeRect(bounds geometry.Rect) geometry.Rect {
	return geometry.NewRect(bounds.Min.X, bounds.Min.Y, scrollbarSafeWidth(bounds.Width()), bounds.Height())
}
