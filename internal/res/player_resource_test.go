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
