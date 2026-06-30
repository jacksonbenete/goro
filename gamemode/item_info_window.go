package gamemode

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

const (
	itemInfoWindowWidth  = 336
	itemInfoWindowHeight = 276
	itemInfoWindowTitleH = 28
	itemInfoWindowPad    = 10
	itemInfoIconPanel    = 44
	itemInfoLineH        = 14
)

type itemInfoWindowState struct {
	open     bool
	x        int
	y        int
	dragging bool
	dragDX   int
	dragDY   int
	item     session.InventoryItem
	title    string
	details  []string
	lines    []string
	scroll   int
	openedAt time.Time
}

func (w *itemInfoWindowState) openItem(ctx Context, item session.InventoryItem, mouseX, mouseY int) {
	if item.ItemID == 0 {
		return
	}
	w.open = true
	w.dragging = false
	w.item = item
	w.title = "Item Information"
	w.details = itemInfoDetailLines(item)
	w.lines = itemInfoDescriptionLines(ctx, item)
	w.scroll = 0
	w.openedAt = time.Now()

	screenW, screenH := ctx.ScreenSize()
	w.x = clampInventoryWindowInt(mouseX+14, 8, maxInt(8, screenW-itemInfoWindowWidth-8))
	w.y = clampInventoryWindowInt(mouseY-22, 8, maxInt(8, screenH-itemInfoWindowHeight-8))
}

func (w *itemInfoWindowState) update(ctx Context) bool {
	if !w.open || ctx.Input == nil {
		return false
	}
	width, height := ctx.ScreenSize()
	if w.dragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			w.x = clampInventoryWindowInt(ctx.Input.MouseX-w.dragDX, 8, maxInt(8, width-itemInfoWindowWidth-8))
			w.y = clampInventoryWindowInt(ctx.Input.MouseY-w.dragDY, 8, maxInt(8, height-itemInfoWindowHeight-8))
			return true
		}
		w.dragging = false
		return true
	}
	inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, itemInfoWindowWidth, itemInfoWindowHeight)
	if inside && ctx.Input.WheelY != 0 {
		w.scrollBy(ctx.Input.WheelY)
		return true
	}
	if ctx.Input.JustPressed(render.KeyEscape) {
		w.open = false
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
	if pointInRect(mx, my, w.x, w.y, itemInfoWindowWidth, itemInfoWindowTitleH) {
		w.dragging = true
		w.dragDX = mx - w.x
		w.dragDY = my - w.y
		return true
	}
	return true
}

func (w *itemInfoWindowState) draw(screen *render.Image, ctx Context, mode *WorldMode) {
	if !w.open || screen == nil {
		return
	}
	x, y := w.x, w.y
	drawUITitledWindowFrame(screen, x, y, itemInfoWindowWidth, itemInfoWindowHeight, itemInfoWindowTitleH)
	drawUIWindowTitle(screen, x, y, itemInfoWindowTitleH, itemInfoWindowPad, w.title, inventoryTitleColor)
	cx, cy, cw, ch := w.closeBounds()
	drawUICloseButton(screen, cx, cy, cw, ch, inventoryButtonColor, inventoryTextColor)

	panelX := x + itemInfoWindowPad
	panelY := y + itemInfoWindowTitleH + itemInfoWindowPad
	drawUISurface(screen, panelX, panelY, itemInfoIconPanel, itemInfoIconPanel, uiPanelBodyColor, uiWindowBorderColor)
	if mode != nil {
		mode.drawInventoryItemIcon(screen, ctx.Resources, w.item, panelX+(itemInfoIconPanel-inventoryIconSize)/2, panelY+(itemInfoIconPanel-inventoryIconSize)/2)
	}

	name := inventoryItemDisplayName(ctx.Resources, w.item)
	if w.item.Refine > 0 {
		name = fmt.Sprintf("+%d %s", w.item.Refine, name)
	}
	nameX := panelX + itemInfoIconPanel + 10
	render.DebugPrintAtColor(screen, trimRunes(name, 34), nameX, panelY+2, inventoryTextColor)
	for i, line := range w.details {
		render.DebugPrintAtColor(screen, trimRunes(line, 34), nameX, panelY+18+i*itemInfoLineH, inventoryMutedColor)
	}

	descX := x + itemInfoWindowPad
	descY := panelY + itemInfoIconPanel + 10
	descW := itemInfoWindowWidth - itemInfoWindowPad*2
	descH := itemInfoWindowHeight - (descY - y) - itemInfoWindowPad
	drawUISurface(screen, descX, descY, descW, descH, uiPanelBodyColor, uiWindowBorderColor)
	visible := w.visibleDescriptionLineCount(descH)
	lines := w.wrappedLines()
	if len(lines) == 0 {
		render.DebugPrintAtColor(screen, "No description available.", descX+7, descY+7, inventoryMutedColor)
		return
	}
	w.clampScroll()
	end := minInt(len(lines), w.scroll+visible)
	for i, line := range lines[w.scroll:end] {
		render.DebugPrintAtColor(screen, line, descX+7, descY+7+i*itemInfoLineH, itemInfoDescriptionColor(line))
	}
	if len(lines) > visible {
		drawItemInfoScrollBar(screen, descX+descW-8, descY+5, descH-10, w.scroll, visible, len(lines))
	}
}

