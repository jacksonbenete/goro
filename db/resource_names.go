package db

var defaultJobResourceNames = map[int]string{
	45:   "WARPNPC",
	46:   "1_ETC_01",
	47:   "1_M_01",
	48:   "1_M_02",
	49:   "1_M_03",
	50:   "1_M_04",
	66:   "1_F_01",
	67:   "1_F_02",
	68:   "1_F_03",
	69:   "1_F_04",
	81:   "4_DOG01",
	82:   "4_KID01",
	83:   "4_M_01",
	84:   "4_M_02",
	85:   "4_M_03",
	86:   "4_M_04",
	1001: "scorpion",
	1002: "poring",
	1004: "hornet",
	1005: "familiar",
	1007: "fabre",
	1008: "pupa",
	1009: "condor",
	1010: "willow",
	1011: "chontchon",
	1013: "wolf",
	1014: "spore",
	1015: "zombie",
	1016: "archer_skeleton",
	1018: "creamie",
	1020: "mandragora",
	1023: "orc_warrior",
	1024: "worm_tail",
	1025: "snake",
	1026: "munak",
	1028: "soldier_skeleton",
	111:  "HIDDEN_NPC",
	844:  "CLEAR_NPC",
	1911: "OBJ_NEUTRAL",
	1912: "OBJ_FLAG_A",
	1913: "OBJ_FLAG_B",
}

var defaultSkillResourceNames = map[int]string{
	1:   "NV_BASIC",
	2:   "SM_SWORD",
	3:   "SM_TWOHAND",
	4:   "SM_RECOVERY",
	5:   "SM_BASH",
	6:   "SM_PROVOKE",
	7:   "SM_MAGNUM",
	8:   "SM_ENDURE",
	9:   "MG_SRECOVERY",
	10:  "MG_SIGHT",
	11:  "MG_NAPALMBEAT",
	12:  "MG_SAFETYWALL",
	13:  "MG_SOULSTRIKE",
	14:  "MG_COLDBOLT",
	15:  "MG_FROSTDIVER",
	16:  "MG_STONECURSE",
	17:  "MG_FIREBALL",
	18:  "MG_FIREWALL",
	19:  "MG_FIREBOLT",
	20:  "MG_LIGHTNINGBOLT",
	21:  "MG_THUNDERSTORM",
	22:  "AL_DP",
	23:  "AL_DEMONBANE",
	24:  "AL_RUWACH",
	25:  "AL_PNEUMA",
	26:  "AL_TELEPORT",
	27:  "AL_WARP",
	28:  "AL_HEAL",
	29:  "AL_INCAGI",
	30:  "AL_DECAGI",
	31:  "AL_HOLYWATER",
	32:  "AL_CRUCIS",
	33:  "AL_ANGELUS",
	34:  "AL_BLESSING",
	35:  "AL_CURE",
	142: "NV_FIRSTAID",
	143: "NV_TRICKDEAD",
}

func JobResourceNames() map[int]string {
	return copyStringMap(defaultJobResourceNames)
}

func SkillResourceNames() map[int]string {
	return copyStringMap(defaultSkillResourceNames)
}

func copyStringMap(in map[int]string) map[int]string {
	out := make(map[int]string, len(in))
	for id, name := range in {
		out[id] = name
	}
	return out
}
