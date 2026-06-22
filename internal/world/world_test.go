package world

import (
	"testing"
	"time"
)

func TestUpsertActorMovePreservesAppearance(t *testing.T) {
	w := New()
	w.UpsertActor(Actor{
		ID:         2000001,
		Name:       "remote",
		X:          10,
		Y:          20,
		Job:        3,
		Head:       7,
		Weapon:     1201,
		Shield:     2101,
		HeadTop:    22,
		HeadMid:    33,
		HeadLow:    11,
		Sex:        1,
		Appearance: true,
	})

	w.UpsertActor(Actor{
		ID:     2000001,
		X:      12,
		Y:      24,
		Moving: true,
		FromX:  10,
		FromY:  20,
		ToX:    12,
		ToY:    24,
	})

	actor := w.Actors[2000001]
	if actor.Job != 3 || actor.Head != 7 || actor.Sex != 1 || actor.Weapon != 1201 || actor.Shield != 2101 || actor.HeadTop != 22 || actor.HeadMid != 33 || actor.HeadLow != 11 || !actor.Appearance {
		t.Fatalf("appearance not preserved: %+v", actor)
	}
	if !actor.Moving || actor.FromX != 10 || actor.FromY != 20 || actor.ToX != 12 || actor.ToY != 24 {
		t.Fatalf("movement not stored: %+v", actor)
	}
}

func TestActorRenderPositionInterpolates(t *testing.T) {
	now := time.Now()
	actor := Actor{
		X:            20,
		Y:            30,
		Moving:       true,
		FromX:        10,
		FromY:        20,
		ToX:          20,
		ToY:          30,
		MoveStarted:  now.Add(-750 * time.Millisecond),
		MoveDuration: 1500 * time.Millisecond,
	}

	x, y := actor.RenderPosition(now)
	if x != 15 || y != 25 {
		t.Fatalf("position = %.2f, %.2f, want 15, 25", x, y)
	}
}

func TestSetPlayerMovementInterpolatesLocalPlayer(t *testing.T) {
	w := New()
	w.SetPlayerMovement(10, 20, 12, 24, 3)
	if !w.Player.Moving {
		t.Fatal("player should be moving")
	}
	if w.Player.X != 12 || w.Player.Y != 24 || w.Player.FromX != 10 || w.Player.FromY != 20 || w.Player.ToX != 12 || w.Player.ToY != 24 {
		t.Fatalf("unexpected player movement: %+v", w.Player)
	}
	if w.Dir != 3 || w.Player.Dir != 3 {
		t.Fatalf("direction = world %d player %d, want 3", w.Dir, w.Player.Dir)
	}
}

func TestRemoveActor(t *testing.T) {
	w := New()
	w.UpsertActor(Actor{ID: 2000002, X: 1, Y: 2})
	w.RemoveActor(2000002)
	if _, ok := w.Actors[2000002]; ok {
		t.Fatal("actor was not removed")
	}
}
