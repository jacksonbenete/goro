package res

import "testing"

func TestPlayerSexTokenUsesRagnarokSexEnum(t *testing.T) {
	if got := PlayerSexToken(0); got != playerFemaleSex {
		t.Fatalf("sex 0 token = %q, want female", got)
	}
	if got := PlayerSexToken(1); got != playerMaleSex {
		t.Fatalf("sex 1 token = %q, want male", got)
	}
}

func TestHasPlayerJobToken(t *testing.T) {
	if !HasPlayerJobToken(0) {
		t.Fatal("novice job token missing")
	}
	if HasPlayerJobToken(1002) {
		t.Fatal("unknown job token should not report as renderable")
	}
}

func TestPlayerIMFResourceCandidates(t *testing.T) {
	got := PlayerIMFResourceCandidates(1, 1)
	want := "data\\imf\\\xB0\xCB\xBB\xE7_\xB3\xB2.imf"
	if len(got) == 0 || got[0] != want {
		t.Fatalf("first imf candidate = %q, want %q", got, want)
	}
	if got[len(got)-1] != "data\\imf\\\xC3\xCA\xBA\xB8\xC0\xDA_\xB3\xB2.imf" {
		t.Fatalf("fallback imf candidate = %q", got[len(got)-1])
	}
}

func TestPlayerCartResourceCandidates(t *testing.T) {
	got := PlayerCartResourceCandidates(1, "spr")
	want := "data\\sprite\\\xC0\xCC\xC6\xD1\xC6\xAE\\\xBC\xD5\xBC\xF6\xB7\xB9.spr"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("cart 1 candidate = %q, want %q", got, want)
	}
	got = PlayerCartResourceCandidates(13, "act")
	want = "data\\sprite\\\xC0\xCC\xC6\xD1\xC6\xAE\\\xB8\xB6\xB5\xB5\xC4\xAB\xC6\xAE.act"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("cart 13 candidate = %q, want %q", got, want)
	}
}

func TestPlayerWeaponOverlayResourceCandidates(t *testing.T) {
	got := PlayerWeaponOverlayResourceCandidates(0, 1, 1201, false, "act")
	want := []string{
		"data\\sprite\\\xC0\xCE\xB0\xA3\xC1\xB7\\\xC3\xCA\xBA\xB8\xC0\xDA\\\xC3\xCA\xBA\xB8\xC0\xDA_\xB3\xB2_1201.act",
		"data\\sprite\\\xC0\xCE\xB0\xA3\xC1\xB7\\\xC3\xCA\xBA\xB8\xC0\xDA\\\xC3\xCA\xBA\xB8\xC0\xDA_\xB3\xB2_\xB4\xDC\xB0\xCB.act",
	}
	if len(got) != len(want) {
		t.Fatalf("weapon overlay = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("weapon overlay[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPlayerWeaponViewIDUsesItemClassNumBeforeFallbackRange(t *testing.T) {
	manager := &Manager{
		itemMetadataLoaded: true,
		itemMetadata: map[int]ItemMetadata{
			1607: {ClassNum: 10, ClassNumSet: true},
			1615: {ClassNum: 70, ClassNumSet: true},
		},
	}
	if got := manager.PlayerWeaponViewID(1607); got != 10 {
		t.Fatalf("weapon view id 1607 = %d, want class num 10", got)
	}
	if got := manager.PlayerWeaponViewID(1615); got != 70 {
		t.Fatalf("weapon view id 1615 = %d, want class num 70", got)
	}
	if got := manager.PlayerWeaponViewID(1701); got != 11 {
		t.Fatalf("weapon view id fallback 1701 = %d, want bow type 11", got)
	}
}

func TestPlayerWeaponOverlayTypeForJobMatchesReferenceJobRules(t *testing.T) {
	if got := PlayerWeaponOverlayTypeForJob(2, 5, false); got != 10 {
		t.Fatalf("mage overlay type for weapon 5 = %d, want rod type 10", got)
	}
	if got := PlayerWeaponOverlayTypeForJob(16, 5, false); got != 10 {
		t.Fatalf("sage overlay type for weapon 5 = %d, want rod type 10", got)
	}
	if got := PlayerWeaponOverlayTypeForJob(1, 5, false); got != 5 {
		t.Fatalf("swordman overlay type for weapon 5 = %d, want spear type 5", got)
	}
}

func TestPlayerWeaponOverlayTokenUsesClientRodSpelling(t *testing.T) {
	if got := PlayerWeaponOverlayToken(10); got != "\xB7\xD4\xB5\xE5" {
		t.Fatalf("rod token = %q, want client spelling", got)
	}
}

func TestNormalizePlayerWeaponShieldMovesLeftHandWeapon(t *testing.T) {
	weapon, shield := NormalizePlayerWeaponShield(0, 1601)
	if weapon != 1601 || shield != 0 {
		t.Fatalf("normalized left-hand weapon = weapon %d shield %d, want 1601/0", weapon, shield)
	}
	weapon, shield = NormalizePlayerWeaponShield(0, 2101)
	if weapon != 0 || shield != 2101 {
		t.Fatalf("normalized real shield = weapon %d shield %d, want 0/2101", weapon, shield)
	}
}

func TestPlayerShieldOverlayResourceCandidates(t *testing.T) {
	got := PlayerShieldOverlayResourceCandidates(0, 1, 2101, "spr")
	want := "data\\sprite\\\xB9\xE6\xC6\xD0\\\xC3\xCA\xBA\xB8\xC0\xDA\\\xC3\xCA\xBA\xB8\xC0\xDA_\xB3\xB2_\xB0\xA1\xB5\xE5.spr"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("shield overlay = %q, want %q", got, want)
	}
}

func TestPlayerAccessoryResourceCandidates(t *testing.T) {
	got := PlayerAccessoryResourceCandidates(0, 3, 0, 100, "sample", "act")
	want := "data\\sprite\\\xBE\xC7\xBC\xBC\xBB\xE7\xB8\xAE\\\xBF\xA9\\\xBF\xA9_sample.act"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("accessory overlay = %q, want %q", got, want)
	}
}
