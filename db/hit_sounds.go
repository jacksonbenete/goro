package db

func WeaponHitSounds(weaponType int) []string {
	switch weaponType {
	case 0:
		return []string{"_hit_fist1.wav", "_hit_fist2.wav", "_hit_fist3.wav", "_hit_fist4.wav"}
	case 1, 2, 3:
		return []string{"_hit_sword.wav"}
	case 4, 5:
		return []string{"_hit_spear.wav"}
	case 6, 7:
		return []string{"_hit_axe.wav"}
	case 8, 9, 13, 14, 15, 16, 22:
		return []string{"_hit_mace.wav"}
	case 10, 23:
		return []string{"_hit_rod.wav"}
	case 11:
		return []string{"_hit_arrow.wav"}
	case 12:
		return []string{"_HIT_FIST2.wav"}
	case 17:
		return []string{"_hit_\xB1\xC7\xC3\xD1.wav"}
	case 18:
		return []string{"_hit_\xB6\xF3\xC0\xCC\xC7\xC3.wav"}
	case 19:
		return []string{"_hit_\xB0\xB3\xC6\xB2\xB8\xB5\xC7\xD1\xB9\xDF.wav"}
	case 20:
		return []string{"_hit_\xBC\xA6\xB0\xC7.wav"}
	case 21:
		return []string{"_hit_\xB1\xD7\xB7\xB9\xB3\xD7\xC0\xCC\xB5\xE5\xB7\xB1\xC3\xC4.wav"}
	case 102:
		return []string{"_hit_fist4.wav"}
	default:
		return []string{"_hit_mace.wav"}
	}
}

func JobHitSounds(job int) []string {
	switch job {
	case 1, 7, 13, 14, 21, 23, 4008, 4015, 4022, 4028, 4036, 4044, 4054, 4066, 4080:
		return []string{"player_metal.wav"}
	case 3, 6, 11, 12, 17, 19, 20, 24, 25, 4047, 4048, 4056, 4057, 4069, 4070, 4083, 4084:
		return []string{"player_wooden_male.wav"}
	default:
		return []string{"player_clothes.wav"}
	}
}
