package gamemode

import (
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
	worldstate "github.com/kivutar/goro/world"
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

	projection := newSceneProjectionForTarget(800, 600, 10.5, 20.5, 5)

	scale := actorBillboardScreenScale(projection, 10.5, 20.5, 5)
	if scale <= 0 || scale >= 1 {
		t.Fatalf("camera billboard scale = %.3f, want between 0 and 1", scale)
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
	if time.Until(mode.pendingAttack.readyAt) > 100*time.Millisecond {
		t.Fatalf("readyAt too far in future: %s", time.Until(mode.pendingAttack.readyAt))
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

func TestApplyParameterChangeUpdatesInventory(t *testing.T) {
	sessionState := &session.Session{}
	ctx := Context{Session: sessionState}

	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusZeny, Value: 1234567})
	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusWeight, Value: 240})
	applyParameterChange(ctx, network.ParameterChange{VarID: network.StatusMaxWeight, Value: 2000})

	if sessionState.Inventory.Zeny != 1234567 || sessionState.Inventory.Weight != 240 || sessionState.Inventory.MaxWeight != 2000 {
		t.Fatalf("inventory = %+v", sessionState.Inventory)
	}
}

func TestSetSelectedCharacterSeedsInventoryZeny(t *testing.T) {
	sessionState := &session.Session{}

	setSelectedCharacter(sessionState, session.Character{ID: 1234, Money: 95000})

	if sessionState.Inventory.Zeny != 95000 {
		t.Fatalf("zeny = %d, want 95000", sessionState.Inventory.Zeny)
	}
}

func TestFormatHUDNumberGroupsThousands(t *testing.T) {
	if got := formatHUDNumber(123456789); got != "123,456,789" {
		t.Fatalf("formatted number = %q", got)
	}
}

func TestSessionProgressFromCharacterUsesBaseLevel(t *testing.T) {
	progress := sessionProgressFromCharacter(session.Character{Level: 12, JobLevel: 7})
	if progress.BaseLevel != 12 || progress.JobLevel != 7 {
		t.Fatalf("progress = %+v", progress)
	}
}

func TestFormatEXPPercent(t *testing.T) {
	if got := formatEXPPercent(123, 1000); got != "12.3%" {
		t.Fatalf("formatted exp percent = %q", got)
	}
	if got := formatEXPPercent(12, 0); got != "--" {
		t.Fatalf("formatted exp percent without next = %q", got)
	}
	if got := formatEXPPercent(1001, 1000); got != "100.0%" {
		t.Fatalf("formatted capped exp percent = %q", got)
	}
}

func TestDisplayWeightUsesROVisibleUnits(t *testing.T) {
	if got := displayWeight(240); got != 24 {
		t.Fatalf("display weight = %d, want 24", got)
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

func TestApplyActorActionNotifyUpdatesLocalSitState(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 10, Y: 20, Moving: true}
	mode := &WorldMode{}
	ctx := Context{
		Session: &session.Session{AccountID: 2000000, CharID: 150000},
		World:   world,
	}

	mode.applyActorActionNotify(ctx, network.ActorActionNotify{
		SourceID: 2000000,
		Action:   network.ActionSitDown,
	})
	if !world.Player.Sitting {
		t.Fatal("local player did not sit")
	}
	if world.Player.Moving {
		t.Fatal("local player kept moving while sitting")
	}

	mode.applyActorActionNotify(ctx, network.ActorActionNotify{
		SourceID: 2000000,
		Action:   network.ActionStandUp,
	})
	if world.Player.Sitting {
		t.Fatal("local player did not stand")
	}
}

func TestApplyActorActionNotifyUpdatesRemoteSitState(t *testing.T) {
	world := worldstate.New()
	world.UpsertActor(worldstate.Actor{ID: 300, X: 10, Y: 20, Moving: true})
	mode := &WorldMode{}
	ctx := Context{
		Session: &session.Session{AccountID: 2000000, CharID: 150000},
		World:   world,
	}

	mode.applyActorActionNotify(ctx, network.ActorActionNotify{
		SourceID: 300,
		Action:   network.ActionSitDown,
	})
	if actor := world.Actors[300]; !actor.Sitting || actor.Moving {
		t.Fatalf("remote actor sit state = sitting %t moving %t", actor.Sitting, actor.Moving)
	}

	mode.applyActorActionNotify(ctx, network.ActorActionNotify{
		SourceID: 300,
		Action:   network.ActionStandUp,
	})
	if actor := world.Actors[300]; actor.Sitting {
		t.Fatalf("remote actor stayed sitting: %+v", actor)
	}
}

func TestApplyItemPickupAckRemovesRequestedItemAndStartsPickupAnimation(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 10, Y: 20, Dir: 4}
	world.UpsertItem(worldstate.FloorItem{ID: 9001, ItemID: 909, X: 11, Y: 20, Amount: 1})
	mode := &WorldMode{
		pickupReqItemID: 9001,
		actorAnims:      make(map[uint32]actorAnimation),
	}
	ctx := Context{
		Session: &session.Session{
			AccountID: 2000000,
			CharID:    150000,
		},
		World: world,
	}

	mode.applyItemPickupAck(ctx, network.ItemPickupAck{ItemID: 909, Amount: 1})

	if _, ok := world.Items[9001]; ok {
		t.Fatal("picked item should be removed locally after pickup ack")
	}
	anim, ok := mode.actorAnims[150000]
	if !ok {
		t.Fatal("local pickup animation missing")
	}
	if anim.actionFamily != spriteActionPickup {
		t.Fatalf("pickup action = %d, want %d", anim.actionFamily, spriteActionPickup)
	}
	if mode.pickupReqItemID != 0 {
		t.Fatalf("pickup request item id = %d, want cleared", mode.pickupReqItemID)
	}
	if world.Dir != worldstate.DirectionFromDelta(10, 20, 11, 20, 4) {
		t.Fatalf("player dir = %d", world.Dir)
	}
}

