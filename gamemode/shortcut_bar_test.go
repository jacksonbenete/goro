package gamemode

import (
	"testing"

	"github.com/kivutar/goro/session"
)

func TestShortcutSlotPersistRoundTrip(t *testing.T) {
	item := shortcutSlotState{
		kind:       shortcutItem,
		itemIndex:  12,
		itemID:     601,
		identified: true,
	}
	if got := shortcutSlotFromPersist(item.persist()); got != item {
		t.Fatalf("item slot = %+v, want %+v", got, item)
	}

	skill := shortcutSlotState{
		kind:       shortcutSkill,
		skillID:    6,
		skillLevel: 2,
	}
	if got := shortcutSlotFromPersist(skill.persist()); got != skill {
		t.Fatalf("skill slot = %+v, want %+v", got, skill)
	}
}

func TestInventoryItemForShortcutFallsBackToItemID(t *testing.T) {
	s := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{
				{Index: 8, ItemID: 501, Amount: 2},
				{Index: 14, ItemID: 601, Amount: 3},
			},
		},
	}

	item, ok := inventoryItemForShortcut(s, 99, 601)
	if !ok {
		t.Fatal("item not found")
	}
	if item.Index != 14 || item.ItemID != 601 {
		t.Fatalf("item = %+v", item)
	}
}

func TestInventoryItemForShortcutRejectsReusedIndexWithDifferentItem(t *testing.T) {
	s := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{
				{Index: 12, ItemID: 602, Amount: 1},
			},
		},
	}

	item, ok := inventoryItemForShortcut(s, 12, 501)
	if ok {
		t.Fatalf("shortcut resolved reused index to wrong item: %+v", item)
	}
}

func TestShortcutBarClearsDepletedItem(t *testing.T) {
	bar := &shortcutBarState{}
	bar.slots[2] = shortcutSlotState{kind: shortcutItem, itemIndex: 12, itemID: 501}
	bar.slots[3] = shortcutSlotState{kind: shortcutItem, itemIndex: 12, itemID: 602}
	bar.slots[4] = shortcutSlotState{kind: shortcutItem, itemIndex: 14, itemID: 501}

	if !bar.clearDepletedItemSlots(12, 501) {
		t.Fatal("depleted shortcut was not cleared")
	}
	if bar.slots[2].kind != shortcutEmpty {
		t.Fatalf("slot 3 kind = %d, want empty", bar.slots[2].kind)
	}
	if bar.slots[3].kind != shortcutItem || bar.slots[4].kind != shortcutItem {
		t.Fatalf("unrelated shortcuts were cleared: slot4=%+v slot5=%+v", bar.slots[3], bar.slots[4])
	}
}
