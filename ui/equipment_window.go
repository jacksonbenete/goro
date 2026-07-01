package ui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/primitives"
	uiwidget "github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

const (
	equipmentWindowWidth  = 300
	equipmentWindowHeight = 232
	equipmentWindowTitleH = 28
	equipmentWindowPad    = 10
	equipmentContentW     = 280
	equipmentContentH     = 130
	equipmentLeftColW     = 112
	equipmentCenterColW   = 56
	equipmentRightColW    = 112
	equipmentRowH         = 24

	equipmentPreviewWorldDirection = 4
)

const (
	equipLocationHeadBottom uint16 = 1 << 0
	equipLocationWeapon     uint16 = 1 << 1
	equipLocationGarment    uint16 = 1 << 2
	equipLocationAccessory1 uint16 = 1 << 3
	equipLocationArmor      uint16 = 1 << 4
	equipLocationShield     uint16 = 1 << 5
	equipLocationShoes      uint16 = 1 << 6
	equipLocationAccessory2 uint16 = 1 << 7
	equipLocationHeadTop    uint16 = 1 << 8
	equipLocationHeadMid    uint16 = 1 << 9
	equipLocationAmmo       uint16 = 1 << 15
)

type EquipmentWindow struct {
	open       bool
	x          int
	y          int
	positioned bool
	dragging   bool
	dragDX     int
	dragDY     int
	status     string
	statusGood bool
	statusAt   time.Time
}

type equipmentSlotDef struct {
	label    string
	location uint16
	side     equipmentSlotSide
	row      int
}

type equipmentSlotSide int

const (
	equipmentSlotLeft equipmentSlotSide = iota
	equipmentSlotRight
	equipmentSlotCenter
)

var equipmentSlots = []equipmentSlotDef{
	{label: "Head Top", location: equipLocationHeadTop, side: equipmentSlotLeft, row: 0},
	{label: "Head Mid", location: equipLocationHeadMid, side: equipmentSlotRight, row: 0},
	{label: "Head Low", location: equipLocationHeadBottom, side: equipmentSlotLeft, row: 1},
	{label: "Armor", location: equipLocationArmor, side: equipmentSlotRight, row: 1},
	{label: "Weapon", location: equipLocationWeapon, side: equipmentSlotLeft, row: 2},
	{label: "Shield", location: equipLocationShield, side: equipmentSlotRight, row: 2},
	{label: "Garment", location: equipLocationGarment, side: equipmentSlotLeft, row: 3},
	{label: "Shoes", location: equipLocationShoes, side: equipmentSlotRight, row: 3},
	{label: "Accessory", location: equipLocationAccessory1, side: equipmentSlotLeft, row: 4},
	{label: "Accessory", location: equipLocationAccessory2, side: equipmentSlotRight, row: 4},
	{label: "Ammo", location: equipLocationAmmo, side: equipmentSlotCenter, row: 1},
}

var equipmentContentSurface *render.Image

func (w *EquipmentWindow) Toggle(ctx Context) {
	if w.open {
		w.open = false
		w.dragging = false
		return
	}
	w.open = true
	w.EnsurePosition(ctx)
}

func (w *EquipmentWindow) Update(ctx Context, itemInfo *ItemInfoWindow) bool {
	if !w.open || ctx.Input == nil {
		return false
	}
	w.EnsurePosition(ctx)
	width, height := ctx.ScreenSize()
	if w.dragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			w.x = clampInventoryWindowInt(ctx.Input.MouseX-w.dragDX, 8, maxInt(8, width-equipmentWindowWidth-8))
			w.y = clampInventoryWindowInt(ctx.Input.MouseY-w.dragDY, 8, maxInt(8, height-equipmentWindowHeight-8))
			return true
		}
		w.dragging = false
		return true
	}

	inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, equipmentWindowWidth, equipmentWindowHeight)
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
		if inside {
			w.updateHoverStatus(ctx)
		}
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
	if pointInRect(mx, my, w.x, w.y, equipmentWindowWidth, equipmentWindowTitleH) {
		w.dragging = true
		w.dragDX = mx - w.x
		w.dragDY = my - w.y
		return true
	}
	if item, ok := w.itemAt(ctx.Session, mx, my); ok {
		w.activateItem(ctx, item)
		return true
	}
	return true
}

