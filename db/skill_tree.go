package db

type SkillRequirement struct {
	SkillID uint16
	Level   int
}

var SkillRequirements = map[uint16][]SkillRequirement{
	SkillSMTwohand:       {{SkillID: SkillSMSword, Level: 1}},
	SkillSMMagnum:        {{SkillID: SkillSMBash, Level: 5}},
	SkillSMEndure:        {{SkillID: SkillSMProvoke, Level: 5}},
	SkillMGSafetywall:    {{SkillID: SkillMGNapalmbeat, Level: 7}, {SkillID: SkillMGSoulstrike, Level: 5}},
	SkillMGSoulstrike:    {{SkillID: SkillMGNapalmbeat, Level: 4}},
	SkillMGFrostdiver:    {{SkillID: SkillMGColdbolt, Level: 5}},
	SkillMGFireball:      {{SkillID: SkillMGFirebolt, Level: 4}},
	SkillMGFirewall:      {{SkillID: SkillMGSight, Level: 1}, {SkillID: SkillMGFireball, Level: 5}},
	SkillMGThunderstorm:  {{SkillID: SkillMGLightningbolt, Level: 4}},
	SkillALPneuma:        {{SkillID: SkillALWarp, Level: 4}},
	SkillALTeleport:      {{SkillID: SkillALRuwach, Level: 1}},
	SkillALWarp:          {{SkillID: SkillALTeleport, Level: 2}},
	SkillALIncagi:        {{SkillID: SkillALHeal, Level: 3}},
	SkillALDecagi:        {{SkillID: SkillALIncagi, Level: 1}},
	SkillALCrucis:        {{SkillID: SkillALDemonbane, Level: 3}},
	SkillALAngelus:       {{SkillID: SkillALDp, Level: 3}},
	SkillALBlessing:      {{SkillID: SkillALDp, Level: 5}},
	SkillALCure:          {{SkillID: SkillALHeal, Level: 2}},
	SkillALDemonbane:     {{SkillID: SkillALDp, Level: 3}},
	SkillMCDiscount:      {{SkillID: SkillMCInccarry, Level: 3}},
	SkillMCOvercharge:    {{SkillID: SkillMCDiscount, Level: 3}},
	SkillMCPushcart:      {{SkillID: SkillMCInccarry, Level: 5}},
	SkillMCVending:       {{SkillID: SkillMCPushcart, Level: 3}},
	SkillACVulture:       {{SkillID: SkillACOwl, Level: 3}},
	SkillACConcentration: {{SkillID: SkillACVulture, Level: 1}},
	SkillACShower:        {{SkillID: SkillACDouble, Level: 5}},
	SkillTFHiding:        {{SkillID: SkillTFSteal, Level: 5}},
	SkillTFDetoxify:      {{SkillID: SkillTFPoison, Level: 3}},
}

var SkillMaxLevels = map[uint16]int{
	SkillNVBasic:          9,
	SkillNVFirstaid:       1,
	SkillNVTrickdead:      1,
	SkillSMSword:          10,
	SkillSMTwohand:        10,
	SkillSMRecovery:       10,
	SkillSMBash:           10,
	SkillSMProvoke:        10,
	SkillSMMagnum:         10,
	SkillSMEndure:         10,
	SkillSMMovingrecovery: 1,
	SkillSMFatalblow:      1,
	SkillSMAutoberserk:    1,
	SkillMGSrecovery:      10,
	SkillMGSight:          1,
	SkillMGNapalmbeat:     10,
	SkillMGSafetywall:     10,
	SkillMGSoulstrike:     10,
	SkillMGColdbolt:       10,
	SkillMGFrostdiver:     10,
	SkillMGStonecurse:     10,
	SkillMGFireball:       10,
	SkillMGFirewall:       10,
	SkillMGFirebolt:       10,
	SkillMGLightningbolt:  10,
	SkillMGThunderstorm:   10,
	SkillMGEnergycoat:     1,
	SkillALRuwach:         1,
	SkillALPneuma:         1,
	SkillALTeleport:       2,
	SkillALWarp:           4,
	SkillALHeal:           10,
	SkillALIncagi:         10,
	SkillALDecagi:         10,
	SkillALHolywater:      1,
	SkillALCrucis:         10,
	SkillALAngelus:        10,
	SkillALBlessing:       10,
	SkillALCure:           1,
	SkillALDp:             10,
	SkillALDemonbane:      10,
	SkillALHolylight:      1,
	SkillMCInccarry:       10,
	SkillMCDiscount:       10,
	SkillMCOvercharge:     10,
	SkillMCPushcart:       10,
	SkillMCIdentify:       1,
	SkillMCVending:        10,
	SkillMCMammonite:      10,
	SkillMCCartrevolution: 1,
	SkillMCChangecart:     1,
	SkillMCLoud:           1,
	SkillMCCartdecorate:   1,
	SkillACOwl:            10,
	SkillACVulture:        10,
	SkillACConcentration:  10,
	SkillACDouble:         10,
	SkillACShower:         10,
	SkillACMakingarrow:    1,
	SkillACChargearrow:    1,
	SkillTFDouble:         10,
	SkillTFMiss:           10,
	SkillTFSteal:          10,
	SkillTFHiding:         10,
	SkillTFPoison:         10,
	SkillTFDetoxify:       1,
	SkillTFSprinklesand:   1,
	SkillTFBacksliding:    1,
	SkillTFPickstone:      1,
	SkillTFThrowstone:     1,
}

