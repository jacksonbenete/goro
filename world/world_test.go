package world

import (
	"testing"
	"time"

	"github.com/kivutar/goro/res"
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
	if actor.Name != "remote" {
		t.Fatalf("name = %q, want remote", actor.Name)
	}
}

func TestUpsertActorMovePreservesObjectType(t *testing.T) {
	w := New()
	w.UpsertActor(Actor{
		ID:            2000001,
		X:             10,
		Y:             20,
		ObjectType:    5,
		HasObjectType: true,
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
	if actor.ObjectType != 5 || !actor.HasObjectType {
		t.Fatalf("object type not preserved: %+v", actor)
	}
}

func TestUpsertActorMoveUsesActorSpeed(t *testing.T) {
	w := New()
	w.UpsertActor(Actor{
		ID:    2000001,
		X:     10,
		Y:     20,
		Speed: 400,
	})

	w.UpsertActor(Actor{
		ID:     2000001,
		X:      11,
		Y:      20,
		Moving: true,
		FromX:  10,
		FromY:  20,
		ToX:    11,
		ToY:    20,
	})

	actor := w.Actors[2000001]
	if actor.Speed != 400 {
		t.Fatalf("speed = %d, want preserved 400", actor.Speed)
	}
	if actor.MoveDuration != 400*time.Millisecond {
		t.Fatalf("duration = %s, want 400ms", actor.MoveDuration)
	}
}

func TestUpsertActorPreservesSittingUntilMove(t *testing.T) {
	w := New()
	w.UpsertActor(Actor{ID: 2000001, X: 10, Y: 20, Sitting: true})

	w.UpsertActor(Actor{ID: 2000001, X: 10, Y: 20})
	if actor := w.Actors[2000001]; !actor.Sitting {
		t.Fatalf("sitting state was not preserved: %+v", actor)
	}

	w.UpsertActor(Actor{ID: 2000001, X: 11, Y: 20, Moving: true, FromX: 10, FromY: 20, ToX: 11, ToY: 20})
	if actor := w.Actors[2000001]; actor.Sitting {
		t.Fatalf("moving actor stayed sitting: %+v", actor)
	}
}

func TestSetPlayerMovementClearsSitting(t *testing.T) {
	w := New()
	w.Player.Sitting = true

	w.SetPlayerMovement(10, 20, 11, 20, 6)

	if w.Player.Sitting {
		t.Fatal("player stayed sitting after movement")
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

func TestActorRenderPositionFollowsSpeedScaledWalkPath(t *testing.T) {
	now := time.Now()
	actor := Actor{
		X:            2,
		Y:            0,
		Moving:       true,
		FromX:        0,
		FromY:        0,
		ToX:          2,
		ToY:          0,
		Speed:        400,
		MoveStarted:  now.Add(-600 * time.Millisecond),
		MoveDuration: 800 * time.Millisecond,
		MovePath: []WalkStep{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 2, Y: 0},
		},
	}

	x, y := actor.RenderPosition(now)
	if x != 1.5 || y != 0 {
		t.Fatalf("position = %.2f, %.2f, want 1.50, 0.00", x, y)
	}
}

func TestActorRenderWalkDistanceFollowsSpeedScaledPath(t *testing.T) {
	now := time.Now()
	actor := Actor{
		X:            2,
		Y:            0,
		Moving:       true,
		FromX:        0,
		FromY:        0,
		ToX:          2,
		ToY:          0,
		Speed:        400,
		MoveStarted:  now.Add(-600 * time.Millisecond),
		MoveDuration: 800 * time.Millisecond,
		MovePath: []WalkStep{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 2, Y: 0},
		},
	}

	if got := actor.RenderWalkDistance(now); got != 1.5 {
		t.Fatalf("walk distance = %.2f, want 1.50", got)
	}
}

func TestUpsertActorMovePreservesWalkPhaseWhileAlreadyMoving(t *testing.T) {
	w := New()
	started := time.Now().Add(-200 * time.Millisecond)
	w.Actors[2000001] = Actor{
		ID:            2000001,
		X:             1,
		Y:             0,
		Moving:        true,
		FromX:         0,
		FromY:         0,
		ToX:           1,
		ToY:           0,
		Speed:         400,
		MoveStarted:   started,
		MoveDuration:  400 * time.Millisecond,
		MovePath:      []WalkStep{{X: 0, Y: 0}, {X: 1, Y: 0}},
		WalkDistance:  2,
		Appearance:    true,
		HasObjectType: true,
	}

	w.UpsertActor(Actor{
		ID:     2000001,
		X:      2,
		Y:      0,
		Moving: true,
		FromX:  1,
		FromY:  0,
		ToX:    2,
		ToY:    0,
		Speed:  400,
	})

	actor := w.Actors[2000001]
	if actor.WalkDistance <= 2 {
		t.Fatalf("walk distance offset = %.2f, want preserved progress greater than 2", actor.WalkDistance)
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
	wantDir := movementFinalDirection(w.Player.MovePath, 10, 20, 12, 24, 3)
	if w.Dir != wantDir || w.Player.Dir != wantDir {
		t.Fatalf("direction = world %d player %d, want %d", w.Dir, w.Player.Dir, wantDir)
	}
}

func TestSetPlayerMovementUsesPlayerSpeed(t *testing.T) {
	w := New()
	w.Player.Speed = 100

	w.SetPlayerMovement(10, 20, 11, 20, 6)

	if w.Player.MoveDuration != 100*time.Millisecond {
		t.Fatalf("move duration = %s, want 100ms", w.Player.MoveDuration)
	}
}

func TestSetPlayerMovementAtFastForwardsRoute(t *testing.T) {
	w := New()
	w.Player.Speed = 100
	now := time.Unix(100, 0)

	w.SetPlayerMovementAt(0, 0, 3, 0, 6, now, 150*time.Millisecond)

	x, y := w.Player.RenderPosition(now)
	if x != 1.5 || y != 0 {
		t.Fatalf("position = %.2f, %.2f, want 1.50, 0.00", x, y)
	}
}

func TestSetPlayerMovementAtPreservesCurrentRouteProgress(t *testing.T) {
	w := New()
	now := time.Unix(100, 0)
	w.Player = Actor{
		X:            3,
		Y:            0,
		Moving:       true,
		FromX:        0,
		FromY:        0,
		ToX:          3,
		ToY:          0,
		MoveStarted:  now.Add(-150 * time.Millisecond),
		MoveDuration: 300 * time.Millisecond,
		MovePath: []WalkStep{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 2, Y: 0},
			{X: 3, Y: 0},
		},
		Speed: 100,
	}

	w.SetPlayerMovementAt(0, 0, 4, 0, 6, now, 0)

	x, y := w.Player.RenderPosition(now)
	if x != 1.5 || y != 0 {
		t.Fatalf("position = %.2f, %.2f, want preserved 1.50, 0.00", x, y)
	}
}

func TestSetPlayerMovementAtUsesFractionalStartNearAckSource(t *testing.T) {
	w := New()
	now := time.Unix(100, 0)
	w.Player = Actor{
		X:            4,
		Y:            0,
		Moving:       true,
		FromX:        0,
		FromY:        0,
		ToX:          4,
		ToY:          0,
		MoveStarted:  now.Add(-195 * time.Millisecond),
		MoveDuration: 400 * time.Millisecond,
		MovePath: []WalkStep{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 2, Y: 0},
			{X: 3, Y: 0},
			{X: 4, Y: 0},
		},
		Speed: 100,
	}

	w.SetPlayerMovementAt(2, 0, 5, 0, 6, now, 0)

	x, y := w.Player.RenderPosition(now)
	if x != 1.95 || y != 0 {
		t.Fatalf("position = %.2f, %.2f, want preserved 1.95, 0.00", x, y)
	}
	if !w.Player.HasMoveStart || w.Player.MoveStartX != 1.95 || w.Player.MoveStartY != 0 {
		t.Fatalf("fractional start = %.2f, %.2f enabled=%t", w.Player.MoveStartX, w.Player.MoveStartY, w.Player.HasMoveStart)
	}
}

func TestSetPlayerMovementAtDoesNotProjectOntoNearbyParallelRoute(t *testing.T) {
	w := New()
	now := time.Unix(100, 0)
	w.Player = Actor{
		X:            3,
		Y:            0,
		Moving:       true,
		FromX:        0,
		FromY:        0,
		ToX:          3,
		ToY:          0,
		MoveStarted:  now.Add(-150 * time.Millisecond),
		MoveDuration: 300 * time.Millisecond,
		MovePath: []WalkStep{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 2, Y: 0},
			{X: 3, Y: 0},
		},
		Speed: 100,
	}

	w.SetPlayerMovementAt(0, 1, 4, 1, 6, now, 0)

	x, y := w.Player.RenderPosition(now)
	if x != 0 || y != 1 {
		t.Fatalf("position = %.2f, %.2f, want new route start 0.00, 1.00", x, y)
	}
}

