package gamemode

import (
	"testing"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestDeathModalOpenAndClearWhenAlive(t *testing.T) {
	var modal deathModalState
	modal.openDeath()
	if !modal.open {
		t.Fatal("expected modal open")
	}
	if modal.pending != deathModalActionNone {
		t.Fatalf("pending = %d, want none", modal.pending)
	}

	ctx := Context{Session: &session.Session{}}
	ctx.Session.Vitals.HP = 0
	ctx.Session.Selected.HP = 0
	modal.clearIfAlive(ctx)
	if !modal.open {
		t.Fatal("modal cleared while character is still dead")
	}

	ctx.Session.Vitals.HP = 1
	modal.clearIfAlive(ctx)
	if modal.open {
		t.Fatal("modal stayed open after positive HP")
	}
}

func TestDeathModalCharacterSelectAckRequestsModeSwitch(t *testing.T) {
	var modal deathModalState
	modal.openDeath()
	modal.pending = deathModalActionCharSelect

	if !modal.applyRestartAck(network.RestartAck{Allowed: true}) {
		t.Fatal("allowed restart ack should request character-select transition")
	}
	if modal.status != "Returning to character select..." {
		t.Fatalf("status = %q", modal.status)
	}
}

func TestDeathModalCharacterSelectAckDeniedKeepsModalOpen(t *testing.T) {
	var modal deathModalState
	modal.openDeath()
	modal.pending = deathModalActionCharSelect

	if modal.applyRestartAck(network.RestartAck{Allowed: false}) {
		t.Fatal("denied restart ack should not request transition")
	}
	if !modal.open || modal.pending != deathModalActionNone {
		t.Fatalf("modal = %+v, want open and no pending action", modal)
	}
}

func TestNewCharacterSelectModePreparesSavedCharacters(t *testing.T) {
	ctx := Context{Session: &session.Session{
		Characters: []session.Character{
			{ID: 150002, Slot: 4, Name: "Second"},
			{ID: 150001, Slot: 2, Name: "First"},
		},
	}}

	mode := NewCharacterSelectMode(ctx, chatConsole{})
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
