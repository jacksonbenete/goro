package db

func WeaponHitSounds(weaponType int) []string {
	if sounds, ok := weaponHitSoundTable[weaponType]; ok {
		return sounds
	}
	return weaponHitSoundTable[MaxWeaponType]
}

func JobHitSounds(job int) []string {
	if sounds, ok := jobHitSoundTable[job]; ok {
		return sounds
	}
	return jobHitSoundTable[JobNovice]
}

var weaponHitSoundTable = map[int][]string{
	WeaponNone:         {"_hit_fist1.wav", "_hit_fist2.wav", "_hit_fist3.wav", "_hit_fist4.wav"},
	WeaponShortsword:   {"_hit_sword.wav"},
	WeaponSword:        {"_hit_sword.wav"},
	WeaponTwoHandSword: {"_hit_sword.wav"},
	WeaponSpear:        {"_hit_spear.wav"},
	WeaponTwoHandSpear: {"_hit_spear.wav"},
	WeaponAxe:          {"_hit_axe.wav"},
	WeaponTwoHandAxe:   {"_hit_axe.wav"},
	WeaponMace:         {"_hit_mace.wav"},
	WeaponTwoHandMace:  {"_hit_mace.wav"},
	WeaponRod:          {"_hit_rod.wav"},
	WeaponBow:          {"_hit_arrow.wav"},
	WeaponKnuckle:      {"_HIT_FIST2.wav"},
	WeaponInstrument:   {"_hit_mace.wav"},
	WeaponWhip:         {"_hit_mace.wav"},
	WeaponBook:         {"_hit_mace.wav"},
	WeaponKatar:        {"_hit_mace.wav"},
	WeaponGunHandgun:   {"_hit_\xB1\xC7\xC3\xD1.wav"},
	WeaponGunRifle:     {"_hit_\xB6\xF3\xC0\xCC\xC7\xC3.wav"},
	WeaponGunGatling:   {"_hit_\xB0\xB3\xC6\xB2\xB8\xB5\xC7\xD1\xB9\xDF.wav"},
	WeaponGunShotgun:   {"_hit_\xBC\xA6\xB0\xC7.wav"},
	WeaponGunGrenade:   {"_hit_\xB1\xD7\xB7\xB9\xB3\xD7\xC0\xCC\xB5\xE5\xB7\xB1\xC3\xC4.wav"},
	WeaponShuriken:     {"_hit_mace.wav"},
	WeaponTwoHandRod:   {"_hit_rod.wav"},
	24:                 {"_hit_fist4.wav"},
	25:                 {"_hit_sword.wav"},
	26:                 {"_hit_sword.wav"},
	27:                 {"_hit_axe.wav"},
	28:                 {"_hit_sword.wav"},
	29:                 {"_hit_sword.wav"},
	30:                 {"_hit_sword.wav"},
	MaxWeaponType:      {"_hit_mace.wav"},
}

var jobHitSoundTable = map[int][]string{}

func init() {
	setJobHitSounds([]string{"player_clothes.wav"},
		JobNovice, JobMagician, JobAcolyte, JobMerchant, JobPriest, JobWizard,
		JobBlacksmith, JobSage, JobAlchemist, JobSuperNovice, JobLinker,
		JobMarried, JobXmas, JobSummer,
		JobNoviceH, JobMagicianH, JobAcolyteH, JobMerchantH, JobPriestH,
		JobWizardH, JobBlacksmithH, JobSageH, JobAlchemistH, JobSuperNoviceB,
		JobNoviceB, JobMagicianB, JobAcolyteB, JobMerchantB, JobPriestB,
		JobWizardB, JobBlacksmithB, JobSageB, JobAlchemistB)
	setJobHitSounds([]string{"player_metal.wav"},
		JobSwordman, JobKnight, JobKnight2, JobCrusader, JobMonk, JobCrusader2,
		JobStar, JobStar2, JobSwordmanH, JobKnightH, JobKnight2H, JobCrusaderH,
		JobMonkH, JobCrusader2H, JobSwordmanB, JobKnightB, JobKnight2B,
		JobCrusaderB, JobMonkB, JobCrusader2B)
	setJobHitSounds([]string{"player_wooden_male.wav"},
		JobArcher, JobThief, JobHunter, JobAssassin, JobRogue, JobBard,
		JobDancer, JobGunslinger, JobNinja, JobTaekwon, JobArcherH, JobThiefH,
		JobHunterH, JobAssassinH, JobRogueH, JobBardH, JobDancerH, JobArcherB,
		JobThiefB, JobHunterB, JobAssassinB, JobRogueB, JobBardB, JobDancerB)
}

func setJobHitSounds(sounds []string, jobs ...int) {
	for _, job := range jobs {
		jobHitSoundTable[job] = sounds
	}
}
