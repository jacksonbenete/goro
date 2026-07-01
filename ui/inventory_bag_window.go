package ui

import (
	"fmt"
	"image/color"
	"log"
	"sort"
	"time"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

const (
	inventoryBagWidth  = 320
	inventoryBagHeight = 286
	inventoryBagTitleH = 28
	inventoryBagPad    = 10
	inventoryBagTabW   = 42
	inventoryBagTabH   = 32
	inventoryBagCell   = 32
	inventoryBagIcon   = 24
	inventoryBagCols   = 7
	inventoryBagRows   = 5
)

const (
	inventoryBagTabItem = iota
	inventoryBagTabEquip
	inventoryBagTabEtc
)

var inventoryBagTabs = []struct {
	label string
	tab   int
}{
	{label: "Item", tab: inventoryBagTabItem},
	{label: "Equip", tab: inventoryBagTabEquip},
	{label: "Etc", tab: inventoryBagTabEtc},
}

type InventoryBagWindow struct {
	open          bool
	x             int
	y             int
	positioned    bool
	dragging      bool
	dragDX        int
	dragDY        int
	tab           int
	scroll        int
	status        string
	statusGood    bool
	statusAt      time.Time
	lastClickItem uint16
	lastClickAt   time.Time
	dragItem      session.InventoryItem
	dragActive    bool
	dragFrom      time.Time
}

func (w *InventoryBagWindow) Toggle(ctx Context) {
	if w.open {
		w.open = false
		w.dragging = false
		return
	}
	w.open = true
	w.EnsurePosition(ctx)
	w.selectFirstNonEmptyTab(ctx.Session)
	w.ClampScroll(ctx.Session)
}

func (w *InventoryBagWindow) Update(ctx Context, shortcuts *ShortcutBar, storage *StorageWindow, itemInfo *ItemInfoWindow) bool {
	if !w.open || ctx.Input == nil {
		return false
	}
	w.EnsurePosition(ctx)
	width, height := ctx.ScreenSize()
	if w.dragActive {
		if ctx.Input.MouseJustReleased(render.MouseButtonLeft) || !ctx.Input.MousePressed(render.MouseButtonLeft) {
			item := w.dragItem
			w.dragActive = false
			w.dragItem = session.InventoryItem{}
			if storage != nil && storage.AcceptInventoryDrop(ctx, item, ctx.Input.MouseX, ctx.Input.MouseY) {
				return true
			}
			if shortcuts != nil && shortcuts.AcceptItemDrop(ctx, item, ctx.Input.MouseX, ctx.Input.MouseY) {
				return true
			}
			return pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, inventoryBagWidth, inventoryBagHeight)
		}
	}
	if w.dragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			w.x = clampInventoryWindowInt(ctx.Input.MouseX-w.dragDX, 8, maxInt(8, width-inventoryBagWidth-8))
			w.y = clampInventoryWindowInt(ctx.Input.MouseY-w.dragDY, 8, maxInt(8, height-inventoryBagHeight-8))
			return true
		}
		w.dragging = false
		return true
	}

	inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, inventoryBagWidth, inventoryBagHeight)
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
			w.dragActive = false
			w.dragItem = session.InventoryItem{}
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
	for _, tab := range inventoryBagTabs {
		tx, ty, tw, th := w.tabBounds(tab.tab)
		if pointInRect(mx, my, tx, ty, tw, th) {
			w.tab = tab.tab
			w.scroll = 0
			w.lastClickItem = 0
			return true
		}
	}
	if pointInRect(mx, my, w.x, w.y, inventoryBagWidth, inventoryBagTitleH) {
		w.dragging = true
		w.dragDX = mx - w.x
		w.dragDY = my - w.y
		return true
	}
	if item, ok := w.itemAt(ctx.Session, mx, my); ok {
		now := time.Now()
		if w.lastClickItem == item.Index && now.Sub(w.lastClickAt) <= 360*time.Millisecond {
			w.dragActive = false
			w.dragItem = session.InventoryItem{}
			w.activateItem(ctx, item)
			w.lastClickItem = 0
			return true
		}
		w.dragItem = item
		w.dragActive = true
		w.dragFrom = now
		w.lastClickItem = item.Index
		w.lastClickAt = now
		w.status = inventoryItemDisplayName(ctx.Resources, item)
		w.statusGood = true
		w.statusAt = now
		return true
	}
	return true
}