var superNoviceSkillTree = []uint16{
	SkillSMSword, SkillSMBash, SkillSMProvoke, SkillTFDouble, SkillTFSteal, SkillTFPoison,
	SkillSMRecovery, SkillSMMagnum, SkillSMEndure, SkillTFMiss, SkillTFHiding, SkillTFDetoxify,
	SkillMGStonecurse, SkillMGColdbolt, SkillMGLightningbolt, SkillMGNapalmbeat, SkillMGFirebolt, SkillMGSight,
	SkillMGSrecovery, SkillMGFrostdiver, SkillMGThunderstorm, SkillMGSoulstrike, SkillMGFireball,
	SkillALRuwach, SkillALHeal, SkillALHolywater, SkillALDp, SkillMGSafetywall, SkillMGFirewall,
	SkillACOwl, SkillALTeleport, SkillALCure, SkillALIncagi, SkillALBlessing, SkillALDemonbane, SkillALAngelus,
	SkillACVulture, SkillALWarp, SkillMCInccarry, SkillALDecagi, SkillMCIdentify, SkillALCrucis,
	SkillMCMammonite, SkillACConcentration, SkillALPneuma, SkillMCDiscount, SkillMCOvercharge,
	SkillMCPushcart, SkillMCVending,
}

var skillTreeByJob = map[int][]uint16{
	JobNovice:       {SkillNVBasic, SkillNVFirstaid, SkillNVTrickdead},
	JobSwordman:     {SkillSMSword, SkillSMRecovery, SkillSMBash, SkillSMProvoke, SkillSMAutoberserk, SkillSMMovingrecovery, SkillSMTwohand, SkillSMMagnum, SkillSMEndure, SkillSMFatalblow},
	JobMagician:     {SkillMGStonecurse, SkillMGColdbolt, SkillMGLightningbolt, SkillMGNapalmbeat, SkillMGFirebolt, SkillMGSight, SkillMGSrecovery, SkillMGFrostdiver, SkillMGThunderstorm, SkillMGSoulstrike, SkillMGFireball, SkillMGEnergycoat, SkillMGSafetywall, SkillMGFirewall},
	JobArcher:       {SkillACDouble, SkillACOwl, SkillACChargearrow, SkillACShower, SkillACVulture, SkillACMakingarrow, SkillACConcentration},
	JobAcolyte:      {SkillALRuwach, SkillALHeal, SkillALHolywater, SkillALDp, SkillALHolylight, SkillALTeleport, SkillALCure, SkillALIncagi, SkillALBlessing, SkillALDemonbane, SkillALAngelus, SkillALWarp, SkillALDecagi, SkillALCrucis, SkillALPneuma},
	JobMerchant:     {SkillMCInccarry, SkillMCMammonite, SkillMCIdentify, SkillMCLoud, SkillMCDiscount, SkillMCPushcart, SkillMCChangecart, SkillMCCartdecorate, SkillMCOvercharge, SkillMCVending, SkillMCCartrevolution},
	JobThief:        {SkillTFDouble, SkillTFSteal, SkillTFPoison, SkillTFSprinklesand, SkillTFThrowstone, SkillTFMiss, SkillTFHiding, SkillTFDetoxify, SkillTFBacksliding, SkillTFPickstone},
	JobSuperNovice:  superNoviceSkillTree,
	JobSuperNoviceB: superNoviceSkillTree,
}

func SkillTreeSkillIDs(job int) []uint16 {
	out := append([]uint16(nil), skillTreeByJob[JobNovice]...)
	if job != JobNovice {
		out = append(out, skillTreeByJob[job]...)
	}
	return out
}

func SkillMaxLevel(skillID uint16) (int, bool) {
	level, ok := SkillMaxLevels[skillID]
	return level, ok && level > 0
}
