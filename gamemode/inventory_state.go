package gamemode

import (
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
)

func applyInventoryItemList(ctx Context, items []network.InventoryItem) {
	if ctx.Session == nil {
		return
	}
	for _, item := range items {
		addOrReplaceSessionInventoryItem(ctx.Session, sessionItemFromNetwork(item))
	}
	rebuildLocalEquipmentAppearance(ctx)
}

func applyStorageItemList(ctx Context, items []network.InventoryItem) {
	if ctx.Session == nil {
		return
	}
	ctx.Session.Storage.Open = true
	for _, item := range items {
		addOrReplaceSessionStorageItem(ctx.Session, sessionItemFromNetwork(item))
	}
}

func applyStorageAmount(ctx Context, amount network.StorageAmount) {
	if ctx.Session == nil {
		return
	}
	ctx.Session.Storage.Open = true
	ctx.Session.Storage.Amount = int(amount.Amount)
	ctx.Session.Storage.MaxAmount = int(amount.MaxAmount)
}

func applyStorageItemAdded(ctx Context, item network.InventoryItem) {
	if ctx.Session == nil {
		return
	}
	ctx.Session.Storage.Open = true
	addOrReplaceSessionStorageItem(ctx.Session, sessionItemFromNetwork(item))
}

func applyStorageItemRemoved(ctx Context, item network.StorageItemRemoved) {
	removeSessionStorageItem(ctx.Session, item.Index, int(item.Amount))
}

func applyStorageClosed(ctx Context) {
	if ctx.Session == nil {
		return
	}
	ctx.Session.Storage.Open = false
	ctx.Session.Storage.Items = nil
	ctx.Session.Storage.Amount = 0
	ctx.Session.Storage.MaxAmount = 0
}

func applyInventoryItemDelete(ctx Context, item network.InventoryItemDelete) {
	removeSessionInventoryItem(ctx.Session, item.Index, int(item.Amount))
	rebuildLocalEquipmentAppearance(ctx)
}

func applyUseItemAck(ctx Context, ack network.UseItemAck) {
	if ack.Result == 0 {
		return
	}
	setSessionInventoryItemAmount(ctx.Session, ack.Index, int(ack.Amount))
}

func sessionItemFromNetwork(item network.InventoryItem) session.InventoryItem {
	amount := int(item.Amount)
	if amount <= 0 {
		amount = 1
	}
	return session.InventoryItem{
		Index:      item.Index,
		ItemID:     item.ItemID,
		Type:       item.Type,
		Location:   item.Location,
		Identified: item.Identified,
		Amount:     amount,
		Equip:      item.Equip || inventoryItemTypeIsEquipment(item.Type),
		Equipped:   item.Equipped,
		Damaged:    item.Damaged,
		Refine:     item.Refine,
	}
}

func addOrReplaceSessionInventoryItem(s *session.Session, item session.InventoryItem) {
	if s == nil || item.Index == 0 {
		return
	}
	if item.Amount <= 0 {
		item.Amount = 1
	}
	for i := range s.Inventory.Items {
		if s.Inventory.Items[i].Index != item.Index {
			continue
		}
		s.Inventory.Items[i] = item
		return
	}
	s.Inventory.Items = append(s.Inventory.Items, item)
}

func addOrReplaceSessionStorageItem(s *session.Session, item session.InventoryItem) {
	if s == nil || item.Index == 0 {
		return
	}
	if item.Amount <= 0 {
		item.Amount = 1
	}
	for i := range s.Storage.Items {
		if s.Storage.Items[i].Index != item.Index {
			continue
		}
		s.Storage.Items[i] = item
		return
	}
	s.Storage.Items = append(s.Storage.Items, item)
}

func applyInventoryEquipAck(ctx Context, ack network.InventoryEquipAck) {
	if ctx.Session == nil || !ack.Success || ack.Index == 0 {
		return
	}
	for i := range ctx.Session.Inventory.Items {
		item := &ctx.Session.Inventory.Items[i]
		if item.Index != ack.Index {
			if !ack.Unequip && ack.Location != 0 && item.Equipped && item.Location&ack.Location != 0 {
				item.Equipped = false
			}
			continue
		}
		item.Equipped = !ack.Unequip
		if !ack.Unequip && ack.Location != 0 {
			item.Location = ack.Location
		}
	}
	rebuildLocalEquipmentAppearance(ctx)
}