func TestApplyActorPickupActionNotifyStartsPickupInsteadOfAttack(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 10, Y: 20, Dir: 4}
	world.UpsertItem(worldstate.FloorItem{ID: 9001, ItemID: 909, X: 11, Y: 20, Amount: 1})
	mode := &WorldMode{actorAnims: make(map[uint32]actorAnimation)}
	ctx := Context{
		Session: &session.Session{
			AccountID: 2000000,
			CharID:    150000,
		},
		World: world,
	}

	mode.applyActorActionNotify(ctx, network.ActorActionNotify{
		SourceID: 2000000,
		TargetID: 9001,
		Action:   network.ActorActionPickupItem,
	})

	anim, ok := mode.actorAnims[150000]
	if !ok {
		t.Fatal("local pickup animation missing")
	}
	if anim.actionFamily != spriteActionPickup {
		t.Fatalf("pickup action = %d, want %d", anim.actionFamily, spriteActionPickup)
	}
	if len(mode.damageFloaters) != 0 {
		t.Fatalf("pickup notify should not create damage floaters: %+v", mode.damageFloaters)
	}
	if world.Dir != worldstate.DirectionFromDelta(10, 20, 11, 20, 4) {
		t.Fatalf("player dir = %d", world.Dir)
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

func TestCombatLifeFallbackDoesNotSubtractRawDamageFromTinyHPGauge(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 10, Y: 20, Dir: 4}
	world.UpsertActor(worldstate.Actor{
		ID:            300,
		Job:           1008,
		X:             11,
		Y:             20,
		HasObjectType: true,
		ObjectType:    actorObjectTypeMob,
	})
	mode := &WorldMode{}
	ctx := Context{
		Session: &session.Session{AccountID: 2000000, CharID: 150000},
		World:   world,
	}

	mode.applyActorHPUpdate(network.ActorHPUpdate{ID: 300, HP: 95, MaxHP: 100, Tiny: true})
	mode.applyActorActionNotify(ctx, network.ActorActionNotify{
		SourceID:    2000000,
		TargetID:    300,
		SourceSpeed: 580,
		TargetSpeed: 480,
		Damage:      42,
		Action:      0,
	})

	life, ok := mode.actorLife[300]
	if !ok {
		t.Fatal("life missing")
	}
	if life.hp != 95 || life.maxHP != 100 || !life.fromTiny {
		t.Fatalf("tiny life = %+v, want unchanged 95/100", life)
	}
}

func TestCombatLifeFallbackSubtractsRawDamageFromExactHPGauge(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 10, Y: 20, Dir: 4}
	world.UpsertActor(worldstate.Actor{
		ID:            300,
		Job:           1002,
		X:             11,
		Y:             20,
		HasObjectType: true,
		ObjectType:    actorObjectTypeMob,
	})
	mode := &WorldMode{}
	ctx := Context{
		Session: &session.Session{AccountID: 2000000, CharID: 150000},
		World:   world,
	}

	mode.applyActorHPUpdate(network.ActorHPUpdate{ID: 300, HP: 50, MaxHP: 100})
	mode.applyActorActionNotify(ctx, network.ActorActionNotify{
		SourceID:    2000000,
		TargetID:    300,
		SourceSpeed: 580,
		TargetSpeed: 480,
		Damage:      12,
		Action:      0,
	})

	life, ok := mode.actorLife[300]
	if !ok {
		t.Fatal("life missing")
	}
	if life.hp != 38 || life.maxHP != 100 || life.fromTiny {
		t.Fatalf("exact life = %+v, want 38/100", life)
	}
}

func TestActorLifeForDisplayUsesLocalPlayerHPAndSP(t *testing.T) {
	mode := &WorldMode{}
	ctx := Context{
		Session: &session.Session{
			AccountID: 2000000,
			CharID:    150000,
			Vitals: session.Vitals{
				HP:    75,
				MaxHP: 100,
				SP:    8,
				MaxSP: 20,
			},
		},
	}

	life, ok := mode.actorLifeForDisplay(ctx, worldstate.Actor{ID: 150000})
	if !ok {
		t.Fatal("local player life missing")
	}
	if life.hp != 75 || life.maxHP != 100 || life.sp != 8 || life.maxSP != 20 || !life.hasSP || !life.player {
		t.Fatalf("local player life = %+v", life)
	}
}

