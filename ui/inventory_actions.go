package ui

import (
	"fmt"

	"github.com/kivutar/goro/session"
)

func useInventoryItem(ctx Context, item session.InventoryItem) error {
	if ctx.Network == nil {
		return fmt.Errorf("not connected")
	}
	if !inventoryItemIsUsable(item) {
		return fmt.Errorf("item cannot be used")
	}
	target := uint32(0)
	if ctx.Session != nil {
		target = ctx.Session.AccountID
		if target == 0 {
			target = ctx.Session.CharID
		}
	}
	if target == 0 {
		return fmt.Errorf("missing player id")
	}
	return ctx.Network.SendUseInventoryItem(item.Index, target)
}

func dropInventoryItem(ctx Context, item session.InventoryItem) error {
	if ctx.Network == nil {
		return fmt.Errorf("not connected")
	}
	return ctx.Network.SendDropInventoryItem(item.Index, inventoryDropAmount(item))
}

func inventoryDropAmount(_ session.InventoryItem) uint16 {
	return 1
}
