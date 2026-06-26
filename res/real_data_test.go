package res

import (
	"os"
	"testing"
)

func realDataManager(t *testing.T) *Manager {
	t.Helper()
	root := os.Getenv("GORO_DATA_DIR")
	if root == "" {
		t.Skip("set GORO_DATA_DIR to run against real client data")
	}
	manager, err := NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, archive := range manager.Archives {
		t.Cleanup(func() {
			_ = archive.Close()
		})
	}
	return manager
}

func realDataArchive(t *testing.T) *GRF {
	t.Helper()
	manager := realDataManager(t)
	if len(manager.Archives) == 0 {
		t.Skip("GORO_DATA_DIR has no GRF/GPF archives")
	}
	return manager.Archives[0]
}

func realDataArchiveFile(t *testing.T, name string) (*GRF, string) {
	t.Helper()
	manager := realDataManager(t)
	for _, archive := range manager.Archives {
		if archive.Has(name) {
			return archive, name
		}
		matches := archive.NamesWithSuffix(name)
		if len(matches) > 0 {
			return archive, matches[0]
		}
	}
	t.Skipf("%s not present in GRF/GPF archives under GORO_DATA_DIR", name)
	return nil, ""
}

func readRealDataFile(t *testing.T, name string) []byte {
	t.Helper()
	manager := realDataManager(t)
	data, err := manager.ReadFile(name)
	if err != nil {
		t.Skipf("%s not present under GORO_DATA_DIR: %v", name, err)
	}
	return data
}
