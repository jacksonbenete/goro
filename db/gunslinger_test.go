package db

import (
	"reflect"
	"testing"
)

func TestGunslingerJobAliasesMatchRobrowser(t *testing.T) {
	if got := JobDisplayName(JobGunslingerB); got != "Baby Gunslinger" {
		t.Fatalf("baby Gunslinger display = %q, want Baby Gunslinger", got)
	}
	gotResource, gotOK := JobSpriteResourceName(JobGunslingerB)
	wantResource, wantOK := JobSpriteResourceName(JobGunslinger)
	if !gotOK || !wantOK || gotResource != wantResource {
		t.Fatalf("baby Gunslinger resource = %q/%t, want %q/%t", gotResource, gotOK, wantResource, wantOK)
	}
	if got, want := JobHitSounds(JobGunslingerB), JobHitSounds(JobGunslinger); !reflect.DeepEqual(got, want) {
		t.Fatalf("baby Gunslinger hit sounds = %v, want %v", got, want)
	}
	for _, job := range []int{JobGunslinger, JobGunslingerB} {
		for _, weapon := range []int{WeaponNone, WeaponGunHandgun, WeaponGunShotgun} {
			if got := PlayerWeaponAction(job, 0, weapon); got != PlayerWeaponActionAttack2 {
				t.Fatalf("job %d weapon %d action = %d, want ATTACK2", job, weapon, got)
			}
		}
		for _, weapon := range []int{WeaponGunGatling, WeaponGunRifle, WeaponGunGrenade} {
			if got := PlayerWeaponAction(job, 0, weapon); got != PlayerWeaponActionAttack3 {
				t.Fatalf("job %d weapon %d action = %d, want ATTACK3", job, weapon, got)
			}
		}
	}
}

func TestGunslingerSkillTreesMatchRobrowser(t *testing.T) {
	want := []uint16{
		SkillGSGlittering, SkillGSSingleaction, SkillGSCracker, SkillGSMagicalbullet,
		SkillGSChainaction, SkillGSTracking, SkillGSDust, SkillGSSpreadattack,
		SkillGSIncreasing, SkillGSFling, SkillGSRapidshower, SkillGSPiercingshot,
		SkillGSFullbuster, SkillGSGrounddrift, SkillGSMadnesscancel, SkillGSTripleaction,
		SkillGSDesperado, SkillGSDisarm, SkillGSAdjustment, SkillGSGatlingfever,
		SkillGSSnakeeye, SkillGSBullseye,
	}
	for _, job := range []int{JobGunslinger, JobGunslingerB} {
		groups := SkillTreeSkillGroups(job)
		if len(groups) != 1 || !reflect.DeepEqual(groups[0].SkillIDs[len(noviceSkillTree):], want) {
			t.Fatalf("job %d groups = %+v, want Gunslinger tree", job, groups)
		}
	}
}

func TestGunslingerSkillRequirementsMatchRobrowser(t *testing.T) {
	tests := []struct {
		skill uint16
		want  []SkillRequirement
	}{
		{SkillGSFling, []SkillRequirement{{SkillGSGlittering, 1}}},
		{SkillGSTripleaction, []SkillRequirement{{SkillGSGlittering, 1}}},
		{SkillGSBullseye, []SkillRequirement{{SkillGSGlittering, 5}}},
		{SkillGSMadnesscancel, []SkillRequirement{{SkillGSGlittering, 4}}},
		{SkillGSAdjustment, []SkillRequirement{{SkillGSGlittering, 4}}},
		{SkillGSIncreasing, []SkillRequirement{{SkillGSGlittering, 2}}},
		{SkillGSMagicalbullet, []SkillRequirement{{SkillGSGlittering, 1}}},
		{SkillGSCracker, []SkillRequirement{{SkillGSGlittering, 1}}},
		{SkillGSChainaction, []SkillRequirement{{SkillGSSingleaction, 1}}},
		{SkillGSTracking, []SkillRequirement{{SkillGSSingleaction, 5}}},
		{SkillGSDisarm, []SkillRequirement{{SkillGSTracking, 7}}},
		{SkillGSPiercingshot, []SkillRequirement{{SkillGSTracking, 5}}},
		{SkillGSRapidshower, []SkillRequirement{{SkillGSChainaction, 3}}},
		{SkillGSDesperado, []SkillRequirement{{SkillGSRapidshower, 5}}},
		{SkillGSGatlingfever, []SkillRequirement{{SkillGSRapidshower, 7}, {SkillGSDesperado, 5}}},
		{SkillGSDust, []SkillRequirement{{SkillGSSingleaction, 5}}},
		{SkillGSFullbuster, []SkillRequirement{{SkillGSDust, 3}}},
		{SkillGSSpreadattack, []SkillRequirement{{SkillGSSingleaction, 5}}},
		{SkillGSGrounddrift, []SkillRequirement{{SkillGSSpreadattack, 7}}},
	}
	for _, tc := range tests {
		if got := SkillRequirementsForJob(JobGunslinger, tc.skill); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("skill %d requirements = %v, want %v", tc.skill, got, tc.want)
		}
	}
}

func TestGunslingerMaxSelectableLevelsAndRangesMatchRobrowser(t *testing.T) {
	tests := []struct {
		skill       uint16
		max         int
		selectable  bool
		attackRange int
	}{
		{SkillGSGlittering, 5, false, 1},
		{SkillGSFling, 1, false, 9},
		{SkillGSTripleaction, 1, false, 9},
		{SkillGSBullseye, 1, false, 9},
		{SkillGSMadnesscancel, 1, false, 1},
		{SkillGSAdjustment, 1, false, 1},
		{SkillGSIncreasing, 1, false, 1},
		{SkillGSMagicalbullet, 1, false, 1},
		{SkillGSCracker, 1, false, 9},
		{SkillGSSingleaction, 10, false, 1},
		{SkillGSSnakeeye, 10, false, 1},
		{SkillGSChainaction, 10, false, 1},
		{SkillGSTracking, 10, true, 9},
		{SkillGSDisarm, 5, true, 9},
		{SkillGSPiercingshot, 5, true, 9},
		{SkillGSRapidshower, 10, true, 9},
		{SkillGSDesperado, 10, true, 1},
		{SkillGSGatlingfever, 10, true, 1},
		{SkillGSDust, 10, true, 2},
		{SkillGSFullbuster, 10, true, 9},
		{SkillGSSpreadattack, 10, true, 9},
		{SkillGSGrounddrift, 10, true, 9},
	}
	for _, tc := range tests {
		if got, ok := SkillMaxLevel(tc.skill); !ok || got != tc.max {
			t.Fatalf("skill %d max = %d/%t, want %d/true", tc.skill, got, ok, tc.max)
		}
		if got, known := SkillLevelSelectable(tc.skill); !known || got != tc.selectable {
			t.Fatalf("skill %d selectable = %t/%t, want %t/true", tc.skill, got, known, tc.selectable)
		}
		if got, ok := SkillAttackRange(tc.skill, 1); !ok || got != tc.attackRange {
			t.Fatalf("skill %d level-one range = %d/%t, want %d/true", tc.skill, got, ok, tc.attackRange)
		}
		if got, ok := SkillAttackRange(tc.skill, tc.max); !ok || got != tc.attackRange {
			t.Fatalf("skill %d max-level range = %d/%t, want %d/true", tc.skill, got, ok, tc.attackRange)
		}
	}
}
