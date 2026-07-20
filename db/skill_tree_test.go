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

func containsSkillID(skills []uint16, skillID uint16) bool {
	for _, id := range skills {
		if id == skillID {
			return true
		}
	}
	return false
}