func (w *EquipmentWindow) Draw(screen *render.Image, ctx Context, assets AssetRenderer) {
	if !w.open || screen == nil {
		return
	}
	w.EnsurePosition(ctx)
	x, y := w.x, w.y
	DrawTitledWindowFrame(screen, x, y, equipmentWindowWidth, equipmentWindowHeight, equipmentWindowTitleH)
	DrawWindowTitle(screen, x, y, equipmentWindowTitleH, equipmentWindowPad, "Equipment", inventoryTitleColor)
	cx, cy, cw, ch := w.closeBounds()
	DrawCloseButton(screen, cx, cy, cw, ch, inventoryButtonColor, inventoryTextColor)

	contentX, contentY := w.contentOrigin()
	drawEquipmentContentSurface(screen, contentX, contentY)
	px, py, pw, ph := w.previewBounds()
	if assets != nil {
		assets.DrawEquipmentPreview(screen, ctx, px, py, pw, ph)
	}

	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	for _, slot := range equipmentSlots {
		sx, sy, sw, sh := w.slotBounds(slot)
		if pointInRect(mx, my, sx, sy, sw, sh) {
			render.DrawRect(screen, float64(sx), float64(sy), float64(sw), float64(sh), color.RGBA{R: 118, G: 150, B: 204, A: 68})
		}
		item, ok := equippedItemForSlot(ctx.Session, slot.location)
		if ok && assets != nil {
			w.drawSlotItem(screen, ctx, assets, slot, item, sx, sy, sw, sh)
			continue
		}
		w.drawEmptySlotLabel(screen, slot, sx, sy)
	}

	if ctx.Session != nil {
		character := selectedCharacter(ctx.Session)
		name := character.Name
		if name == "" {
			name = "Player"
		}
		footerY := y + equipmentWindowTitleH + equipmentWindowPad + equipmentContentH + 10
		render.DebugPrintAtColor(screen, trimRunes(name, 24), x+equipmentWindowPad, footerY, inventoryTextColor)
		render.DebugPrintAtColor(screen, JobName(int(character.Job)), x+equipmentWindowPad, footerY+16, inventoryMutedColor)
	}
	if w.status != "" && time.Since(w.statusAt) < 2200*time.Millisecond {
		statusColor := inventoryMutedColor
		if !w.statusGood {
			statusColor = shopErrorColor
		}
		render.DebugPrintAtColor(screen, trimRunes(w.status, 26), x+118, y+equipmentWindowHeight-21, statusColor)
	}
}

func (w *EquipmentWindow) CursorAction(ctx Context) (int, bool) {
	if !w.open || ctx.Input == nil {
		return 0, false
	}
	if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, equipmentWindowWidth, equipmentWindowHeight) {
		return CursorActionClick, true
	}
	return 0, false
}

func (w *EquipmentWindow) EnsurePosition(ctx Context) {
	if w.positioned {
		return
	}
	width, _ := ctx.ScreenSize()
	w.x = maxInt(8, width-equipmentWindowWidth-24)
	w.y = 104
	w.positioned = true
}

func (w *EquipmentWindow) closeBounds() (int, int, int, int) {
	return w.x + equipmentWindowWidth - 23, w.y + 7, 16, 16
}

func (w *EquipmentWindow) contentOrigin() (int, int) {
	return w.x + equipmentWindowPad, w.y + equipmentWindowTitleH + equipmentWindowPad
}

func (w *EquipmentWindow) previewBounds() (int, int, int, int) {
	x, y := w.contentOrigin()
	return x + equipmentLeftColW, y + 2, equipmentCenterColW, 125
}

func (w *EquipmentWindow) slotBounds(slot equipmentSlotDef) (int, int, int, int) {
	x, y := w.contentOrigin()
	switch slot.side {
	case equipmentSlotLeft:
		return x, y + slot.row*equipmentRowH, equipmentLeftColW, equipmentRowH
	case equipmentSlotRight:
		return x + equipmentLeftColW + equipmentCenterColW, y + slot.row*equipmentRowH, equipmentRightColW, equipmentRowH
	case equipmentSlotCenter:
		return x + equipmentLeftColW, y + 30, equipmentCenterColW, equipmentRowH
	default:
		return x, y, 0, 0
	}
}

func (w *EquipmentWindow) itemAt(s *session.Session, mx, my int) (session.InventoryItem, bool) {
	for _, slot := range equipmentSlots {
		x, y, width, height := w.slotBounds(slot)
		if !pointInRect(mx, my, x, y, width, height) {
			continue
		}
		return equippedItemForSlot(s, slot.location)
	}
	return session.InventoryItem{}, false
}

func (w *EquipmentWindow) updateHoverStatus(ctx Context) {
	item, ok := w.itemAt(ctx.Session, ctx.Input.MouseX, ctx.Input.MouseY)
	if !ok {
		return
	}
	w.status = inventoryItemDisplayName(ctx.Resources, item)
	w.statusGood = true
	w.statusAt = time.Now()
}

func (w *EquipmentWindow) activateItem(ctx Context, item session.InventoryItem) {
	if ctx.Network == nil {
		w.setStatus("Not connected", false)
		return
	}
	if err := ctx.Network.SendTakeoffEquip(item.Index); err != nil {
		w.setStatus(err.Error(), false)
		return
	}
	w.setStatus("Unequip requested", true)
}

func (w *EquipmentWindow) setStatus(text string, good bool) {
	w.status = text
	w.statusGood = good
	w.statusAt = time.Now()
}

