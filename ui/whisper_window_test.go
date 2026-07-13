package ui

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
)

func TestWhisperWindowOpensForTarget(t *testing.T) {
	var window WhisperWindow

	window.Open(client.Context{}, "Alice Smith")

	if !window.IsOpen() {
		t.Fatal("whisper window did not open")
	}
	if window.target != "Alice Smith" {
		t.Fatalf("target = %q, want Alice Smith", window.target)
	}
}

func TestWhisperWindowSubmitKeepsTargetWithSpaces(t *testing.T) {
	var window WhisperWindow
	window.Open(client.Context{}, "Alice Smith")
	window.input = "hello"

	window.submit(client.Context{})

	action := window.PopAction()
	if action.Target != "Alice Smith" || action.Message != "hello" {
		t.Fatalf("action = %+v, want target Alice Smith and message hello", action)
	}
}

func TestWhisperWindowFocusedEnterSubmitsImmediately(t *testing.T) {
	inputState := input.NewState()
	var window WhisperWindow
	ctx := client.Context{Input: inputState}
	window.Open(ctx, "Alice Smith")
	window.input = "hello"
	window.inputWidget(ctx).SetFocused(true)
	inputState.SetKey(input.KeyEnter, true)

	if !window.Update(ctx) {
		t.Fatal("focused enter did not consume the whisper window update")
	}
	action := window.PopAction()
	if action.Target != "Alice Smith" || action.Message != "hello" {
		t.Fatalf("action = %+v, want target Alice Smith and message hello", action)
	}
}

func TestWhisperWindowRebindRecreatesInputField(t *testing.T) {
	var window WhisperWindow
	ctx := client.Context{}
	window.Open(ctx, "Alice Smith")
	oldInput := window.inputField
	if oldInput == nil {
		t.Fatal("whisper window did not create input field")
	}

	window.Rebind(ctx)

	if window.inputField == nil {
		t.Fatal("whisper window did not recreate input field")
	}
	if window.inputField == oldInput {
		t.Fatal("whisper window rebind reused stale input field")
	}
}
