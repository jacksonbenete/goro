package db

import (
	"reflect"
	"testing"
)

func TestNinjaJobAliasesMatchRobrowser(t *testing.T) {
	if got := JobDisplayName(JobNinjaB); got != "Baby Ninja" {
		t.Fatalf("baby Ninja display = %q, want Baby Ninja", got)
	}
	gotResource, gotOK := JobSpriteResourceName(JobNinjaB)
	wantResource, wantOK := JobSpriteResourceName(JobNinja)
	if !gotOK || !wantOK || gotResource != wantResource {
		t.Fatalf("baby Ninja resource = %q/%t, want %q/%t", gotResource, gotOK, wantResource, wantOK)
	}
	if got, want := JobHitSounds(JobNinjaB), JobHitSounds(JobNinja); !reflect.DeepEqual(got, want) {
		t.Fatalf("baby Ninja hit sounds = %v, want %v", got, want)
	}
	for _, job := range []int{JobNinja, JobNinjaB} {
		if got := PlayerWeaponAction(job, 0, WeaponNone); got != PlayerWeaponActionAttack1 {
			t.Fatalf("job %d unarmed action = %d, want ATTACK1", job, got)
		}
		if got := PlayerWeaponAction(job, 0, WeaponShortsword); got != PlayerWeaponActionAttack2 {
			t.Fatalf("job %d dagger action = %d, want ATTACK2", job, got)
		}
		if got := PlayerWeaponAction(job, 0, WeaponShuriken); got != PlayerWeaponActionAttack3 {
			t.Fatalf("job %d huuma action = %d, want ATTACK3", job, got)
		}
	}
}

func TestNinjaSkillTreesMatchRobrowser(t *testing.T) {
	want := []uint16{
		SkillNJTobidougu, SkillNJTatamigaeshi, SkillNJNinpou, SkillNJSyuriken,
		SkillNJShadowjump, SkillNJNen, SkillNJKouenka, SkillNJHyousensou,
		SkillNJHuujin, SkillNJKunai, SkillNJKasumikiri, SkillNJUtsusemi,
		SkillNJKaensin, SkillNJSuiton, SkillNJRaigekisai, SkillNJHuuma,
		SkillNJKirikage, SkillNJBakuenryu, SkillNJHyousyouraku, SkillNJKamaitachi,
		SkillNJZenynage, SkillNJBunsinjyutsu, SkillNJIssen,
	}
	for _, job := range []int{JobNinja, JobNinjaB} {
		groups := SkillTreeSkillGroups(job)
		if len(groups) != 1 || !reflect.DeepEqual(groups[0].SkillIDs[len(noviceSkillTree):], want) {
			t.Fatalf("job %d groups = %+v, want Ninja tree", job, groups)
		}
	}
}

