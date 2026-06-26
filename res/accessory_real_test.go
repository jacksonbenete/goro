package res

import "testing"

func TestAccessoryResourceNameRealWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	name, ok := manager.AccessoryResourceName(1)
	if !ok || name == "" {
		t.Fatalf("accessory view 1 unresolved: ok=%v name=%q", ok, name)
	}
	t.Logf("view 1 accessory resource = %q", name)
	for _, candidates := range [][]string{
		PlayerAccessoryResourceCandidates(0, 1, 1, 1, name, "act"),
		PlayerAccessoryResourceCandidates(0, 1, 1, 1, name, "spr"),
	} {
		source, _, ok := manager.ReadFirst(candidates)
		if !ok {
			t.Fatalf("accessory resource missing for %q: %q", name, candidates)
		}
		t.Logf("resolved accessory resource %s", source)
	}
}
