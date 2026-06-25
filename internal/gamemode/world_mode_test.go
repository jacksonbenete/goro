package gamemode

import (
	"math"
	"os"
	"path/filepath"
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

	if actor, ok := clickedAttackTarget(ctx, projection, int(npcPoint.x), int(npcPoint.y), now, nil); ok {
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

	actor, ok := clickedAttackTarget(ctx, projection, int(mobPoint.x), int(mobPoint.y), now, nil)
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

func TestApplyParameterChangeUpdatesProgress(t *testing.T) {
	sessionState := &session.Session{
		Selected: session.Character{Level: 4},
		Progress: session.Progress{
			BaseLevel: 4,
		},
	}
	ctx := Context{Session: sessionState}

	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusBaseExp, Value: 123})
	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusNextBaseExp, Value: 1000})
	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusJobExp, Value: 45})
	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusNextJobExp, Value: 500})
	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusBaseLevel, Value: 5})
	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusJobLevel, Value: 3})

	if sessionState.Progress.BaseLevel != 5 || sessionState.Progress.JobLevel != 3 {
		t.Fatalf("levels = base %d job %d", sessionState.Progress.BaseLevel, sessionState.Progress.JobLevel)
	}
	if sessionState.Progress.BaseExp != 123 || sessionState.Progress.NextBaseExp != 1000 || sessionState.Progress.JobExp != 45 || sessionState.Progress.NextJobExp != 500 {
		t.Fatalf("progress = %+v", sessionState.Progress)
	}
	if sessionState.Selected.Level != 5 {
		t.Fatalf("selected level = %d, want 5", sessionState.Selected.Level)
	}
}

func TestSessionProgressFromCharacterUsesBaseLevel(t *testing.T) {
	progress := sessionProgressFromCharacter(session.Character{Level: 12})
	if progress.BaseLevel != 12 {
		t.Fatalf("base level = %d, want 12", progress.BaseLevel)
	}
}

func TestFormatProgressValue(t *testing.T) {
	if got := formatProgressValue(12, 100); got != "12 / 100" {
		t.Fatalf("formatted progress = %q", got)
	}
	if got := formatProgressValue(12, 0); got != "12" {
		t.Fatalf("formatted progress without next = %q", got)
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
	life, ok := mode.actorLife[300]
	if !ok {
		t.Fatal("target life fallback missing")
	}
	if life.hp != 58 || life.maxHP != 100 {
		t.Fatalf("target life = %+v, want 58/100", life)
	}
}

func TestApplyActorHPUpdateStoresExactLife(t *testing.T) {
	mode := &WorldMode{}
	mode.applyActorHPUpdate(network.ActorHPUpdate{ID: 300, HP: 12, MaxHP: 48})

	life, ok := mode.actorLife[300]
	if !ok {
		t.Fatal("life missing")
	}
	if life.hp != 12 || life.maxHP != 48 || life.fromTiny {
		t.Fatalf("life = %+v, want exact 12/48", life)
	}
}

func TestCombatHitDelayUsesActionSoundMotion(t *testing.T) {
	action := res.ACTAction{Animations: []res.ACTAnimation{
		{Sound: -1},
		{Sound: -1},
		{Sound: 0},
		{Sound: -1},
	}}
	if got := combatHitDelayFromAction(action, 800*time.Millisecond); got != 400*time.Millisecond {
		t.Fatalf("hit delay = %s, want 400ms", got)
	}
}

func TestCombatHitDelayFallsBackToMidpoint(t *testing.T) {
	action := res.ACTAction{Animations: []res.ACTAnimation{
		{Sound: -1},
		{Sound: -1},
		{Sound: -1},
		{Sound: -1},
	}}
	if got := combatHitDelayFromAction(action, 800*time.Millisecond); got != 400*time.Millisecond {
		t.Fatalf("hit delay = %s, want midpoint", got)
	}
}

func TestActionSoundNameResolvesACTSound(t *testing.T) {
	act := &res.ACT{Sounds: []string{"attack.wav"}}
	action := res.ACTAction{Animations: []res.ACTAnimation{{Sound: -1}, {Sound: 0}}}
	if got := actionSoundName(act, action, 1); got != "attack.wav" {
		t.Fatalf("sound = %q, want attack.wav", got)
	}
}

func TestApplyActorActionNotifyUsesMobACTHitPhase(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 11, Y: 20, Dir: 4}
	world.UpsertActor(worldstate.Actor{
		ID:            300,
		X:             10,
		Y:             20,
		Dir:           4,
		Job:           1002,
		ObjectType:    actorObjectTypeMob,
		HasObjectType: true,
	})
	mode := &WorldMode{
		nonPCViews: map[int]*playerSpriteView{
			1002: {
				act: &res.ACT{
					Actions: []res.ACTAction{
						{},
						{},
						{Animations: []res.ACTAnimation{
							{Sound: -1},
							{Sound: -1},
							{Sound: 0},
							{Sound: -1},
						}},
					},
					Sounds: []string{"poring_attack.wav"},
				},
			},
		},
	}
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
		SourceID:    300,
		TargetID:    2000000,
		SourceSpeed: 800,
		TargetSpeed: 480,
		Damage:      1,
		Action:      0,
	})

	sourceAnim, ok := mode.actorAnims[300]
	if !ok {
		t.Fatal("source animation missing")
	}
	targetAnim, ok := mode.actorAnims[150000]
	if !ok {
		t.Fatal("local target animation missing")
	}
	if got := targetAnim.started.Sub(sourceAnim.started); got != 400*time.Millisecond {
		t.Fatalf("hit delay = %s, want ACT sound phase", got)
	}
	if len(mode.scheduledSounds) != 2 {
		t.Fatalf("scheduled sounds = %+v, want attack and hit sounds", mode.scheduledSounds)
	}
	if !mode.scheduledSounds[0].at.Equal(targetAnim.started) || mode.scheduledSounds[0].paths[0] != "poring_attack.wav" {
		t.Fatalf("attack sound = %+v targetStarted=%s", mode.scheduledSounds[0], targetAnim.started)
	}
}

