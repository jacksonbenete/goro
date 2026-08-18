package db

import "fmt"

const (
	JobNovice      = 0
	JobSwordman    = 1
	JobMagician    = 2
	JobArcher      = 3
	JobAcolyte     = 4
	JobMerchant    = 5
	JobThief       = 6
	JobKnight      = 7
	JobPriest      = 8
	JobWizard      = 9
	JobBlacksmith  = 10
	JobHunter      = 11
	JobAssassin    = 12
	JobKnight2     = 13
	JobCrusader    = 14
	JobMonk        = 15
	JobSage        = 16
	JobRogue       = 17
	JobAlchemist   = 18
	JobBard        = 19
	JobDancer      = 20
	JobCrusader2   = 21
	JobMarried     = 22
	JobSuperNovice = 23
	JobGunslinger  = 24
	JobNinja       = 25
	JobXmas        = 26
	JobSummer      = 27

	JobNoviceH     = 4001
	JobSwordmanH   = 4002
	JobMagicianH   = 4003
	JobArcherH     = 4004
	JobAcolyteH    = 4005
	JobMerchantH   = 4006
	JobThiefH      = 4007
	JobKnightH     = 4008
	JobPriestH     = 4009
	JobWizardH     = 4010
	JobBlacksmithH = 4011
	JobHunterH     = 4012
	JobAssassinH   = 4013
	JobKnight2H    = 4014
	JobCrusaderH   = 4015
	JobMonkH       = 4016
	JobSageH       = 4017
	JobRogueH      = 4018
	JobAlchemistH  = 4019
	JobBardH       = 4020
	JobDancerH     = 4021
	JobCrusader2H  = 4022

	JobNoviceB      = 4023
	JobSwordmanB    = 4024
	JobMagicianB    = 4025
	JobArcherB      = 4026
	JobAcolyteB     = 4027
	JobMerchantB    = 4028
	JobThiefB       = 4029
	JobKnightB      = 4030
	JobPriestB      = 4031
	JobWizardB      = 4032
	JobBlacksmithB  = 4033
	JobHunterB      = 4034
	JobAssassinB    = 4035
	JobKnight2B     = 4036
	JobCrusaderB    = 4037
	JobMonkB        = 4038
	JobSageB        = 4039
	JobRogueB       = 4040
	JobAlchemistB   = 4041
	JobBardB        = 4042
	JobDancerB      = 4043
	JobCrusader2B   = 4044
	JobSuperNoviceB = 4045

	JobTaekwon = 4046
	JobStar    = 4047
	JobStar2   = 4048
	JobLinker  = 4049

	JobRuneKnight      = 4054
	JobWarlock         = 4055
	JobRanger          = 4056
	JobArchbishop      = 4057
	JobMechanic        = 4058
	JobGuillotine      = 4059
	JobRoyalGuard      = 4066
	JobSorcerer        = 4067
	JobMinstrel        = 4068
	JobWanderer        = 4069
	JobSura            = 4070
	JobGenetic         = 4071
	JobShadowChaser    = 4072
	JobRuneKnight2     = 4080
	JobRoyalGuard2     = 4082
	JobRanger2         = 4084
	JobMechanic2       = 4086
	JobKagerou         = 4211
	JobOboro           = 4212
	JobRebellion       = 4215
	JobSummoner        = 4218
	JobNinjaB          = 4222
	JobTaekwonB        = 4225
	JobStarB           = 4226
	JobLinkerB         = 4227
	JobGunslingerB     = 4228
	JobStar2B          = 4238
	JobStarEmperor     = 4239
	JobSoulReaper      = 4240
	JobDragonKnight    = 4252
	JobMeister         = 4253
	JobShadowCross     = 4254
	JobArchMage        = 4255
	JobCardinal        = 4256
	JobWindhawk        = 4257
	JobImperialGuard   = 4258
	JobBiolo           = 4259
	JobAbyssChaser     = 4260
	JobElementalMaster = 4261
	JobInquisitor      = 4262
	JobTroubadour      = 4263
	JobTrouvere        = 4264
)

