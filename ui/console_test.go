package ui

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/session"
)

func TestConsoleNoShiftCommandTogglesSessionPreference(t *testing.T) {
	console := &ChatConsole{input: "/ns", active: true}
	sessionState := &session.Session{}
	ctx := client.Context{Session: sessionState}

	if !console.SubmitCommand(ctx, "/ns") {
		t.Fatal("noshift command was not handled")
	}
	if !sessionState.NoShift {
		t.Fatal("noshift was not enabled")
	}
	if console.active || console.input != "" {
		t.Fatalf("console active=%t input=%q, want closed empty input", console.active, console.input)
	}

	if !console.SubmitCommand(ctx, "/noshift") {
		t.Fatal("noshift command was not handled")
	}
	if sessionState.NoShift {
		t.Fatal("noshift was not disabled")
	}
}

func TestConsoleMemoCommandWithoutNetwork(t *testing.T) {
	console := &ChatConsole{input: "/memo", active: true}

	if !console.SubmitCommand(client.Context{}, "/memo") {
		t.Fatal("memo command was not handled")
	}
	if console.active || console.input != "" {
		t.Fatalf("console active=%t input=%q, want closed empty input", console.active, console.input)
	}
	if len(console.messages) != 1 || console.messages[0].Text != "send failed: not connected" {
		t.Fatalf("console messages = %+v", console.messages)
	}
}