func (w *InventoryBagWindow) Draw(screen *render.Image, ctx Context, assets AssetRenderer) {
	if !w.open || screen == nil {
		return
	}
	w.EnsurePosition(ctx)
	w.ClampScroll(ctx.Session)
	x, y := w.x, w.y
	DrawTitledWindowFrame(screen, x, y, inventoryBagWidth, inventoryBagHeight, inventoryBagTitleH)
	DrawWindowTitle(screen, x, y, inventoryBagTitleH, inventoryWindowPad, "Inventory", inventoryTitleColor)
	cx, cy, cw, ch := w.closeBounds()
	DrawCloseButton(screen, cx, cy, cw, ch, inventoryButtonColor, inventoryTextColor)

	gx, gy, gw, gh := w.gridBounds()
	px, py, pw, ph := w.panelBounds()
	DrawSurface(screen, px, py, pw, ph, PanelBodyColor, WindowBorderColor)
	DrawSurface(screen, gx, gy, gw, gh, WindowBodyColor, WindowBorderColor)
	for _, tab := range inventoryBagTabs {
		tx, ty, tw, th := w.tabBounds(tab.tab)
		fill := inventoryButtonColor
		if tab.tab == w.tab {
			fill = inventoryHoverColor
		}
		DrawButtonLabel(screen, tx, ty, tw, th, tab.label, fill, inventoryTextColor)
	}
	for row := 0; row < inventoryBagRows; row++ {
		for col := 0; col < inventoryBagCols; col++ {
			cx := gx + col*inventoryBagCell
			cy := gy + row*inventoryBagCell
			fill := color.RGBA{R: 255, G: 255, B: 250, A: 64}
			render.DrawRect(screen, float64(cx), float64(cy), inventoryBagCell-1, inventoryBagCell-1, fill)
		}
	}

	items := w.visibleItems(ctx.Session)
	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for i, item := range items {
		col := i % inventoryBagCols
		row := i / inventoryBagCols
		cx := gx + col*inventoryBagCell
		cy := gy + row*inventoryBagCell
		if pointInRect(mx, my, cx, cy, inventoryBagCell, inventoryBagCell) {
			render.DrawRect(screen, float64(cx), float64(cy), inventoryBagCell-1, inventoryBagCell-1, color.RGBA{R: 118, G: 150, B: 204, A: 92})
		}
		if assets != nil {
			assets.DrawInventoryItemIcon(screen, ctx.Resources, item, cx+4, cy+4)
		}
		if item.Amount > 1 {
			render.DebugPrintAtColor(screen, fmt.Sprintf("%d", item.Amount), cx+inventoryBagCell-16, cy+inventoryBagCell-14, color.RGBA{R: 40, G: 36, B: 32, A: 255})
		}
		if item.Equipped {
			render.DebugPrintAtColor(screen, "E", cx+2, cy+2, shopGoodColor)
		}
	}
	w.drawScrollBar(screen, len(w.tabItems(ctx.Session)))

	if ctx.Session != nil {
		inv := ctx.Session.Inventory
		footerY := y + inventoryBagHeight - 23
		render.DebugPrintAtColor(screen, fmt.Sprintf("Num:%d/%d", len(ctx.Session.Inventory.Items), 100), x+inventoryBagPad, footerY, inventoryMutedColor)
		render.DebugPrintAtColor(screen, fmt.Sprintf("Weight %d/%d", displayWeight(inv.Weight), displayWeight(inv.MaxWeight)), x+92, footerY, inventoryMutedColor)
		render.DebugPrintAtColor(screen, formatHUDNumber(inv.Zeny)+" z", x+inventoryBagWidth-94, footerY, inventoryMutedColor)
	}
	if w.status != "" && time.Since(w.statusAt) < 2200*time.Millisecond {
		statusColor := inventoryMutedColor
		if !w.statusGood {
			statusColor = shopErrorColor
		}
		render.DebugPrintAtColor(screen, trimRunes(w.status, 34), x+inventoryBagPad, y+inventoryBagHeight-41, statusColor)
	}
	if w.dragActive && ctx.Input != nil && time.Since(w.dragFrom) > 80*time.Millisecond && assets != nil {
		dx := ctx.Input.MouseX - inventoryIconSize/2
		dy := ctx.Input.MouseY - inventoryIconSize/2
		assets.DrawInventoryItemIcon(screen, ctx.Resources, w.dragItem, dx, dy)
	}
}

