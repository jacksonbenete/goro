package ui

import (
	"fmt"
	"image"
	"time"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	equipmentChoiceWindowWidth  = 312
	equipmentChoiceTableHeaderH = 24
	equipmentChoiceRowH         = 32
	equipmentChoiceRows         = 6
	equipmentChoiceWindowHeight = ROWindowTitleHeight + equipmentChoiceTableHeaderH + equipmentChoiceRows*equipmentChoiceRowH + ROWindowFooterHeight
)

type equipmentChoice struct {
	Index  uint16
	ItemID uint16
	Refine uint8
	Cards  [4]uint16
}

type equipmentChoiceWindow struct {
	Window
	scrollY      state.Signal[float32]
	selectedRow  int
	items        []equipmentChoice
	title        string
	emptyText    string
	lastClickAt  time.Time
	lastClickRow int
	icons        map[identifyItemIconKey]image.Image
	iconMiss     map[identifyItemIconKey]struct{}
	onConfirm    func(Context, equipmentChoice)
	onCancel     func(Context)
}

func (w *equipmentChoiceWindow) OpenChoices(ctx Context, title, emptyText string, items []equipmentChoice, onConfirm func(Context, equipmentChoice), onCancel func(Context)) {
	w.EnsureWindow(equipmentChoiceWindowWidth, equipmentChoiceWindowHeight)
	w.title = title
	w.emptyText = emptyText
	w.onConfirm = onConfirm
	w.onCancel = onCancel
	w.items = append(w.items[:0], items...)
	w.selectedRow = 0
	w.lastClickRow = -1
	w.lastClickAt = time.Time{}
	w.ensureScrollSignal().Set(0)
	w.clampScroll()
	if len(w.items) == 0 {
		w.Close()
		w.Publish(ctx)
		return
	}
	w.Open(ctx, w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *equipmentChoiceWindow) ApplyAck(ctx Context, failure bool, label string) {
	w.EnsureWindow(equipmentChoiceWindowWidth, equipmentChoiceWindowHeight)
	if failure {
		glog.Warnf("%s failed", label)
	}
	w.Close()
	w.Publish(ctx)
}

func (w *equipmentChoiceWindow) Update(ctx Context) bool {
	w.EnsureWindow(equipmentChoiceWindowWidth, equipmentChoiceWindowHeight)
	if !w.IsOpen() {
		return false
	}
	w.clampScroll()
	if ctx.Input != nil && ctx.Input.JustPressed(input.KeyEscape) {
		w.cancel(ctx)
		w.Publish(ctx)
		return true
	}
	consumed := w.Window.Update(ctx)
	keyboardConsumed := false
	if w.IsOpen() {
		keyboardConsumed = updateSelectionTableKeyboard(ctx, &w.selectedRow, len(w.items), equipmentChoiceRows, equipmentChoiceRowH, w.ensureScrollSignal(), func() {
			w.confirm(ctx)
		})
		if keyboardConsumed && w.IsOpen() {
			w.RebindContent(ctx, w.widgetTree(ctx))
		}
	}
	if w.IsOpen() {
		w.updateDoubleClick(ctx)
	}
	w.Publish(ctx)
	return consumed || keyboardConsumed
}

func (w *equipmentChoiceWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title(w.title),
		CloseButton(true),
		OnClose(func() {
			w.cancel(ctx)
			w.Publish(ctx)
		}),
		Size(equipmentChoiceWindowWidth, equipmentChoiceWindowHeight),
		Content(
			primitives.Box(w.tableWidget(ctx)).
				Height(equipmentChoiceTableHeight()).
				Background(rotheme.Default.Colors.PanelBody),
		),
		Footer(
			primitives.Expanded(primitives.Box()),
			rotheme.Button("Cancel", func() {
				w.cancel(ctx)
				w.Publish(ctx)
			}),
			rotheme.Button("OK", func() {
				w.confirm(ctx)
				w.Publish(ctx)
			}),
		),
	)
}

