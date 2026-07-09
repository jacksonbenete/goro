package game

import (
	"image/color"
	"log"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

const (
	actorBodyStateStone     uint16 = 1
	actorBodyStateFreeze    uint16 = 2
	actorBodyStateStun      uint16 = 3
	actorBodyStateSleep     uint16 = 4
	actorBodyStateStoneWait uint16 = 6

	actorHealthPoison uint16 = 0x0001
	actorHealthCurse  uint16 = 0x0002
	actorHealthBlind  uint16 = 0x0010
)

func (m *WorldMode) applyActorStateChange(ctx client.Context, change network.ActorStateChange) {
	if change.ID == 0 || ctx.World == nil {
		return
	}
	if isLocalActor(ctx, change.ID) {
		setActorRenderState(&ctx.World.Player, change.BodyState, change.HealthState, change.EffectState)
		log.Printf("actor state local id=%d body=%d health=0x%04X effect=0x%08X", change.ID, change.BodyState, change.HealthState, change.EffectState)
		return
	}
	actor, ok := ctx.World.Actors[change.ID]
	if !ok {
		return
	}
	setActorRenderState(&actor, change.BodyState, change.HealthState, change.EffectState)
	ctx.World.UpsertActor(actor)
	log.Printf("actor state id=%d body=%d health=0x%04X effect=0x%08X", change.ID, change.BodyState, change.HealthState, change.EffectState)
}

func (m *WorldMode) applyActorBladeStop(ctx client.Context, blade network.ActorBladeStop) {
	if blade.SourceID == 0 || blade.TargetID == 0 || ctx.World == nil {
		return
	}
	m.applyActorBladeStopSide(ctx, blade.SourceID, blade.TargetID, blade.Active)
	m.applyActorBladeStopSide(ctx, blade.TargetID, blade.SourceID, blade.Active)
	log.Printf("actor blade stop src=%d target=%d active=%t", blade.SourceID, blade.TargetID, blade.Active)
}

func (m *WorldMode) applyActorBladeStopSide(ctx client.Context, actorID, lookID uint32, active bool) {
	actor, ok, local := actorForCombatID(ctx, actorID)
	if !ok {
		return
	}
	target, targetOK, _ := actorForCombatID(ctx, lookID)
	if targetOK {
		actor.Dir = directionFromDelta(actor.X, actor.Y, target.X, target.Y, actor.Dir)
		if local {
			ctx.World.Player.Dir = actor.Dir
			ctx.World.Dir = actor.Dir
		} else {
			ctx.World.UpsertActor(actor)
		}
	}
	action := spriteActionIdle
	if active {
		action = readyFightActionFamily(actor)
	}
	anim := actorAnimation{
		actionFamily: action,
		started:      time.Now(),
		play:         true,
		hasPlay:      true,
		loop:         active,
	}
	if local && ctx.Session != nil {
		m.setActorAction(ctx, ctx.Session.AccountID, anim)
		m.setActorAction(ctx, ctx.Session.CharID, anim)
		return
	}
	m.setActorAction(ctx, actorID, anim)
}

func setActorRenderState(actor *worldstate.Actor, bodyState, healthState uint16, effectState uint32) {
	actor.BodyState = bodyState
	actor.HealthState = healthState
	actor.EffectState = effectState
	actor.HasState = true
}

func applyActorBodyState(actor worldstate.Actor, state *spriteState) {
	switch actor.BodyState {
	case actorBodyStateFreeze:
		state.actionFamily = freezeActionFamily(actor)
		state.moving = false
		state.loop = false
		state.loopIdle = false
		state.play = false
		state.hasPlay = true
		state.fixedMotion = 0
		state.hasFixedMotion = true
	case actorBodyStateStone:
		state.moving = false
		state.loop = false
		state.loopIdle = false
		state.play = false
		state.hasPlay = true
	}
}

func actorStateTint(actor worldstate.Actor) color.RGBA {
	r, g, b := 1.0, 1.0, 1.0
	switch actor.BodyState {
	case actorBodyStateStone:
		r, g, b = 0.1, 0.1, 0.1
	case actorBodyStateStoneWait:
		r, g, b = 0.3, 0.3, 0.3
	case actorBodyStateFreeze:
		r, g, b = 0.0, 0.4, 0.8
	}
	if actor.HealthState&actorHealthCurse != 0 {
		r *= 0.5
		g *= 0.15
		b *= 0.1
	}
	if actor.HealthState&actorHealthPoison != 0 {
		r *= 0.9
		g *= 0.4
		b *= 0.8
	}
	if actor.HealthState&actorHealthBlind != 0 {
		r *= 0.2
		g *= 0.2
		b *= 0.2
	}
	return color.RGBA{R: byte(clampUnit(r) * 255), G: byte(clampUnit(g) * 255), B: byte(clampUnit(b) * 255), A: 255}
}

func freezeActionFamily(actor worldstate.Actor) int {
	if res.HasPlayerJobToken(int(actor.Job)) {
		return spriteActionPCFreeze2
	}
	return spriteActionIdle
}

func readyFightActionFamily(actor worldstate.Actor) int {
	if res.HasPlayerJobToken(int(actor.Job)) {
		return spriteActionPCReadyFight
	}
	return spriteActionIdle
}