func (w *InventoryBagWindow) CursorAction(ctx Context) (int, bool) {
	if !w.open || ctx.Input == nil {
		return 0, false
	}
	if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, inventoryBagWidth, inventoryBagHeight) {
		return CursorActionClick, true
	}
	return 0, false
}

func (w *InventoryBagWindow) EnsurePosition(ctx Context) {
	if w.positioned {
		return
	}
	width, _ := ctx.ScreenSize()
	w.x = maxInt(8, width-inventoryBagWidth-24)
	w.y = 86
	w.positioned = true
}

func (w *InventoryBagWindow) closeBounds() (int, int, int, int) {
	return w.x + inventoryBagWidth - 23, w.y + 7, 16, 16
}

func (w *InventoryBagWindow) tabBounds(tab int) (int, int, int, int) {
	return w.x + inventoryBagPad + 2, w.y + inventoryBagTitleH + 12 + tab*(inventoryBagTabH+4), inventoryBagTabW, inventoryBagTabH
}

func (w *InventoryBagWindow) panelBounds() (int, int, int, int) {
	x := w.x + inventoryBagPad
	y := w.y + inventoryBagTitleH + 8
	return x, y, inventoryBagWidth - inventoryBagPad*2, inventoryBagRows*inventoryBagCell + 10
}

func (w *InventoryBagWindow) gridBounds() (int, int, int, int) {
	x := w.x + inventoryBagPad + inventoryBagTabW + 9
	y := w.y + inventoryBagTitleH + 13
	return x, y, inventoryBagCols * inventoryBagCell, inventoryBagRows * inventoryBagCell
}

func (w *InventoryBagWindow) itemAt(s *session.Session, mx, my int) (session.InventoryItem, bool) {
	gx, gy, gw, gh := w.gridBounds()
	if !pointInRect(mx, my, gx, gy, gw, gh) {
		return session.InventoryItem{}, false
	}
	col := (mx - gx) / inventoryBagCell
	row := (my - gy) / inventoryBagCell
	if col < 0 || col >= inventoryBagCols || row < 0 || row >= inventoryBagRows {
		return session.InventoryItem{}, false
	}
	index := row*inventoryBagCols + col
	items := w.visibleItems(s)
	if index < 0 || index >= len(items) {
		return session.InventoryItem{}, false
	}
	return items[index], true
}

func (w *InventoryBagWindow) activateItem(ctx Context, item session.InventoryItem) {
	if inventoryItemIsEquipment(item) {
		if ctx.Network == nil {
			w.setStatus("Not connected", false)
			return
		}
		if item.Equipped {
			if err := ctx.Network.SendTakeoffEquip(item.Index); err != nil {
				w.setStatus(err.Error(), false)
				return
			}
			w.setStatus("Unequip requested", true)
			return
		}
		if item.Location == 0 {
			w.setStatus("Missing equip location", false)
			return
		}
		if err := ctx.Network.SendWearEquip(item.Index, item.Location); err != nil {
			w.setStatus(err.Error(), false)
			return
		}
		w.setStatus("Equip requested", true)
		return
	}
	if !inventoryItemIsUsable(item) {
		w.setStatus("Item cannot be used", false)
		return
	}
	if err := useInventoryItem(ctx, item); err != nil {
		w.setStatus(err.Error(), false)
		return
	}
	w.setStatus("Use requested", true)
	log.Printf("inventory use requested index=%d item=%d type=%d", item.Index, item.ItemID, item.Type)
}