func TestApplyActorVanishDeathKeepsMobForDeathAnimation(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 10, Y: 20}
	world.UpsertActor(worldstate.Actor{
		ID:            300,
		X:             11,
		Y:             20,
		Job:           1002,
		ObjectType:    actorObjectTypeMob,
		HasObjectType: true,
	})
	mode := &WorldMode{
		nonPCViews: map[int]*playerSpriteView{
			1002: {
				act: &res.ACT{
					Actions: []res.ACTAction{
						{},
						{},
						{},
						{},
						{Animations: []res.ACTAnimation{{Sound: 0}, {Sound: -1}}, DelayMS: 100},
					},
					Sounds: []string{"poring_die.wav"},
				},
			},
		},
	}
	ctx := Context{
		Session: &session.Session{AccountID: 2000000, CharID: 150000, Selected: session.Character{ID: 150000}},
		World:   world,
	}

	mode.applyActorVanish(ctx, network.ActorVanish{ID: 300, Reason: 1})

	if _, ok := world.Actors[300]; !ok {
		t.Fatal("dead actor was removed immediately")
	}
	anim, ok := mode.actorAnims[300]
	if !ok {
		t.Fatal("death animation missing")
	}
	if anim.actionFamily != spriteActionNonPCDeath {
		t.Fatalf("death action = %d, want %d", anim.actionFamily, spriteActionNonPCDeath)
	}
	if removeAt, ok := mode.actorDeaths[300]; !ok || !removeAt.After(anim.started) {
		t.Fatalf("death removal time = %s ok=%t", removeAt, ok)
	}
	if got := mode.actorDeaths[300].Sub(anim.started); got != nonPCDeathFadeDuration {
		t.Fatalf("death visible duration = %s, want %s", got, nonPCDeathFadeDuration)
	}
	mode.processNonPCMotionSound(ctx, world.Actors[300], anim.started)
	if len(mode.scheduledSounds) != 1 || mode.scheduledSounds[0].paths[0] != "poring_die.wav" {
		t.Fatalf("death sounds = %+v", mode.scheduledSounds)
	}

	mode.cleanupDeadActors(ctx, mode.actorDeaths[300].Add(time.Millisecond))
	if _, ok := world.Actors[300]; ok {
		t.Fatal("dead actor was not removed after death hold")
	}
}

func TestActorDeathAlphaFadesOverVisibleDuration(t *testing.T) {
	started := time.Unix(10, 0)
	mode := &WorldMode{
		actorAnims: map[uint32]actorAnimation{
			300: {actionFamily: spriteActionNonPCDeath, started: started, duration: nonPCDeathFadeDuration},
		},
		actorDeaths: map[uint32]time.Time{
			300: started.Add(nonPCDeathFadeDuration),
		},
	}
	if got := mode.actorDeathAlpha(300, started); got != 1 {
		t.Fatalf("alpha at start = %.2f, want 1", got)
	}
	if got := mode.actorDeathAlpha(300, started.Add(nonPCDeathFadeDuration/2)); math.Abs(got-0.5) > 0.001 {
		t.Fatalf("alpha halfway = %.2f, want 0.5", got)
	}
	if got := mode.actorDeathAlpha(300, started.Add(nonPCDeathFadeDuration)); got != 0 {
		t.Fatalf("alpha at end = %.2f, want 0", got)
	}
}