func rebuildLocalEquipmentAppearance(ctx Context) {
	if ctx.Session == nil {
		return
	}
	sawEquipment := false
	hasWeapon := false
	hasShield := false
	weapon := int(ctx.Session.Selected.Weapon)
	shield := int(ctx.Session.Selected.Shield)
	for _, item := range ctx.Session.Inventory.Items {
		if !item.Equip {
			continue
		}
		sawEquipment = true
		if !item.Equipped || item.Location == 0 {
			continue
		}
		occupiesRightHand := item.Location&equipLocationWeapon != 0
		occupiesLeftHand := item.Location&equipLocationShield != 0
		if occupiesRightHand {
			hasWeapon = true
			weapon = res.PlayerWeaponViewID(ctx.Resources, int(item.ItemID))
		}
		if occupiesLeftHand && !occupiesRightHand {
			hasShield = true
			shield = int(item.ItemID)
		}
	}
	if !sawEquipment {
		return
	}
	if !hasWeapon {
		weapon = 0
	}
	if !hasShield {
		shield = 0
	}
	updateLocalWeaponAppearance(ctx, weapon, shield)
}

func updateLocalWeaponAppearance(ctx Context, weapon, shield int) {
	if ctx.Session == nil {
		return
	}
	weapon, shield = res.NormalizePlayerWeaponShield(weapon, shield)
	ctx.Session.Selected.Weapon = int16(weapon)
	ctx.Session.Selected.Shield = int16(shield)
	for i := range ctx.Session.Characters {
		if ctx.Session.Characters[i].ID == ctx.Session.CharID || ctx.Session.Characters[i].ID == ctx.Session.Selected.ID {
			ctx.Session.Characters[i].Weapon = int16(weapon)
			ctx.Session.Characters[i].Shield = int16(shield)
		}
	}
	if ctx.World != nil {
		ctx.World.Player.Weapon = int16(weapon)
		ctx.World.Player.Shield = int16(shield)
	}
}

func addPickedSessionInventoryItem(s *session.Session, item session.InventoryItem) {
	if s == nil || item.Index == 0 {
		return
	}
	if item.Amount <= 0 {
		item.Amount = 1
	}
	for i := range s.Inventory.Items {
		if s.Inventory.Items[i].Index != item.Index {
			continue
		}
		if s.Inventory.Items[i].ItemID == item.ItemID && !s.Inventory.Items[i].Equip {
			s.Inventory.Items[i].Amount += item.Amount
			if item.Location != 0 {
				s.Inventory.Items[i].Location = item.Location
			}
			return
		}
		s.Inventory.Items[i] = item
		return
	}
	s.Inventory.Items = append(s.Inventory.Items, item)
}

func removeSessionInventoryItem(s *session.Session, index uint16, amount int) {
	if s == nil || index == 0 {
		return
	}
	if amount <= 0 {
		amount = 1
	}
	for i := range s.Inventory.Items {
		if s.Inventory.Items[i].Index != index {
			continue
		}
		s.Inventory.Items[i].Amount -= amount
		if s.Inventory.Items[i].Amount > 0 {
			return
		}
		s.Inventory.Items = append(s.Inventory.Items[:i], s.Inventory.Items[i+1:]...)
		return
	}
}

func removeSessionStorageItem(s *session.Session, index uint16, amount int) {
	if s == nil || index == 0 {
		return
	}
	if amount <= 0 {
		amount = 1
	}
	for i := range s.Storage.Items {
		if s.Storage.Items[i].Index != index {
			continue
		}
		s.Storage.Items[i].Amount -= amount
		if s.Storage.Items[i].Amount > 0 {
			return
		}
		s.Storage.Items = append(s.Storage.Items[:i], s.Storage.Items[i+1:]...)
		return
	}
}

func setSessionInventoryItemAmount(s *session.Session, index uint16, amount int) {
	if s == nil || index == 0 {
		return
	}
	for i := range s.Inventory.Items {
		if s.Inventory.Items[i].Index != index {
			continue
		}
		if amount > 0 {
			s.Inventory.Items[i].Amount = amount
			return
		}
		s.Inventory.Items = append(s.Inventory.Items[:i], s.Inventory.Items[i+1:]...)
		return
	}
}
