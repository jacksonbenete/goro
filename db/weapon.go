package db

const (
	WeaponNone                 = 0
	WeaponShortsword           = 1
	WeaponSword                = 2
	WeaponTwoHandSword         = 3
	WeaponSpear                = 4
	WeaponTwoHandSpear         = 5
	WeaponAxe                  = 6
	WeaponTwoHandAxe           = 7
	WeaponMace                 = 8
	WeaponTwoHandMace          = 9
	WeaponRod                  = 10
	WeaponBow                  = 11
	WeaponKnuckle              = 12
	WeaponInstrument           = 13
	WeaponWhip                 = 14
	WeaponBook                 = 15
	WeaponKatar                = 16
	WeaponGunHandgun           = 17
	WeaponGunRifle             = 18
	WeaponGunGatling           = 19
	WeaponGunShotgun           = 20
	WeaponGunGrenade           = 21
	WeaponShuriken             = 22
	WeaponTwoHandRod           = 23
	WeaponShortswordShortsword = 25
	WeaponSwordSword           = 26
	WeaponAxeAxe               = 27
	WeaponShortswordSword      = 28
	WeaponShortswordAxe        = 29
	WeaponSwordAxe             = 30
	MaxWeaponType              = 103
)

func PlayerWeaponType(weaponValue int) int {
	if weaponValue <= 0 {
		return WeaponNone
	}
	if weaponType, ok := playerWeaponTypeExpansion[weaponValue]; ok {
		return weaponType
	}
	if weaponValue < MaxWeaponType {
		return weaponValue
	}
	if weaponValue < 1100 {
		return WeaponNone
	}
	if weaponValue >= 1116 && weaponValue <= 1118 {
		return WeaponTwoHandSword
	}
	if weaponValue >= 1314 && weaponValue <= 1315 {
		return WeaponTwoHandAxe
	}
	if weaponValue >= 1410 && weaponValue <= 1412 {
		return WeaponTwoHandSpear
	}
	if weaponValue >= 1472 && weaponValue <= 1473 {
		return WeaponRod
	}
	if weaponValue == 1599 {
		return WeaponMace
	}
	if containsInt(robrGunGatling, weaponValue) {
		return WeaponGunGatling
	}
	if containsInt(robrGunShotgun, weaponValue) {
		return WeaponGunShotgun
	}
	if containsInt(robrGunGrenade, weaponValue) {
		return WeaponGunGrenade
	}
	switch {
	case weaponValue < 1150:
		return WeaponSword
	case weaponValue < 1200:
		return WeaponTwoHandSword
	case weaponValue < 1250:
		return WeaponShortsword
	case weaponValue < 1300:
		return WeaponKatar
	case weaponValue < 1350:
		return WeaponAxe
	case weaponValue < 1400:
		return WeaponTwoHandAxe
	case weaponValue < 1450:
		return WeaponSpear
	case weaponValue < 1500:
		return WeaponTwoHandSpear
	case weaponValue < 1550:
		return WeaponMace
	case weaponValue < 1600:
		return WeaponBook
	case weaponValue < 1700:
		return WeaponRod
	case weaponValue < 1750:
		return WeaponBow
	case weaponValue < 1800:
		return WeaponNone
	case weaponValue < 1850:
		return WeaponKnuckle
	case weaponValue < 1900:
		return WeaponNone
	case weaponValue < 1950:
		return WeaponInstrument
	case weaponValue < 2000:
		return WeaponWhip
	case weaponValue < 2050:
		return WeaponTwoHandRod
	case weaponValue < 13000:
		return WeaponNone
	case weaponValue < 13100:
		return WeaponShortsword
	case weaponValue < 13150:
		return WeaponGunHandgun
	case weaponValue < 13200:
		return WeaponGunRifle
	case weaponValue < 13300:
		return WeaponNone
	case weaponValue < 13400:
		return WeaponShuriken
	case weaponValue < 13500:
		return WeaponSword
	case weaponValue < 18100:
		return WeaponNone
	case weaponValue < 18500:
		return WeaponBow
	case weaponValue < 20000:
		return WeaponNone
	case weaponValue < 21000:
		return WeaponTwoHandRod
	case weaponValue < 22000:
		return WeaponTwoHandSword
	default:
		return WeaponNone
	}
}

var playerWeaponTypeExpansion = map[int]int{
	31:  WeaponShortsword,
	32:  WeaponShortsword,
	33:  WeaponShortsword,
	34:  WeaponShortsword,
	35:  WeaponShortsword,
	36:  WeaponShortsword,
	37:  WeaponShortsword,
	38:  WeaponShortsword,
	39:  WeaponSword,
	40:  WeaponSword,
	41:  WeaponSword,
	42:  WeaponSword,
	43:  WeaponSword,
	44:  WeaponSword,
	45:  WeaponSword,
	46:  WeaponSword,
	47:  WeaponSword,
	48:  WeaponTwoHandSword,
	49:  WeaponTwoHandSword,
	50:  WeaponTwoHandSword,
	51:  WeaponTwoHandSword,
	52:  WeaponSpear,
	53:  WeaponSpear,
	54:  WeaponSpear,
	55:  WeaponSpear,
	56:  WeaponSpear,
	57:  WeaponSpear,
	58:  WeaponAxe,
	59:  WeaponAxe,
	60:  WeaponAxe,
	61:  WeaponAxe,
	62:  WeaponMace,
	63:  WeaponMace,
	64:  WeaponMace,
	65:  WeaponMace,
	66:  WeaponMace,
	67:  WeaponMace,
	68:  WeaponMace,
	69:  WeaponRod,
	70:  WeaponRod,
	71:  WeaponRod,
	72:  WeaponRod,
	73:  WeaponBow,
	74:  WeaponBow,
	75:  WeaponBow,
	76:  WeaponBow,
	77:  WeaponBow,
	78:  WeaponKnuckle,
	79:  WeaponKnuckle,
	80:  WeaponKnuckle,
	81:  WeaponKnuckle,
	82:  WeaponKnuckle,
	83:  WeaponKnuckle,
	84:  WeaponKnuckle,
	85:  WeaponKnuckle,
	86:  WeaponWhip,
	87:  WeaponWhip,
	88:  WeaponWhip,
	89:  WeaponBook,
	90:  WeaponBook,
	91:  WeaponBook,
	92:  WeaponBook,
	93:  WeaponBook,
	94:  WeaponBook,
	95:  WeaponBook,
	96:  WeaponTwoHandRod,
	97:  WeaponTwoHandRod,
	98:  WeaponMace,
	99:  WeaponRod,
	100: WeaponRod,
	101: WeaponRod,
	102: WeaponRod,
}

var (
	robrGunGatling = []int{13157, 13158, 13159, 13172, 13177}
	robrGunShotgun = []int{13154, 13155, 13156, 13167, 13168, 13169, 13173, 13178}
	robrGunGrenade = []int{13160, 13161, 13162, 13174, 13179}
)

func containsInt(values []int, needle int) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
