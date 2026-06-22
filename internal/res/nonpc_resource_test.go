package res

import (
	"os"
	"testing"
)

func TestNonPCSpriteResourceCandidatesNPC(t *testing.T) {
	got := NonPCSpriteResourceCandidates(47, "1_M_01", "act")
	want := []string{
		"data\\sprite\\NPC\\1_M_01.act",
		"data\\sprite\\NPC\\1_m_01.act",
		"data\\sprite\\npc\\1_M_01.act",
		"data\\sprite\\npc\\1_m_01.act",
	}
	if len(got) != len(want) {
		t.Fatalf("candidate count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("candidate[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestNonPCSpriteResourceCandidatesMonster(t *testing.T) {
	got := NonPCSpriteResourceCandidates(1002, "PORING", "spr")
	want := []string{
		"data\\sprite\\monster\\PORING.spr",
		"data\\sprite\\monster\\poring.spr",
		legacyMonsterSpriteRoot + "PORING.spr",
		legacyMonsterSpriteRoot + "poring.spr",
		"data\\sprite\\PORING.spr",
		"data\\sprite\\poring.spr",
	}
	if len(got) != len(want) {
		t.Fatalf("candidate count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("candidate[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFallbackJobResourceName(t *testing.T) {
	manager := &Manager{}
	if got, ok := manager.JobResourceName(1002); !ok || got != "poring" {
		t.Fatalf("job resource name = %q, %v, want poring, true", got, ok)
	}
}

func TestNonPCSpriteResourceRealWhenConfigured(t *testing.T) {
	root := os.Getenv("GORO_TEST_DATA_DIR")
	if root == "" {
		t.Skip("set GORO_TEST_DATA_DIR to run against a real client data directory")
	}
	manager, err := NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	cases := []int{47, 1002}
	for _, job := range cases {
		name, ok := manager.JobResourceName(job)
		if !ok {
			t.Fatalf("job %d resource name missing", job)
		}
		if source, _, ok := manager.ReadFirst(NonPCSpriteResourceCandidates(job, name, "act")); !ok {
			t.Fatalf("job %d act not found for %q", job, name)
		} else {
			t.Logf("job %d act=%s", job, source)
		}
		if source, _, ok := manager.ReadFirst(NonPCSpriteResourceCandidates(job, name, "spr")); !ok {
			t.Fatalf("job %d spr not found for %q", job, name)
		} else {
			t.Logf("job %d spr=%s", job, source)
		}
	}
}
