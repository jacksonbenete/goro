package ui

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/render"
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

func TestConsoleInputHistoryUsesArrowKeys(t *testing.T) {
	console := &ChatConsole{input: "draft", active: true}
	console.rememberInput("/sit")
	console.rememberInput("hello")

	inputState := input.NewState()
	inputState.SetKey(render.KeyArrowUp, true)
	if !console.Update(client.Context{Input: inputState, ScreenW: 800, ScreenH: 600}) {
		t.Fatal("up key was not handled")
	}
	if console.input != "hello" {
		t.Fatalf("first history input = %q, want hello", console.input)
	}

	inputState = input.NewState()
	inputState.SetKey(render.KeyArrowUp, true)
	console.Update(client.Context{Input: inputState, ScreenW: 800, ScreenH: 600})
	if console.input != "/sit" {
		t.Fatalf("second history input = %q, want /sit", console.input)
	}

	inputState = input.NewState()
	inputState.SetKey(render.KeyArrowUp, true)
	console.Update(client.Context{Input: inputState, ScreenW: 800, ScreenH: 600})
	if console.input != "/sit" {
		t.Fatalf("oldest history input = %q, want /sit", console.input)
	}

	inputState = input.NewState()
	inputState.SetKey(render.KeyArrowDown, true)
	console.Update(client.Context{Input: inputState, ScreenW: 800, ScreenH: 600})
	if console.input != "hello" {
		t.Fatalf("newer history input = %q, want hello", console.input)
	}

	inputState = input.NewState()
	inputState.SetKey(render.KeyArrowDown, true)
	console.Update(client.Context{Input: inputState, ScreenW: 800, ScreenH: 600})
	if console.input != "draft" {
		t.Fatalf("restored draft input = %q, want draft", console.input)
	}
}
