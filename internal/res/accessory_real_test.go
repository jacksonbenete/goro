package res

import (
	"os"
	"testing"
)

func TestAccessoryResourceNameRealWhenConfigured(t *testing.T) {
	root := "/home/kivutar/Téléchargements/OldRO"
	if _, err := os.Stat(root); err != nil {
		t.Skip("OldRO data not available")
	}
	manager, err := NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	name, ok := manager.AccessoryResourceName(1)
	if !ok || name == "" {
		t.Fatalf("accessory view 1 unresolved: ok=%v name=%q", ok, name)
	}
	t.Logf("view 1 accessory resource = %q", name)
}
