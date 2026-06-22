package gamemode

import (
	"testing"
	"time"

	"github.com/kivutar/goro/internal/network"
	"github.com/kivutar/goro/internal/res"
	"github.com/kivutar/goro/internal/session"
	worldstate "github.com/kivutar/goro/internal/world"
)

func TestApplyLocalActorLookChangeUpdatesSelectedCharacter(t *testing.T) {
	sessionState := &session.Session{
		AccountID: 100,
		CharID:    200,
		Selected:  session.Character{ID: 200, Job: 0, Hair: 1},
		Characters: []session.Character{
			{ID: 200, Job: 0, Hair: 1},
		},
	}
	ctx := Context{
		Session: sessionState,
		World:   worldstate.New(),
	}
	look := network.ActorLookChange{
		ID:    200,
		Type:  2,
		Value: uint32(2101)<<16 | 1201,
	}

	if !applyActorLookChange(ctx, look) {
		t.Fatal("local look change should request player sprite reload")
	}
	if sessionState.Selected.Weapon != 1201 || sessionState.Selected.Shield != 2101 {
		t.Fatalf("selected appearance = weapon %d shield %d", sessionState.Selected.Weapon, sessionState.Selected.Shield)
	}
	if sessionState.Characters[0].Weapon != 1201 || sessionState.Characters[0].Shield != 2101 {
		t.Fatalf("character appearance = weapon %d shield %d", sessionState.Characters[0].Weapon, sessionState.Characters[0].Shield)
	}
	if ctx.World.Player.Weapon != 1201 || ctx.World.Player.Shield != 2101 {
		t.Fatalf("world player appearance = weapon %d shield %d", ctx.World.Player.Weapon, ctx.World.Player.Shield)
	}
}

func TestApplyRemoteActorLookChangeUpdatesWorldActor(t *testing.T) {
	world := worldstate.New()
	world.UpsertActor(worldstate.Actor{ID: 300, Job: 0, Head: 1, Appearance: true})
	ctx := Context{
		Session: &session.Session{AccountID: 100, CharID: 200},
		World:   world,
	}

	if applyActorLookChange(ctx, network.ActorLookChange{ID: 300, Type: 4, Value: 7}) {
		t.Fatal("remote look change should not request local player sprite reload")
	}
	actor := world.Actors[300]
	if actor.HeadTop != 7 {
		t.Fatalf("remote head top = %d, want 7", actor.HeadTop)
	}
}

func TestDirectionFromDeltaUsesRathenaDirectionOrder(t *testing.T) {
	cases := []struct {
		name string
		toX  int
		toY  int
		want int
	}{
		{name: "north", toX: 10, toY: 11, want: 0},
		{name: "northwest", toX: 9, toY: 11, want: 1},
		{name: "west", toX: 9, toY: 10, want: 2},
		{name: "southwest", toX: 9, toY: 9, want: 3},
		{name: "south", toX: 10, toY: 9, want: 4},
		{name: "southeast", toX: 11, toY: 9, want: 5},
		{name: "east", toX: 11, toY: 10, want: 6},
		{name: "northeast", toX: 11, toY: 11, want: 7},
		{name: "long mostly north", toX: 11, toY: 20, want: 0},
		{name: "long mostly west", toX: 0, toY: 11, want: 2},
		{name: "long mostly south", toX: 11, toY: 0, want: 4},
		{name: "long mostly east", toX: 20, toY: 11, want: 6},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := directionFromDelta(10, 10, tt.toX, tt.toY, 6); got != tt.want {
				t.Fatalf("directionFromDelta = %d, want %d", got, tt.want)
			}
		})
	}
	if got := directionFromDelta(10, 10, 10, 10, -1); got != 7 {
		t.Fatalf("stationary fallback = %d, want 7", got)
	}
}

func TestActorBillboardScreenScaleUsesProjectedReferenceHeight(t *testing.T) {
	if actorBillboardWorldHeightUnit != 5 {
		t.Fatalf("actor billboard world height = %.1f, want 5.0", actorBillboardWorldHeightUnit)
	}

	projection := sceneProjection{
		screenW:        800,
		screenH:        600,
		camera:         true,
		viewProjection: sceneCameraMatrix(800, 600, 10.5, 20.5, 5),
	}

	scale := actorBillboardScreenScale(projection, 10.5, 20.5, 5)
	if scale <= 0 || scale >= 1 {
		t.Fatalf("camera billboard scale = %.3f, want between 0 and 1", scale)
	}
}