func (w *equipmentChoiceWindow) tableWidget(ctx Context) *rotheme.TableViewWidget {
	items := append([]equipmentChoice(nil), w.items...)
	return itemTableView(
		w.itemTableRows(ctx, items),
		"Item",
		equipmentChoiceRowH,
		equipmentChoiceTableHeaderH,
		w.emptyText,
		w.ensureScrollSignal(),
		w.selectedRow,
		func(row int) {
			w.selectedRow = row
		},
	)
}

func (w *equipmentChoiceWindow) confirm(ctx Context) {
	if w.selectedRow < 0 || w.selectedRow >= len(w.items) {
		return
	}
	if w.onConfirm != nil {
		w.onConfirm(ctx, w.items[w.selectedRow])
		return
	}
	w.Close()
}

func (w *equipmentChoiceWindow) cancel(ctx Context) {
	if w.onCancel != nil {
		w.onCancel(ctx)
		return
	}
	w.Close()
}

func (w *equipmentChoiceWindow) updateDoubleClick(ctx Context) {
	if ctx.Input == nil || !ctx.Input.MouseJustPressed(input.MouseButtonLeft) {
		return
	}
	row, ok := w.rowAtMouse(ctx.Input.MouseX, ctx.Input.MouseY)
	if !ok {
		return
	}
	now := time.Now()
	if w.lastClickRow == row && now.Sub(w.lastClickAt) <= 360*time.Millisecond {
		w.selectedRow = row
		w.lastClickRow = -1
		w.lastClickAt = time.Time{}
		w.confirm(ctx)
		return
	}
	w.lastClickRow = row
	w.lastClickAt = now
}

func (w *equipmentChoiceWindow) rowAtMouse(mouseX, mouseY int) (int, bool) {
	tableX := w.x
	tableY := w.y + ROWindowTitleHeight
	rowY := tableY + equipmentChoiceTableHeaderH
	if !pointInRect(mouseX, mouseY, tableX, rowY, scrollbarSafeIntWidth(equipmentChoiceWindowWidth), equipmentChoiceRows*equipmentChoiceRowH) {
		return 0, false
	}
	row := int((float32(mouseY-rowY) + w.ensureScrollSignal().Get()) / equipmentChoiceRowH)
	if row < 0 || row >= len(w.items) {
		return 0, false
	}
	return row, true
}

func (w *equipmentChoiceWindow) itemTableRows(ctx Context, items []equipmentChoice) []itemTableRow {
	rows := make([]itemTableRow, len(items))
	for i, item := range items {
		rows[i] = itemTableRow{
			name: equipmentChoiceDisplayName(ctx, item),
			icon: w.itemIconImage(ctx.Resources, item.ItemID),
		}
	}
	return rows
}

func equipmentChoiceDisplayName(ctx Context, item equipmentChoice) string {
	name := inventoryItemDisplayName(ctx.Resources, session.InventoryItem{ItemID: item.ItemID, Identified: true, Refine: item.Refine, Cards: item.Cards})
	if item.Refine > 0 {
		name = fmt.Sprintf("+%d %s", item.Refine, name)
	}
	return name
}

func (w *equipmentChoiceWindow) itemIconImage(manager *res.Manager, itemID uint16) image.Image {
	return cachedIdentifiedItemIcon(manager, itemID, &w.icons, &w.iconMiss)
}

func (w *equipmentChoiceWindow) clampScroll() {
	if w.selectedRow >= len(w.items) {
		w.selectedRow = -1
	}
	scroll := w.ensureScrollSignal()
	maxScroll := float32(maxInt(0, len(w.items)-equipmentChoiceRows) * equipmentChoiceRowH)
	switch value := scroll.Get(); {
	case value < 0:
		scroll.Set(0)
	case value > maxScroll:
		scroll.Set(maxScroll)
	}
}

func (w *equipmentChoiceWindow) ensureScrollSignal() state.Signal[float32] {
	if w.scrollY == nil {
		w.scrollY = state.NewSignal[float32](0)
	}
	return w.scrollY
}

