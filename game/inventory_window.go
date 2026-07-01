package game

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
)

const (
	inventoryWindowWidth  = 312
	inventoryWindowHeight = 356
	inventoryWindowTitleH = 28
	inventoryWindowPad    = 10
	inventoryRowH         = 32
	inventoryIconSize     = 24
)

var (
	inventoryTitleColor  = uiTitleTextColor
	inventoryTextColor   = uiTextColor
	inventoryMutedColor  = uiMutedTextColor
	inventoryButtonColor = uiButtonColor
	inventoryHoverColor  = uiButtonHoverColor
	inventoryDragColor   = uiButtonDownColor
)

type inventoryWindowState struct {
	open       bool
	x          int
	y          int
	positioned bool
	dragging   bool
	dragDX     int
	dragDY     int
	scroll     int
	dragItem   session.InventoryItem
	dragActive bool
	dragFrom   time.Time
}

func (w *inventoryWindowState) update(ctx Context, shop *shopWindowState, itemInfo *itemInfoWindowState) bool {
	if !w.open || ctx.Input == nil {
		return false
	}
	if shop == nil || !shop.open || shop.mode != shopModeSell {
		w.open = false
		w.dragging = false
		w.dragActive = false
		return false
	}
	w.ensurePosition(ctx)
	width, height := ctx.ScreenSize()
	if w.dragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			w.x = clampInventoryWindowInt(ctx.Input.MouseX-w.dragDX, 8, maxInt(8, width-inventoryWindowWidth-8))
			w.y = clampInventoryWindowInt(ctx.Input.MouseY-w.dragDY, 8, maxInt(8, height-inventoryWindowHeight-8))
			return true
		}
		w.dragging = false
		return true
	}
	if w.dragActive {
		if ctx.Input.MouseJustReleased(render.MouseButtonLeft) || !ctx.Input.MousePressed(render.MouseButtonLeft) {
			if shop != nil && shop.acceptInventoryDrop(ctx, w.dragItem, ctx.Input.MouseX, ctx.Input.MouseY) {
				w.dragActive = false
				return true
			}
			w.dragActive = false
			return true
		}
		return true
	}
	inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, inventoryWindowWidth, inventoryWindowHeight)
	if inside && ctx.Input.WheelY != 0 {
		w.scrollBy(ctx.Input.WheelY, ctx.Session)
		return true
	}
	if ctx.Input.JustPressed(render.KeyEscape) {
		w.open = false
		return true
	}
	if ctx.Input.MouseJustPressed(render.MouseButtonRight) {
		mx, my := ctx.Input.MouseX, ctx.Input.MouseY
		if !inside {
			return false
		}
		if item, ok := w.itemAt(ctx.Session, mx, my); ok {
			if itemInfo != nil {
				itemInfo.openItem(ctx, item, mx, my)
			}
			return true
		}
		return true
	}
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return inside
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	if !inside {
		return false
	}
	cx, cy, cw, ch := w.closeBounds()
	if pointInRect(mx, my, cx, cy, cw, ch) {
		w.open = false
		return true
	}
	if pointInRect(mx, my, w.x, w.y, inventoryWindowWidth, inventoryWindowTitleH) {
		w.dragging = true
		w.dragDX = mx - w.x
		w.dragDY = my - w.y
		return true
	}
	if item, ok := w.itemAt(ctx.Session, mx, my); ok {
		w.dragItem = item
		w.dragActive = true
		w.dragFrom = time.Now()
		return true
	}
	return true
}

func (w *inventoryWindowState) draw(screen *render.Image, ctx Context, mode *WorldMode) {
	if !w.open || screen == nil {
		return
	}
	w.ensurePosition(ctx)
	w.clampScroll(ctx.Session)
	x, y := w.x, w.y
	drawUITitledWindowFrame(screen, x, y, inventoryWindowWidth, inventoryWindowHeight, inventoryWindowTitleH)
	drawUIWindowTitle(screen, x, y, inventoryWindowTitleH, inventoryWindowPad, "Sell Inventory", inventoryTitleColor)
	cx, cy, cw, ch := w.closeBounds()
	drawUICloseButton(screen, cx, cy, cw, ch, inventoryButtonColor, inventoryTextColor)

	items := sortedInventoryItems(ctx.Session)
	if len(items) == 0 {
		render.DebugPrintAtColor(screen, "No items", x+inventoryWindowPad, y+inventoryWindowTitleH+18, inventoryMutedColor)
	} else {
		mx, my := -1, -1
		if ctx.Input != nil {
			mx, my = ctx.Input.MouseX, ctx.Input.MouseY
		}
		for row, item := range visibleInventoryItems(items, w.scroll) {
			rx, ry, rw, rh := w.rowBounds(row)
			fill := uiPanelAltColor
			if pointInRect(mx, my, rx, ry, rw, rh) {
				fill = inventoryHoverColor
			}
			if w.dragActive && w.dragItem.Index == item.Index {
				fill = inventoryDragColor
			}
			drawUISurface(screen, rx, ry, rw, rh, fill, uiWindowBorderColor)
			if mode != nil {
				mode.drawInventoryItemIcon(screen, ctx.Resources, item, rx+3, ry+3)
			}
			name := inventoryItemDisplayName(ctx.Resources, item)
			if item.Refine > 0 {
				name = fmt.Sprintf("+%d %s", item.Refine, name)
			}
			if item.Equipped {
				name += " [E]"
			}
			render.DebugPrintAtColor(screen, trimRunes(name, 28), rx+inventoryIconSize+10, ry+5, inventoryTextColor)
			render.DebugPrintAtColor(screen, fmt.Sprintf("x%d", item.Amount), rx+rw-42, ry+5, inventoryMutedColor)
		}
		w.drawScrollBar(screen, len(items))
	}
	if ctx.Session != nil {
		inv := ctx.Session.Inventory
		render.DebugPrintAtColor(screen, fmt.Sprintf("Weight %d / %d", displayWeight(inv.Weight), displayWeight(inv.MaxWeight)), x+inventoryWindowPad, y+inventoryWindowHeight-22, inventoryMutedColor)
		render.DebugPrintAtColor(screen, formatHUDNumber(inv.Zeny)+" z", x+inventoryWindowWidth-112, y+inventoryWindowHeight-22, inventoryMutedColor)
	}
	if w.dragActive {
		label := trimRunes(inventoryItemDisplayName(ctx.Resources, w.dragItem), 22)
		dx, dy := ctx.Input.MouseX+12, ctx.Input.MouseY+10
		width := len([]rune(label))*7 + inventoryIconSize + 18
		drawUISurface(screen, dx, dy, width, inventoryIconSize+6, uiPanelBodyColor, uiWindowBorderColor)
		if mode != nil {
			mode.drawInventoryItemIcon(screen, ctx.Resources, w.dragItem, dx+3, dy+3)
		}
		render.DebugPrintAtColor(screen, label, dx+inventoryIconSize+9, dy+(inventoryIconSize-13)/2+2, inventoryTextColor)
	}
}

