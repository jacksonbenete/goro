package game

const inventoryIconSize = 24

const (
	equipLocationWeapon uint16 = 1 << 1
	equipLocationShield uint16 = 1 << 5
)

func inventoryItemTypeIsEquipment(itemType uint8) bool {
	switch itemType {
	case 4, 5, 6, 7, 8, 10:
		return true
	default:
		return false
	}
}
