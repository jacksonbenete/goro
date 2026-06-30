package gamemode

import (
	"testing"

	"github.com/kivutar/goro/session"
)

func TestConsoleNoShiftCommandTogglesSessionPreference(t *testing.T) {
	console := &chatConsole{input: "/ns", active: true}
	sessionState := &session.Session{}
	ctx := Context{Session: sessionState}

	if !console.submitCommand(ctx, "/ns") {
		t.Fatal("noshift command was not handled")
	}
	if !sessionState.NoShift {
		t.Fatal("noshift was not enabled")
	}
	if console.active || console.input != "" {
		t.Fatalf("console active=%t input=%q, want closed empty input", console.active, console.input)
	}

	if !console.submitCommand(ctx, "/noshift") {
		t.Fatal("noshift command was not handled")
	}
	if sessionState.NoShift {
		t.Fatal("noshift was not disabled")
	}
}