func TestActorOverlayLifeBarIsBelowNameLabel(t *testing.T) {
	nameY := actorNameLabelY(100, 1.2)
	barY := actorLifeBarY(100, 1.2)
	if barY <= nameY+10 {
		t.Fatalf("bar y = %.1f, name y = %.1f; want bar below name", barY, nameY)
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
	if anim.holdFinal {
		t.Fatal("non-player death animation should not hold forever")
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

func TestLocalDeathAnimationHoldsUntilPlayerAlive(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 2000000, X: 10, Y: 20, Dir: 4}
	mode := &WorldMode{}
	ctx := Context{
		Session: &session.Session{
			AccountID: 2000000,
			CharID:    150000,
			Selected:  session.Character{ID: 150000, Job: 0, HP: 0},
			Vitals:    session.Vitals{HP: 0},
		},
		World: world,
	}

	mode.startActorDeath(ctx, 150000)

	if !mode.deathModal.open {
		t.Fatal("death modal should open for local death")
	}
	anim, ok := mode.actorAnims[150000]
	if !ok {
		t.Fatal("character death animation missing")
	}
	if anim.actionFamily != spriteActionPCDeath {
		t.Fatalf("death action = %d, want %d", anim.actionFamily, spriteActionPCDeath)
	}
	if !anim.holdFinal {
		t.Fatal("local death animation should hold final frame")
	}
	accountAnim, ok := mode.actorAnims[2000000]
	if !ok || !accountAnim.holdFinal || accountAnim.actionFamily != spriteActionPCDeath {
		t.Fatalf("account death animation = %+v ok=%t", accountAnim, ok)
	}
	if held, ok := mode.actorAnimation(150000, anim.started.Add(anim.duration+time.Second)); !ok || held.actionFamily != spriteActionPCDeath {
		t.Fatalf("expired local death animation = %+v ok=%t", held, ok)
	}

	ctx.Session.Vitals.HP = 1
	mode.clearLocalDeathStateIfAlive(ctx)

	if mode.deathModal.open {
		t.Fatal("death modal should clear when player is alive")
	}
	if _, ok := mode.actorAnims[150000]; ok {
		t.Fatal("character death animation should clear when player is alive")
	}
	if _, ok := mode.actorAnims[2000000]; ok {
		t.Fatal("account death animation should clear when player is alive")
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
	world := worldstate.New()
	world.Player = worldstate.Actor{X: 10, Y: 20}
	ctx := Context{World: world}

	camera := followCamera{}
	now := time.Now()
	camera.Update(ctx, now)
	world.Player = worldstate.Actor{X: 14, Y: 20}
	camera.Update(ctx, now.Add(time.Second/60))

	if math.Abs(camera.x-10.9) > 0.001 || camera.y != 20.5 {
		t.Fatalf("camera target = %.2f, %.2f, want 10.9, 20.5", camera.x, camera.y)
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

	entries := appendActorDrawEntry(nil, world, projection, actor, false, now, 800, 600)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if got, want := entries[0].actor.Dir, worldstate.DirectionFromDelta(0, 1, 1, 1, 4); got != want {
		t.Fatalf("entry direction = %d, want %d", got, want)
	}
}

func TestActorBillboardSortDepthUsesTopInCameraProjection(t *testing.T) {
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

func TestCameraFollowFactorUsesReferenceDefault(t *testing.T) {
	if got := cameraFollowFactor(); got != defaultCameraFollowFactor {
		t.Fatalf("camera follow factor = %.2f, want %.2f", got, defaultCameraFollowFactor)
	}
}

func TestCameraYawForIndoorMapIsLocked(t *testing.T) {
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
	if got := cameraYawForMap(ctx); got != defaultSceneCameraYaw {
		t.Fatalf("outdoor camera yaw = %.1f, want %.1f", got, defaultSceneCameraYaw)
	}
}

func TestCameraYawForFixedViewPointMapIsLocked(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.Mkdir(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "viewpointtable.txt"), []byte("fixed_view.rsw#150#50#170#30#30#30#60#30#45#\nfree_view.rsw#150#50#170#-360#360#0#60#30#45#\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manager, err := res.NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := Context{
		Resources: manager,
		World:     &worldstate.World{MapName: "fixed_view"},
	}
	if got := cameraYawForMap(ctx); got != -30 {
		t.Fatalf("fixed viewpoint yaw = %.1f, want -30.0", got)
	}
	if !cameraRotationLockedForMap(ctx) {
		t.Fatal("fixed viewpoint should lock camera rotation")
	}
	ctx.World.MapName = "free_view"
	if got := cameraYawForMap(ctx); got != defaultSceneCameraYaw {
		t.Fatalf("free viewpoint yaw = %.1f, want %.1f", got, defaultSceneCameraYaw)
	}
	if cameraRotationLockedForMap(ctx) {
		t.Fatal("free viewpoint should not lock camera rotation")
	}
}

func TestFollowCameraProjectionIncludesRuntimeYawOffset(t *testing.T) {
	world := worldstate.New()
	world.MapName = "prontera"
	ctx := Context{World: world}
	camera := followCamera{initialized: true, x: 10.5, y: 20.5, z: 0}

	camera.Rotate(90)
	projection := camera.Projection(ctx, 800, 600, time.Now())
	if got := projection.cameraYaw; got != 90 {
		t.Fatalf("projection yaw = %.1f, want 90.0", got)
	}

	camera.ResetRotation()
	projection = camera.Projection(ctx, 800, 600, time.Now())
	if got := projection.cameraYaw; got != defaultSceneCameraYaw {
		t.Fatalf("reset projection yaw = %.1f, want %.1f", got, defaultSceneCameraYaw)
	}
}

func TestFollowCameraProjectionKeepsIndoorBaseYaw(t *testing.T) {
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
	camera := followCamera{initialized: true, x: 10.5, y: 20.5, z: 0}

	camera.Rotate(90)
	projection := camera.Projection(ctx, 800, 600, time.Now())
	if got := projection.cameraYaw; got != -45 {
		t.Fatalf("indoor projection yaw = %.1f, want -45.0", got)
	}
	if camera.yawOffset != 0 {
		t.Fatalf("indoor projection left yaw offset = %.1f, want reset", camera.yawOffset)
	}
}

func TestCameraRotationIsDisabledOnIndoorMap(t *testing.T) {
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
	inputState := input.NewState()
	inputState.SetMousePosition(100, 100)
	inputState.SetMouseButton(input.MouseButtonRight, true)
	inputState.SetMousePosition(200, 100)
	mode := &WorldMode{}
	mode.camera.Rotate(90)
	ctx := Context{
		Resources: manager,
		World:     &worldstate.World{MapName: "geffen_in"},
		Input:     inputState,
		ScreenW:   800,
		ScreenH:   600,
	}

	mode.updateCameraRotation(ctx)
	if mode.camera.yawOffset != 0 {
		t.Fatalf("indoor camera yaw offset = %.1f, want reset", mode.camera.yawOffset)
	}
}

func TestCameraDragYawDeltaMatchesRobrowserScale(t *testing.T) {
	if got := cameraDragYawDelta(100, 1000); got != -72 {
		t.Fatalf("drag yaw delta = %.1f, want -72.0", got)
	}
	if got := cameraDragYawDelta(-100, 1000); got != 72 {
		t.Fatalf("reverse drag yaw delta = %.1f, want 72.0", got)
	}
	if got := cameraDragYawDelta(100, 0); got != 0 {
		t.Fatalf("zero width drag yaw delta = %.1f, want 0", got)
	}
}

func TestCameraWheelZoomFactorZoomsInOnWheelUp(t *testing.T) {
	if got := cameraWheelZoomFactor(1); got >= 1 {
		t.Fatalf("wheel up factor = %.3f, want zoom-in factor below 1", got)
	}
	if got := cameraWheelZoomFactor(-1); got <= 1 {
		t.Fatalf("wheel down factor = %.3f, want zoom-out factor above 1", got)
	}
}

func TestCameraWheelZoomDeltaMatchesRobrowserStep(t *testing.T) {
	if got := cameraWheelZoomDelta(1); got != -15 {
		t.Fatalf("wheel up delta = %.1f, want -15", got)
	}
	if got := cameraWheelZoomDelta(-2); got != 30 {
		t.Fatalf("wheel down delta = %.1f, want 30", got)
	}
}

func TestCameraPinchZoomFactorZoomsInWhenFingersSpread(t *testing.T) {
	if got := cameraPinchZoomFactor(25); got >= 1 {
		t.Fatalf("pinch spread factor = %.3f, want zoom-in factor below 1", got)
	}
	if got := cameraPinchZoomFactor(-25); got <= 1 {
		t.Fatalf("pinch close factor = %.3f, want zoom-out factor above 1", got)
	}
}

func TestFollowCameraZoomIsClampedAndProjected(t *testing.T) {
	world := worldstate.New()
	ctx := Context{World: world}
	camera := followCamera{initialized: true, x: 10.5, y: 20.5, z: 0}

	camera.ZoomBy(0.1)
	if got := camera.currentZoom(); got != defaultCameraMinZoom {
		t.Fatalf("zoom in clamp = %.1f, want %.1f", got, defaultCameraMinZoom)
	}
	camera.ZoomBy(10)
	if got := camera.currentZoom(); got != defaultCameraMaxZoom {
		t.Fatalf("zoom out clamp = %.1f, want %.1f", got, defaultCameraMaxZoom)
	}
	projection := camera.Projection(ctx, 800, 600, time.Now())
	if got := projection.cameraZoom; got != defaultCameraMaxZoom {
		t.Fatalf("projection zoom = %.1f, want %.1f", got, defaultCameraMaxZoom)
	}
}

func TestCursorRotateInfoMatchesRobrowser(t *testing.T) {
	info := cursorInfo(cursorActionRotate)
	if info.delayMult != 1 {
		t.Fatalf("rotate cursor info = %+v", info)
	}
}

func TestWorldSceneClearColorMatchesReferenceDefaults(t *testing.T) {
	if got := worldSceneClearColor("geffen_in"); got != (color.RGBA{A: 255}) {
		t.Fatalf("default map clear color = %#v, want black", got)
	}
	if got := worldSceneClearColor("data/yuno.gat"); got != (color.RGBA{R: 0x99, G: 0xcc, B: 0xff, A: 255}) {
		t.Fatalf("yuno clear color = %#v", got)
	}
	if got := worldSceneClearColor("5@tower.rsw"); got != (color.RGBA{R: 0x33, G: 0x00, B: 0x33, A: 255}) {
		t.Fatalf("tower clear color = %#v", got)
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

func TestSceneLightingModelScaleIgnoresOpacityLikeRobrowser(t *testing.T) {
	opaque := sceneLightingFromRSW(&res.RSW{Light: res.RSWLight{
		Longitude: 45,
		Latitude:  45,
		Diffuse:   [3]float32{0.3, 0.3, 0.3},
		Ambient:   [3]float32{0.4, 0.4, 0.4},
		Opacity:   1,
	}})
	half := sceneLightingFromRSW(&res.RSW{Light: res.RSWLight{
		Longitude: 45,
		Latitude:  45,
		Diffuse:   [3]float32{0.3, 0.3, 0.3},
		Ambient:   [3]float32{0.4, 0.4, 0.4},
		Opacity:   0.5,
	}})
	normal := modelPoint3{x: 0, y: 1, z: 0}
	if got, want := half.modelScale(normal), opaque.modelScale(normal); got != want {
		t.Fatalf("model scale changed with opacity: got %+v want %+v", got, want)
	}
}

func TestSmoothGNDTopNormalsKeepsFlatTilesUniform(t *testing.T) {
	gnd := testGNDWithTopHeights(2, 2, func(_, _ int) [4]float32 {
		return [4]float32{0, 0, 0, 0}
	})
	normals := buildSmoothGNDTopNormals(gnd)
	if len(normals) != 4 {
		t.Fatalf("normal count = %d, want 4", len(normals))
	}
	center := normals[0]
	for i := 1; i < 4; i++ {
		if !modelPointNear(center[0], center[i], 0.0001) {
			t.Fatalf("flat tile normals differ: %v vs %v", center[0], center[i])
		}
	}
}

func TestSmoothGNDTopNormalsVariesSlopedTileCorners(t *testing.T) {
	gnd := testGNDWithTopHeights(3, 3, func(x, y int) [4]float32 {
		return [4]float32{
			float32(x*y + y),
			float32(x*3 + y*y + 2),
			float32(x*x + y*4 + 1),
			float32(x*x + y*y + x*y + 7),
		}
	})
	normals := buildSmoothGNDTopNormals(gnd)
	tile := normals[1+1*gnd.Width]
	same := true
	for i := 1; i < 4; i++ {
		if !modelPointNear(tile[0], tile[i], 0.0001) {
			same = false
		}
	}
	if same {
		t.Fatalf("sloped tile normals are all equal: %+v", tile)
	}
}

func TestSurfaceVertexTintsUsePerVertexNormals(t *testing.T) {
	lighting := sceneLightingFromRSW(&res.RSW{Light: res.RSWLight{
		Longitude: 45,
		Latitude:  45,
		Diffuse:   [3]float32{1, 1, 1},
		Ambient:   [3]float32{0, 0, 0},
		Opacity:   1,
	}})
	tints := surfaceVertexTints(nil, res.GNDSurface{}, uniformGNDSurfaceBaseTints(color.RGBA{}), [4]int{0, 1, 2, 3}, [4]float32{}, [4]modelPoint3{
		{x: -0.5, y: -math.Sqrt2 / 2, z: -0.5},
		{x: 0.5, y: -math.Sqrt2 / 2, z: -0.5},
		{x: 0.5, y: -math.Sqrt2 / 2, z: 0.5},
		{x: -0.5, y: -math.Sqrt2 / 2, z: 0.5},
	}, lighting)
	if tints[0] == tints[1] && tints[0] == tints[2] && tints[0] == tints[3] {
		t.Fatalf("vertex tints are uniform: %+v", tints)
	}
}

func TestTopGNDSurfaceBaseTintsUseNeighborTileColors(t *testing.T) {
	gnd := &res.GND{
		Width:  2,
		Height: 2,
		Surfaces: []res.GNDSurface{
			{Color: color.RGBA{R: 10, G: 20, B: 30, A: 255}},
			{Color: color.RGBA{R: 40, G: 50, B: 60, A: 255}},
			{Color: color.RGBA{R: 70, G: 80, B: 90, A: 255}},
			{Color: color.RGBA{R: 100, G: 110, B: 120, A: 255}},
		},
		Cells: []res.GNDCell{
			{Top: 0, Front: -1, Right: -1},
			{Top: 1, Front: -1, Right: -1},
			{Top: 2, Front: -1, Right: -1},
			{Top: 3, Front: -1, Right: -1},
		},
	}
	tints := topGNDSurfaceBaseTints(gnd, 0, 0, color.RGBA{})
	want := [4]color.RGBA{
		{R: 10, G: 20, B: 30, A: 255},
		{R: 40, G: 50, B: 60, A: 255},
		{R: 100, G: 110, B: 120, A: 255},
		{R: 70, G: 80, B: 90, A: 255},
	}
	if tints != want {
		t.Fatalf("top GND tints = %+v, want %+v", tints, want)
	}
}

func testGNDWithTopHeights(width, height int, fn func(x, y int) [4]float32) *res.GND {
	gnd := &res.GND{
		Width:    width,
		Height:   height,
		Cells:    make([]res.GNDCell, width*height),
		Surfaces: []res.GNDSurface{{TextureID: 0, LightmapID: -1}},
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gnd.Cells[x+y*width] = res.GNDCell{Top: 0, Front: -1, Right: -1, Heights: fn(x, y)}
		}
	}
	return gnd
}

func modelPointNear(a, b modelPoint3, epsilon float64) bool {
	return math.Abs(a.x-b.x) <= epsilon && math.Abs(a.y-b.y) <= epsilon && math.Abs(a.z-b.z) <= epsilon
}

func TestGNDDrawBoundsUseCameraFootprint(t *testing.T) {
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
	if endX-startX >= gnd.Width-1 || endY-startY >= gnd.Height-1 {
		t.Fatalf("camera bounds %d..%d,%d..%d should not cover the full map", startX, endX, startY, endY)
	}
}

func TestGNDShadowMapPointMatchesROBrowserCellCenterMapping(t *testing.T) {
	x, y := gndShadowMapPoint(10, 20)
	if x != 42 || y != 82 {
		t.Fatalf("shadow map point for even cell = %d,%d, want 42,82", x, y)
	}
	x, y = gndShadowMapPoint(11, 21)
	if x != 46 || y != 86 {
		t.Fatalf("shadow map point for odd cell = %d,%d, want 46,86", x, y)
	}
}

func TestActorShadowFactorAveragesGNDLightmapAlpha(t *testing.T) {
	var lightmap res.GNDLightmap
	for y := range lightmap.Alpha {
		for x := range lightmap.Alpha[y] {
			lightmap.Alpha[y][x] = 128
		}
	}
	gnd := &res.GND{
		Width:     4,
		Height:    4,
		Lightmaps: []res.GNDLightmap{lightmap},
		Surfaces:  []res.GNDSurface{{LightmapID: 0}},
		Cells:     make([]res.GNDCell, 16),
	}
	for i := range gnd.Cells {
		gnd.Cells[i].Top = 0
	}

	got := actorShadowFactor(&worldstate.World{GND: gnd}, 3, 3)
	want := float64(128) / 255
	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("shadow factor = %.4f, want %.4f", got, want)
	}
}

func TestActorShadowFactorDefaultsToLitWithoutGroundLightmap(t *testing.T) {
	if got := actorShadowFactor(nil, 3, 3); got != 1 {
		t.Fatalf("nil world shadow = %.2f, want 1", got)
	}
	gnd := &res.GND{
		Width:    1,
		Height:   1,
		Surfaces: []res.GNDSurface{{LightmapID: 0}},
		Cells:    []res.GNDCell{{Top: -1}},
	}
	if got := actorShadowFactor(&worldstate.World{GND: gnd}, 0, 0); got != 1 {
		t.Fatalf("missing top surface shadow = %.2f, want 1", got)
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
	projection := newSceneProjectionForTarget(800, 600, 5.5, 5.5, 0)
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
	projection := newSceneProjectionForTarget(800, 600, 5.5, 5.5, 0)
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
	projection := newSceneProjectionForTarget(800, 600, 5.5, 5.5, 0)
	point := projection.Project(6.5, 5.5, 0)

	x, y, ok := hoveredWalkCell(Context{World: world}, projection, int(point.x), int(point.y))
	if !ok || x != 6 || y != 5 {
		t.Fatalf("hovered cell = %d,%d ok=%t, want 6,5 true", x, y, ok)
	}
	if _, _, ok := hoveredWalkCell(Context{World: world}, projection, -10000, -10000); ok {
		t.Fatal("hover should not fall back to nearest cell outside the projected map")
	}
}

func TestTileCursorCellVertsUseGATHeightsWithLift(t *testing.T) {
	gat := &res.GAT{
		Width:  4,
		Height: 4,
		Cells:  make([]res.GATCell, 16),
	}
	gat.Cells[2*gat.Width+1] = res.GATCell{Heights: [4]float32{2, 2, 2, 2}, Type: res.GATTypeWalkable}
	now := time.Unix(0, 0)
	verts, ok := tileCursorCellVerts(gat, 1, 2, now)
	if !ok {
		t.Fatal("missing cursor cell")
	}
	wantY := 2 + tileCursorLift(now)
	if math.Abs(verts[0].y-wantY) > 0.001 {
		t.Fatalf("cursor vertex y = %.4f, want %.4f", verts[0].y, wantY)
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

func TestHoveredActorDisplayNameUsesServerNameForNPC(t *testing.T) {
	ctx := Context{Resources: &res.Manager{}}
	mode := &WorldMode{}
	actor := worldstate.Actor{
		Job:           84,
		ObjectType:    actorObjectTypeNPC,
		HasObjectType: true,
	}

	if got := mode.hoveredActorDisplayName(ctx, actor, time.Now()); got != "4 M 02" {
		t.Fatalf("hovered NPC name = %q, want resource fallback", got)
	}
	actor.Name = "Kafra Employee#izlude"
	if got := mode.hoveredActorDisplayName(ctx, actor, time.Now()); got != "Kafra Employee" {
		t.Fatalf("hovered NPC server name = %q, want Kafra Employee", got)
	}
}

func TestHoveredActorDisplayNameUsesServerNameForMonster(t *testing.T) {
	ctx := Context{Resources: &res.Manager{}}
	mode := &WorldMode{}
	actor := worldstate.Actor{
		Job:           1002,
		ObjectType:    actorObjectTypeMob,
		HasObjectType: true,
	}

	if got := mode.hoveredActorDisplayName(ctx, actor, time.Now()); got != "Poring" {
		t.Fatalf("hovered monster name = %q, want resource fallback", got)
	}
	actor.Name = "Poring"
	if got := mode.hoveredActorDisplayName(ctx, actor, time.Now()); got != "Poring" {
		t.Fatalf("hovered monster server name = %q, want Poring", got)
	}
}

func TestFormatConsoleMessageUsesMsgStringTable(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "msgstringtable.txt"), []byte("ignored#\nYou got %d items.#\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manager, err := res.NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	got := formatConsoleMessage(manager, network.ChatMessage{MessageID: 1, Value: 3})
	if got != "You got 3 items." {
		t.Fatalf("message = %q", got)
	}
}

func TestFormatPickupConsoleMessageUsesMsgStringAndItemName(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	msgTable := strings.Repeat("ignored#\n", 153) + "You got %s %d.#\n"
	if err := os.WriteFile(filepath.Join(dataDir, "msgstringtable.txt"), []byte(msgTable), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "idnum2itemdisplaynametable.txt"), []byte("938#Apple#\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manager, err := res.NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	got := formatPickupConsoleMessage(manager, network.ItemPickupAck{ItemID: 938, Amount: 2, Identified: true})
	if got != "You got Apple 2." {
		t.Fatalf("pickup message = %q", got)
	}
}

func TestFormatPickupConsoleMessageFallback(t *testing.T) {
	got := formatPickupConsoleMessage(nil, network.ItemPickupAck{ItemID: 938, Amount: 0})
	if got != "You got item 938 1." {
		t.Fatalf("pickup message = %q", got)
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

func TestMapFadeAlphaTransitionsThroughBlack(t *testing.T) {
	start := time.Unix(100, 0)
	mode := &WorldMode{}
	mode.startMapFadeOut(network.MapChange{MapName: "geffen"}, start)

	if got := mode.mapFadeAlpha(start); got != 0 {
		t.Fatalf("fade-out start alpha = %d, want 0", got)
	}
	if got := mode.mapFadeAlpha(start.Add(mapFadeOutDuration)); got != 255 {
		t.Fatalf("fade-out end alpha = %d, want 255", got)
	}

	mode.mapFade = mapFadeState{phase: mapFadeHold, started: start}
	if got := mode.mapFadeAlpha(start.Add(time.Second)); got != 255 {
		t.Fatalf("hold alpha = %d, want 255", got)
	}

	mode.startMapFadeIn(start)
	if got := mode.mapFadeAlpha(start); got != 255 {
		t.Fatalf("fade-in start alpha = %d, want 255", got)
	}
	if got := mode.mapFadeAlpha(start.Add(mapFadeInDuration)); got != 0 {
		t.Fatalf("fade-in end alpha = %d, want 0", got)
	}
}

func TestApplyInventoryItemListReplacesExistingAmount(t *testing.T) {
	sessionState := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{{Index: 7, ItemID: 938, Amount: 3}},
		},
	}
	ctx := Context{Session: sessionState}

	applyInventoryItemList(ctx, []network.InventoryItem{{
		Index:      7,
		ItemID:     938,
		Type:       3,
		Identified: true,
		Amount:     5,
	}})

	if len(sessionState.Inventory.Items) != 1 {
		t.Fatalf("inventory item count = %d, want 1", len(sessionState.Inventory.Items))
	}
	if got := sessionState.Inventory.Items[0]; got.Amount != 5 || !got.Identified || got.Type != 3 {
		t.Fatalf("inventory item = %+v, want replaced amount/type", got)
	}
}

func TestInventoryItemDeleteDecrementsAndRemoves(t *testing.T) {
	sessionState := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{{Index: 7, ItemID: 938, Amount: 3}},
		},
	}
	ctx := Context{Session: sessionState}

	applyInventoryItemDelete(ctx, network.InventoryItemDelete{Index: 7, Amount: 2})
	if got := sessionState.Inventory.Items[0].Amount; got != 1 {
		t.Fatalf("amount after partial delete = %d, want 1", got)
	}
	applyInventoryItemDelete(ctx, network.InventoryItemDelete{Index: 7, Amount: 1})
	if len(sessionState.Inventory.Items) != 0 {
		t.Fatalf("inventory item count = %d, want 0", len(sessionState.Inventory.Items))
	}
}

func TestPickedInventoryItemAddsToExistingStack(t *testing.T) {
	sessionState := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{{Index: 7, ItemID: 512, Type: 0, Amount: 3, Identified: true}},
		},
	}

	addPickedSessionInventoryItem(sessionState, session.InventoryItem{Index: 7, ItemID: 512, Type: 3, Amount: 2})

	if got := sessionState.Inventory.Items[0].Amount; got != 5 {
		t.Fatalf("picked stack amount = %d, want 5", got)
	}
	if got := sessionState.Inventory.Items[0].Type; got != 0 {
		t.Fatalf("picked stack type = %d, want preserved healing type", got)
	}
}

func TestShopAcceptInventoryDropAddsSellableItem(t *testing.T) {
	window := shopWindowState{
		open: true,
		mode: shopModeSell,
		x:    100,
		y:    100,
		sellable: map[uint16]network.ShopSellItem{
			7: {Index: 7, Price: 10, OverchargePrice: 12},
		},
	}

	ok := window.acceptInventoryDrop(Context{}, session.InventoryItem{Index: 7, ItemID: 938, Amount: 3}, 120, 150)
	if !ok {
		t.Fatal("drop was not accepted")
	}
	if len(window.cart) != 1 || window.cart[0].amount != 1 || window.cart[0].max != 3 || window.cart[0].over != 12 {
		t.Fatalf("cart = %+v", window.cart)
	}
}

func TestInventoryDragReleaseOverShopAddsCartItem(t *testing.T) {
	inputState := input.NewState()
	inputState.SetMousePosition(120, 150)
	sessionState := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{{Index: 7, ItemID: 938, Amount: 3}},
		},
	}
	ctx := Context{Input: inputState, Session: sessionState}
	inventory := inventoryWindowState{
		open:       true,
		positioned: true,
		x:          500,
		y:          100,
		dragActive: true,
		dragItem:   session.InventoryItem{Index: 7, ItemID: 938, Amount: 3},
	}
	shop := shopWindowState{
		open: true,
		mode: shopModeSell,
		x:    100,
		y:    100,
		sellable: map[uint16]network.ShopSellItem{
			7: {Index: 7, Price: 10, OverchargePrice: 12},
		},
	}

	if !inventory.update(ctx, &shop) {
		t.Fatal("inventory update did not consume drag release")
	}
	if inventory.dragActive {
		t.Fatal("drag still active after release")
	}
	if len(shop.cart) != 1 || shop.cart[0].item.Index != 7 {
		t.Fatalf("shop cart = %+v, want dropped item", shop.cart)
	}
}

func TestShopBuyCartTracksQuantityAndTotal(t *testing.T) {
	window := shopWindowState{mode: shopModeBuy}
	item := network.ShopBuyItem{ItemID: 501, Price: 100, DiscountPrice: 80}

	window.addBuyItem(item)
	window.addBuyItem(item)
	if got := window.buyAmount(501); got != 2 {
		t.Fatalf("buy amount = %d, want 2", got)
	}
	if got := window.total(); got != 160 {
		t.Fatalf("total = %d, want 160", got)
	}

	window.decrementBuyItem(501)
	if got := window.buyAmount(501); got != 1 {
		t.Fatalf("buy amount after decrement = %d, want 1", got)
	}
}

func TestInventoryBagClassifiesTabs(t *testing.T) {
	tests := []struct {
		name string
		item session.InventoryItem
		tab  int
	}{
		{name: "healing item", item: session.InventoryItem{Type: 0}, tab: inventoryBagTabItem},
		{name: "usable item", item: session.InventoryItem{Type: 2}, tab: inventoryBagTabItem},
		{name: "equipment flag", item: session.InventoryItem{Type: 4, Equip: true}, tab: inventoryBagTabEquip},
		{name: "weapon type", item: session.InventoryItem{Type: 5}, tab: inventoryBagTabEquip},
		{name: "pet egg type", item: session.InventoryItem{Type: 7}, tab: inventoryBagTabEquip},
		{name: "etc", item: session.InventoryItem{Type: 3}, tab: inventoryBagTabEtc},
		{name: "card", item: session.InventoryItem{Type: 6}, tab: inventoryBagTabEtc},
		{name: "ammo", item: session.InventoryItem{Type: 10}, tab: inventoryBagTabEtc},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := inventoryItemTab(tc.item); got != tc.tab {
				t.Fatalf("tab = %d, want %d", got, tc.tab)
			}
		})
	}
}

func TestSessionItemFromNetworkMarksEquipmentByType(t *testing.T) {
	item := sessionItemFromNetwork(network.InventoryItem{
		Index:      7,
		ItemID:     1201,
		Type:       5,
		Location:   0x0002,
		Identified: true,
		Amount:     1,
	})
	if !item.Equip || inventoryItemTab(item) != inventoryBagTabEquip {
		t.Fatalf("item = %+v, want equipment tab item", item)
	}
}

func TestInventoryItemListReplacesDifferentItemAtReusedIndex(t *testing.T) {
	sessionState := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{{
				Index:    11,
				ItemID:   1201,
				Type:     5,
				Location: 0x0002,
				Amount:   1,
				Equip:    true,
			}},
		},
	}
	ctx := Context{Session: sessionState}

	applyInventoryItemList(ctx, []network.InventoryItem{{
		Index:  11,
		ItemID: 938,
		Type:   3,
		Amount: 2,
	}})

	item := sessionState.Inventory.Items[0]
	if item.Equip || item.Location != 0 || item.Type != 3 || item.ItemID != 938 || inventoryItemTab(item) != inventoryBagTabEtc {
		t.Fatalf("item = %+v, want clean replacement", item)
	}
}

