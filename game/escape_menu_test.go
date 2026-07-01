package game

import (
	"testing"

	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
)

func TestEscapeMenuCharacterSelectButtonRequestsAction(t *testing.T) {
	inputState := input.NewState()
	menu := escapeMenuState{open: true}
	ctx := Context{Input: inputState, ScreenW: 800, ScreenH: 600}
	x, y, w, _ := escapeMenuBounds(800, 600)
	bx, by, bw, bh := escapeMenuButtonBounds(x, y, w, 0)
	inputState.SetMousePosition(bx+bw/2, by+bh/2)
	inputState.SetMouseButton(input.MouseButtonLeft, true)

	if !menu.update(ctx) {
		t.Fatal("escape menu did not consume character-select click")
	}
	if got := menu.consumeAction(); got != escapeMenuActionCharacterSelect {
		t.Fatalf("action = %d, want character select", got)
	}
}

func TestEscapeMenuCharacterSelectAckRequestsModeSwitch(t *testing.T) {
	menu := escapeMenuState{open: true, pending: true}

	if !menu.applyRestartAck(network.RestartAck{Allowed: true}) {
		t.Fatal("allowed restart ack should request character-select transition")
	}
	if menu.status != "Returning to character select..." {
		t.Fatalf("status = %q", menu.status)
	}
}

func TestEscapeMenuCharacterSelectAckDeniedKeepsMenuOpen(t *testing.T) {
	menu := escapeMenuState{open: true, pending: true}

	if menu.applyRestartAck(network.RestartAck{Allowed: false}) {
		t.Fatal("denied restart ack should not request transition")
	}
	if !menu.open || menu.pending {
		t.Fatalf("menu = %+v, want open without pending request", menu)
	}
}

func TestEscapeMenuCharacterSelectWithoutNetworkShowsError(t *testing.T) {
	menu := escapeMenuState{open: true}
	menu.requestCharacterSelect(Context{})

	if menu.pending {
		t.Fatal("menu stayed pending without a network connection")
	}
	if menu.status != "Character select failed: not connected" {
		t.Fatalf("status = %q", menu.status)
	}
}
