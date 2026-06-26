package res

import "testing"

func TestMsgStringRealWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	text, ok := manager.MsgString(0)
	if !ok || text == "" {
		t.Fatalf("msgstring 0 not found")
	}
	t.Logf("msgstring[0]=%q", text)
}
