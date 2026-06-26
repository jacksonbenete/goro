package gamemode

import (
	"image/color"
	"math"
	"time"

	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/primitives"
	uiwidget "github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/internal/render"
	"github.com/kivutar/goro/internal/session"
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

type equipmentWindowState struct {
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

func (w *equipmentWindowState) toggle(ctx Context) {
	if w.open {
		w.open = false
		w.dragging = false
		return
	}
	w.open = true
	w.ensurePosition(ctx)
}

func (w *equipmentWindowState) update(ctx Context) bool {
	if !w.open || ctx.Input == nil {
		return false
	}
	w.ensurePosition(ctx)
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

func (w *equipmentWindowState) draw(screen *render.Image, ctx Context, mode *WorldMode) {
	if !w.open || screen == nil {
		return
	}
	w.ensurePosition(ctx)
	x, y := w.x, w.y
	drawNPCWindowFrame(screen, x, y, equipmentWindowWidth, equipmentWindowHeight)
	render.DebugPrintAtColor(screen, "Equipment", x+equipmentWindowPad, y+9, inventoryTitleColor)
	cx, cy, cw, ch := w.closeBounds()
	drawUIButtonSurface(screen, cx, cy, cw, ch, inventoryButtonColor)
	render.DebugPrintAtColor(screen, "x", cx+5, cy+(ch-13)/2-1, inventoryTextColor)
	render.DrawRect(screen, float64(x+8), float64(y+equipmentWindowTitleH), float64(equipmentWindowWidth-16), 1, color.RGBA{R: 210, G: 200, B: 170, A: 80})

	contentX, contentY := w.contentOrigin()
	drawEquipmentContentSurface(screen, contentX, contentY)
	px, py, pw, ph := w.previewBounds()
	if mode != nil {
		mode.drawEquipmentPreview(screen, ctx, px, py, pw, ph)
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
		if ok && mode != nil {
			w.drawSlotItem(screen, ctx, mode, slot, item, sx, sy, sw, sh)
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
		render.DebugPrintAtColor(screen, jobName(int(character.Job)), x+equipmentWindowPad, footerY+16, inventoryMutedColor)
	}
	if w.status != "" && time.Since(w.statusAt) < 2200*time.Millisecond {
		statusColor := inventoryMutedColor
		if !w.statusGood {
			statusColor = shopErrorColor
		}
		render.DebugPrintAtColor(screen, trimRunes(w.status, 26), x+118, y+equipmentWindowHeight-21, statusColor)
	}
}

func (w *equipmentWindowState) cursorAction(ctx Context) (int, bool) {
	if !w.open || ctx.Input == nil {
		return 0, false
	}
	if pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, equipmentWindowWidth, equipmentWindowHeight) {
		return cursorActionClick, true
	}
	return 0, false
}

func (w *equipmentWindowState) ensurePosition(ctx Context) {
	if w.positioned {
		return
	}
	width, _ := ctx.ScreenSize()
	w.x = maxInt(8, width-equipmentWindowWidth-24)
	w.y = 104
	w.positioned = true
}

func (w *equipmentWindowState) closeBounds() (int, int, int, int) {
	return w.x + equipmentWindowWidth - 23, w.y + 7, 16, 16
}

func (w *equipmentWindowState) contentOrigin() (int, int) {
	return w.x + equipmentWindowPad, w.y + equipmentWindowTitleH + equipmentWindowPad
}

func (w *equipmentWindowState) previewBounds() (int, int, int, int) {
	x, y := w.contentOrigin()
	return x + equipmentLeftColW, y + 2, equipmentCenterColW, 125
}

func (w *equipmentWindowState) slotBounds(slot equipmentSlotDef) (int, int, int, int) {
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

func (w *equipmentWindowState) itemAt(s *session.Session, mx, my int) (session.InventoryItem, bool) {
	for _, slot := range equipmentSlots {
		x, y, width, height := w.slotBounds(slot)
		if !pointInRect(mx, my, x, y, width, height) {
			continue
		}
		return equippedItemForSlot(s, slot.location)
	}
	return session.InventoryItem{}, false
}

func (w *equipmentWindowState) updateHoverStatus(ctx Context) {
	item, ok := w.itemAt(ctx.Session, ctx.Input.MouseX, ctx.Input.MouseY)
	if !ok {
		return
	}
	w.status = inventoryItemDisplayName(ctx.Resources, item)
	w.statusGood = true
	w.statusAt = time.Now()
}

func (w *equipmentWindowState) activateItem(ctx Context, item session.InventoryItem) {
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

func (w *equipmentWindowState) setStatus(text string, good bool) {
	w.status = text
	w.statusGood = good
	w.statusAt = time.Now()
}

func (w *equipmentWindowState) drawSlotItem(screen *render.Image, ctx Context, mode *WorldMode, slot equipmentSlotDef, item session.InventoryItem, x, y, width, height int) {
	name := inventoryItemDisplayName(ctx.Resources, item)
	if slot.side == equipmentSlotCenter {
		iconX := x + (width-inventoryIconSize)/2
		mode.drawInventoryItemIcon(screen, ctx.Resources, item, iconX, y)
		return
	}
	if slot.side == equipmentSlotLeft {
		mode.drawInventoryItemIcon(screen, ctx.Resources, item, x+4, y)
		render.DebugPrintAtColor(screen, trimRunes(name, 10), x+32, y+6, color.RGBA{R: 66, G: 60, B: 54, A: 255})
		if item.Refine > 0 {
			render.DebugPrintAtColor(screen, "+"+formatHUDNumber(int64(item.Refine)), x+2, y+1, shopGoodColor)
		}
		return
	}
	iconX := x + width - inventoryIconSize - 4
	mode.drawInventoryItemIcon(screen, ctx.Resources, item, iconX, y)
	render.DebugPrintAtColor(screen, trimRunes(name, 10), x+4, y+6, color.RGBA{R: 66, G: 60, B: 54, A: 255})
	if item.Refine > 0 {
		render.DebugPrintAtColor(screen, "+"+formatHUDNumber(int64(item.Refine)), iconX-2, y+1, shopGoodColor)
	}
}

func (w *equipmentWindowState) drawEmptySlotLabel(screen *render.Image, slot equipmentSlotDef, x, y int) {
	if slot.side == equipmentSlotCenter {
		render.DebugPrintAtColor(screen, "Ammo", x+14, y+6, color.RGBA{R: 98, G: 94, B: 86, A: 180})
		return
	}
	label := slot.label
	if slot.side == equipmentSlotLeft {
		render.DebugPrintAtColor(screen, label, x+8, y+6, color.RGBA{R: 98, G: 94, B: 86, A: 180})
		return
	}
	render.DebugPrintAtColor(screen, label, x+widthForRightSlotLabel(label), y+6, color.RGBA{R: 98, G: 94, B: 86, A: 180})
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

func (m *WorldMode) drawEquipmentPreview(screen *render.Image, ctx Context, x, y, width, height int) {
	if screen == nil || width <= 0 || height <= 0 {
		return
	}
	view := m.playerView
	if view == nil && ctx.Resources != nil && ctx.Session != nil {
		loaded, _ := loadPlayerHumanoidSpriteView(ctx.Resources, selectedCharacter(ctx.Session), ctx.Session.Sex)
		view = loaded
		if loaded != nil {
			m.playerView = loaded
		}
	}
	state := spriteState{
		actionFamily: spriteActionIdle,
		direction:    equipmentPreviewWorldDirection,
	}
	billboard, ok := humanoidBillboardForState(view, state, time.Now())
	if !ok || billboard == nil || billboard.image == nil {
		drawPanel(screen, float64(x+width/2-14), float64(y+height/2-24), 28, 48)
		return
	}
	bounds := visibleImageBounds(billboard.image)
	if bounds.Empty() {
		return
	}
	srcW, srcH := float64(bounds.Dx()), float64(bounds.Dy())
	if srcW <= 0 || srcH <= 0 {
		return
	}
	scale := math.Min(float64(width-4)/srcW, float64(height-4)/srcH)
	scale = math.Min(scale, 1.6)
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	dstW, dstH := srcW*scale, srcH*scale
	dstX := float64(x) + (float64(width)-dstW)/2
	dstY := float64(y+height) - dstH - 7
	vertices := []render.Vertex{
		{DstX: float32(dstX), DstY: float32(dstY), SrcX: float32(bounds.Min.X), SrcY: float32(bounds.Min.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(dstX + dstW), DstY: float32(dstY), SrcX: float32(bounds.Max.X), SrcY: float32(bounds.Min.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(dstX), DstY: float32(dstY + dstH), SrcX: float32(bounds.Min.X), SrcY: float32(bounds.Max.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(dstX + dstW), DstY: float32(dstY + dstH), SrcX: float32(bounds.Max.X), SrcY: float32(bounds.Max.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
	screen.DrawTriangles(vertices, []uint16{0, 1, 2, 2, 1, 3}, billboard.image, &render.DrawTrianglesOptions{Filter: spriteDrawFilter(), Address: render.AddressClampToZero})
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
		Background(uiColor(color.RGBA{R: 226, G: 222, B: 208, A: 224})).
		BorderStyle(1, uiColor(color.RGBA{R: 95, G: 85, B: 72, A: 120}))
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
		Background(uiColor(color.RGBA{R: 244, G: 244, B: 236, A: 176}))
	if border {
		box.BorderStyle(1, uiColor(color.RGBA{R: 95, G: 85, B: 72, A: 70}))
	}
	return box
}

func equipmentCenterCellBox(row int) *primitives.BoxWidget {
	bg := color.RGBA{R: 232, G: 229, B: 216, A: 120}
	if row == 1 {
		bg = color.RGBA{R: 244, G: 244, B: 236, A: 132}
	}
	return primitives.Box().
		Width(equipmentCenterColW).
		Height(equipmentRowH).
		Background(uiColor(bg))
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