func TestActorBillboardScreenScaleKeepsFlatProjectionNative(t *testing.T) {
	projection := sceneProjection{}
	if got := actorBillboardScreenScale(projection, 10.5, 20.5, 5); got != 1 {
		t.Fatalf("flat billboard scale = %.3f, want 1", got)
	}
}

func TestClickedAttackTargetPicksMobOnly(t *testing.T) {
	now := time.Now()
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 200, X: 10, Y: 20}
	world.UpsertActor(worldstate.Actor{
		ID:            300,
		X:             11,
		Y:             20,
		ObjectType:    6,
		HasObjectType: true,
	})
	ctx := Context{
		Session: &session.Session{AccountID: 100, CharID: 200},
		World:   world,
	}
	projection := newSceneProjectionForTarget(800, 600, cellCenter(10), cellCenter(20), 0)
	npcPoint := projection.Project(cellCenter(11), cellCenter(20), 0)

	if actor, ok := clickedAttackTarget(ctx, projection, int(npcPoint.x), int(npcPoint.y), now); ok {
		t.Fatalf("npc should not be attack-clickable: %+v", actor)
	}

	world.UpsertActor(worldstate.Actor{
		ID:            400,
		X:             12,
		Y:             20,
		ObjectType:    5,
		HasObjectType: true,
	})
	mobPoint := projection.Project(cellCenter(12), cellCenter(20), 0)

	actor, ok := clickedAttackTarget(ctx, projection, int(mobPoint.x), int(mobPoint.y), now)
	if !ok {
		t.Fatal("expected mob hit")
	}
	if actor.ID != 400 {
		t.Fatalf("target id = %d, want 400", actor.ID)
	}
}

func TestAttackTargetWithinRangeUsesMeleeAdjacency(t *testing.T) {
	if !attackTargetWithinRange(10, 20, 11, 21) {
		t.Fatal("diagonal adjacent target should be in melee range")
	}
	if attackTargetWithinRange(10, 20, 12, 20) {
		t.Fatal("two cells away should be out of melee range")
	}
}

func TestAttackApproachCellChoosesClosestWalkableNeighbor(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{X: 112, Y: 302}
	world.GAT = &res.GAT{
		Width:  200,
		Height: 400,
		Cells:  make([]res.GATCell, 200*400),
	}
	for i := range world.GAT.Cells {
		world.GAT.Cells[i] = res.GATCell{Type: res.GATTypeWalkable}
	}
	ctx := Context{World: world}
	actor := worldstate.Actor{ID: 300, X: 116, Y: 303}

	x, y, ok := attackApproachCell(ctx, actor)
	if !ok {
		t.Fatal("expected approach cell")
	}
	if x != 115 || y != 302 {
		t.Fatalf("approach = %d,%d, want 115,302", x, y)
	}

	world.GAT.Cells[302*world.GAT.Width+115] = res.GATCell{}
	x, y, ok = attackApproachCell(ctx, actor)
	if !ok {
		t.Fatal("expected fallback approach cell")
	}
	if x == 115 && y == 302 {
		t.Fatalf("blocked approach cell was selected")
	}
}

func TestContinuePendingAttackSchedulesDelayedAction(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{X: 10, Y: 20}
	world.UpsertActor(worldstate.Actor{
		ID:            300,
		X:             11,
		Y:             20,
		ObjectType:    actorObjectTypeMob,
		HasObjectType: true,
	})
	mode := &WorldMode{
		pendingAttack: attackIntent{
			targetID: 300,
			expires:  time.Now().Add(time.Second),
		},
	}
	ctx := Context{
		Session: &session.Session{AccountID: 100, CharID: 200},
		World:   world,
	}

	mode.continuePendingAttack(ctx, "test")

	if mode.pendingAttack.targetID != 300 {
		t.Fatalf("pending target cleared")
	}
	if mode.pendingAttack.readyAt.IsZero() {
		t.Fatal("pending attack was not scheduled")
	}
	if mode.pendingAttack.readyAt.Sub(time.Now()) > 100*time.Millisecond {
		t.Fatalf("readyAt too far in future: %s", mode.pendingAttack.readyAt.Sub(time.Now()))
	}
}