func TestPickedEquipmentKeepsEquipMetadata(t *testing.T) {
	sessionState := &session.Session{}
	addPickedSessionInventoryItem(sessionState, session.InventoryItem{
		Index:      11,
		ItemID:     1201,
		Type:       5,
		Location:   0x0002,
		Identified: true,
		Amount:     1,
		Equip:      inventoryItemTypeIsEquipment(5),
	})
	if len(sessionState.Inventory.Items) != 1 {
		t.Fatalf("item count = %d, want 1", len(sessionState.Inventory.Items))
	}
	item := sessionState.Inventory.Items[0]
	if !item.Equip || item.Location != 0x0002 || inventoryItemTab(item) != inventoryBagTabEquip {
		t.Fatalf("picked item = %+v, want equipment metadata", item)
	}
}

func TestApplyInventoryEquipAckUpdatesEquippedState(t *testing.T) {
	sessionState := &session.Session{
		Inventory: session.Inventory{
			Items: []session.InventoryItem{
				{Index: 1, ItemID: 1201, Type: 4, Location: 0x0002, Equip: true},
				{Index: 2, ItemID: 1202, Type: 4, Location: 0x0002, Equip: true, Equipped: true},
			},
		},
	}
	ctx := Context{Session: sessionState}

	applyInventoryEquipAck(ctx, network.InventoryEquipAck{Index: 1, Location: 0x0002, Success: true})
	if !sessionState.Inventory.Items[0].Equipped {
		t.Fatal("equipped item was not marked equipped")
	}
	if sessionState.Inventory.Items[1].Equipped {
		t.Fatal("previous item in same location stayed equipped")
	}

	applyInventoryEquipAck(ctx, network.InventoryEquipAck{Index: 1, Location: 0x0002, Success: true, Unequip: true})
	if sessionState.Inventory.Items[0].Equipped {
		t.Fatal("unequipped item stayed equipped")
	}
}
