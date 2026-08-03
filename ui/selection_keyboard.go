package ui

import (
	"github.com/gogpu/ui/state"
	"github.com/kivutar/goro/input"
)

func updateSelectionTableKeyboard(ctx Context, selected *int, itemCount, visibleRows int, rowHeight float32, scroll state.Signal[float32], onEnter func()) bool {
	if ctx.Input == nil || itemCount <= 0 || selected == nil {
		return false
	}
	consumed := false
	switch {
	case ctx.Input.JustPressed(input.KeyArrowUp):
		if *selected <= 0 {
			*selected = 0
		} else {
			*selected--
		}
		consumed = true
	case ctx.Input.JustPressed(input.KeyArrowDown):
		if *selected < 0 {
			*selected = 0
		} else if *selected < itemCount-1 {
			*selected++
		}
		consumed = true
	}
	if consumed {
		ensureSelectionTableRowVisible(*selected, itemCount, visibleRows, rowHeight, scroll)
	}
	if ctx.Input.JustPressed(input.KeyEnter) {
		if onEnter != nil {
			onEnter()
		}
		return true
	}
	return consumed
}

func ensureSelectionTableRowVisible(row, itemCount, visibleRows int, rowHeight float32, scroll state.Signal[float32]) {
	if scroll == nil || row < 0 || itemCount <= 0 || visibleRows <= 0 || rowHeight <= 0 {
		return
	}
	top := int(scroll.Get() / rowHeight)
	switch {
	case row < top:
		scroll.Set(float32(row) * rowHeight)
	case row >= top+visibleRows:
		scroll.Set(float32(row-visibleRows+1) * rowHeight)
	}
	maxScroll := float32(maxInt(0, itemCount-visibleRows)) * rowHeight
	switch value := scroll.Get(); {
	case value < 0:
		scroll.Set(0)
	case value > maxScroll:
		scroll.Set(maxScroll)
	}
}