var jobDisplayNames = map[int]string{
	0:  "Novice",
	1:  "Swordman",
	2:  "Magician",
	3:  "Archer",
	4:  "Acolyte",
	5:  "Merchant",
	6:  "Thief",
	7:  "Knight",
	8:  "Priest",
	9:  "Wizard",
	10: "Blacksmith",
	11: "Hunter",
	12: "Assassin",
	13: "Peco Knight",
	14: "Crusader",
	15: "Monk",
	16: "Sage",
	17: "Rogue",
	18: "Alchemist",
	19: "Bard",
	20: "Dancer",
	21: "Peco Crusader",
	22: "Married",
	23: "Super Novice",
	24: "Gunslinger",
	25: "Ninja",
	26: "Santa",
	27: "Summer",

	4001: "High Novice",
	4002: "High Swordman",
	4003: "High Magician",
	4004: "High Archer",
	4005: "High Acolyte",
	4006: "High Merchant",
	4007: "High Thief",
	4008: "Lord Knight",
	4009: "High Priest",
	4010: "High Wizard",
	4011: "Whitesmith",
	4012: "Sniper",
	4013: "Assassin Cross",
	4014: "Peco Lord Knight",
	4015: "Paladin",
	4016: "Champion",
	4017: "Professor",
	4018: "Stalker",
	4019: "Creator",
	4020: "Clown",
	4021: "Gypsy",
	4022: "Peco Paladin",
	4023: "Baby Novice",
	4024: "Baby Swordman",
	4025: "Baby Magician",
	4026: "Baby Archer",
	4027: "Baby Acolyte",
	4028: "Baby Merchant",
	4029: "Baby Thief",
	4030: "Baby Knight",
	4031: "Baby Priest",
	4032: "Baby Wizard",
	4033: "Baby Blacksmith",
	4034: "Baby Hunter",
	4035: "Baby Assassin",
	4036: "Baby Peco Knight",
	4037: "Baby Crusader",
	4038: "Baby Monk",
	4039: "Baby Sage",
	4040: "Baby Rogue",
	4041: "Baby Alchemist",
	4042: "Baby Bard",
	4043: "Baby Dancer",
	4044: "Baby Peco Crusader",
	4045: "Baby Super Novice",
	4046: "Taekwon",
	4047: "Star Gladiator",
	4048: "Star Gladiator",
	4049: "Soul Linker",
	4050: "Gangsi",
	4051: "Death Knight",
	4052: "Dark Collector",

	4054: "Rune Knight",
	4055: "Warlock",
	4056: "Ranger",
	4057: "Arch Bishop",
	4058: "Mechanic",
	4059: "Guillotine Cross",
	4060: "Rune Knight",
	4061: "Warlock",
	4062: "Ranger",
	4063: "Arch Bishop",
	4064: "Mechanic",
	4065: "Guillotine Cross",
	4066: "Royal Guard",
	4067: "Sorcerer",
	4068: "Minstrel",
	4069: "Wanderer",
	4070: "Sura",
	4071: "Genetic",
	4072: "Shadow Chaser",
	4073: "Royal Guard",
	4074: "Sorcerer",
	4075: "Minstrel",
	4076: "Wanderer",
	4077: "Sura",
	4078: "Genetic",
	4079: "Shadow Chaser",
	4080: "Dragon Rune Knight",
	4081: "Dragon Rune Knight",
	4082: "Gryphon Royal Guard",
	4083: "Gryphon Royal Guard",
	4084: "Wug Ranger",
	4085: "Wug Ranger",
	4086: "Mado Mechanic",
	4087: "Mado Mechanic",

	4096: "Baby Rune Knight",
	4097: "Baby Warlock",
	4098: "Baby Ranger",
	4099: "Baby Arch Bishop",
	4100: "Baby Mechanic",
	4101: "Baby Guillotine Cross",
	4102: "Baby Royal Guard",
	4103: "Baby Sorcerer",
	4104: "Baby Minstrel",
	4105: "Baby Wanderer",
	4106: "Baby Sura",
	4107: "Baby Genetic",
	4108: "Baby Shadow Chaser",
	4109: "Baby Dragon Rune Knight",
	4110: "Baby Gryphon Royal Guard",
	4111: "Baby Wug Ranger",
	4112: "Baby Mado Mechanic",

	4211: "Kagerou",
	4212: "Oboro",
	4215: "Rebellion",
	4218: "Summoner",
	4220: "Baby Summoner",
	4222: "Baby Ninja",
	4223: "Baby Kagerou",
	4224: "Baby Oboro",
	4225: "Baby Taekwon",
	4226: "Baby Star Gladiator",
	4227: "Baby Soul Linker",
	4228: "Baby Gunslinger",
	4229: "Baby Rebellion",
	4238: "Baby Star Gladiator",
	4239: "Star Emperor",
	4240: "Soul Reaper",
	4241: "Baby Star Emperor",
	4242: "Baby Soul Reaper",

	4252: "Dragon Knight",
	4253: "Meister",
	4254: "Shadow Cross",
	4255: "Arch Mage",
	4256: "Cardinal",
	4257: "Wind Hawk",
	4258: "Imperial Guard",
	4259: "Biolo",
	4260: "Abyss Chaser",
	4261: "Elemental Master",
	4262: "Inquisitor",
	4263: "Troubadour",
	4264: "Trouvere",
}