func equipmentChoiceTableHeight() float32 {
	return equipmentChoiceTableHeaderH + equipmentChoiceRows*equipmentChoiceRowH
}

type RepairItemWindow struct {
	equipmentChoiceWindow
}

func (w *RepairItemWindow) OpenList(ctx Context, list network.RepairItemList) {
	items := make([]equipmentChoice, len(list.Items))
	for i, item := range list.Items {
		items[i] = equipmentChoice{Index: item.Index, ItemID: item.ItemID, Refine: item.Refine, Cards: item.Cards}
	}
	w.OpenChoices(ctx, "Repair Item", "No damaged items", items, func(ctx Context, item equipmentChoice) {
		w.repairItem(ctx, item)
	}, func(ctx Context) {
		w.cancelRepair(ctx)
	})
}

func (w *RepairItemWindow) ApplyAck(ctx Context, ack network.RepairItemAck) {
	glog.Debugf("item repair ack index=%d success=%v", ack.Index, ack.Success())
	w.equipmentChoiceWindow.ApplyAck(ctx, ack.Failure, "item repair")
}

func (w *RepairItemWindow) repairItem(ctx Context, item equipmentChoice) {
	if ctx.Network == nil {
		glog.Warnf("item repair failed: not connected")
		return
	}
	request := network.RepairItem{Index: item.Index, ItemID: item.ItemID, Refine: item.Refine, Cards: item.Cards}
	if err := ctx.Network.SendRepairItem(request); err != nil {
		glog.Warnf("item repair failed: %v", err)
		return
	}
	glog.Debugf("item repair requested index=%d item=%d refine=%d cards=%v", item.Index, item.ItemID, item.Refine, item.Cards)
	w.Close()
}

func (w *RepairItemWindow) cancelRepair(ctx Context) {
	if ctx.Network != nil {
		if err := ctx.Network.SendRepairItem(network.RepairItem{Index: 0xFFFF}); err != nil {
			glog.Warnf("item repair cancel failed: %v", err)
		}
	}
	w.Close()
}

type WeaponRefineWindow struct {
	equipmentChoiceWindow
}

func (w *WeaponRefineWindow) OpenList(ctx Context, list network.WeaponRefineList) {
	items := make([]equipmentChoice, len(list.Items))
	for i, item := range list.Items {
		items[i] = equipmentChoice{Index: item.Index, ItemID: item.ItemID, Refine: item.Refine, Cards: item.Cards}
	}
	w.OpenChoices(ctx, "Weapon Refine", "No weapons", items, func(ctx Context, item equipmentChoice) {
		w.refineWeapon(ctx, item)
	}, func(ctx Context) {
		w.cancelRefine(ctx)
	})
}

func (w *WeaponRefineWindow) ApplyAck(ctx Context, ack network.WeaponRefineAck) {
	glog.Debugf("weapon refine ack item=%d result=%d success=%v", ack.ItemID, ack.Result, ack.Success())
	w.equipmentChoiceWindow.ApplyAck(ctx, !ack.Success(), "weapon refine")
}

func (w *WeaponRefineWindow) refineWeapon(ctx Context, item equipmentChoice) {
	if ctx.Network == nil {
		glog.Warnf("weapon refine failed: not connected")
		return
	}
	if err := ctx.Network.SendWeaponRefine(uint32(item.Index)); err != nil {
		glog.Warnf("weapon refine failed: %v", err)
		return
	}
	glog.Debugf("weapon refine requested index=%d item=%d refine=%d cards=%v", item.Index, item.ItemID, item.Refine, item.Cards)
	w.Close()
}

func (w *WeaponRefineWindow) cancelRefine(ctx Context) {
	if ctx.Network != nil {
		if err := ctx.Network.SendWeaponRefine(0xFFFFFFFF); err != nil {
			glog.Warnf("weapon refine cancel failed: %v", err)
		}
	}
	w.Close()
}
