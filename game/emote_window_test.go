package game

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
)

func TestAltLTogglesEmoteWindow(t *testing.T) {
	inputState := input.NewState()
	inputState.SetKey(input.KeyAlt, true)
	inputState.SetKey(input.KeyL, true)
	mode := &WorldMode{}
	ctx := client.Context{
		Input:   inputState,
		ScreenW: 800,
		ScreenH: 600,
	}

	if !mode.toggleEmoteWindowFromInput(ctx) {
		t.Fatal("Alt+L was not consumed")
	}
	if !mode.ui.emoteWindow.IsOpen() {
		t.Fatal("emote window did not open")
	}
}
