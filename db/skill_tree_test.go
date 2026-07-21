package db

import (
	"reflect"
	"testing"
)

func TestWizardSkillTreeIncludesRobrowserBeforeJobs(t *testing.T) {
	wizard := SkillTreeSkillIDs(JobWizard)
	if !containsSkillID(wizard, SkillMGFirebolt) || !containsSkillID(wizard, SkillWZMeteor) {
		t.Fatalf("wizard tree = %v, want magician and wizard skills", wizard)
	}
	if containsSkillID(wizard, SkillHWMagicpower) {
		t.Fatalf("wizard tree = %v, should not include high wizard skills", wizard)
	}

	babyWizard := SkillTreeSkillIDs(JobWizardB)
	if !containsSkillID(babyWizard, SkillWZStormgust) || containsSkillID(babyWizard, SkillHWMagicpower) {
		t.Fatalf("baby wizard tree = %v, want wizard duplicate without high wizard skills", babyWizard)
	}

	highWizard := SkillTreeSkillIDs(JobWizardH)
	if !containsSkillID(highWizard, SkillMGFirebolt) || !containsSkillID(highWizard, SkillWZStormgust) || !containsSkillID(highWizard, SkillHWMagicpower) {
		t.Fatalf("high wizard tree = %v, want magician, wizard, and high wizard skills", highWizard)
	}
}

func TestKnightSkillTreeIncludesRobrowserBeforeJobs(t *testing.T) {
	knight := SkillTreeSkillIDs(JobKnight)
	if !containsSkillID(knight, SkillSMBash) || !containsSkillID(knight, SkillKNPierce) || !containsSkillID(knight, SkillKNChargeatk) {
		t.Fatalf("knight tree = %v, want swordman and knight skills", knight)
	}
	if containsSkillID(knight, SkillLKSpiralpierce) {
		t.Fatalf("knight tree = %v, should not include lord knight skills", knight)
	}

	mountedKnight := SkillTreeSkillIDs(JobKnight2)
	if !containsSkillID(mountedKnight, SkillKNRiding) || !containsSkillID(mountedKnight, SkillKNBrandishspear) {
		t.Fatalf("mounted knight tree = %v, want knight duplicate skills", mountedKnight)
	}

	babyKnight := SkillTreeSkillIDs(JobKnightB)
	if !containsSkillID(babyKnight, SkillKNBowlingbash) || containsSkillID(babyKnight, SkillLKAurablade) {
		t.Fatalf("baby knight tree = %v, want knight duplicate without lord knight skills", babyKnight)
	}

	lordKnight := SkillTreeSkillIDs(JobKnightH)
	if !containsSkillID(lordKnight, SkillSMBash) || !containsSkillID(lordKnight, SkillKNBowlingbash) || !containsSkillID(lordKnight, SkillLKSpiralpierce) {
		t.Fatalf("lord knight tree = %v, want swordman, knight, and lord knight skills", lordKnight)
	}

	mountedLordKnight := SkillTreeSkillIDs(JobKnight2H)
	if !containsSkillID(mountedLordKnight, SkillLKJointbeat) || !containsSkillID(mountedLordKnight, SkillKNOnehand) {
		t.Fatalf("mounted lord knight tree = %v, want lord knight duplicate skills", mountedLordKnight)
	}
}

func TestPriestSkillTreeIncludesRobrowserBeforeJobs(t *testing.T) {
	acolyte := SkillTreeSkillIDs(JobAcolyteH)
	if !containsSkillID(acolyte, SkillALHeal) || !containsSkillID(acolyte, SkillALPneuma) || containsSkillID(acolyte, SkillPRKyrie) {
		t.Fatalf("high acolyte tree = %v, want acolyte skills only", acolyte)
	}

	priest := SkillTreeSkillIDs(JobPriest)
	for _, skillID := range []uint16{
		SkillALHeal,
		SkillPRKyrie,
		SkillMGSrecovery,
		SkillALLResurrection,
		SkillMGSafetywall,
		SkillPRRedemptio,
	} {
		if !containsSkillID(priest, skillID) {
			t.Fatalf("priest tree = %v, missing robr skill %d", priest, skillID)
		}
	}
	if containsSkillID(priest, SkillHPAssumptio) {
		t.Fatalf("priest tree = %v, should not include high priest skills", priest)
	}

	babyPriest := SkillTreeSkillIDs(JobPriestB)
	if !containsSkillID(babyPriest, SkillPRMagnus) || containsSkillID(babyPriest, SkillHPMeditatio) {
		t.Fatalf("baby priest tree = %v, want priest duplicate without high priest skills", babyPriest)
	}

	highPriest := SkillTreeSkillIDs(JobPriestH)
	if !containsSkillID(highPriest, SkillALHeal) || !containsSkillID(highPriest, SkillPRMagnus) || !containsSkillID(highPriest, SkillHPAssumptio) || !containsSkillID(highPriest, SkillHPManarecharge) {
		t.Fatalf("high priest tree = %v, want acolyte, priest, and high priest skills", highPriest)
	}
}

