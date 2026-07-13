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

func TestSmallPromptLinesWrapLongDisconnectMessage(t *testing.T) {
	lines := smallPromptLines("You have been forced to disconnect by the Game Master Team.", alertPromptMaxLines)
	if len(lines) != alertPromptMaxLines {
		t.Fatalf("line count = %d, want %d", len(lines), alertPromptMaxLines)
	}
	if lines[0] == "" || lines[1] == "" {
		t.Fatalf("message was not wrapped into visible rows: %#v", lines)
	}
}
