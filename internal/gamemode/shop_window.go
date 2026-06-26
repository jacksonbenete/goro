package gamemode

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/kivutar/goro/internal/network"
	"github.com/kivutar/goro/internal/render"
	"github.com/kivutar/goro/internal/session"
)

const (
	shopWindowWidth  = 360
	shopWindowHeight = 320
	shopWindowTitleH = 28
	shopWindowPad    = 10
	shopCartRowH     = 28

	shopDealWidth  = 244
	shopDealHeight = 108
)

var (
	shopTitleColor  = color.RGBA{R: 255, G: 230, B: 150, A: 255}
	shopTextColor   = color.RGBA{R: 236, G: 232, B: 220, A: 255}
	shopMutedColor  = color.RGBA{R: 166, G: 174, B: 184, A: 255}
	shopGoodColor   = color.RGBA{R: 144, G: 210, B: 142, A: 255}
	shopErrorColor  = color.RGBA{R: 255, G: 116, B: 116, A: 255}
	shopButtonColor = color.RGBA{R: 56, G: 62, B: 72, A: 235}
	shopHoverColor  = color.RGBA{R: 72, G: 84, B: 104, A: 238}
	shopDropColor   = color.RGBA{R: 36, G: 48, B: 64, A: 225}
)

type shopWindowState struct {
	dealOpen        bool
	dealNPCID       uint32
	open            bool
	x               int
	y               int
	positioned      bool
	dragging        bool
	dragDX          int
	dragDY          int
	sellable        map[uint16]network.ShopSellItem
	cart            []shopSellCartItem
	status          string
	statusGood      bool
	statusAt        time.Time
	closePacketSent bool
}

type shopSellCartItem struct {
	item    session.InventoryItem
	price   uint32
	over    uint32
	amount  uint16
	max     uint16
	addedAt time.Time
}

func (w *shopWindowState) openDeal(selection network.ShopDealSelection, ctx Context) {
	w.dealOpen = true
	w.dealNPCID = selection.NPCID
	w.ensureSellPosition(ctx)
}

func (w *shopWindowState) openSell(list []network.ShopSellItem, ctx Context) {
	w.dealOpen = false
	w.open = true
	w.ensureSellPosition(ctx)
	w.sellable = make(map[uint16]network.ShopSellItem, len(list))
	for _, item := range list {
		w.sellable[item.Index] = item
	}
	w.cart = nil
	w.status = "Drag items here to sell"
	w.statusGood = true
	w.statusAt = time.Now()
	w.closePacketSent = false
}

func (w *shopWindowState) applyResult(ctx Context, result network.ShopResult) {
	if !result.Sell {
		w.status = fmt.Sprintf("Buy result %d", result.Result)
		w.statusGood = result.Result == 0
		w.statusAt = time.Now()
		return
	}
	if result.Result == 0 {
		w.status = "Deal completed"
		w.statusGood = true
		w.statusAt = time.Now()
		w.open = false
		w.cart = nil
		w.sellable = nil
		w.closePacketSent = true
		return
	}
	w.status = "Sell failed"
	w.statusGood = false
	w.statusAt = time.Now()
}

func (w *shopWindowState) update(ctx Context) bool {
	if ctx.Input == nil {
		return false
	}
	if w.dealOpen {
		return w.updateDeal(ctx)
	}
	if !w.open {
		return false
	}
	w.ensureSellPosition(ctx)
	width, height := ctx.ScreenSize()
	if w.dragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			w.x = clampShopWindowInt(ctx.Input.MouseX-w.dragDX, 8, maxInt(8, width-shopWindowWidth-8))
			w.y = clampShopWindowInt(ctx.Input.MouseY-w.dragDY, 8, maxInt(8, height-shopWindowHeight-8))
			return true
		}
		w.dragging = false
		return true
	}
	if ctx.Input.JustPressed(render.KeyEscape) {
		w.cancel(ctx)
		return true
	}
	inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, shopWindowWidth, shopWindowHeight)
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return inside
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	if !inside {
		return false
	}
	cx, cy, cw, ch := w.closeBounds()
	if pointInRect(mx, my, cx, cy, cw, ch) {
		w.cancel(ctx)
		return true
	}
	if pointInRect(mx, my, w.x, w.y, shopWindowWidth, shopWindowTitleH) {
		w.dragging = true
		w.dragDX = mx - w.x
		w.dragDY = my - w.y
		return true
	}
	sx, sy, sw, sh := w.sellButtonBounds()
	if pointInRect(mx, my, sx, sy, sw, sh) {
		w.submit(ctx)
		return true
	}
	bx, by, bw, bh := w.cancelButtonBounds()
	if pointInRect(mx, my, bx, by, bw, bh) {
		w.cancel(ctx)
		return true
	}
	for i := range w.cart {
		if w.handleCartButton(ctx, i, mx, my) {
			return true
		}
	}
	return true
}