func TestHunterSkillTreeIncludesRobrowserBeforeJobs(t *testing.T) {
	archer := SkillTreeSkillIDs(JobArcherH)
	if !containsSkillID(archer, SkillACDouble) || !containsSkillID(archer, SkillACConcentration) || containsSkillID(archer, SkillHTFalcon) {
		t.Fatalf("high archer tree = %v, want archer skills only", archer)
	}

	hunter := SkillTreeSkillIDs(JobHunter)
	for _, skillID := range []uint16{
		SkillACDouble,
		SkillHTBeastbane,
		SkillHTSkidtrap,
		SkillHTPhantasmic,
		SkillHTClaymoretrap,
	} {
		if !containsSkillID(hunter, skillID) {
			t.Fatalf("hunter tree = %v, missing robr skill %d", hunter, skillID)
		}
	}
	if containsSkillID(hunter, SkillSNSharpshooting) {
		t.Fatalf("hunter tree = %v, should not include sniper skills", hunter)
	}

	babyHunter := SkillTreeSkillIDs(JobHunterB)
	if !containsSkillID(babyHunter, SkillHTBlitzbeat) || containsSkillID(babyHunter, SkillSNWindwalk) {
		t.Fatalf("baby hunter tree = %v, want hunter duplicate without sniper skills", babyHunter)
	}

	sniper := SkillTreeSkillIDs(JobHunterH)
	if !containsSkillID(sniper, SkillACVulture) || !containsSkillID(sniper, SkillHTSteelcrow) || !containsSkillID(sniper, SkillSNFalconassault) || !containsSkillID(sniper, SkillSNWindwalk) {
		t.Fatalf("sniper tree = %v, want archer, hunter, and sniper skills", sniper)
	}
}

func TestAssassinSkillTreeIncludesRobrowserBeforeJobs(t *testing.T) {
	thief := SkillTreeSkillIDs(JobThiefH)
	if !containsSkillID(thief, SkillTFDouble) || !containsSkillID(thief, SkillTFPickstone) || containsSkillID(thief, SkillASSonicblow) {
		t.Fatalf("high thief tree = %v, want thief skills only", thief)
	}

	assassin := SkillTreeSkillIDs(JobAssassin)
	for _, skillID := range []uint16{
		SkillTFDouble,
		SkillASRight,
		SkillASCloaking,
		SkillASVenomknife,
		SkillASSonicaccel,
		SkillASSplasher,
	} {
		if !containsSkillID(assassin, skillID) {
			t.Fatalf("assassin tree = %v, missing robr skill %d", assassin, skillID)
		}
	}
	if containsSkillID(assassin, SkillASCBreaker) {
		t.Fatalf("assassin tree = %v, should not include assassin cross skills", assassin)
	}

	babyAssassin := SkillTreeSkillIDs(JobAssassinB)
	if !containsSkillID(babyAssassin, SkillASGrimtooth) || containsSkillID(babyAssassin, SkillASCEdp) {
		t.Fatalf("baby assassin tree = %v, want assassin duplicate without assassin cross skills", babyAssassin)
	}

	assassinCross := SkillTreeSkillIDs(JobAssassinH)
	if !containsSkillID(assassinCross, SkillTFPoison) || !containsSkillID(assassinCross, SkillASKatar) || !containsSkillID(assassinCross, SkillASCBreaker) || !containsSkillID(assassinCross, SkillASCMeteorassault) {
		t.Fatalf("assassin cross tree = %v, want thief, assassin, and assassin cross skills", assassinCross)
	}
}

func TestWizardSkillRequirementsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		skillID uint16
		want    []SkillRequirement
	}{
		{SkillWZFirepillar, []SkillRequirement{{SkillID: SkillMGFirewall, Level: 1}}},
		{SkillWZSightrasher, []SkillRequirement{{SkillID: SkillMGSight, Level: 1}, {SkillID: SkillMGLightningbolt, Level: 1}}},
		{SkillWZMeteor, []SkillRequirement{{SkillID: SkillMGThunderstorm, Level: 1}, {SkillID: SkillWZSightrasher, Level: 2}}},
		{SkillWZJupitel, []SkillRequirement{{SkillID: SkillMGNapalmbeat, Level: 1}, {SkillID: SkillMGLightningbolt, Level: 1}}},
		{SkillWZVermilion, []SkillRequirement{{SkillID: SkillMGThunderstorm, Level: 1}, {SkillID: SkillWZJupitel, Level: 5}}},
		{SkillWZWaterball, []SkillRequirement{{SkillID: SkillMGColdbolt, Level: 1}, {SkillID: SkillMGLightningbolt, Level: 1}}},
		{SkillWZIcewall, []SkillRequirement{{SkillID: SkillMGStonecurse, Level: 1}, {SkillID: SkillMGFrostdiver, Level: 1}}},
		{SkillWZFrostnova, []SkillRequirement{{SkillID: SkillWZIcewall, Level: 1}}},
		{SkillWZStormgust, []SkillRequirement{{SkillID: SkillMGFrostdiver, Level: 1}, {SkillID: SkillWZJupitel, Level: 3}}},
		{SkillWZEarthspike, []SkillRequirement{{SkillID: SkillMGStonecurse, Level: 1}}},
		{SkillWZHeavendrive, []SkillRequirement{{SkillID: SkillWZEarthspike, Level: 3}}},
		{SkillWZQuagmire, []SkillRequirement{{SkillID: SkillWZHeavendrive, Level: 1}}},
		{SkillHWSouldrain, []SkillRequirement{{SkillID: SkillMGSrecovery, Level: 5}, {SkillID: SkillMGSoulstrike, Level: 7}}},
		{SkillHWMagiccrasher, []SkillRequirement{{SkillID: SkillMGSrecovery, Level: 1}}},
		{SkillHWNapalmvulcan, []SkillRequirement{{SkillID: SkillMGNapalmbeat, Level: 5}}},
		{SkillHWGanbantein, []SkillRequirement{{SkillID: SkillWZEstimation, Level: 1}, {SkillID: SkillWZIcewall, Level: 1}}},
		{SkillHWGravitation, []SkillRequirement{{SkillID: SkillWZQuagmire, Level: 1}, {SkillID: SkillHWMagiccrasher, Level: 1}, {SkillID: SkillHWMagicpower, Level: 10}}},
	} {
		if got := SkillRequirements[tc.skillID]; !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("requirements for skill %d = %+v, want %+v", tc.skillID, got, tc.want)
		}
	}
}

func TestKnightSkillRequirementsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		skillID uint16
		want    []SkillRequirement
	}{
		{SkillKNPierce, []SkillRequirement{{SkillID: SkillKNSpearmastery, Level: 1}}},
		{SkillKNBrandishspear, []SkillRequirement{{SkillID: SkillKNRiding, Level: 1}, {SkillID: SkillKNSpearstab, Level: 3}}},
		{SkillKNSpearstab, []SkillRequirement{{SkillID: SkillKNPierce, Level: 5}}},
		{SkillKNSpearboomerang, []SkillRequirement{{SkillID: SkillKNPierce, Level: 3}}},
		{SkillKNTwohandquicken, []SkillRequirement{{SkillID: SkillSMTwohand, Level: 1}}},
		{SkillKNAutocounter, []SkillRequirement{{SkillID: SkillSMTwohand, Level: 1}}},
		{SkillKNBowlingbash, []SkillRequirement{
			{SkillID: SkillSMBash, Level: 10},
			{SkillID: SkillSMMagnum, Level: 3},
			{SkillID: SkillSMTwohand, Level: 5},
			{SkillID: SkillKNTwohandquicken, Level: 10},
			{SkillID: SkillKNAutocounter, Level: 5},
		}},
		{SkillKNRiding, []SkillRequirement{{SkillID: SkillSMEndure, Level: 1}}},
		{SkillKNCavaliermastery, []SkillRequirement{{SkillID: SkillKNRiding, Level: 1}}},
		{SkillKNOnehand, []SkillRequirement{{SkillID: SkillKNTwohandquicken, Level: 10}}},
		{SkillLKSpiralpierce, []SkillRequirement{
			{SkillID: SkillKNSpearmastery, Level: 5},
			{SkillID: SkillKNPierce, Level: 5},
			{SkillID: SkillKNRiding, Level: 1},
			{SkillID: SkillKNSpearstab, Level: 5},
		}},
		{SkillLKHeadcrush, []SkillRequirement{{SkillID: SkillKNSpearmastery, Level: 9}, {SkillID: SkillKNRiding, Level: 1}}},
		{SkillLKJointbeat, []SkillRequirement{{SkillID: SkillKNCavaliermastery, Level: 3}, {SkillID: SkillLKHeadcrush, Level: 3}}},
		{SkillLKAurablade, []SkillRequirement{{SkillID: SkillSMMagnum, Level: 5}, {SkillID: SkillSMTwohand, Level: 5}}},
		{SkillLKParrying, []SkillRequirement{{SkillID: SkillSMProvoke, Level: 5}, {SkillID: SkillSMTwohand, Level: 10}, {SkillID: SkillKNTwohandquicken, Level: 3}}},
		{SkillLKConcentration, []SkillRequirement{{SkillID: SkillSMRecovery, Level: 5}, {SkillID: SkillKNSpearmastery, Level: 5}, {SkillID: SkillKNRiding, Level: 1}}},
		{SkillLKTensionrelax, []SkillRequirement{{SkillID: SkillSMProvoke, Level: 5}, {SkillID: SkillSMRecovery, Level: 10}, {SkillID: SkillSMEndure, Level: 3}}},
	} {
		if got := SkillRequirements[tc.skillID]; !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("requirements for skill %d = %+v, want %+v", tc.skillID, got, tc.want)
		}
	}
}

func TestPriestSkillRequirementsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		job     int
		skillID uint16
		want    []SkillRequirement
	}{
		{JobWizard, SkillMGSafetywall, []SkillRequirement{{SkillID: SkillMGNapalmbeat, Level: 7}, {SkillID: SkillMGSoulstrike, Level: 5}}},
		{JobPriest, SkillMGSafetywall, []SkillRequirement{{SkillID: SkillPRSanctuary, Level: 3}, {SkillID: SkillPRAspersio, Level: 4}}},
		{JobPriestH, SkillMGSafetywall, []SkillRequirement{{SkillID: SkillPRSanctuary, Level: 3}, {SkillID: SkillPRAspersio, Level: 4}}},
		{JobPriest, SkillALLResurrection, []SkillRequirement{{SkillID: SkillMGSrecovery, Level: 4}, {SkillID: SkillPRStrecovery, Level: 1}}},
		{JobPriest, SkillPRSuffragium, []SkillRequirement{{SkillID: SkillPRImpositio, Level: 2}}},
		{JobPriest, SkillPRAspersio, []SkillRequirement{{SkillID: SkillALHolywater, Level: 1}, {SkillID: SkillPRImpositio, Level: 3}}},
		{JobPriest, SkillPRBenedictio, []SkillRequirement{{SkillID: SkillPRAspersio, Level: 5}, {SkillID: SkillPRGloria, Level: 3}}},
		{JobPriest, SkillPRSanctuary, []SkillRequirement{{SkillID: SkillALHeal, Level: 1}}},
		{JobPriest, SkillPRKyrie, []SkillRequirement{{SkillID: SkillALAngelus, Level: 2}}},
		{JobPriest, SkillPRGloria, []SkillRequirement{{SkillID: SkillPRKyrie, Level: 4}, {SkillID: SkillPRMagnificat, Level: 3}}},
		{JobPriest, SkillPRLexdivina, []SkillRequirement{{SkillID: SkillALRuwach, Level: 1}}},
		{JobPriest, SkillPRTurnundead, []SkillRequirement{{SkillID: SkillALLResurrection, Level: 1}, {SkillID: SkillPRLexdivina, Level: 3}}},
		{JobPriest, SkillPRLexaeterna, []SkillRequirement{{SkillID: SkillPRLexdivina, Level: 5}}},
		{JobPriest, SkillPRMagnus, []SkillRequirement{{SkillID: SkillMGSafetywall, Level: 1}, {SkillID: SkillPRLexaeterna, Level: 1}, {SkillID: SkillPRTurnundead, Level: 3}}},
		{JobPriestH, SkillHPManarecharge, []SkillRequirement{{SkillID: SkillPRMacemastery, Level: 10}, {SkillID: SkillALDemonbane, Level: 10}}},
		{JobPriestH, SkillHPAssumptio, []SkillRequirement{{SkillID: SkillALAngelus, Level: 1}, {SkillID: SkillMGSrecovery, Level: 3}, {SkillID: SkillPRImpositio, Level: 3}}},
		{JobPriestH, SkillHPBasilica, []SkillRequirement{{SkillID: SkillPRGloria, Level: 2}, {SkillID: SkillMGSrecovery, Level: 1}, {SkillID: SkillPRKyrie, Level: 3}}},
		{JobPriestH, SkillHPMeditatio, []SkillRequirement{{SkillID: SkillMGSrecovery, Level: 5}, {SkillID: SkillPRLexdivina, Level: 5}, {SkillID: SkillPRAspersio, Level: 3}}},
	} {
		if got := SkillRequirementsForJob(tc.job, tc.skillID); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("requirements for job %d skill %d = %+v, want %+v", tc.job, tc.skillID, got, tc.want)
		}
	}
}

func TestHunterSkillRequirementsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		skillID uint16
		want    []SkillRequirement
	}{
		{SkillHTPower, []SkillRequirement{{SkillID: SkillACDouble, Level: 10}}},
		{SkillHTAnklesnare, []SkillRequirement{{SkillID: SkillHTSkidtrap, Level: 1}}},
		{SkillHTShockwave, []SkillRequirement{{SkillID: SkillHTAnklesnare, Level: 1}}},
		{SkillHTSandman, []SkillRequirement{{SkillID: SkillHTFlasher, Level: 1}}},
		{SkillHTFlasher, []SkillRequirement{{SkillID: SkillHTSkidtrap, Level: 1}}},
		{SkillHTFreezingtrap, []SkillRequirement{{SkillID: SkillHTFlasher, Level: 1}}},
		{SkillHTBlastmine, []SkillRequirement{{SkillID: SkillHTLandmine, Level: 1}, {SkillID: SkillHTSandman, Level: 1}, {SkillID: SkillHTFreezingtrap, Level: 1}}},
		{SkillHTClaymoretrap, []SkillRequirement{{SkillID: SkillHTShockwave, Level: 1}, {SkillID: SkillHTBlastmine, Level: 1}}},
		{SkillHTRemovetrap, []SkillRequirement{{SkillID: SkillHTLandmine, Level: 1}}},
		{SkillHTTalkiebox, []SkillRequirement{{SkillID: SkillHTRemovetrap, Level: 1}, {SkillID: SkillHTShockwave, Level: 1}}},
		{SkillHTFalcon, []SkillRequirement{{SkillID: SkillHTBeastbane, Level: 1}}},
		{SkillHTSteelcrow, []SkillRequirement{{SkillID: SkillHTBlitzbeat, Level: 5}}},
		{SkillHTBlitzbeat, []SkillRequirement{{SkillID: SkillHTFalcon, Level: 1}}},
		{SkillHTDetecting, []SkillRequirement{{SkillID: SkillACConcentration, Level: 1}, {SkillID: SkillHTFalcon, Level: 1}}},
		{SkillHTSpringtrap, []SkillRequirement{{SkillID: SkillHTFalcon, Level: 1}, {SkillID: SkillHTRemovetrap, Level: 1}}},
		{SkillSNSight, []SkillRequirement{{SkillID: SkillACOwl, Level: 10}, {SkillID: SkillACVulture, Level: 10}, {SkillID: SkillACConcentration, Level: 10}, {SkillID: SkillHTFalcon, Level: 1}}},
		{SkillSNFalconassault, []SkillRequirement{{SkillID: SkillACVulture, Level: 5}, {SkillID: SkillHTFalcon, Level: 1}, {SkillID: SkillHTBlitzbeat, Level: 5}, {SkillID: SkillHTSteelcrow, Level: 3}}},
		{SkillSNSharpshooting, []SkillRequirement{{SkillID: SkillACDouble, Level: 5}, {SkillID: SkillACConcentration, Level: 10}}},
		{SkillSNWindwalk, []SkillRequirement{{SkillID: SkillACConcentration, Level: 9}}},
	} {
		if got := SkillRequirements[tc.skillID]; !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("requirements for skill %d = %+v, want %+v", tc.skillID, got, tc.want)
		}
	}
}

func TestAssassinSkillRequirementsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		skillID uint16
		want    []SkillRequirement
	}{
		{SkillASLeft, []SkillRequirement{{SkillID: SkillASRight, Level: 2}}},
		{SkillASCloaking, []SkillRequirement{{SkillID: SkillTFHiding, Level: 2}}},
		{SkillASSonicblow, []SkillRequirement{{SkillID: SkillASKatar, Level: 4}}},
		{SkillASGrimtooth, []SkillRequirement{{SkillID: SkillASCloaking, Level: 2}, {SkillID: SkillASSonicblow, Level: 5}}},
		{SkillASEnchantpoison, []SkillRequirement{{SkillID: SkillTFPoison, Level: 1}}},
		{SkillASPoisonreact, []SkillRequirement{{SkillID: SkillASEnchantpoison, Level: 3}}},
		{SkillASVenomdust, []SkillRequirement{{SkillID: SkillASEnchantpoison, Level: 5}}},
		{SkillASSplasher, []SkillRequirement{{SkillID: SkillASVenomdust, Level: 5}, {SkillID: SkillASPoisonreact, Level: 5}}},
		{SkillASCKatar, []SkillRequirement{{SkillID: SkillTFDouble, Level: 5}, {SkillID: SkillASKatar, Level: 7}}},
		{SkillASCEdp, []SkillRequirement{{SkillID: SkillASCCdp, Level: 1}}},
		{SkillASCBreaker, []SkillRequirement{{SkillID: SkillTFDouble, Level: 5}, {SkillID: SkillTFPoison, Level: 5}, {SkillID: SkillASCloaking, Level: 3}, {SkillID: SkillASEnchantpoison, Level: 6}}},
		{SkillASCMeteorassault, []SkillRequirement{{SkillID: SkillASKatar, Level: 5}, {SkillID: SkillASRight, Level: 3}, {SkillID: SkillASSonicblow, Level: 5}, {SkillID: SkillASCBreaker, Level: 1}}},
		{SkillASCCdp, []SkillRequirement{{SkillID: SkillTFPoison, Level: 10}, {SkillID: SkillTFDetoxify, Level: 1}, {SkillID: SkillASEnchantpoison, Level: 5}}},
	} {
		if got := SkillRequirements[tc.skillID]; !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("requirements for skill %d = %+v, want %+v", tc.skillID, got, tc.want)
		}
	}
}

func TestWizardSkillMaxLevelsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		skillID uint16
		want    int
	}{
		{SkillWZFirepillar, 10},
		{SkillWZSightrasher, 10},
		{SkillWZFireivy, 0},
		{SkillWZMeteor, 10},
		{SkillWZJupitel, 10},
		{SkillWZVermilion, 10},
		{SkillWZWaterball, 5},
		{SkillWZIcewall, 10},
		{SkillWZFrostnova, 10},
		{SkillWZStormgust, 10},
		{SkillWZEarthspike, 5},
		{SkillWZHeavendrive, 5},
		{SkillWZQuagmire, 5},
		{SkillWZEstimation, 1},
		{SkillWZSightblaster, 1},
		{SkillHWSouldrain, 10},
		{SkillHWMagiccrasher, 1},
		{SkillHWMagicpower, 10},
		{SkillHWNapalmvulcan, 5},
		{SkillHWGanbantein, 1},
		{SkillHWGravitation, 5},
	} {
		got, ok := SkillMaxLevel(tc.skillID)
		if tc.want == 0 {
			if ok {
				t.Fatalf("skill %d max level = %d ok=%t, want unavailable", tc.skillID, got, ok)
			}
			continue
		}
		if !ok || got != tc.want {
			t.Fatalf("skill %d max level = %d ok=%t, want %d", tc.skillID, got, ok, tc.want)
		}
	}
}

func TestKnightSkillMaxLevelsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		skillID uint16
		want    int
	}{
		{SkillKNSpearmastery, 10},
		{SkillKNPierce, 10},
		{SkillKNBrandishspear, 10},
		{SkillKNSpearstab, 10},
		{SkillKNSpearboomerang, 5},
		{SkillKNTwohandquicken, 10},
		{SkillKNAutocounter, 5},
		{SkillKNBowlingbash, 10},
		{SkillKNChargeatk, 1},
		{SkillKNRiding, 1},
		{SkillKNCavaliermastery, 5},
		{SkillKNOnehand, 1},
		{SkillLKSpiralpierce, 5},
		{SkillLKHeadcrush, 5},
		{SkillLKJointbeat, 10},
		{SkillLKAurablade, 5},
		{SkillLKParrying, 10},
		{SkillLKConcentration, 5},
		{SkillLKTensionrelax, 1},
		{SkillLKBerserk, 1},
	} {
		got, ok := SkillMaxLevel(tc.skillID)
		if !ok || got != tc.want {
			t.Fatalf("skill %d max level = %d ok=%t, want %d", tc.skillID, got, ok, tc.want)
		}
	}
}

func TestPriestSkillMaxLevelsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		skillID uint16
		want    int
	}{
		{SkillALLResurrection, 4},
		{SkillPRMacemastery, 10},
		{SkillPRImpositio, 5},
		{SkillPRSuffragium, 3},
		{SkillPRAspersio, 5},
		{SkillPRBenedictio, 5},
		{SkillPRSanctuary, 10},
		{SkillPRSlowpoison, 4},
		{SkillPRStrecovery, 1},
		{SkillPRKyrie, 10},
		{SkillPRMagnificat, 5},
		{SkillPRGloria, 5},
		{SkillPRLexdivina, 10},
		{SkillPRTurnundead, 10},
		{SkillPRLexaeterna, 1},
		{SkillPRMagnus, 10},
		{SkillPRRedemptio, 1},
		{SkillHPAssumptio, 5},
		{SkillHPBasilica, 5},
		{SkillHPMeditatio, 10},
		{SkillHPManarecharge, 5},
	} {
		got, ok := SkillMaxLevel(tc.skillID)
		if !ok || got != tc.want {
			t.Fatalf("skill %d max level = %d ok=%t, want %d", tc.skillID, got, ok, tc.want)
		}
	}
}

func TestHunterSkillMaxLevelsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		skillID uint16
		want    int
	}{
		{SkillHTPower, 1},
		{SkillHTPhantasmic, 1},
		{SkillHTSkidtrap, 5},
		{SkillHTLandmine, 5},
		{SkillHTAnklesnare, 5},
		{SkillHTShockwave, 5},
		{SkillHTSandman, 5},
		{SkillHTFlasher, 5},
		{SkillHTFreezingtrap, 5},
		{SkillHTBlastmine, 5},
		{SkillHTClaymoretrap, 5},
		{SkillHTRemovetrap, 1},
		{SkillHTTalkiebox, 1},
		{SkillHTBeastbane, 10},
		{SkillHTFalcon, 1},
		{SkillHTSteelcrow, 10},
		{SkillHTBlitzbeat, 5},
		{SkillHTDetecting, 4},
		{SkillHTSpringtrap, 5},
		{SkillSNSight, 10},
		{SkillSNFalconassault, 5},
		{SkillSNSharpshooting, 5},
		{SkillSNWindwalk, 10},
	} {
		got, ok := SkillMaxLevel(tc.skillID)
		if !ok || got != tc.want {
			t.Fatalf("skill %d max level = %d ok=%t, want %d", tc.skillID, got, ok, tc.want)
		}
	}
}

func TestAssassinSkillMaxLevelsMirrorRobrowser(t *testing.T) {
	for _, tc := range []struct {
		skillID uint16
		want    int
	}{
		{SkillASRight, 5},
		{SkillASLeft, 5},
		{SkillASKatar, 10},
		{SkillASCloaking, 10},
		{SkillASSonicblow, 10},
		{SkillASGrimtooth, 5},
		{SkillASEnchantpoison, 10},
		{SkillASPoisonreact, 10},
		{SkillASVenomdust, 10},
		{SkillASSplasher, 10},
		{SkillASSonicaccel, 1},
		{SkillASVenomknife, 1},
		{SkillASCKatar, 5},
		{SkillASCEdp, 5},
		{SkillASCBreaker, 10},
		{SkillASCMeteorassault, 10},
		{SkillASCCdp, 1},
	} {
		got, ok := SkillMaxLevel(tc.skillID)
		if !ok || got != tc.want {
			t.Fatalf("skill %d max level = %d ok=%t, want %d", tc.skillID, got, ok, tc.want)
		}
	}
}

func containsSkillID(skills []uint16, skillID uint16) bool {
	for _, id := range skills {
		if id == skillID {
			return true
		}
	}
	return false
}
