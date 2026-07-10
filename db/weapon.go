package db

const MaxWeaponType = 103

func PlayerWeaponType(weaponValue int) int {
	if weaponValue <= 0 {
		return 0
	}
	if weaponType, ok := playerWeaponTypeExpansion[weaponValue]; ok {
		return weaponType
	}
	if weaponValue < MaxWeaponType {
		return weaponValue
	}
	switch {
	case weaponValue >= 1100 && weaponValue <= 1149:
		return 2
	case weaponValue >= 1150 && weaponValue <= 1199:
		return 3
	case weaponValue >= 1200 && weaponValue <= 1249:
		return 1
	case weaponValue >= 1250 && weaponValue <= 1299:
		return 16
	case weaponValue >= 1300 && weaponValue <= 1349:
		return 6
	case weaponValue >= 1350 && weaponValue <= 1399:
		return 7
	case weaponValue >= 1400 && weaponValue <= 1449:
		return 4
	case weaponValue >= 1450 && weaponValue <= 1499:
		return 5
	case weaponValue >= 1500 && weaponValue <= 1549:
		return 8
	case weaponValue >= 1550 && weaponValue <= 1599:
		return 15
	case weaponValue >= 1600 && weaponValue <= 1699:
		return 10
	case weaponValue >= 1700 && weaponValue <= 1749:
		return 11
	case weaponValue >= 1800 && weaponValue <= 1849:
		return 12
	case weaponValue >= 1900 && weaponValue <= 1949:
		return 13
	case weaponValue >= 1950 && weaponValue <= 1999:
		return 14
	case weaponValue >= 13000 && weaponValue <= 13099:
		return 1
	case weaponValue >= 13100 && weaponValue <= 13149:
		return 17
	case weaponValue >= 13150 && weaponValue <= 13199:
		return 18
	case weaponValue >= 13300 && weaponValue <= 13399:
		return 22
	case weaponValue >= 13400 && weaponValue <= 13499:
		return 2
	case weaponValue >= 18100 && weaponValue <= 18499:
		return 11
	case weaponValue >= 20000 && weaponValue <= 20999:
		return 23
	case weaponValue >= 21000 && weaponValue <= 21999:
		return 3
	default:
		return 0
	}
}

var playerWeaponTypeExpansion = map[int]int{
	31:  1,
	32:  1,
	33:  1,
	34:  1,
	35:  1,
	36:  1,
	37:  1,
	38:  1,
	39:  2,
	40:  2,
	41:  2,
	42:  2,
	43:  2,
	44:  2,
	45:  2,
	46:  2,
	47:  2,
	48:  3,
	49:  3,
	50:  3,
	51:  3,
	52:  4,
	53:  4,
	54:  4,
	55:  4,
	56:  4,
	57:  4,
	58:  6,
	59:  6,
	60:  6,
	61:  6,
	62:  8,
	63:  8,
	64:  8,
	65:  8,
	66:  8,
	67:  8,
	68:  8,
	69:  10,
	70:  10,
	71:  10,
	72:  10,
	73:  11,
	74:  11,
	75:  11,
	76:  11,
	77:  11,
	78:  12,
	79:  12,
	80:  12,
	81:  12,
	82:  12,
	83:  12,
	84:  12,
	85:  12,
	86:  14,
	87:  14,
	88:  14,
	89:  15,
	90:  15,
	91:  15,
	92:  15,
	93:  15,
	94:  15,
	95:  15,
	96:  23,
	97:  23,
	98:  8,
	99:  10,
	100: 10,
	101: 10,
}