func (w *itemInfoWindowState) cursorAction(ctx Context) (int, bool) {
	if !w.open || ctx.Input == nil {
		return 0, false
	}
	if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, itemInfoWindowWidth, itemInfoWindowHeight) {
		return cursorActionClick, true
	}
	return 0, false
}

func (w *itemInfoWindowState) closeBounds() (int, int, int, int) {
	return w.x + itemInfoWindowWidth - 23, w.y + 7, 16, 16
}

func (w *itemInfoWindowState) scrollBy(wheelY float64) {
	if wheelY > 0 {
		w.scroll--
	} else if wheelY < 0 {
		w.scroll++
	}
	w.clampScroll()
}

func (w *itemInfoWindowState) clampScroll() {
	maxScroll := maxInt(0, len(w.wrappedLines())-w.visibleDescriptionLineCount(itemInfoDescriptionHeight()))
	if w.scroll < 0 {
		w.scroll = 0
	}
	if w.scroll > maxScroll {
		w.scroll = maxScroll
	}
}

func (w *itemInfoWindowState) visibleDescriptionLineCount(height int) int {
	return maxInt(1, (height-12)/itemInfoLineH)
}

func (w *itemInfoWindowState) wrappedLines() []string {
	return wrapItemInfoLines(w.lines, 42)
}

func itemInfoDescriptionHeight() int {
	descY := itemInfoWindowTitleH + itemInfoWindowPad + itemInfoIconPanel + 10
	return itemInfoWindowHeight - descY - itemInfoWindowPad
}

func itemInfoDetailLines(item session.InventoryItem) []string {
	lines := []string{fmt.Sprintf("Item ID: %d", item.ItemID)}
	if item.Amount > 1 {
		lines = append(lines, fmt.Sprintf("Amount: %d", item.Amount))
	}
	if item.Type != 0 {
		lines = append(lines, fmt.Sprintf("Type: %s", itemInfoTypeLabel(item.Type)))
	}
	if item.Equipped {
		lines = append(lines, "Equipped")
	}
	if !item.Identified {
		lines = append(lines, "Unidentified")
	}
	return lines
}

func itemInfoDescriptionLines(ctx Context, item session.InventoryItem) []string {
	if ctx.Resources == nil {
		return nil
	}
	lines, ok := ctx.Resources.ItemDescription(int(item.ItemID), item.Identified)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(stripItemInfoColorCodes(strings.ReplaceAll(line, "_", " ")))
		if line == "" {
			out = append(out, "")
			continue
		}
		out = append(out, line)
	}
	return out
}

func stripItemInfoColorCodes(text string) string {
	runes := []rune(text)
	out := make([]rune, 0, len(runes))
	for i := 0; i < len(runes); i++ {
		if runes[i] == '^' && i+6 < len(runes) && isHexRunes(runes[i+1:i+7]) {
			i += 6
			continue
		}
		out = append(out, runes[i])
	}
	return string(out)
}

func isHexRunes(runes []rune) bool {
	if len(runes) != 6 {
		return false
	}
	for _, r := range runes {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

func wrapItemInfoLines(lines []string, maxRunes int) []string {
	var out []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			out = append(out, "")
			continue
		}
		for _, wrapped := range wrapItemInfoLine(line, maxRunes) {
			out = append(out, wrapped)
		}
	}
	return out
}

func wrapItemInfoLine(line string, maxRunes int) []string {
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}
	var out []string
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if runeLen(current)+1+runeLen(word) <= maxRunes {
			current += " " + word
			continue
		}
		out = append(out, current)
		current = word
	}
	if current != "" {
		out = append(out, current)
	}
	return out
}

func runeLen(text string) int {
	return len([]rune(text))
}

func itemInfoTypeLabel(itemType uint8) string {
	switch itemType {
	case 0:
		return "Healing"
	case 2:
		return "Usable"
	case 3:
		return "Etc"
	case 4:
		return "Weapon"
	case 5:
		return "Armor"
	case 6:
		return "Card"
	case 7:
		return "Pet Egg"
	case 8:
		return "Pet Equipment"
	case 10:
		return "Ammo"
	case 11:
		return "Delayed Consumable"
	case 12:
		return "Shadow Gear"
	case 18:
		return "Cash Usable"
	default:
		return fmt.Sprintf("%d", itemType)
	}
}

func itemInfoDescriptionColor(line string) color.RGBA {
	if strings.HasPrefix(strings.TrimSpace(line), "Class :") || strings.HasPrefix(strings.TrimSpace(line), "Weight :") {
		return inventoryMutedColor
	}
	return inventoryTextColor
}

func drawItemInfoScrollBar(screen *render.Image, x, y, h, scroll, visible, total int) {
	if screen == nil || total <= visible || h <= 0 {
		return
	}
	render.DrawRect(screen, float64(x), float64(y), 4, float64(h), uiPanelAltColor)
	maxScroll := maxInt(1, total-visible)
	thumbH := maxInt(18, h*visible/total)
	thumbTravel := h - thumbH
	thumbY := y + thumbTravel*scroll/maxScroll
	render.DrawRect(screen, float64(x), float64(thumbY), 4, float64(thumbH), inventoryMutedColor)
}