func TestPendingAttackReadyAtWaitsForWalkEnd(t *testing.T) {
	now := time.Unix(100, 0)
	player := worldstate.Actor{
		Moving:       true,
		MoveStarted:  now.Add(-100 * time.Millisecond),
		MoveDuration: 600 * time.Millisecond,
	}

	got := pendingAttackReadyAt(player, now)
	want := now.Add(560 * time.Millisecond)
	if !got.Equal(want) {
		t.Fatalf("readyAt = %s, want %s", got.Sub(now), want.Sub(now))
	}
}

func TestAttackRetryDueUsesOpenMidgardInterval(t *testing.T) {
	now := time.Unix(100, 0)
	if !attackRetryDue(time.Time{}, now) {
		t.Fatal("zero last attack should be due")
	}
	if attackRetryDue(now.Add(-attackRetryInterval+time.Millisecond), now) {
		t.Fatal("attack should not retry before the interval")
	}
	if !attackRetryDue(now.Add(-attackRetryInterval), now) {
		t.Fatal("attack should retry at the interval")
	}
}

func TestLockAttackKeepsExistingRetryTimersForSameTarget(t *testing.T) {
	firstAttack := time.Unix(100, 0)
	firstChase := time.Unix(101, 0)
	mode := &WorldMode{
		lockedAttackID: 300,
		lastAttackAt:   firstAttack,
		lastChaseAt:    firstChase,
	}

	mode.lockAttack(300)
	if mode.lastAttackAt != firstAttack || mode.lastChaseAt != firstChase {
		t.Fatal("same target lock reset retry timers")
	}

	mode.lockAttack(400)
	if mode.lockedAttackID != 400 {
		t.Fatalf("locked target = %d, want 400", mode.lockedAttackID)
	}
	if !mode.lastAttackAt.IsZero() || !mode.lastChaseAt.IsZero() {
		t.Fatal("new target lock should reset retry timers")
	}
}

func TestApplyParameterChangeUpdatesVitals(t *testing.T) {
	sessionState := &session.Session{
		Selected: session.Character{HP: 70, MaxHP: 100, SP: 20, MaxSP: 30},
		Vitals:   session.Vitals{HP: 70, MaxHP: 100, SP: 20, MaxSP: 30},
	}
	ctx := Context{Session: sessionState}

	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusHP, Value: 42})
	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusMaxSP, Value: 55})

	if sessionState.Vitals.HP != 42 || sessionState.Vitals.MaxHP != 100 || sessionState.Vitals.SP != 20 || sessionState.Vitals.MaxSP != 55 {
		t.Fatalf("vitals = %+v", sessionState.Vitals)
	}
	if sessionState.Selected.HP != 42 || sessionState.Selected.MaxSP != 55 {
		t.Fatalf("selected vitals = hp %d maxsp %d", sessionState.Selected.HP, sessionState.Selected.MaxSP)
	}
}

func TestApplyActorActionNotifySchedulesAttackAndHitAnimations(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 10, Y: 20, Dir: 4}
	world.UpsertActor(worldstate.Actor{
		ID:            300,
		X:             11,
		Y:             20,
		Job:           1002,
		ObjectType:    actorObjectTypeMob,
		HasObjectType: true,
	})
	mode := &WorldMode{}
	ctx := Context{
		Session: &session.Session{
			AccountID: 2000000,
			CharID:    150000,
			Sex:       0,
			Selected:  session.Character{ID: 150000, Job: 0, Hair: 1, Weapon: 1201},
		},
		World: world,
	}

	mode.applyActorActionNotify(ctx, network.ActorActionNotify{
		SourceID:    2000000,
		TargetID:    300,
		SourceSpeed: 580,
		TargetSpeed: 480,
		Damage:      42,
		Action:      0,
	})

	sourceAnim, ok := mode.actorAnims[150000]
	if !ok {
		t.Fatal("local character animation missing")
	}
	if sourceAnim.actionFamily != spriteActionPCAttack3 {
		t.Fatalf("source action = %d, want %d", sourceAnim.actionFamily, spriteActionPCAttack3)
	}
	targetAnim, ok := mode.actorAnims[300]
	if !ok {
		t.Fatal("target animation missing")
	}
	if targetAnim.actionFamily != spriteActionNonPCHurt {
		t.Fatalf("target action = %d, want %d", targetAnim.actionFamily, spriteActionNonPCHurt)
	}
	if targetAnim.started.Sub(sourceAnim.started) != 580*time.Millisecond {
		t.Fatalf("hit delay = %s, want 580ms", targetAnim.started.Sub(sourceAnim.started))
	}
	if world.Dir != worldstate.DirectionFromDelta(10, 20, 11, 20, 4) {
		t.Fatalf("player dir = %d", world.Dir)
	}
	if len(mode.damageFloaters) != 1 || !mode.damageFloaters[0].starts.Equal(targetAnim.started) {
		t.Fatalf("damage floater = %+v targetStarted=%s", mode.damageFloaters, targetAnim.started)
	}
}