func TestNinjaSkillRequirementsMatchRobrowser(t *testing.T) {
	tests := []struct {
		skill uint16
		want  []SkillRequirement
	}{
		{SkillNJSyuriken, []SkillRequirement{{SkillNJTobidougu, 1}}},
		{SkillNJKunai, []SkillRequirement{{SkillNJSyuriken, 5}}},
		{SkillNJHuuma, []SkillRequirement{{SkillNJTobidougu, 5}, {SkillNJKunai, 5}}},
		{SkillNJZenynage, []SkillRequirement{{SkillNJTobidougu, 10}, {SkillNJHuuma, 5}}},
		{SkillNJKasumikiri, []SkillRequirement{{SkillNJShadowjump, 1}}},
		{SkillNJShadowjump, []SkillRequirement{{SkillNJTatamigaeshi, 1}}},
		{SkillNJKirikage, []SkillRequirement{{SkillNJKasumikiri, 5}}},
		{SkillNJUtsusemi, []SkillRequirement{{SkillNJShadowjump, 5}}},
		{SkillNJBunsinjyutsu, []SkillRequirement{{SkillNJNen, 1}, {SkillNJUtsusemi, 4}, {SkillNJKirikage, 3}}},
		{SkillNJKouenka, []SkillRequirement{{SkillNJNinpou, 1}}},
		{SkillNJKaensin, []SkillRequirement{{SkillNJKouenka, 5}}},
		{SkillNJBakuenryu, []SkillRequirement{{SkillNJNinpou, 10}, {SkillNJKaensin, 7}}},
		{SkillNJHyousensou, []SkillRequirement{{SkillNJNinpou, 1}}},
		{SkillNJSuiton, []SkillRequirement{{SkillNJHyousensou, 5}}},
		{SkillNJHyousyouraku, []SkillRequirement{{SkillNJNinpou, 10}, {SkillNJSuiton, 7}}},
		{SkillNJHuujin, []SkillRequirement{{SkillNJNinpou, 1}}},
		{SkillNJRaigekisai, []SkillRequirement{{SkillNJHuujin, 5}}},
		{SkillNJKamaitachi, []SkillRequirement{{SkillNJNinpou, 10}, {SkillNJRaigekisai, 5}}},
		{SkillNJNen, []SkillRequirement{{SkillNJNinpou, 5}}},
		{SkillNJIssen, []SkillRequirement{{SkillNJTobidougu, 7}, {SkillNJNen, 1}, {SkillNJKirikage, 5}}},
	}
	for _, tc := range tests {
		if got := SkillRequirementsForJob(JobNinja, tc.skill); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("skill %d requirements = %v, want %v", tc.skill, got, tc.want)
		}
	}
}

func TestNinjaMaxSelectableLevelsAndRangesMatchRobrowser(t *testing.T) {
	tests := []struct {
		skill      uint16
		max        int
		selectable bool
		firstRange int
		lastRange  int
	}{
		{SkillNJTobidougu, 10, false, 1, 1},
		{SkillNJSyuriken, 10, false, 9, 9},
		{SkillNJKunai, 5, false, 9, 9},
		{SkillNJHuuma, 5, true, 9, 9},
		{SkillNJZenynage, 10, true, 7, 7},
		{SkillNJTatamigaeshi, 5, false, 1, 1},
		{SkillNJKasumikiri, 10, true, 1, 1},
		{SkillNJShadowjump, 5, false, 6, 14},
		{SkillNJKirikage, 5, true, 1, 1},
		{SkillNJUtsusemi, 5, true, 1, 1},
		{SkillNJBunsinjyutsu, 10, true, 1, 1},
		{SkillNJNinpou, 10, false, 1, 1},
		{SkillNJKouenka, 10, true, 9, 9},
		{SkillNJKaensin, 10, false, 1, 1},
		{SkillNJBakuenryu, 5, true, 9, 9},
		{SkillNJHyousensou, 10, true, 9, 9},
		{SkillNJSuiton, 10, true, 9, 9},
		{SkillNJHyousyouraku, 5, true, 1, 1},
		{SkillNJHuujin, 10, true, 9, 9},
		{SkillNJRaigekisai, 5, true, 9, 9},
		{SkillNJKamaitachi, 5, true, 5, 9},
		{SkillNJNen, 5, true, 1, 1},
		{SkillNJIssen, 10, true, 5, 5},
	}
	for _, tc := range tests {
		if got, ok := SkillMaxLevel(tc.skill); !ok || got != tc.max {
			t.Fatalf("skill %d max = %d/%t, want %d/true", tc.skill, got, ok, tc.max)
		}
		if got, known := SkillLevelSelectable(tc.skill); !known || got != tc.selectable {
			t.Fatalf("skill %d selectable = %t/%t, want %t/true", tc.skill, got, known, tc.selectable)
		}
		if got, ok := SkillAttackRange(tc.skill, 1); !ok || got != tc.firstRange {
			t.Fatalf("skill %d level-one range = %d/%t, want %d/true", tc.skill, got, ok, tc.firstRange)
		}
		if got, ok := SkillAttackRange(tc.skill, tc.max); !ok || got != tc.lastRange {
			t.Fatalf("skill %d max-level range = %d/%t, want %d/true", tc.skill, got, ok, tc.lastRange)
		}
	}
}
