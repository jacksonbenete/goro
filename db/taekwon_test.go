package db

import (
	"reflect"
	"testing"
)

func TestTaekwonFamilyJobAliasesMatchRobrowser(t *testing.T) {
	tests := []struct {
		job          int
		display      string
		resourceFrom int
		hitFrom      int
	}{
		{JobTaekwonB, "Baby Taekwon", JobTaekwon, JobTaekwon},
		{JobStarB, "Baby Star Gladiator", JobStar, JobStar},
		{JobStar2B, "Baby Star Gladiator", JobStar2, JobStar2},
		{JobLinkerB, "Baby Soul Linker", JobLinker, JobLinker},
	}
	for _, tc := range tests {
		if got := JobDisplayName(tc.job); got != tc.display {
			t.Fatalf("job %d display = %q, want %q", tc.job, got, tc.display)
		}
		gotResource, gotOK := JobSpriteResourceName(tc.job)
		wantResource, wantOK := JobSpriteResourceName(tc.resourceFrom)
		if !gotOK || !wantOK || gotResource != wantResource {
			t.Fatalf("job %d resource = %q/%t, want %q/%t", tc.job, gotResource, gotOK, wantResource, wantOK)
		}
		if got, want := JobHitSounds(tc.job), JobHitSounds(tc.hitFrom); !reflect.DeepEqual(got, want) {
			t.Fatalf("job %d hit sounds = %v, want %v", tc.job, got, want)
		}
	}
}

func TestTaekwonFamilySkillTreesMatchRobrowser(t *testing.T) {
	taekwon := []uint16{
		SkillTKRun, SkillTKStormkick, SkillTKDownkick, SkillTKTurnkick, SkillTKCounter,
		SkillTKJumpkick, SkillTKHighjump, SkillTKReadystorm, SkillTKReadydown,
		SkillTKReadyturn, SkillTKReadycounter, SkillTKDodge, SkillTKHptime,
		SkillTKSptime, SkillTKPower, SkillTKSevenwind, SkillTKMission,
	}
	star := []uint16{
		SkillSGFeel, SkillSGHate, SkillSGDevil, SkillSGKnowledge, SkillSGSunWarm,
		SkillSGSunComfort, SkillSGSunAnger, SkillSGSunBless, SkillSGFriend,
		SkillSGFusion, SkillSGMoonWarm, SkillSGMoonComfort, SkillSGMoonAnger,
		SkillSGMoonBless, SkillSGStarWarm, SkillSGStarComfort, SkillSGStarAnger,
		SkillSGStarBless,
	}
	linker := []uint16{
		SkillSLAlchemist, SkillSLStar, SkillSLAssasin, SkillSLCrusader,
		SkillSLBarddancer, SkillSLSupernovice, SkillSLBlacksmith, SkillSLSoullinker,
		SkillSLRogue, SkillSLKnight, SkillSLHunter, SkillSLHigh, SkillSLMonk,
		SkillSLKaupe, SkillSLSke, SkillSLSage, SkillSLKaina, SkillSLPriest,
		SkillSLSka, SkillSLWizard, SkillSLKaite, SkillSLKaahi, SkillSLKaizel,
		SkillSLSwoo, SkillSLStin, SkillSLStun, SkillSLSma,
	}

	for _, job := range []int{JobTaekwon, JobTaekwonB} {
		groups := SkillTreeSkillGroups(job)
		if len(groups) != 1 || !reflect.DeepEqual(groups[0].SkillIDs[len(noviceSkillTree):], taekwon) {
			t.Fatalf("job %d groups = %+v, want Taekwon tree", job, groups)
		}
	}
	for _, job := range []int{JobStar, JobStar2, JobStarB, JobStar2B} {
		groups := SkillTreeSkillGroups(job)
		if len(groups) != 2 || !reflect.DeepEqual(groups[0].SkillIDs[len(noviceSkillTree):], taekwon) || !reflect.DeepEqual(groups[1].SkillIDs, star) {
			t.Fatalf("job %d groups = %+v, want Taekwon and Star Gladiator trees", job, groups)
		}
	}
	for _, job := range []int{JobLinker, JobLinkerB} {
		groups := SkillTreeSkillGroups(job)
		if len(groups) != 2 || !reflect.DeepEqual(groups[0].SkillIDs[len(noviceSkillTree):], taekwon) || !reflect.DeepEqual(groups[1].SkillIDs, linker) {
			t.Fatalf("job %d groups = %+v, want Taekwon and Soul Linker trees", job, groups)
		}
	}
}

