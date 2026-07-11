package game

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/db"
)

type actorEffectStateEffect struct {
	bit      uint32
	effectID int
}

var actorEffectStateEffects = []actorEffectStateEffect{
	{bit: db.EffectStateSight, effectID: effectSight},
	{bit: db.EffectStateRuwach, effectID: effectRuwach},
}

func (m *WorldMode) applyActorEffectStateEffects(ctx client.Context, actorID uint32, oldState, newState uint32) {
	if actorID == 0 {
		return
	}
	for _, stateEffect := range actorEffectStateEffects {
		if oldState&stateEffect.bit != 0 {
			m.removeWorldEffect(stateEffect.effectID, actorID)
		}
		if newState&stateEffect.bit != 0 {
			m.addWorldEffect(ctx, stateEffect.effectID, actorID)
		}
	}
}

func (m *WorldMode) syncCurrentActorEffectStateEffects(ctx client.Context) {
	if ctx.World == nil {
		return
	}
	if ctx.World.Player.HasState {
		m.applyActorEffectStateEffects(ctx, ctx.World.Player.ID, 0, ctx.World.Player.EffectState)
	}
	for _, actor := range ctx.World.Actors {
		if actor.HasState {
			m.applyActorEffectStateEffects(ctx, actor.ID, 0, actor.EffectState)
		}
	}
}

func (m *WorldMode) removeActorEffectStateEffects(actorID uint32) {
	for _, stateEffect := range actorEffectStateEffects {
		m.removeWorldEffect(stateEffect.effectID, actorID)
	}
}
