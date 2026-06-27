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
