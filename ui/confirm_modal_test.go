package ui

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
)

func TestConfirmModalOpenAlertEscapeConfirms(t *testing.T) {
	var modal ConfirmModal
	inputState := input.NewState()
	confirmed := false
	ctx := client.Context{
		Input:   inputState,
		ScreenW: 800,
		ScreenH: 600,
	}
	modal.OpenAlert(ctx, "Disconnected", "Disconnected from Server.", func() {
		confirmed = true
	})

	inputState.SetKey(input.KeyEscape, true)
	if !modal.Update(ctx) {
		t.Fatal("alert did not consume escape")
	}
	if modal.IsOpen() {
		t.Fatal("alert remained open after escape")
	}
	if !confirmed {
		t.Fatal("escape did not confirm alert")
	}
}

func TestConfirmModalUsesCompactHeightForOneLinePrompt(t *testing.T) {
	var modal ConfirmModal
	modal.Open(client.Context{ScreenW: 800, ScreenH: 600}, "Expel Party Member", "Expel Alice from the party?", nil, nil)

	want := ROWindowTitleHeight + smallPromptContentH + smallPromptFooterH
	if modal.height != want {
		t.Fatalf("modal height = %d, want %d", modal.height, want)
	}
}

func TestConfirmModalKeepsRoomForWrappedPrompt(t *testing.T) {
	var oneLine ConfirmModal
	oneLine.Open(client.Context{ScreenW: 800, ScreenH: 600}, "Confirm", "Expel Alice from the party?", nil, nil)

	var wrapped ConfirmModal
	wrapped.Open(client.Context{ScreenW: 800, ScreenH: 600}, "Confirm", "Would you like to invite Some Very Long Character Name to join your party?", nil, nil)

	if wrapped.height <= oneLine.height {
		t.Fatalf("wrapped height = %d, want greater than one-line height %d", wrapped.height, oneLine.height)
	}
}

func TestSmallPromptLinesWrapLongDisconnectMessage(t *testing.T) {
	lines := smallPromptLines("You have been forced to disconnect by the Game Master Team.", alertPromptMaxLines)
	if len(lines) != alertPromptMaxLines {
		t.Fatalf("line count = %d, want %d", len(lines), alertPromptMaxLines)
	}
	if lines[0] == "" || lines[1] == "" {
		t.Fatalf("message was not wrapped into visible rows: %#v", lines)
	}
}