func (w *shopWindowState) updateDeal(ctx Context) bool {
	width, height := ctx.ScreenSize()
	x := (width - shopDealWidth) / 2
	y := (height - shopDealHeight) * 2 / 3
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, x, y, shopDealWidth, shopDealHeight)
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	if !pointInRect(mx, my, x, y, shopDealWidth, shopDealHeight) {
		return false
	}
	if pointInRect(mx, my, x+18, y+64, 60, 24) {
		w.sendDealSelection(ctx, 0)
		return true
	}
	if pointInRect(mx, my, x+92, y+64, 60, 24) {
		w.sendDealSelection(ctx, 1)
		return true
	}
	if pointInRect(mx, my, x+166, y+64, 60, 24) {
		w.dealOpen = false
		return true
	}
	return true
}

func (w *shopWindowState) draw(screen *render.Image, ctx Context) {
	if screen == nil {
		return
	}
	if w.dealOpen {
		w.drawDeal(screen, ctx)
	}
	if !w.open {
		return
	}
	w.ensureSellPosition(ctx)
	x, y := w.x, w.y
	drawNPCWindowFrame(screen, x, y, shopWindowWidth, shopWindowHeight)
	render.DebugPrintAtColor(screen, "Sell Items", x+shopWindowPad, y+9, shopTitleColor)
	cx, cy, cw, ch := w.closeBounds()
	drawUIButtonSurface(screen, cx, cy, cw, ch, shopButtonColor)
	render.DebugPrintAtColor(screen, "x", cx+5, cy+(ch-13)/2-1, shopTextColor)
	render.DrawRect(screen, float64(x+8), float64(y+shopWindowTitleH), float64(shopWindowWidth-16), 1, color.RGBA{R: 210, G: 200, B: 170, A: 80})

	dx, dy, dw, dh := w.dropBounds()
	fill := color.RGBA{R: 28, G: 32, B: 40, A: 205}
	if ctx.Input != nil && pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, dx, dy, dw, dh) {
		fill = shopDropColor
	}
	drawUISurface(screen, dx, dy, dw, dh, fill, color.RGBA{R: 210, G: 200, B: 170, A: 70})
	if len(w.cart) == 0 {
		render.DebugPrintAtColor(screen, "Drop inventory items here", dx+46, dy+72, shopMutedColor)
	} else {
		for i, item := range w.visibleCartItems() {
			w.drawCartRow(screen, ctx, i, item)
		}
	}
	render.DebugPrintAtColor(screen, fmt.Sprintf("Total: %s z", formatHUDNumber(int64(w.total()))), x+shopWindowPad, y+shopWindowHeight-48, shopTextColor)
	sx, sy, sw, sh := w.sellButtonBounds()
	w.drawButton(screen, sx, sy, sw, sh, "Sell", len(w.cart) > 0)
	bx, by, bw, bh := w.cancelButtonBounds()
	w.drawButton(screen, bx, by, bw, bh, "Cancel", true)
	if w.status != "" && time.Since(w.statusAt) < 2500*time.Millisecond {
		statusColor := shopErrorColor
		if w.statusGood {
			statusColor = shopGoodColor
		}
		render.DebugPrintAtColor(screen, trimRunes(w.status, 42), x+shopWindowPad, y+shopWindowHeight-20, statusColor)
	}
}

