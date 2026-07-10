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
	JobNovice:      "\xC3\xCA\xBA\xB8\xC0\xDA",
	JobSwordman:    "\xB0\xCB\xBB\xE7",
	JobMagician:    "\xB8\xB6\xB9\xFD\xBB\xE7",
	JobArcher:      "\xB1\xC3\xBC\xF6",
	JobAcolyte:     "\xBC\xBA\xC1\xF7\xC0\xDA",
	JobMerchant:    "\xBB\xF3\xC0\xCE",
	JobThief:       "\xB5\xB5\xB5\xCF",
	JobKnight:      "\xB1\xE2\xBB\xE7",
	JobPriest:      "\xC7\xC1\xB8\xAE\xBD\xBA\xC6\xAE",
	JobWizard:      "\xC0\xA7\xC0\xFA\xB5\xE5",
	JobBlacksmith:  "\xC1\xA6\xC3\xB6\xB0\xF8",
	JobHunter:      "\xC7\xE5\xC5\xCD",
	JobAssassin:    "\xBE\xEE\xBC\xBC\xBD\xC5",
	JobKnight2:     "\xC6\xE4\xC4\xDA\xC6\xE4\xC4\xDA_\xB1\xE2\xBB\xE7",
	JobCrusader:    "\xC5\xA9\xB7\xE7\xBC\xBC\xC0\xCC\xB4\xF5",
	JobMonk:        "\xB8\xF9\xC5\xA9",
	JobSage:        "\xBC\xBC\xC0\xCC\xC1\xF6",
	JobRogue:       "\xB7\xCE\xB1\xD7",
	JobAlchemist:   "\xBF\xAC\xB1\xDD\xBC\xFA\xBB\xE7",
	JobBard:        "\xB9\xD9\xB5\xE5",
	JobDancer:      "\xB9\xAB\xC8\xF1",
	JobCrusader2:   "\xBD\xC5\xC6\xE4\xC4\xDA\xC5\xA9\xB7\xE7\xBC\xBC\xC0\xCC\xB4\xF5",
	JobSuperNovice: "\xBD\xB4\xC6\xDB\xB3\xEB\xBA\xF1\xBD\xBA",
	JobGunslinger:  "\xB0\xC7\xB3\xCA",
	JobNinja:       "\xB4\xD1\xC0\xDA",
	JobTaekwon:     "\xc5\xc2\xb1\xc7\xbc\xd2\xb3\xe2",
	JobStar:        "\xb1\xc7\xbc\xba",
	JobStar2:       "\xb1\xc7\xbc\xba\xc0\xb6\xc7\xd5",
	JobLinker:      "\xbc\xd2\xbf\xef\xb8\xb5\xc4\xbf",
	JobMarried:     "\xB0\xE1\xC8\xA5",
	JobXmas:        "\xBB\xEA\xC5\xB8",
	JobSummer:      "\xBF\xA9\xB8\xA7",
	JobKnightH:     "\xB7\xCE\xB5\xE5\xB3\xAA\xC0\xCC\xC6\xAE",
	JobPriestH:     "\xC7\xCF\xC0\xCC\xC7\xC1\xB8\xAE",
	JobWizardH:     "\xC7\xCF\xC0\xCC\xC0\xA7\xC0\xFA\xB5\xE5",
	JobBlacksmithH: "\xC8\xAD\xC0\xCC\xC6\xAE\xBD\xBA\xB9\xCC\xBD\xBA",
	JobHunterH:     "\xBD\xBA\xB3\xAA\xC0\xCC\xC6\xDB",
	JobAssassinH:   "\xBE\xEE\xBD\xD8\xBD\xC5\xC5\xA9\xB7\xCE\xBD\xBA",
	JobKnight2H:    "\xB7\xCE\xB5\xE5\xC6\xE4\xC4\xDA",
	JobCrusaderH:   "\xC6\xC8\xB6\xF3\xB5\xF2",
	JobMonkH:       "\xC3\xA8\xC7\xC7\xBF\xC2",
	JobSageH:       "\xC7\xC1\xB7\xCE\xC6\xE4\xBC\xAD",
	JobRogueH:      "\xBD\xBA\xC5\xE4\xC4\xBF",
	JobAlchemistH:  "\xC5\xA9\xB8\xAE\xBF\xA1\xC0\xCC\xC5\xCD",
	JobBardH:       "\xC5\xAC\xB6\xF3\xBF\xEE",
	JobDancerH:     "\xC1\xFD\xBD\xC3",
	JobCrusader2H:  "\xC6\xE4\xC4\xDA\xC6\xC8\xB6\xF3\xB5\xF2",
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