func (w *EquipmentWindow) drawSlotItem(screen *render.Image, ctx Context, assets AssetRenderer, slot equipmentSlotDef, item session.InventoryItem, x, y, width, height int) {
	name := inventoryItemDisplayName(ctx.Resources, item)
	if slot.side == equipmentSlotCenter {
		iconX := x + (width-inventoryIconSize)/2
		assets.DrawInventoryItemIcon(screen, ctx.Resources, item, iconX, y)
		if equipmentSlotShowsAmount(slot, item) {
			render.DebugPrintAtColor(screen, fmt.Sprintf("%d", item.Amount), x+width-18, y+height-14, TextColor)
		}
		return
	}
	if slot.side == equipmentSlotLeft {
		assets.DrawInventoryItemIcon(screen, ctx.Resources, item, x+4, y)
		render.DebugPrintAtColor(screen, trimRunes(name, 10), x+32, y+6, TextColor)
		if item.Refine > 0 {
			render.DebugPrintAtColor(screen, "+"+formatHUDNumber(int64(item.Refine)), x+2, y+1, shopGoodColor)
		}
		return
	}
	iconX := x + width - inventoryIconSize - 4
	assets.DrawInventoryItemIcon(screen, ctx.Resources, item, iconX, y)
	render.DebugPrintAtColor(screen, trimRunes(name, 10), x+4, y+6, TextColor)
	if item.Refine > 0 {
		render.DebugPrintAtColor(screen, "+"+formatHUDNumber(int64(item.Refine)), iconX-2, y+1, shopGoodColor)
	}
}

func equipmentSlotShowsAmount(slot equipmentSlotDef, item session.InventoryItem) bool {
	return slot.location == equipLocationAmmo && item.Amount > 0
}

func (w *EquipmentWindow) drawEmptySlotLabel(screen *render.Image, slot equipmentSlotDef, x, y int) {
	if slot.side == equipmentSlotCenter {
		render.DebugPrintAtColor(screen, "Ammo", x+14, y+6, MutedTextColor)
		return
	}
	label := slot.label
	if slot.side == equipmentSlotLeft {
		render.DebugPrintAtColor(screen, label, x+8, y+6, MutedTextColor)
		return
	}
	render.DebugPrintAtColor(screen, label, x+widthForRightSlotLabel(label), y+6, MutedTextColor)
}

func equippedItemForSlot(s *session.Session, location uint16) (session.InventoryItem, bool) {
	if s == nil || location == 0 {
		return session.InventoryItem{}, false
	}
	for _, item := range s.Inventory.Items {
		if !item.Equip || !item.Equipped || item.Location&location == 0 {
			continue
		}
		return item, true
	}
	return session.InventoryItem{}, false
}

func drawEquipmentContentSurface(screen *render.Image, x, y int) {
	if screen == nil {
		return
	}
	img := cachedEquipmentContentSurface()
	if img == nil {
		return
	}
	var opts render.DrawImageOptions
	opts.GeoM.Translate(float64(x), float64(y))
	opts.Filter = render.FilterNearest
	screen.DrawImage(img, &opts)
}

func cachedEquipmentContentSurface() *render.Image {
	if equipmentContentSurface != nil {
		return equipmentContentSurface
	}
	rows := make([]uiwidget.Widget, 0, 5)
	for row := 0; row < 5; row++ {
		rows = append(rows, primitives.HBox(
			equipmentCellBox(equipmentLeftColW, equipmentRowH, true),
			equipmentCenterCellBox(row),
			equipmentCellBox(equipmentRightColW, equipmentRowH, true),
		).Width(equipmentContentW).Height(equipmentRowH))
	}
	root := primitives.VBox(rows...).
		Width(equipmentContentW).
		Height(equipmentContentH).
		Background(Color(PanelBodyColor)).
		BorderStyle(1, Color(WindowBorderColor))
	r := offscreen.NewRenderer(equipmentContentW, equipmentContentH, offscreen.WithBackground(uiwidget.ColorTransparent))
	r.Render(root)
	src := r.Image()
	if src == nil {
		return nil
	}
	equipmentContentSurface = render.NewImageFromImage(src)
	return equipmentContentSurface
}

func equipmentCellBox(width, height int, border bool) *primitives.BoxWidget {
	box := primitives.Box().
		Width(float32(width)).
		Height(float32(height)).
		Background(Color(WindowBodyColor))
	if border {
		box.BorderStyle(1, Color(WindowBorderColor))
	}
	return box
}

func equipmentCenterCellBox(row int) *primitives.BoxWidget {
	bg := PanelBodyColor
	if row == 1 {
		bg = WindowBodyColor
	}
	return primitives.Box().
		Width(equipmentCenterColW).
		Height(equipmentRowH).
		Background(Color(bg))
}

func widthForRightSlotLabel(label string) int {
	return maxInt(4, equipmentRightColW-8-len([]rune(label))*7)
}

func equipmentSlotByLocation(location uint16) (equipmentSlotDef, bool) {
	for _, slot := range equipmentSlots {
		if location&slot.location != 0 {
			return slot, true
		}
	}
	return equipmentSlotDef{}, false
}