func (w *shopWindowState) drawDeal(screen *render.Image, ctx Context) {
	width, height := ctx.ScreenSize()
	x := (width - shopDealWidth) / 2
	y := (height - shopDealHeight) * 2 / 3
	drawNPCWindowFrame(screen, x, y, shopDealWidth, shopDealHeight)
	render.DebugPrintAtColor(screen, "What do you want to do?", x+20, y+20, shopTextColor)
	w.drawButton(screen, x+18, y+64, 60, 24, "Buy", true)
	w.drawButton(screen, x+92, y+64, 60, 24, "Sell", true)
	w.drawButton(screen, x+166, y+64, 60, 24, "Cancel", true)
}

func (w *shopWindowState) cursorAction(ctx Context) (int, bool) {
	if ctx.Input == nil {
		return 0, false
	}
	if w.dealOpen {
		width, height := ctx.ScreenSize()
		x := (width - shopDealWidth) / 2
		y := (height - shopDealHeight) * 2 / 3
		if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, x, y, shopDealWidth, shopDealHeight) {
			return cursorActionClick, true
		}
	}
	if w.open && pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, shopWindowWidth, shopWindowHeight) {
		return cursorActionClick, true
	}
	return 0, false
}

func (w *shopWindowState) acceptInventoryDrop(ctx Context, item session.InventoryItem, mouseX, mouseY int) bool {
	dx, dy, dw, dh := w.dropBounds()
	if !w.open || !pointInRect(mouseX, mouseY, dx, dy, dw, dh) {
		return false
	}
	sell, ok := w.sellable[item.Index]
	if !ok {
		w.status = "That item cannot be sold"
		w.statusGood = false
		w.statusAt = time.Now()
		return true
	}
	w.addCartItem(item, sell)
	return true
}

func (w *shopWindowState) addCartItem(item session.InventoryItem, sell network.ShopSellItem) {
	maxAmount := uint16(maxInt(1, item.Amount))
	for i := range w.cart {
		if w.cart[i].item.Index == item.Index {
			if w.cart[i].amount < w.cart[i].max {
				w.cart[i].amount++
			}
			w.status = "Item added"
			w.statusGood = true
			w.statusAt = time.Now()
			return
		}
	}
	w.cart = append(w.cart, shopSellCartItem{
		item:    item,
		price:   sell.Price,
		over:    sell.OverchargePrice,
		amount:  1,
		max:     maxAmount,
		addedAt: time.Now(),
	})
	w.status = "Item added"
	w.statusGood = true
	w.statusAt = time.Now()
}

func (w *shopWindowState) submit(ctx Context) {
	if len(w.cart) == 0 {
		w.status = "No items selected"
		w.statusGood = false
		w.statusAt = time.Now()
		return
	}
	if ctx.Network == nil {
		w.status = "Not connected"
		w.statusGood = false
		w.statusAt = time.Now()
		return
	}
	items := make([]network.SellRequestItem, 0, len(w.cart))
	for _, item := range w.cart {
		items = append(items, network.SellRequestItem{Index: item.item.Index, Amount: item.amount})
	}
	if err := ctx.Network.SendShopSellItems(items); err != nil {
		w.status = err.Error()
		w.statusGood = false
		w.statusAt = time.Now()
		return
	}
	w.closePacketSent = true
	w.status = "Sell requested"
	w.statusGood = true
	w.statusAt = time.Now()
}

func (w *shopWindowState) cancel(ctx Context) {
	if w.open && !w.closePacketSent && ctx.Network != nil {
		if err := ctx.Network.SendShopSellItems(nil); err != nil {
			log.Printf("send empty sell list on shop close failed: %v", err)
		}
	}
	w.open = false
	w.dealOpen = false
	w.cart = nil
	w.sellable = nil
	w.closePacketSent = true
}

func (w *shopWindowState) sendDealSelection(ctx Context, dealType uint8) {
	if ctx.Network == nil {
		w.status = "Not connected"
		w.statusGood = false
		w.statusAt = time.Now()
		return
	}
	if err := ctx.Network.SendShopDealSelection(w.dealNPCID, dealType); err != nil {
		w.status = err.Error()
		w.statusGood = false
		w.statusAt = time.Now()
		return
	}
	w.dealOpen = false
	if dealType == 1 {
		w.status = "Waiting for sell list"
	} else {
		w.status = "Buy requested"
	}
	w.statusGood = true
	w.statusAt = time.Now()
}

