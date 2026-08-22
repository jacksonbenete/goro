package game

import (
	"testing"

	"github.com/kivutar/goro/network"
)

func TestMapInfoNotifyAddsServerRestrictionToConsole(t *testing.T) {
	mode := &WorldMode{}

	mode.applyMapInfoNotify(network.MapInfoNotify{Result: 1})

	messages := mode.ui.console.Messages()
	if len(messages) != 1 || messages[0].Text != "Saved point cannot be memorized." {
		t.Fatalf("console messages = %+v", messages)
	}
}

func TestMapInfoNotifyMessages(t *testing.T) {
	tests := []struct {
		result uint16
		want   string
	}{
		{0, "Unable to teleport in this area."},
		{1, "Saved point cannot be memorized."},
		{2, "This skill cannot be used in this area."},
		{3, "This item cannot be used in this area."},
		{99, "Action cannot be used in this area."},
	}
	for _, test := range tests {
		if got := mapInfoNotifyMessage(test.result); got != test.want {
			t.Errorf("result %d = %q, want %q", test.result, got, test.want)
		}
	}
}