func TestTaekwonFamilySkillRequirementsMatchRobrowser(t *testing.T) {
	tests := []struct {
		skill uint16
		want  []SkillRequirement
	}{
		{SkillTKReadystorm, []SkillRequirement{{SkillTKStormkick, 1}}},
		{SkillTKReadydown, []SkillRequirement{{SkillTKDownkick, 1}}},
		{SkillTKReadyturn, []SkillRequirement{{SkillTKTurnkick, 1}}},
		{SkillTKReadycounter, []SkillRequirement{{SkillTKCounter, 1}}},
		{SkillTKDodge, []SkillRequirement{{SkillTKJumpkick, 7}}},
		{SkillTKSevenwind, []SkillRequirement{{SkillTKHptime, 5}, {SkillTKSptime, 5}, {SkillTKPower, 5}}},
		{SkillTKMission, []SkillRequirement{{SkillTKPower, 5}}},
		{SkillSGSunWarm, []SkillRequirement{{SkillSGFeel, 1}}},
		{SkillSGMoonWarm, []SkillRequirement{{SkillSGFeel, 2}}},
		{SkillSGStarWarm, []SkillRequirement{{SkillSGFeel, 3}}},
		{SkillSGSunComfort, []SkillRequirement{{SkillSGFeel, 1}}},
		{SkillSGMoonComfort, []SkillRequirement{{SkillSGFeel, 2}}},
		{SkillSGStarComfort, []SkillRequirement{{SkillSGFeel, 3}}},
		{SkillSGSunAnger, []SkillRequirement{{SkillSGHate, 1}}},
		{SkillSGMoonAnger, []SkillRequirement{{SkillSGHate, 2}}},
		{SkillSGStarAnger, []SkillRequirement{{SkillSGHate, 3}}},
		{SkillSGSunBless, []SkillRequirement{{SkillSGFeel, 1}, {SkillSGHate, 1}}},
		{SkillSGMoonBless, []SkillRequirement{{SkillSGFeel, 2}, {SkillSGHate, 2}}},
		{SkillSGStarBless, []SkillRequirement{{SkillSGFeel, 3}, {SkillSGHate, 3}}},
		{SkillSGFusion, []SkillRequirement{{SkillSGKnowledge, 9}}},
		{SkillSLSupernovice, []SkillRequirement{{SkillSLStar, 1}}},
		{SkillSLKnight, []SkillRequirement{{SkillSLCrusader, 1}}},
		{SkillSLWizard, []SkillRequirement{{SkillSLSage, 1}}},
		{SkillSLPriest, []SkillRequirement{{SkillSLMonk, 1}}},
		{SkillSLRogue, []SkillRequirement{{SkillSLAssasin, 1}}},
		{SkillSLBlacksmith, []SkillRequirement{{SkillSLAlchemist, 1}}},
		{SkillSLHunter, []SkillRequirement{{SkillSLBarddancer, 1}}},
		{SkillSLSoullinker, []SkillRequirement{{SkillSLStar, 1}}},
		{SkillSLKaizel, []SkillRequirement{{SkillSLPriest, 1}}},
		{SkillSLKaahi, []SkillRequirement{{SkillSLCrusader, 1}, {SkillSLMonk, 1}, {SkillSLPriest, 1}}},
		{SkillSLKaupe, []SkillRequirement{{SkillSLAssasin, 1}, {SkillSLRogue, 1}}},
		{SkillSLKaite, []SkillRequirement{{SkillSLSage, 1}, {SkillSLWizard, 1}}},
		{SkillSLKaina, []SkillRequirement{{SkillTKSptime, 1}}},
		{SkillSLStin, []SkillRequirement{{SkillSLWizard, 1}}},
		{SkillSLStun, []SkillRequirement{{SkillSLWizard, 1}}},
		{SkillSLSma, []SkillRequirement{{SkillSLStin, 7}, {SkillSLStun, 7}}},
		{SkillSLSwoo, []SkillRequirement{{SkillSLPriest, 1}}},
		{SkillSLSke, []SkillRequirement{{SkillSLKnight, 1}}},
		{SkillSLSka, []SkillRequirement{{SkillSLMonk, 1}}},
		{SkillSLHigh, []SkillRequirement{{SkillSLSupernovice, 5}}},
	}
	for _, tc := range tests {
		if got := SkillRequirementsForJob(JobLinker, tc.skill); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("skill %d requirements = %v, want %v", tc.skill, got, tc.want)
		}
	}
}