func (w *shopWindowState) ensureSellPosition(ctx Context) {
	if w.positioned {
		return
	}
	width, _ := ctx.ScreenSize()
	w.x = maxInt(8, (width-shopWindowWidth)/2)
	w.y = 120
	w.positioned = true
}

func (w *shopWindowState) closeBounds() (int, int, int, int) {
	return w.x + shopWindowWidth - 23, w.y + 7, 16, 16
}

func (w *shopWindowState) dropBounds() (int, int, int, int) {
	return w.x + shopWindowPad, w.y + shopWindowTitleH + 12, shopWindowWidth - shopWindowPad*2, 194
}

func (w *shopWindowState) sellButtonBounds() (int, int, int, int) {
	return w.x + shopWindowWidth - 146, w.y + shopWindowHeight - 54, 58, 24
}

func (w *shopWindowState) cancelButtonBounds() (int, int, int, int) {
	return w.x + shopWindowWidth - 80, w.y + shopWindowHeight - 54, 62, 24
}

func (w *shopWindowState) cartRowBounds(row int) (int, int, int, int) {
	x, y, width, _ := w.dropBounds()
	return x + 5, y + 5 + row*shopCartRowH, width - 10, shopCartRowH - 3
}

func (w *shopWindowState) visibleCartItems() []shopSellCartItem {
	visible := minInt(len(w.cart), 6)
	return w.cart[:visible]
}

func (w *shopWindowState) drawCartRow(screen *render.Image, ctx Context, row int, item shopSellCartItem) {
	x, y, width, height := w.cartRowBounds(row)
	drawUISurface(screen, x, y, width, height, color.RGBA{R: 34, G: 38, B: 46, A: 210}, color.RGBA{R: 210, G: 200, B: 170, A: 45})
	name := inventoryItemDisplayName(ctx.Resources, item.item)
	render.DebugPrintAtColor(screen, trimRunes(name, 22), x+7, y+5, shopTextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("x%d", item.amount), x+170, y+5, shopMutedColor)
	render.DebugPrintAtColor(screen, formatHUDNumber(int64(item.over)*int64(item.amount)), x+210, y+5, shopMutedColor)
	w.drawTinyButton(screen, x+width-52, y+4, "-", true)
	w.drawTinyButton(screen, x+width-34, y+4, "+", item.amount < item.max)
	w.drawTinyButton(screen, x+width-16, y+4, "x", true)
}

func (w *shopWindowState) handleCartButton(ctx Context, row int, mx, my int) bool {
	if row >= len(w.cart) {
		return false
	}
	x, y, width, _ := w.cartRowBounds(row)
	minus := [4]int{x + width - 52, y + 4, 15, 17}
	plus := [4]int{x + width - 34, y + 4, 15, 17}
	remove := [4]int{x + width - 16, y + 4, 15, 17}
	switch {
	case pointInRect(mx, my, minus[0], minus[1], minus[2], minus[3]):
		if w.cart[row].amount > 1 {
			w.cart[row].amount--
		} else {
			w.cart = append(w.cart[:row], w.cart[row+1:]...)
		}
		return true
	case pointInRect(mx, my, plus[0], plus[1], plus[2], plus[3]):
		if w.cart[row].amount < w.cart[row].max {
			w.cart[row].amount++
		}
		return true
	case pointInRect(mx, my, remove[0], remove[1], remove[2], remove[3]):
		w.cart = append(w.cart[:row], w.cart[row+1:]...)
		return true
	default:
		return false
	}
}

func (w *shopWindowState) drawButton(screen *render.Image, x, y, width, height int, label string, enabled bool) {
	fill := shopButtonColor
	text := shopTextColor
	if !enabled {
		fill = color.RGBA{R: 42, G: 46, B: 54, A: 205}
		text = shopMutedColor
	}
	drawUIButtonSurface(screen, x, y, width, height, fill)
	render.DebugPrintAtColor(screen, label, x+(width-len([]rune(label))*7)/2, y+6, text)
}

func (w *shopWindowState) drawTinyButton(screen *render.Image, x, y int, label string, enabled bool) {
	w.drawButton(screen, x, y, 15, 17, label, enabled)
}

func (w *shopWindowState) total() int64 {
	var total int64
	for _, item := range w.cart {
		total += int64(item.over) * int64(item.amount)
	}
	return total
}

func clampShopWindowInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