func TestProcessNonPCMotionSoundSchedulesIdleACTSound(t *testing.T) {
	now := time.Unix(10, 0)
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 10, Y: 20}
	actor := worldstate.Actor{
		ID:            300,
		X:             11,
		Y:             20,
		Job:           1002,
		ObjectType:    actorObjectTypeMob,
		HasObjectType: true,
	}
	world.UpsertActor(actor)
	mode := &WorldMode{
		nonPCViews: map[int]*playerSpriteView{
			1002: {
				started: now,
				act: &res.ACT{
					Actions: []res.ACTAction{
						{Animations: []res.ACTAnimation{{Sound: 0}, {Sound: -1}}, DelayMS: 100},
					},
					Sounds: []string{"poring_idle.wav"},
				},
			},
		},
	}
	ctx := Context{World: world}

	mode.processNonPCMotionSound(ctx, actor, now)
	mode.processNonPCMotionSound(ctx, actor, now)

	if len(mode.scheduledSounds) != 1 {
		t.Fatalf("scheduled sounds = %+v, want one idle sound", mode.scheduledSounds)
	}
	if mode.scheduledSounds[0].paths[0] != "poring_idle.wav" {
		t.Fatalf("idle sound = %+v", mode.scheduledSounds[0])
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

func TestActorBillboardSortDepthUsesTopInCameraProjection(t *testing.T) {
	t.Setenv("GORO_SCENE_PROJECTION", "")
	t.Setenv("GORO_CAMERA_ZOOM", "150")
	t.Setenv("GORO_CAMERA_PITCH", "230")
	t.Setenv("GORO_CAMERA_FOV", "15")
	projection := newSceneProjectionForTarget(800, 600, 10.5, 20.5, 0)
	footDepth := projection.Depth(10.5, 20.5, 0)
	topDepth := projection.Depth(10.5, 20.5, actorBillboardWorldHeightUnit)
	got := actorBillboardSortDepth(projection, 10.5, 20.5, 0)
	want := math.Min(footDepth, topDepth)
	if got != want {
		t.Fatalf("billboard depth = %.4f, want closer of foot %.4f and top %.4f", got, footDepth, topDepth)
	}
	if got >= footDepth {
		t.Fatalf("billboard depth = %.4f, want closer than foot depth %.4f", got, footDepth)
	}
}

func TestActorBillboardSortDepthUsesFootInFlatProjection(t *testing.T) {
	projection := sceneProjection{
		playerX:     10.5,
		playerY:     20.5,
		centerX:     400,
		centerY:     300,
		tileW:       sceneTileW,
		tileH:       sceneTileH,
		heightScale: 2,
	}
	footDepth := projection.Depth(10.5, 20.5, 0)
	if got := actorBillboardSortDepth(projection, 10.5, 20.5, 0); got != footDepth {
		t.Fatalf("flat billboard depth = %.4f, want foot depth %.4f", got, footDepth)
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

func TestCameraYawForIndoorMapIsLocked(t *testing.T) {
	t.Setenv("GORO_CAMERA_YAW", "123")
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.Mkdir(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "indoorrswtable.txt"), []byte("geffen_in.rsw#\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manager, err := res.NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := Context{
		Resources: manager,
		World:     &worldstate.World{MapName: "geffen_in"},
	}
	if got := cameraYawForMap(ctx); got != -45 {
		t.Fatalf("indoor camera yaw = %.1f, want -45.0", got)
	}
	ctx.World.MapName = "prontera"
	if got := cameraYawForMap(ctx); got != 123 {
		t.Fatalf("outdoor camera yaw = %.1f, want env override 123.0", got)
	}
}

func TestSceneLightingFromRSWMatchesReferenceDirection(t *testing.T) {
	lighting := sceneLightingFromRSW(&res.RSW{Light: res.RSWLight{
		Longitude: 45,
		Latitude:  45,
		Diffuse:   [3]float32{1, 1, 1},
		Opacity:   1,
	}})
	want := modelPoint3{x: -0.5, y: -math.Sqrt2 / 2, z: -0.5}
	if math.Abs(lighting.direction.x-want.x) > 0.0001 ||
		math.Abs(lighting.direction.y-want.y) > 0.0001 ||
		math.Abs(lighting.direction.z-want.z) > 0.0001 {
		t.Fatalf("light direction = %+v, want %+v", lighting.direction, want)
	}
}

func TestSortGNDSurfacesDrawsFarBeforeNear(t *testing.T) {
	surfaces := []gndSurfaceDraw{
		{depth: 2},
		{depth: 8},
		{depth: 4},
	}
	sortGNDSurfaces(surfaces)
	got := []float64{surfaces[0].depth, surfaces[1].depth, surfaces[2].depth}
	want := []float64{8, 4, 2}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("surface order = %v, want %v", got, want)
		}
	}
}

func TestGNDDrawBoundsUseCameraFootprint(t *testing.T) {
	t.Setenv("GORO_SCENE_PROJECTION", "")
	t.Setenv("GORO_CAMERA_ZOOM", "150")
	t.Setenv("GORO_CAMERA_PITCH", "230")
	t.Setenv("GORO_CAMERA_FOV", "15")

	gnd := &res.GND{Width: 200, Height: 200}
	projection := newSceneProjectionForTargetYaw(1024, 768, 200.5, 260.5, 8, 0)
	startX, endX, startY, endY, ok := gndDrawBounds(gnd, projection, 1024, 768)
	if !ok {
		t.Fatal("missing GND bounds")
	}
	centerX := gndTileFromWorld(projection.playerX)
	centerY := gndTileFromWorld(projection.playerY)
	if startX > centerX || endX < centerX || startY > centerY || endY < centerY {
		t.Fatalf("bounds %d..%d,%d..%d do not include center %d,%d", startX, endX, startY, endY, centerX, centerY)
	}
	oldSymmetricStartY := centerY - (int(768/sceneTileH) + 12)
	if startY <= oldSymmetricStartY {
		t.Fatalf("camera bounds startY=%d, want tighter than old symmetric startY=%d", startY, oldSymmetricStartY)
	}
}

func TestQuadHasInvalidPointDetectsCameraSentinel(t *testing.T) {
	points := [4]screenPoint{{x: -1 << 20, y: -1 << 20}, {x: 1, y: 1}, {x: 2, y: 1}, {x: 1, y: 2}}
	if !quadHasInvalidPoint(points) {
		t.Fatal("expected camera sentinel point to invalidate GND quad")
	}
}

func TestMapWaterPrefersGNDOverride(t *testing.T) {
	gnd := &res.GND{Water: res.GNDWater{Present: true, Level: -4, Type: 3, WaveHeight: 2, WaveSpeed: 5, WavePitch: 20, AnimSpeed: 6}}
	rsw := &res.RSW{Water: res.RSWWater{Level: -1, Type: 1}}
	water, ok := mapWater(gnd, rsw)
	if !ok {
		t.Fatal("missing water")
	}
	if water.Level != -4 || water.Type != 3 || water.WaveHeight != 2 || water.AnimSpeed != 6 {
		t.Fatalf("water = %+v, want GND override", water)
	}
}

func TestWaterUVsUseFourTileRepeat(t *testing.T) {
	uvs := waterUVs(3, 4)
	if uvs[0] != (texturePoint{u: 0.75, v: 0}) || uvs[2] != (texturePoint{u: 1.0, v: 0.25}) {
		t.Fatalf("water uvs = %+v", uvs)
	}
}

func TestWaterVisibleForCellUsesInvertedHeightConvention(t *testing.T) {
	water := res.RSWWater{Level: -2, WaveHeight: 0.5}
	if !waterVisibleForCell(res.GNDCell{Heights: [4]float32{-3, -2, -2, -2}}, water) {
		t.Fatal("expected water where terrain is below the water threshold")
	}
	if waterVisibleForCell(res.GNDCell{Heights: [4]float32{-1, -1.2, -0.5, 0}}, water) {
		t.Fatal("unexpected water where all terrain vertices are above the water threshold")
	}
}

func TestClickedWalkCellByProjectedPolygonUsesWalkableGATCell(t *testing.T) {
	world := worldstate.New()
	world.Player.X = 5
	world.Player.Y = 5
	world.GAT = &res.GAT{
		Width:  12,
		Height: 12,
		Cells:  make([]res.GATCell, 12*12),
	}
	for i := range world.GAT.Cells {
		world.GAT.Cells[i] = res.GATCell{Type: res.GATTypeWalkable}
	}
	projection := sceneProjection{
		playerX:     5.5,
		playerY:     5.5,
		centerX:     400,
		centerY:     300,
		tileW:       sceneTileW,
		tileH:       sceneTileH,
		heightScale: 2,
	}
	point := projection.Project(6.5, 5.5, 0)

	x, y, ok := clickedWalkCellByProjectedPolygon(Context{World: world}, projection, int(point.x), int(point.y), 0, 11, 0, 11)
	if !ok || x != 6 || y != 5 {
		t.Fatalf("clicked cell = %d,%d ok=%t, want 6,5 true", x, y, ok)
	}
}

func TestClickedWalkCellByProjectedPolygonSkipsBlockedGATCell(t *testing.T) {
	world := worldstate.New()
	world.Player.X = 5
	world.Player.Y = 5
	world.GAT = &res.GAT{
		Width:  12,
		Height: 12,
		Cells:  make([]res.GATCell, 12*12),
	}
	world.GAT.Cells[5*world.GAT.Width+6] = res.GATCell{Type: res.GATTypeNone}
	projection := sceneProjection{
		playerX:     5.5,
		playerY:     5.5,
		centerX:     400,
		centerY:     300,
		tileW:       sceneTileW,
		tileH:       sceneTileH,
		heightScale: 2,
	}
	point := projection.Project(6.5, 5.5, 0)

	_, _, ok := clickedWalkCellByProjectedPolygon(Context{World: world}, projection, int(point.x), int(point.y), 0, 11, 0, 11)
	if ok {
		t.Fatal("blocked cell should not be picked")
	}
}

func TestHoveredWalkCellRequiresProjectedWalkableCell(t *testing.T) {
	world := worldstate.New()
	world.Player.X = 5
	world.Player.Y = 5
	world.GAT = &res.GAT{
		Width:  12,
		Height: 12,
		Cells:  make([]res.GATCell, 12*12),
	}
	for i := range world.GAT.Cells {
		world.GAT.Cells[i] = res.GATCell{Type: res.GATTypeWalkable}
	}
	projection := sceneProjection{
		playerX:     5.5,
		playerY:     5.5,
		centerX:     400,
		centerY:     300,
		tileW:       sceneTileW,
		tileH:       sceneTileH,
		heightScale: 2,
	}
	point := projection.Project(6.5, 5.5, 0)

	x, y, ok := hoveredWalkCell(Context{World: world}, projection, int(point.x), int(point.y))
	if !ok || x != 6 || y != 5 {
		t.Fatalf("hovered cell = %d,%d ok=%t, want 6,5 true", x, y, ok)
	}
	if _, _, ok := hoveredWalkCell(Context{World: world}, projection, -10000, -10000); ok {
		t.Fatal("hover should not fall back to nearest cell outside the projected map")
	}
}

func TestProjectedTileCursorCellUsesGATHeightsWithLift(t *testing.T) {
	gat := &res.GAT{
		Width:  4,
		Height: 4,
		Cells:  make([]res.GATCell, 16),
	}
	gat.Cells[2*gat.Width+1] = res.GATCell{Heights: [4]float32{2, 2, 2, 2}, Type: res.GATTypeWalkable}
	projection := sceneProjection{
		playerX:     1.5,
		playerY:     2.5,
		centerX:     400,
		centerY:     300,
		tileW:       sceneTileW,
		tileH:       sceneTileH,
		heightScale: 2,
	}
	now := time.Unix(0, 0)
	points, ok := projectedTileCursorCell(projection, gat, 1, 2, now)
	if !ok {
		t.Fatal("missing cursor cell")
	}
	ground := projection.Project(1, 2, 2)
	wantY := float64(ground.y) - tileCursorLift(now)*projection.heightScale
	if math.Abs(float64(points[0].y)-wantY) > 0.001 {
		t.Fatalf("cursor point y = %.4f, want %.4f", points[0].y, wantY)
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

func TestActorDisplayNameDoesNotLabelWarpPortal(t *testing.T) {
	ctx := Context{Resources: &res.Manager{}}
	actor := worldstate.Actor{Job: actorJobWarpPortal}

	if got := actorDisplayName(ctx, actor, false); got != "" {
		t.Fatalf("display name = %q, want empty", got)
	}
	if !isWarpActor(actor) {
		t.Fatal("expected warp actor classification")
	}
}