func (w *inventoryWindowState) cursorAction(ctx Context) (int, bool) {
	if !w.open || ctx.Input == nil {
		return 0, false
	}
	if w.dragActive {
		return cursorActionPick, true
	}
	if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, inventoryWindowWidth, inventoryWindowHeight) {
		return cursorActionClick, true
	}
	return 0, false
}

func (w *inventoryWindowState) ensurePosition(ctx Context) {
	if w.positioned {
		return
	}
	width, _ := ctx.ScreenSize()
	w.x = maxInt(8, width-inventoryWindowWidth-24)
	w.y = 86
	w.positioned = true
}

func (w *inventoryWindowState) closeBounds() (int, int, int, int) {
	return w.x + inventoryWindowWidth - 23, w.y + 7, 16, 16
}

func (w *inventoryWindowState) rowBounds(row int) (int, int, int, int) {
	x := w.x + inventoryWindowPad
	y := w.y + inventoryWindowTitleH + 10 + row*inventoryRowH
	return x, y, inventoryWindowWidth - inventoryWindowPad*2 - 8, inventoryRowH - 4
}

func (w *inventoryWindowState) itemAt(s *session.Session, mx, my int) (session.InventoryItem, bool) {
	items := visibleInventoryItems(sortedInventoryItems(s), w.scroll)
	for row, item := range items {
		x, y, width, height := w.rowBounds(row)
		if pointInRect(mx, my, x, y, width, height) {
			return item, true
		}
	}
	return session.InventoryItem{}, false
}

func (w *inventoryWindowState) scrollBy(wheelY float64, s *session.Session) {
	if wheelY > 0 {
		w.scroll--
	} else if wheelY < 0 {
		w.scroll++
	}
	w.clampScroll(s)
}

func (w *inventoryWindowState) clampScroll(s *session.Session) {
	maxScroll := maxInt(0, len(sortedInventoryItems(s))-visibleInventoryRows())
	if w.scroll < 0 {
		w.scroll = 0
	}
	if w.scroll > maxScroll {
		w.scroll = maxScroll
	}
}

func (w *inventoryWindowState) drawScrollBar(screen *render.Image, total int) {
	visible := visibleInventoryRows()
	if total <= visible {
		return
	}
	trackX := w.x + inventoryWindowWidth - 14
	trackY := w.y + inventoryWindowTitleH + 10
	trackH := visible*inventoryRowH - 4
	render.DrawRect(screen, float64(trackX), float64(trackY), 4, float64(trackH), uiPanelAltColor)
	maxScroll := maxInt(1, total-visible)
	thumbH := maxInt(18, trackH*visible/total)
	thumbTravel := trackH - thumbH
	thumbY := trackY + thumbTravel*w.scroll/maxScroll
	render.DrawRect(screen, float64(trackX), float64(thumbY), 4, float64(thumbH), inventoryMutedColor)
}

func visibleInventoryRows() int {
	return (inventoryWindowHeight - inventoryWindowTitleH - 44) / inventoryRowH
}

func visibleInventoryItems(items []session.InventoryItem, scroll int) []session.InventoryItem {
	if scroll < 0 {
		scroll = 0
	}
	if scroll >= len(items) {
		return nil
	}
	end := minInt(len(items), scroll+visibleInventoryRows())
	return items[scroll:end]
}

func sortedInventoryItems(s *session.Session) []session.InventoryItem {
	if s == nil || len(s.Inventory.Items) == 0 {
		return nil
	}
	items := append([]session.InventoryItem(nil), s.Inventory.Items...)
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Index < items[j].Index
	})
	return items
}

func inventoryItemDisplayName(manager *res.Manager, item session.InventoryItem) string {
	if manager != nil {
		if name, ok := manager.ItemDisplayName(int(item.ItemID), item.Identified); ok && strings.TrimSpace(name) != "" {
			return name
		}
	}
	return fmt.Sprintf("item %d", item.ItemID)
}

func clampInventoryWindowInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
