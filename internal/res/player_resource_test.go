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