func TestTaekwonFamilyMaxAndSelectableLevelsMatchRobrowser(t *testing.T) {
	maxLevels := map[uint16]int{
		SkillTKRun: 10, SkillTKReadystorm: 1, SkillTKStormkick: 7, SkillTKReadydown: 1,
		SkillTKDownkick: 7, SkillTKReadyturn: 1, SkillTKTurnkick: 7, SkillTKReadycounter: 1,
		SkillTKCounter: 7, SkillTKDodge: 1, SkillTKJumpkick: 7, SkillTKHptime: 10,
		SkillTKSptime: 10, SkillTKPower: 5, SkillTKSevenwind: 7, SkillTKHighjump: 5,
		SkillTKMission: 1, SkillSGFeel: 3, SkillSGSunWarm: 3, SkillSGMoonWarm: 3,
		SkillSGStarWarm: 3, SkillSGSunComfort: 4, SkillSGMoonComfort: 4,
		SkillSGStarComfort: 4, SkillSGHate: 3, SkillSGSunAnger: 3, SkillSGMoonAnger: 3,
		SkillSGStarAnger: 3, SkillSGSunBless: 5, SkillSGMoonBless: 5, SkillSGStarBless: 5,
		SkillSGDevil: 10, SkillSGFriend: 3, SkillSGKnowledge: 10, SkillSGFusion: 1,
		SkillSLAlchemist: 5, SkillSLMonk: 5, SkillSLStar: 5, SkillSLSage: 5,
		SkillSLCrusader: 5, SkillSLSupernovice: 5, SkillSLKnight: 5, SkillSLWizard: 5,
		SkillSLPriest: 5, SkillSLBarddancer: 5, SkillSLRogue: 5, SkillSLAssasin: 5,
		SkillSLBlacksmith: 5, SkillSLHunter: 5, SkillSLSoullinker: 5, SkillSLKaizel: 7,
		SkillSLKaahi: 7, SkillSLKaupe: 3, SkillSLKaite: 7, SkillSLKaina: 7,
		SkillSLStin: 7, SkillSLStun: 7, SkillSLSma: 10, SkillSLSwoo: 7,
		SkillSLSke: 3, SkillSLSka: 3, SkillSLHigh: 5,
	}
	selectable := map[uint16]bool{
		SkillTKSevenwind: true, SkillTKHighjump: true, SkillSGFeel: true, SkillSGHate: true,
		SkillSLKaahi: true, SkillSLStin: true, SkillSLStun: true, SkillSLSma: true,
	}
	for skillID, wantMax := range maxLevels {
		if got, ok := SkillMaxLevel(skillID); !ok || got != wantMax {
			t.Fatalf("skill %d max = %d/%t, want %d", skillID, got, ok, wantMax)
		}
		if got, known := SkillLevelSelectable(skillID); !known || got != selectable[skillID] {
			t.Fatalf("skill %d selectable = %t/%t, want %t/true", skillID, got, known, selectable[skillID])
		}
		if _, ok := SkillAttackRange(skillID, 1); !ok {
			t.Fatalf("skill %d has no level-one attack range", skillID)
		}
		if _, ok := SkillAttackRange(skillID, wantMax); !ok {
			t.Fatalf("skill %d has no max-level attack range", skillID)
		}
	}
	for level, want := range []int{2, 4, 6, 8, 10} {
		if got, ok := SkillAttackRange(SkillTKHighjump, level+1); !ok || got != want {
			t.Fatalf("high jump level %d range = %d/%t, want %d/true", level+1, got, ok, want)
		}
	}
	if got, _ := SkillAttackRange(SkillTKJumpkick, 7); got != 9 {
		t.Fatalf("flying kick range = %d, want 9", got)
	}
	if got, _ := SkillAttackRange(SkillSGHate, 3); got != 9 {
		t.Fatalf("opposition range = %d, want 9", got)
	}
	if got, _ := SkillAttackRange(SkillSLHigh, 5); got != 9 {
		t.Fatalf("transcendent spirit range = %d, want 9", got)
	}
}