func JobDisplayName(id int) string {
	if name, ok := jobDisplayNames[id]; ok {
		return name
	}
	return fmt.Sprintf("Job %d", id)
}

var JobResourceName = map[int]string{
	JobNovice:      "초보자",
	JobSwordman:    "검사",
	JobMagician:    "마법사",
	JobArcher:      "궁수",
	JobAcolyte:     "성직자",
	JobMerchant:    "상인",
	JobThief:       "도둑",
	JobKnight:      "기사",
	JobPriest:      "프리스트",
	JobWizard:      "위저드",
	JobBlacksmith:  "제철공",
	JobHunter:      "헌터",
	JobAssassin:    "어세신",
	JobKnight2:     "페코페코_기사",
	JobCrusader:    "크루세이더",
	JobMonk:        "몽크",
	JobSage:        "세이지",
	JobRogue:       "로그",
	JobAlchemist:   "연금술사",
	JobBard:        "바드",
	JobDancer:      "무희",
	JobCrusader2:   "신페코크루세이더",
	JobSuperNovice: "슈퍼노비스",
	JobGunslinger:  "건너",
	JobNinja:       "닌자",
	JobTaekwon:     "태권소년",
	JobStar:        "권성",
	JobStar2:       "권성융합",
	JobLinker:      "소울링커",
	JobMarried:     "결혼",
	JobXmas:        "산타",
	JobSummer:      "여름",
	JobKnightH:     "로드나이트",
	JobPriestH:     "하이프리",
	JobWizardH:     "하이위저드",
	JobBlacksmithH: "화이트스미스",
	JobHunterH:     "스나이퍼",
	JobAssassinH:   "어쌔신크로스",
	JobKnight2H:    "로드페코",
	JobCrusaderH:   "팔라딘",
	JobMonkH:       "챔피온",
	JobSageH:       "프로페서",
	JobRogueH:      "스토커",
	JobAlchemistH:  "크리에이터",
	JobBardH:       "클라운",
	JobDancerH:     "집시",
	JobCrusader2H:  "페코팔라딘",
}

func init() {
	duplicateJobResourceName(JobNovice, JobNoviceH, JobNoviceB)
	duplicateJobResourceName(JobSwordman, JobSwordmanH, JobSwordmanB)
	duplicateJobResourceName(JobMagician, JobMagicianH, JobMagicianB)
	duplicateJobResourceName(JobArcher, JobArcherH, JobArcherB)
	duplicateJobResourceName(JobAcolyte, JobAcolyteH, JobAcolyteB)
	duplicateJobResourceName(JobMerchant, JobMerchantH, JobMerchantB)
	duplicateJobResourceName(JobThief, JobThiefH, JobThiefB)
	duplicateJobResourceName(JobKnight, JobKnightB)
	duplicateJobResourceName(JobKnight2, JobKnight2B)
	duplicateJobResourceName(JobPriest, JobPriestB)
	duplicateJobResourceName(JobWizard, JobWizardB)
	duplicateJobResourceName(JobBlacksmith, JobBlacksmithB)
	duplicateJobResourceName(JobHunter, JobHunterB)
	duplicateJobResourceName(JobAssassin, JobAssassinB)
	duplicateJobResourceName(JobCrusader, JobCrusaderB)
	duplicateJobResourceName(JobCrusader2, JobCrusader2B)
	duplicateJobResourceName(JobMonk, JobMonkB)
	duplicateJobResourceName(JobSage, JobSageB)
	duplicateJobResourceName(JobRogue, JobRogueB)
	duplicateJobResourceName(JobAlchemist, JobAlchemistB)
	duplicateJobResourceName(JobBard, JobBardB)
	duplicateJobResourceName(JobDancer, JobDancerB)
	duplicateJobResourceName(JobGunslinger, JobGunslingerB)
	duplicateJobResourceName(JobNinja, JobNinjaB)
	duplicateJobResourceName(JobTaekwon, JobTaekwonB)
	duplicateJobResourceName(JobStar, JobStarB)
	duplicateJobResourceName(JobStar2, JobStar2B)
	duplicateJobResourceName(JobLinker, JobLinkerB)
}

func duplicateJobResourceName(origin int, jobs ...int) {
	value := JobResourceName[origin]
	for _, job := range jobs {
		JobResourceName[job] = value
	}
}

func JobSpriteResourceName(id int) (string, bool) {
	name, ok := JobResourceName[id]
	return name, ok && name != ""
}
