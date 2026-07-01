package game

import (
	"github.com/kivutar/goro/client"
	"testing"

	"github.com/kivutar/goro/session"
	gameui "github.com/kivutar/goro/ui"
)

func TestNewCharacterSelectModePreparesSavedCharacters(t *testing.T) {
	ctx := client.Context{Session: &session.Session{
		Characters: []session.Character{
			{ID: 150002, Slot: 4, Name: "Second"},
			{ID: 150001, Slot: 2, Name: "First"},
		},
	}}

	mode := NewCharacterSelectMode(ctx, gameui.ChatConsole{})
	if mode.phase != loginPhaseCharacter {
		t.Fatalf("phase = %d, want character select", mode.phase)
	}
	if mode.selectedSlot != 2 {
		t.Fatalf("selected slot = %d, want first occupied slot 2", mode.selectedSlot)
	}
	if mode.maxSlots < 6 {
		t.Fatalf("max slots = %d, want enough slots for saved characters", mode.maxSlots)
	}
}
