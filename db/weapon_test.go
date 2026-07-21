package db

import "testing"

func TestPlayerWeaponTypeMatchesRobrowserFallbackRanges(t *testing.T) {
	for _, tc := range []struct {
		itemID int
		want   int
	}{
		{1600, WeaponRod},
		{1649, WeaponRod},
		{1650, WeaponRod},
		{1699, WeaponRod},
		{1700, WeaponBow},
		{2000, WeaponTwoHandRod},
		{2049, WeaponTwoHandRod},
		{2050, WeaponNone},
		{13099, WeaponShortsword},
		{13399, WeaponShuriken},
		{13499, WeaponSword},
		{18499, WeaponBow},
		{20000, WeaponTwoHandRod},
		{20999, WeaponTwoHandRod},
		{21000, WeaponTwoHandSword},
		{21999, WeaponTwoHandSword},
		{22000, WeaponNone},
	} {
		if got := PlayerWeaponType(tc.itemID); got != tc.want {
			t.Fatalf("weapon type for item %d = %d, want %d", tc.itemID, got, tc.want)
		}
	}
}

func TestPlayerWeaponActionMatchesRobrowserWizardRodBySex(t *testing.T) {
	if got := PlayerWeaponAction(JobWizard, 0, WeaponRod); got != PlayerWeaponActionAttack3 {
		t.Fatalf("female Wizard rod action = %d, want ATTACK3", got)
	}
	if got := PlayerWeaponAction(JobWizard, 0, 1699); got != PlayerWeaponActionAttack3 {
		t.Fatalf("female Wizard item-id rod action = %d, want ATTACK3", got)
	}
	if got := PlayerWeaponAction(JobWizardH, 0, 1601); got != PlayerWeaponActionAttack3 {
		t.Fatalf("female High Wizard rod action = %d, want ATTACK3", got)
	}
	if got := PlayerWeaponAction(JobWizard, 1, WeaponRod); got != PlayerWeaponActionAttack2 {
		t.Fatalf("male Wizard rod action = %d, want ATTACK2", got)
	}
	if got := PlayerWeaponAction(JobWizard, 0, WeaponNone); got != PlayerWeaponActionAttack1 {
		t.Fatalf("female Wizard unarmed action = %d, want ATTACK1", got)
	}
	if got := PlayerWeaponAction(JobWizard, 1, WeaponShortsword); got != PlayerWeaponActionAttack3 {
		t.Fatalf("male Wizard shortsword action = %d, want ATTACK3", got)
	}
}
