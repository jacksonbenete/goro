package ui

import (
	"testing"

	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestShopAcceptInventoryDropAddsSellableItem(t *testing.T) {
	window := ShopWindow{
		open: true,
		mode: shopModeSell,
		x:    100,
		y:    100,
		sellable: map[uint16]network.ShopSellItem{
			7: {Index: 7, Price: 10, OverchargePrice: 12},
		},
	}

	ok := window.AcceptInventoryDrop(Context{}, session.InventoryItem{Index: 7, ItemID: 938, Amount: 3}, 120, 150)
	if !ok {
		t.Fatal("drop was not accepted")
	}
	if len(window.cart) != 1 || window.cart[0].amount != 1 || window.cart[0].max != 3 || window.cart[0].over != 12 {
		t.Fatalf("cart = %+v", window.cart)
	}
}

func TestInventoryDragReleaseOverShopAddsCartItem(t *testing.T) {
	inputState := input.NewState()
	inputState.SetMousePosition(120, 150)
	sessionState := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{{Index: 7, ItemID: 938, Amount: 3}},
		},
	}
	ctx := Context{Input: inputState, Session: sessionState}
	inventory := InventoryWindow{
		open:       true,
		positioned: true,
		x:          500,
		y:          100,
		dragActive: true,
		dragItem:   session.InventoryItem{Index: 7, ItemID: 938, Amount: 3},
	}
	shop := ShopWindow{
		open: true,
		mode: shopModeSell,
		x:    100,
		y:    100,
		sellable: map[uint16]network.ShopSellItem{
			7: {Index: 7, Price: 10, OverchargePrice: 12},
		},
	}

	if !inventory.Update(ctx, &shop, nil) {
		t.Fatal("inventory update did not consume drag release")
	}
	if inventory.dragActive {
		t.Fatal("drag still active after release")
	}
	if len(shop.cart) != 1 || shop.cart[0].item.Index != 7 {
		t.Fatalf("shop cart = %+v, want dropped item", shop.cart)
	}
}

func TestShopBuyCartTracksQuantityAndTotal(t *testing.T) {
	window := ShopWindow{mode: shopModeBuy}
	item := network.ShopBuyItem{ItemID: 501, Price: 100, DiscountPrice: 80}

	window.addBuyItem(item)
	window.addBuyItem(item)
	if got := window.buyAmount(501); got != 2 {
		t.Fatalf("buy amount = %d, want 2", got)
	}
	if got := window.total(); got != 160 {
		t.Fatalf("total = %d, want 160", got)
	}

	window.decrementBuyItem(501)
	if got := window.buyAmount(501); got != 1 {
		t.Fatalf("buy amount after decrement = %d, want 1", got)
	}
}

func TestInventoryBagClassifiesTabs(t *testing.T) {
	tests := []struct {
		name string
		item session.InventoryItem
		tab  int
	}{
		{name: "healing item", item: session.InventoryItem{Type: 0}, tab: inventoryBagTabItem},
		{name: "usable item", item: session.InventoryItem{Type: 2}, tab: inventoryBagTabItem},
		{name: "equipment flag", item: session.InventoryItem{Type: 4, Equip: true}, tab: inventoryBagTabEquip},
		{name: "weapon type", item: session.InventoryItem{Type: 5}, tab: inventoryBagTabEquip},
		{name: "pet egg type", item: session.InventoryItem{Type: 7}, tab: inventoryBagTabEquip},
		{name: "etc", item: session.InventoryItem{Type: 3}, tab: inventoryBagTabEtc},
		{name: "card", item: session.InventoryItem{Type: 6}, tab: inventoryBagTabEtc},
		{name: "ammo", item: session.InventoryItem{Type: 10}, tab: inventoryBagTabEquip},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := inventoryItemTab(tc.item); got != tc.tab {
				t.Fatalf("tab = %d, want %d", got, tc.tab)
			}
		})
	}
}

func TestInventoryBagRightClickOpensItemInfo(t *testing.T) {
	inputState := input.NewState()
	inputState.SetMousePosition(165, 145)
	inputState.SetMouseButton(input.MouseButtonRight, true)
	sessionState := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{{Index: 3, ItemID: 512, Type: 0, Amount: 2, Identified: true}},
		},
	}
	ctx := Context{Input: inputState, Session: sessionState, ScreenW: 800, ScreenH: 600}
	bag := InventoryBagWindow{
		open:       true,
		x:          100,
		y:          100,
		positioned: true,
		tab:        inventoryBagTabItem,
	}
	info := ItemInfoWindow{}

	if !bag.Update(ctx, nil, nil, &info) {
		t.Fatal("right click did not consume inventory bag input")
	}
	if !info.open || info.item.ItemID != 512 {
		t.Fatalf("item info = open:%v item:%+v, want item 512", info.open, info.item)
	}
	if bag.dragActive {
		t.Fatal("right click should not start item drag")
	}
}

func TestInventoryBagOpensUnderBasicMenu(t *testing.T) {
	bag := InventoryBagWindow{x: 400, y: 200, positioned: true}
	bag.Toggle(Context{ScreenW: 1280, ScreenH: 720})
	menuX, menuY, _, menuH := basicMenuBounds()

	if !bag.open {
		t.Fatal("inventory bag did not open")
	}
	if bag.x != menuX || bag.y != menuY+menuH+8 {
		t.Fatalf("inventory position = %d,%d, want %d,%d", bag.x, bag.y, menuX, menuY+menuH+8)
	}
}

func TestStorageAcceptInventoryDropWithoutNetworkConsumesDrop(t *testing.T) {
	window := StorageWindow{
		open:       true,
		positioned: true,
		x:          100,
		y:          100,
	}
	sessionState := &session.Session{Storage: session.Storage{Open: true}}
	ok := window.AcceptInventoryDrop(Context{Session: sessionState}, session.InventoryItem{Index: 7, ItemID: 938, Amount: 3}, 120, 150)
	if !ok {
		t.Fatal("drop over storage was not consumed")
	}
	if window.status == "" || window.statusGood {
		t.Fatalf("status = %q good=%v, want not connected error", window.status, window.statusGood)
	}
}
