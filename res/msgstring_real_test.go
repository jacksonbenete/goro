package res

import (
	"os"
	"testing"
)

func TestMsgStringRealWhenConfigured(t *testing.T) {
	root := os.Getenv("GORO_TEST_DATA_DIR")
	if root == "" {
		t.Skip("set GORO_TEST_DATA_DIR to run against real client data")
	}
	manager, err := NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, archive := range manager.Archives {
		defer archive.Close()
	}
	text, ok := manager.MsgString(0)
	if !ok || text == "" {
		t.Fatalf("msgstring 0 not found")
	}
	t.Logf("msgstring[0]=%q", text)
}
