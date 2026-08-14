package db

import "testing"

func TestSkillAttackRangeMirrorsRobrowserCompanionRows(t *testing.T) {
	tests := []struct {
		name    string
		skillID uint16
		level   int
		want    int
	}{
		{"vanilmirth caprice", SkillHvanCaprice, 5, 9},
		{"wedding male", SkillWEMale, 1, 9},
		{"wedding female", SkillWEFemale, 1, 9},
		{"wedding call partner", SkillWECallpartner, 1, 1},
		{"wedding baby", SkillWEBaby, 1, 9},
		{"wedding call parent", SkillWECallparent, 1, 1},
		{"wedding call baby", SkillWECallbaby, 1, 1},
		{"filir sbr44", SkillHfliSbr44, 1, 9},
		{"homunculus s needle", SkillMhStahlHorn, 10, 9},
		{"mercenary devotion", SkillMlDevotion, 4, 10},
		{"mercenary lex divina", SkillMerLexdivina, 1, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := SkillAttackRange(tt.skillID, tt.level)
			if !ok || got != tt.want {
				t.Fatalf("SkillAttackRange(%d, %d) = %d, %t; want %d, true", tt.skillID, tt.level, got, ok, tt.want)
			}
		})
	}
}

func TestSkillAttackRangeRejectsUnknownOrInvalidLevel(t *testing.T) {
	if _, ok := SkillAttackRange(SkillHvanCaprice, 0); ok {
		t.Fatal("level 0 should not resolve")
	}
	if _, ok := SkillAttackRange(SkillHvanCaprice, 6); ok {
		t.Fatal("level beyond range table should not resolve")
	}
}

func TestPreRenewalSelectableLevelMetadata(t *testing.T) {
	for _, skillID := range []uint16{
		SkillSMBash,
		SkillCRAutoguard,
		SkillMGSoulstrike,
		SkillPRKyrie,
		SkillHTBlitzbeat,
		SkillASSonicblow,
		SkillAMPotionpitcher,
		SkillCGArrowvulcan,
	} {
		if selectable, known := SkillLevelSelectable(skillID); !known || !selectable {
			t.Fatalf("skill %d selectable=%t known=%t, want true,true", skillID, selectable, known)
		}
	}

	for _, skillID := range []uint16{
		SkillSMMagnum,
		SkillCRShieldboomerang,
		SkillACDouble,
		SkillPRMagnificat,
		SkillHTDetecting,
		SkillASGrimtooth,
		SkillCRAciddemonstration,
		SkillCGTarotcard,
	} {
		if selectable, known := SkillLevelSelectable(skillID); !known || selectable {
			t.Fatalf("skill %d selectable=%t known=%t, want false,true", skillID, selectable, known)
		}
	}

	if selectable, known := SkillLevelSelectable(65000); known || selectable {
		t.Fatalf("unimported skill selectable=%t known=%t, want false,false", selectable, known)
	}
}
