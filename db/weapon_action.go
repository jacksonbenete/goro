package db

const (
	PlayerWeaponActionAttack1 = iota
	PlayerWeaponActionAttack2
	PlayerWeaponActionAttack3
)

type playerWeaponActionTable struct {
	common map[int]int
	bySex  [2]map[int]int
}

// PlayerWeaponAction mirrors ROBrowser DB/Jobs/WeaponAction.js. The return
// value is the index into the PC attack actions: ATTACK1, ATTACK2, ATTACK3.
func PlayerWeaponAction(job int, sex byte, weaponValue int) int {
	table, ok := playerWeaponActions[job]
	if !ok {
		return PlayerWeaponActionAttack1
	}
	weaponType := PlayerWeaponType(weaponValue)
	if table.bySex[0] != nil || table.bySex[1] != nil {
		sexIndex := 0
		if sex != 0 {
			sexIndex = 1
		}
		if action, ok := table.bySex[sexIndex][weaponType]; ok {
			return action
		}
		return PlayerWeaponActionAttack1
	}
	if action, ok := table.common[weaponType]; ok {
		return action
	}
	return PlayerWeaponActionAttack1
}

var playerWeaponActions = func() map[int]playerWeaponActionTable {
	actions := map[int]playerWeaponActionTable{
		JobNovice: sexWeaponAction(
			mergeWeaponActions(
				weaponActions(PlayerWeaponActionAttack2, WeaponRod, WeaponTwoHandRod, WeaponSword, WeaponTwoHandSword, WeaponAxe, WeaponTwoHandAxe, WeaponMace, WeaponTwoHandMace),
				weaponActions(PlayerWeaponActionAttack3, WeaponShortsword),
			),
			mergeWeaponActions(
				weaponActions(PlayerWeaponActionAttack2, WeaponShortsword),
				weaponActions(PlayerWeaponActionAttack3, WeaponRod, WeaponTwoHandRod, WeaponSword, WeaponTwoHandSword, WeaponAxe, WeaponTwoHandAxe, WeaponMace, WeaponTwoHandMace),
			),
		),
		JobSwordman: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword, WeaponSword, WeaponTwoHandSword, WeaponAxe, WeaponTwoHandAxe, WeaponMace, WeaponTwoHandMace),
			weaponActions(PlayerWeaponActionAttack3, WeaponSpear, WeaponTwoHandSpear),
		)),
		JobMagician: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponRod, WeaponTwoHandRod),
			weaponActions(PlayerWeaponActionAttack3, WeaponShortsword),
		)),
		JobArcher: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponBow),
			weaponActions(PlayerWeaponActionAttack3, WeaponShortsword),
		)),
		JobAcolyte: commonWeaponAction(weaponActions(PlayerWeaponActionAttack2, WeaponRod, WeaponTwoHandRod, WeaponMace, WeaponTwoHandMace)),
		JobMerchant: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponMace, WeaponTwoHandMace, WeaponAxe, WeaponTwoHandAxe, WeaponSword, WeaponTwoHandSword),
			weaponActions(PlayerWeaponActionAttack3, WeaponShortsword),
		)),
		JobThief: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponSword, WeaponTwoHandSword, WeaponShortsword),
			weaponActions(PlayerWeaponActionAttack3, WeaponBow),
		)),
		JobKnight: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword, WeaponSword, WeaponTwoHandSword, WeaponTwoHandMace, WeaponAxe, WeaponTwoHandAxe, WeaponMace),
			weaponActions(PlayerWeaponActionAttack3, WeaponSpear, WeaponTwoHandSpear),
		)),
		JobPriest: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponRod, WeaponTwoHandRod, WeaponMace, WeaponTwoHandMace),
			weaponActions(PlayerWeaponActionAttack3, WeaponBook),
		)),
		JobWizard: sexWeaponAction(
			mergeWeaponActions(
				weaponActions(PlayerWeaponActionAttack2, WeaponShortsword),
				weaponActions(PlayerWeaponActionAttack3, WeaponRod, WeaponTwoHandRod),
			),
			mergeWeaponActions(
				weaponActions(PlayerWeaponActionAttack2, WeaponRod, WeaponTwoHandRod),
				weaponActions(PlayerWeaponActionAttack3, WeaponShortsword),
			),
		),
		JobBlacksmith: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword),
			weaponActions(PlayerWeaponActionAttack3, WeaponSword, WeaponTwoHandSword, WeaponAxe, WeaponTwoHandAxe, WeaponMace, WeaponTwoHandMace),
		)),
		JobHunter: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword),
			weaponActions(PlayerWeaponActionAttack3, WeaponBow),
		)),
		JobAssassin: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponAxe, WeaponSword, WeaponShortsword),
			weaponActions(PlayerWeaponActionAttack3, WeaponShortswordShortsword, WeaponSwordSword, WeaponAxeAxe, WeaponShortswordSword, WeaponShortswordAxe, WeaponSwordAxe, WeaponKatar),
		)),
		JobKnight2: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword, WeaponSword, WeaponTwoHandSword, WeaponAxe, WeaponTwoHandAxe, WeaponMace, WeaponTwoHandMace),
			weaponActions(PlayerWeaponActionAttack3, WeaponSpear, WeaponTwoHandSpear),
		)),
		JobCrusader: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword, WeaponSword, WeaponTwoHandSword, WeaponAxe, WeaponTwoHandAxe, WeaponMace, WeaponTwoHandMace),
			weaponActions(PlayerWeaponActionAttack3, WeaponSpear, WeaponTwoHandSpear),
		)),
		JobMonk: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponRod, WeaponTwoHandRod, WeaponMace, WeaponTwoHandMace),
			weaponActions(PlayerWeaponActionAttack3, WeaponKnuckle),
		)),
		JobSage: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword),
			weaponActions(PlayerWeaponActionAttack3, WeaponRod, WeaponTwoHandRod, WeaponBook),
		)),
		JobRogue: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponSword, WeaponTwoHandSword, WeaponShortsword),
			weaponActions(PlayerWeaponActionAttack3, WeaponBow),
		)),
		JobAlchemist: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword),
			weaponActions(PlayerWeaponActionAttack3, WeaponTwoHandSword, WeaponSword, WeaponAxe, WeaponTwoHandAxe, WeaponMace, WeaponTwoHandMace),
		)),
		JobBard: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword, WeaponInstrument),
			weaponActions(PlayerWeaponActionAttack3, WeaponBow),
		)),
		JobDancer: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponWhip),
			weaponActions(PlayerWeaponActionAttack3, WeaponBow),
		)),
		JobCrusader2: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword, WeaponSword, WeaponTwoHandSword, WeaponAxe, WeaponTwoHandAxe, WeaponMace),
			weaponActions(PlayerWeaponActionAttack3, WeaponSpear, WeaponTwoHandSpear),
		)),
		JobSuperNovice: sexWeaponAction(
			mergeWeaponActions(
				weaponActions(PlayerWeaponActionAttack2, WeaponRod, WeaponTwoHandRod, WeaponAxe, WeaponTwoHandAxe, WeaponMace, WeaponTwoHandMace, WeaponSword),
				weaponActions(PlayerWeaponActionAttack3, WeaponShortsword),
			),
			mergeWeaponActions(
				weaponActions(PlayerWeaponActionAttack2, WeaponShortsword),
				weaponActions(PlayerWeaponActionAttack3, WeaponRod, WeaponTwoHandRod, WeaponAxe, WeaponTwoHandAxe, WeaponMace, WeaponTwoHandMace, WeaponSword),
			),
		),
		JobGunslinger: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponNone, WeaponGunHandgun, WeaponGunShotgun),
			weaponActions(PlayerWeaponActionAttack3, WeaponGunGatling, WeaponGunRifle, WeaponGunGrenade),
		)),
		JobNinja: commonWeaponAction(mergeWeaponActions(
			weaponActions(PlayerWeaponActionAttack2, WeaponShortsword),
			weaponActions(PlayerWeaponActionAttack3, WeaponShuriken),
		)),
		JobLinker: sexWeaponAction(
			mergeWeaponActions(
				weaponActions(PlayerWeaponActionAttack2, WeaponShortsword),
				weaponActions(PlayerWeaponActionAttack3, WeaponRod, WeaponTwoHandRod),
			),
			mergeWeaponActions(
				weaponActions(PlayerWeaponActionAttack2, WeaponRod, WeaponTwoHandRod),
				weaponActions(PlayerWeaponActionAttack3, WeaponShortsword),
			),
		),
	}

	duplicatePlayerWeaponAction(actions, JobNovice, JobNoviceH, JobNoviceB)
	duplicatePlayerWeaponAction(actions, JobSwordman, JobSwordmanH, JobSwordmanB)
	duplicatePlayerWeaponAction(actions, JobMagician, JobMagicianH, JobMagicianB)
	duplicatePlayerWeaponAction(actions, JobArcher, JobArcherH, JobArcherB)
	duplicatePlayerWeaponAction(actions, JobAcolyte, JobAcolyteH, JobAcolyteB)
	duplicatePlayerWeaponAction(actions, JobMerchant, JobMerchantH, JobMerchantB)
	duplicatePlayerWeaponAction(actions, JobThief, JobThiefH, JobThiefB)
	duplicatePlayerWeaponAction(actions, JobKnight, JobKnightH, JobKnightB, JobRuneKnight, JobDragonKnight)
	duplicatePlayerWeaponAction(actions, JobKnight2, JobKnight2H, JobKnight2B, JobRuneKnight2)
	duplicatePlayerWeaponAction(actions, JobPriest, JobPriestH, JobPriestB, JobArchbishop, JobCardinal)
	duplicatePlayerWeaponAction(actions, JobWizard, JobWizardH, JobWizardB, JobWarlock, JobArchMage)
	duplicatePlayerWeaponAction(actions, JobBlacksmith, JobBlacksmithH, JobBlacksmithB, JobMechanic, JobMeister)
	duplicatePlayerWeaponAction(actions, JobHunter, JobHunterH, JobHunterB, JobRanger, JobRanger2, JobWindhawk)
	duplicatePlayerWeaponAction(actions, JobAssassin, JobAssassinH, JobAssassinB, JobGuillotine, JobShadowCross)
	duplicatePlayerWeaponAction(actions, JobCrusader, JobCrusaderH, JobCrusaderB, JobRoyalGuard, JobImperialGuard)
	duplicatePlayerWeaponAction(actions, JobCrusader2, JobCrusader2H, JobCrusader2B, JobRoyalGuard2)
	duplicatePlayerWeaponAction(actions, JobMonk, JobMonkH, JobMonkB, JobSura, JobInquisitor)
	duplicatePlayerWeaponAction(actions, JobSage, JobSageH, JobSageB, JobSorcerer, JobElementalMaster)
	duplicatePlayerWeaponAction(actions, JobRogue, JobRogueH, JobRogueB, JobShadowChaser, JobAbyssChaser)
	duplicatePlayerWeaponAction(actions, JobAlchemist, JobAlchemistH, JobAlchemistB, JobGenetic, JobBiolo)
	duplicatePlayerWeaponAction(actions, JobBard, JobBardH, JobBardB, JobMinstrel, JobTroubadour)
	duplicatePlayerWeaponAction(actions, JobDancer, JobDancerH, JobDancerB, JobWanderer, JobTrouvere)
	duplicatePlayerWeaponAction(actions, JobSuperNovice, JobSuperNoviceB)
	duplicatePlayerWeaponAction(actions, JobGunslinger, JobGunslingerB, JobRebellion)
	duplicatePlayerWeaponAction(actions, JobNinja, JobNinjaB, JobKagerou, JobOboro)
	duplicatePlayerWeaponAction(actions, JobLinker, JobSoulReaper)
	return actions
}()

func commonWeaponAction(actions map[int]int) playerWeaponActionTable {
	return playerWeaponActionTable{common: actions}
}

func sexWeaponAction(female, male map[int]int) playerWeaponActionTable {
	return playerWeaponActionTable{bySex: [2]map[int]int{female, male}}
}

func weaponActions(action int, weaponTypes ...int) map[int]int {
	out := make(map[int]int, len(weaponTypes))
	for _, weaponType := range weaponTypes {
		out[weaponType] = action
	}
	return out
}

func mergeWeaponActions(maps ...map[int]int) map[int]int {
	out := map[int]int{}
	for _, values := range maps {
		for weaponType, action := range values {
			out[weaponType] = action
		}
	}
	return out
}

func duplicatePlayerWeaponAction(actions map[int]playerWeaponActionTable, origin int, jobs ...int) {
	value := actions[origin]
	for _, job := range jobs {
		actions[job] = value
	}
}
