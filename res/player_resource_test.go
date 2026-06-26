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

func TestPlayerWeaponOverlayResourceCandidates(t *testing.T) {
	got := PlayerWeaponOverlayResourceCandidates(0, 1, 1201, false, "act")
	want := "data\\sprite\\\xC0\xCE\xB0\xA3\xC1\xB7\\\xC3\xCA\xBA\xB8\xC0\xDA\\\xC3\xCA\xBA\xB8\xC0\xDA_\xB3\xB2_\xB4\xDC\xB0\xCB.act"
	if len(got) != 1 || got[0] != want {
		t.Fatalf("weapon overlay = %q, want %q", got, want)
	}
}

func TestPlayerShieldOverlayResourceCandidates(t *testing.T) {
	got := PlayerShieldOverlayResourceCandidates(0, 1, 2101, "spr")
	want := "data\\sprite\\\xB9\xE6\xC6\xD0\\\xC3\xCA\xBA\xB8\xC0\xDA\\\xC3\xCA\xBA\xB8\xC0\xDA_\xB3\xB2_guard.spr"
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
