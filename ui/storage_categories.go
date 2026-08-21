package ui

import (
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/session"
)

type storageCategory uint8

const (
	storageCategoryItem storageCategory = iota
	storageCategoryKafra
	storageCategoryArmor
	storageCategoryArms
	storageCategoryAmmo
	storageCategoryCard
	storageCategoryEtc
	storageCategoryCount
)

var storageCategoryTabs = [...]struct {
	label    string
	category storageCategory
}{
	{label: "Item", category: storageCategoryItem},
	{label: "Kafra", category: storageCategoryKafra},
	{label: "Armor", category: storageCategoryArmor},
	{label: "Arms", category: storageCategoryArms},
	{label: "Ammo", category: storageCategoryAmmo},
	{label: "Card", category: storageCategoryCard},
	{label: "Etc", category: storageCategoryEtc},
}

func storageItemCategory(item session.InventoryItem) storageCategory {
	switch item.Type {
	case db.ItemTypeHealing, db.ItemTypeUsable, db.ItemTypeDelayConsume:
		return storageCategoryItem
	case db.ItemTypeCash:
		return storageCategoryKafra
	case db.ItemTypeArmor, db.ItemTypeShadowGear, db.ItemTypePetEgg:
		return storageCategoryArmor
	case db.ItemTypeWeapon, db.ItemTypePetArmor:
		return storageCategoryArms
	case db.ItemTypeAmmo:
		return storageCategoryAmmo
	case db.ItemTypeCard:
		return storageCategoryCard
	default:
		return storageCategoryEtc
	}
}
