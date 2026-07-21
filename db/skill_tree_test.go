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

func containsSkillID(skills []uint16, skillID uint16) bool {
	for _, id := range skills {
		if id == skillID {
			return true
		}
	}
	return false
}