func TestActorRenderPositionFollowsWalkPath(t *testing.T) {
	now := time.Now()
	actor := Actor{
		X:            2,
		Y:            1,
		Dir:          0,
		Moving:       true,
		FromX:        0,
		FromY:        0,
		ToX:          2,
		ToY:          1,
		MoveStarted:  now.Add(-225 * time.Millisecond),
		MoveDuration: 450 * time.Millisecond,
		MovePath: []WalkStep{
			{X: 0, Y: 0},
			{X: 0, Y: 1},
			{X: 1, Y: 1},
			{X: 2, Y: 1},
		},
	}

	x, y := actor.RenderPosition(now)
	if x != 0.5 || y != 1 {
		t.Fatalf("position = %.2f, %.2f, want 0.50, 1.00", x, y)
	}
	if got := actor.RenderDirection(now); got != DirectionFromDelta(0, 1, 1, 1, 0) {
		t.Fatalf("direction = %d", got)
	}
}

func TestSetPlayerMovementBuildsPathAroundObstacle(t *testing.T) {
	w := New()
	w.GAT = testGAT(3, 3, map[WalkStep]bool{
		{X: 1, Y: 0}: true,
	})

	w.SetPlayerMovement(0, 0, 2, 0, 0)
	if len(w.Player.MovePath) < 4 {
		t.Fatalf("path too short: %+v", w.Player.MovePath)
	}
	for _, step := range w.Player.MovePath {
		if step.X == 1 && step.Y == 0 {
			t.Fatalf("path crosses blocked cell: %+v", w.Player.MovePath)
		}
	}
	if got := w.Player.MoveDuration; got <= 300*time.Millisecond {
		t.Fatalf("duration = %s, want longer than direct two-cell move", got)
	}
}

func TestSetPlayerMovementStoresFinalSegmentDirection(t *testing.T) {
	w := New()
	w.GAT = testGAT(3, 3, map[WalkStep]bool{
		{X: 1, Y: 0}: true,
	})

	w.SetPlayerMovement(0, 0, 2, 0, DirectionFromDelta(0, 0, 2, 0, 0))

	if got, want := w.Player.Dir, DirectionFromDelta(2, 1, 2, 0, 0); got != want {
		t.Fatalf("player dir = %d, want final segment dir %d; path=%+v", got, want, w.Player.MovePath)
	}
	if w.Dir != w.Player.Dir {
		t.Fatalf("world dir = %d, want player dir %d", w.Dir, w.Player.Dir)
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

func testGAT(width, height int, blocked map[WalkStep]bool) *res.GAT {
	gat := &res.GAT{
		Width:  width,
		Height: height,
		Cells:  make([]res.GATCell, width*height),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cellType := res.GATTypeWalkable | res.GATTypeSnipable
			if blocked[WalkStep{X: x, Y: y}] {
				cellType = res.GATTypeNone
			}
			gat.Cells[y*width+x] = res.GATCell{Type: cellType}
		}
	}
	return gat
}
