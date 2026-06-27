package gamemode

import (
	"testing"

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