func TestFollowCameraInitializesToRenderedPlayerPosition(t *testing.T) {
	now := time.Now()
	world := worldstate.New()
	world.Player = worldstate.Actor{
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
	ctx := Context{World: world}

	camera := followCamera{}
	camera.Update(ctx, now)

	if camera.x != 15.5 || camera.y != 25.5 {
		t.Fatalf("camera target = %.2f, %.2f, want rendered player center 15.5, 25.5", camera.x, camera.y)
	}
	if world.Camera.X != camera.x || world.Camera.Y != camera.y {
		t.Fatalf("world camera = %.2f, %.2f, want %.2f, %.2f", world.Camera.X, world.Camera.Y, camera.x, camera.y)
	}
}

func TestFollowCameraInterpolatesTowardPlayerLikeReferenceView(t *testing.T) {
	t.Setenv("GORO_CAMERA_FOLLOW_FACTOR", "0.25")
	world := worldstate.New()
	world.Player = worldstate.Actor{X: 10, Y: 20}
	ctx := Context{World: world}

	camera := followCamera{}
	now := time.Now()
	camera.Update(ctx, now)
	world.Player = worldstate.Actor{X: 14, Y: 20}
	camera.Update(ctx, now.Add(time.Second/60))

	if camera.x != 11.5 || camera.y != 20.5 {
		t.Fatalf("camera target = %.2f, %.2f, want 11.5, 20.5", camera.x, camera.y)
	}
}

func TestAppendActorDrawEntryUsesPathRenderDirection(t *testing.T) {
	now := time.Now()
	world := worldstate.New()
	actor := worldstate.Actor{
		X:            2,
		Y:            1,
		Dir:          4,
		Moving:       true,
		FromX:        0,
		FromY:        0,
		ToX:          2,
		ToY:          1,
		MoveStarted:  now.Add(-225 * time.Millisecond),
		MoveDuration: 450 * time.Millisecond,
		MovePath: []worldstate.WalkStep{
			{X: 0, Y: 0},
			{X: 0, Y: 1},
			{X: 1, Y: 1},
			{X: 2, Y: 1},
		},
	}
	projection := newSceneProjectionForTarget(800, 600, 0.5, 1.5, 0)

	entries := appendActorDrawEntry(nil, world, projection, actor, "actor", false, now, 800, 600)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if got, want := entries[0].actor.Dir, worldstate.DirectionFromDelta(0, 1, 1, 1, 4); got != want {
		t.Fatalf("entry direction = %d, want %d", got, want)
	}
}

func TestCameraFollowFactorIsClamped(t *testing.T) {
	t.Setenv("GORO_CAMERA_FOLLOW_FACTOR", "2")
	if got := cameraFollowFactor(); got != 1 {
		t.Fatalf("camera follow factor = %.2f, want 1", got)
	}
	t.Setenv("GORO_CAMERA_FOLLOW_FACTOR", "-1")
	if got := cameraFollowFactor(); got != 0 {
		t.Fatalf("camera follow factor = %.2f, want 0", got)
	}
}

func TestApplyActorNameAckUpdatesWorldActor(t *testing.T) {
	world := worldstate.New()
	world.UpsertActor(worldstate.Actor{ID: 300, Job: 1002})
	ctx := Context{
		Session: &session.Session{AccountID: 100, CharID: 200},
		World:   world,
	}

	applyActorNameAck(ctx, network.ActorNameAck{ID: 300, Name: "Guide#prontera"})

	if got := world.Actors[300].Name; got != "Guide" {
		t.Fatalf("actor name = %q, want Guide", got)
	}
}

func TestApplyActorNameAckUpdatesLocalPlayer(t *testing.T) {
	world := worldstate.New()
	ctx := Context{
		Session: &session.Session{AccountID: 100, CharID: 200},
		World:   world,
	}

	applyActorNameAck(ctx, network.ActorNameAck{ID: 200, Name: "Kivutar"})

	if got := world.Player.Name; got != "Kivutar" {
		t.Fatalf("player name = %q, want Kivutar", got)
	}
}

func TestHandleMapChangeSameServerUpdatesMapAndResetsActors(t *testing.T) {
	world := worldstate.New()
	world.MapName = "prontera"
	world.SetPlayerPosition(10, 20, 4)
	world.UpsertActor(worldstate.Actor{ID: 300, Name: "Remote", X: 11, Y: 20})
	sessionState := &session.Session{AccountID: 100, CharID: 200, PlayerDir: 4}
	ctx := Context{
		Session: sessionState,
		World:   world,
	}

	next := (&WorldMode{}).handleMapChange(ctx, network.MapChange{MapName: "geffen", X: 120, Y: 80})

	if next == nil || next.Name() != "world" {
		t.Fatalf("next mode = %#v, want world", next)
	}
	if world.MapName != "geffen" || sessionState.Zone.MapName != "geffen" {
		t.Fatalf("map = world %q session %q, want geffen", world.MapName, sessionState.Zone.MapName)
	}
	if world.Player.X != 120 || world.Player.Y != 80 || sessionState.PlayerX != 120 || sessionState.PlayerY != 80 {
		t.Fatalf("position = world %d,%d session %d,%d", world.Player.X, world.Player.Y, sessionState.PlayerX, sessionState.PlayerY)
	}
	if len(world.Actors) != 0 {
		t.Fatalf("actors were not cleared: %+v", world.Actors)
	}
}

func TestHandleMapChangeSameLoadedMapReusesModeAndSnapsCamera(t *testing.T) {
	world := worldstate.New()
	world.MapName = "izlude"
	world.GND = &res.GND{}
	world.SetPlayerPosition(10, 20, 4)
	world.UpsertActor(worldstate.Actor{ID: 300, Name: "Remote", X: 11, Y: 20})
	sessionState := &session.Session{AccountID: 100, CharID: 200, PlayerDir: 4}
	ctx := Context{
		Session: sessionState,
		World:   world,
	}
	mode := &WorldMode{}

	next := mode.handleMapChange(ctx, network.MapChange{MapName: "izlude", X: 114, Y: 145})

	if next != nil {
		t.Fatalf("next mode = %#v, want nil same-mode reuse", next)
	}
	if world.Player.X != 114 || world.Player.Y != 145 || sessionState.PlayerX != 114 || sessionState.PlayerY != 145 {
		t.Fatalf("position = world %d,%d session %d,%d", world.Player.X, world.Player.Y, sessionState.PlayerX, sessionState.PlayerY)
	}
	if len(world.Actors) != 0 {
		t.Fatalf("actors were not cleared: %+v", world.Actors)
	}
	if !mode.camera.initialized || mode.camera.x != 114.5 || mode.camera.y != 145.5 {
		t.Fatalf("camera = initialized %t %.2f,%.2f, want 114.5,145.5", mode.camera.initialized, mode.camera.x, mode.camera.y)
	}
}

func TestActorDisplayNameUsesSelectedCharacterForPlayer(t *testing.T) {
	ctx := Context{Session: &session.Session{CharID: 200, Selected: session.Character{ID: 200, Name: "Kivutar"}}}

	if got := actorDisplayName(ctx, worldstate.Actor{Name: "Player"}, true); got != "Kivutar" {
		t.Fatalf("display name = %q, want Kivutar", got)
	}
}

func TestActorDisplayNameUsesServerNameBeforeFallback(t *testing.T) {
	ctx := Context{Resources: &res.Manager{}}
	actor := worldstate.Actor{Name: "Kafra Employee#izlude", Job: 1002}

	if got := actorDisplayName(ctx, actor, false); got != "Kafra Employee" {
		t.Fatalf("display name = %q, want Kafra Employee", got)
	}
}

func TestActorDisplayNameFallsBackToNonPCResource(t *testing.T) {
	ctx := Context{Resources: &res.Manager{}}
	actor := worldstate.Actor{Job: 1002}

	if got := actorDisplayName(ctx, actor, false); got != "Poring" {
		t.Fatalf("display name = %q, want Poring", got)
	}
}

func TestActorDisplayNameDoesNotLabelUnnamedPlayerJob(t *testing.T) {
	ctx := Context{Resources: &res.Manager{}}
	actor := worldstate.Actor{Job: 0}

	if got := actorDisplayName(ctx, actor, false); got != "" {
		t.Fatalf("display name = %q, want empty", got)
	}
}