func (w *InventoryBagWindow) setStatus(text string, good bool) {
	w.status = text
	w.statusGood = good
	w.statusAt = time.Now()
}

func (w *InventoryBagWindow) scrollBy(wheelY float64, s *session.Session) {
	if wheelY > 0 {
		w.scroll--
	} else if wheelY < 0 {
		w.scroll++
	}
	w.ClampScroll(s)
}

func (w *InventoryBagWindow) ClampScroll(s *session.Session) {
	maxScroll := maxInt(0, (len(w.tabItems(s))+inventoryBagCols-1)/inventoryBagCols-inventoryBagRows)
	if w.scroll < 0 {
		w.scroll = 0
	}
	if w.scroll > maxScroll {
		w.scroll = maxScroll
	}
}

func (w *InventoryBagWindow) selectFirstNonEmptyTab(s *session.Session) {
	if len(w.tabItems(s)) > 0 {
		return
	}
	original := w.tab
	for _, tab := range inventoryBagTabs {
		if tab.tab == w.tab {
			continue
		}
		w.tab = tab.tab
		if len(w.tabItems(s)) > 0 {
			w.scroll = 0
			return
		}
	}
	w.tab = original
}

func (w *InventoryBagWindow) visibleItems(s *session.Session) []session.InventoryItem {
	items := w.tabItems(s)
	start := w.scroll * inventoryBagCols
	if start < 0 {
		start = 0
	}
	if start >= len(items) {
		return nil
	}
	end := minInt(len(items), start+inventoryBagCols*inventoryBagRows)
	return items[start:end]
}

func (w *InventoryBagWindow) tabItems(s *session.Session) []session.InventoryItem {
	items := sortedInventoryItems(s)
	if len(items) == 0 {
		return nil
	}
	filtered := items[:0]
	for _, item := range items {
		if inventoryItemTab(item) == w.tab {
			filtered = append(filtered, item)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].Index < filtered[j].Index
	})
	return filtered
}

func (w *InventoryBagWindow) drawScrollBar(screen *render.Image, total int) {
	if total <= inventoryBagCols*inventoryBagRows {
		return
	}
	gx, gy, gw, gh := w.gridBounds()
	trackX := gx + gw + 7
	render.DrawRect(screen, float64(trackX), float64(gy), 4, float64(gh), PanelAltColor)
	totalRows := (total + inventoryBagCols - 1) / inventoryBagCols
	maxScroll := maxInt(1, totalRows-inventoryBagRows)
	thumbH := maxInt(18, gh*inventoryBagRows/totalRows)
	thumbTravel := gh - thumbH
	thumbY := gy + thumbTravel*w.scroll/maxScroll
	render.DrawRect(screen, float64(trackX), float64(thumbY), 4, float64(thumbH), inventoryMutedColor)
}

func inventoryItemTab(item session.InventoryItem) int {
	if inventoryItemIsEquipment(item) {
		return inventoryBagTabEquip
	}
	if inventoryItemIsUsable(item) {
		return inventoryBagTabItem
	}
	return inventoryBagTabEtc
}

func inventoryItemIsEquipment(item session.InventoryItem) bool {
	return item.Equip || inventoryItemTypeIsEquipment(item.Type)
}

func inventoryItemTypeIsEquipment(itemType uint8) bool {
	switch itemType {
	case 4, 5, 7, 8, 12:
		return true
	default:
		return false
	}
}

func inventoryItemIsUsable(item session.InventoryItem) bool {
	switch item.Type {
	case 0, 2, 11, 18:
		return true
	default:
		return false
	}
}
