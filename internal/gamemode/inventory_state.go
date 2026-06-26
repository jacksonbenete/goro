package gamemode

import (
	"github.com/kivutar/goro/internal/network"
	"github.com/kivutar/goro/internal/session"
)

func applyInventoryItemList(ctx Context, items []network.InventoryItem) {
	if ctx.Session == nil {
		return
	}
	for _, item := range items {
		addOrReplaceSessionInventoryItem(ctx.Session, sessionItemFromNetwork(item))
	}
}

func applyInventoryItemDelete(ctx Context, item network.InventoryItemDelete) {
	removeSessionInventoryItem(ctx.Session, item.Index, int(item.Amount))
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
		Equip:      item.Equip,
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
			s.Inventory.Items[i].Identified = item.Identified
			s.Inventory.Items[i].Type = item.Type
			s.Inventory.Items[i].Damaged = item.Damaged
			s.Inventory.Items[i].Refine = item.Refine
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
