package game

import (
	"log"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func (m *WorldMode) applyStatusEffectChange(ctx client.Context, change network.StatusEffectChange) {
	if ctx.Session == nil || change.StatusID == 0xFFFF {
		return
	}
	if m.applyPushCartStatus(ctx, change) {
		return
	}
	m.applyActorEffectStateStatus(ctx, change)
	m.applyTrickDeadStatus(ctx, change)
	localID := localSkillTarget(ctx)
	if change.ActorID != 0 && localID != 0 && change.ActorID != localID && change.ActorID != ctx.Session.CharID {
		return
	}
	if ctx.Session.Statuses.Active == nil {
		ctx.Session.Statuses.Active = make(map[uint16]session.StatusEffect)
	}
	m.addStatusEffectTransition(ctx, change)
	if !change.Active {
		delete(ctx.Session.Statuses.Active, change.StatusID)
		log.Printf("status effect inactive id=%d actor=%d", change.StatusID, change.ActorID)
		return
	}
	now := time.Now()
	effect := session.StatusEffect{
		ID:          change.StatusID,
		Source:      change.ActorID,
		StartedAt:   now,
		HasDuration: change.HasDuration,
	}
	if change.HasDuration {
		effect.ExpiresAt = now.Add(change.Duration)
	}
	ctx.Session.Statuses.Active[change.StatusID] = effect
	log.Printf("status effect active id=%d actor=%d duration_ms=%d", change.StatusID, change.ActorID, change.Duration.Milliseconds())
}

func (m *WorldMode) applyActorEffectStateStatus(ctx client.Context, change network.StatusEffectChange) {
	bit, ok := actorEffectStateBitForStatus(change.StatusID)
	if !ok || ctx.World == nil {
		return
	}
	id := change.ActorID
	if id == 0 {
		id = localSkillTarget(ctx)
	}
	if id == 0 {
		return
	}
	actor, ok, local := actorForCombatID(ctx, id)
	if !ok {
		return
	}
	oldState := actor.EffectState
	if change.Active {
		actor.EffectState |= bit
	} else {
		actor.EffectState &^= bit
	}
	actor.HasState = true
	m.applyActorEffectStateEffects(ctx, id, oldState, actor.EffectState)
	if local {
		ctx.World.Player.EffectState = actor.EffectState
		ctx.World.Player.HasState = true
		return
	}
	ctx.World.UpsertActor(actor)
}

func actorEffectStateBitForStatus(statusID uint16) (uint32, bool) {
	if statusID != db.StatusEnergycoat {
		return 0, false
	}
	bit, ok := db.StatusOpt3State[statusID]
	if !ok {
		return 0, false
	}
	return bit, true
}

func (m *WorldMode) applyTrickDeadStatus(ctx client.Context, change network.StatusEffectChange) {
	if change.StatusID != db.StatusTrickdead || ctx.World == nil {
		return
	}
	id := change.ActorID
	if id == 0 {
		id = localSkillTarget(ctx)
	}
	actor, ok, _ := actorForCombatID(ctx, id)
	if !ok {
		return
	}
	now := time.Now()
	if change.Active {
		actionFamily := deathActionFamilyForActor(actor)
		m.setTrickDeadStatusAction(ctx, id, actionFamily, now, m.actorActionDuration(ctx, actor, actionFamily, defaultDeathAnimationDuration), true)
		log.Printf("trick dead active actor=%d", id)
		return
	}
	m.setTrickDeadStatusAction(ctx, id, spriteActionIdle, now, defaultAttackAnimationDuration, false)
	log.Printf("trick dead inactive actor=%d", id)
}

func (m *WorldMode) setTrickDeadStatusAction(ctx client.Context, id uint32, actionFamily int, started time.Time, duration time.Duration, holdFinal bool) {
	m.setCombatActorAction(ctx, id, actorAnimation{
		actionFamily: actionFamily,
		started:      started,
		duration:     duration,
		play:         true,
		hasPlay:      true,
		holdFinal:    holdFinal,
	})
}

func (m *WorldMode) addStatusEffectTransition(ctx client.Context, change network.StatusEffectChange) {
	if change.StatusID != db.StatusHiding {
		return
	}
	effectID := effectSummonSlave
	if change.Active {
		effectID = effectBashBegin
	}
	if m.addWorldEffect(ctx, effectID, localSkillTarget(ctx)) {
		log.Printf("status effect transition id=%d active=%t effect=%d", change.StatusID, change.Active, effectID)
	}
}

func removeExpiredStatusEffects(s *session.Session, now time.Time) {
	if s == nil {
		return
	}
	for id, effect := range s.Statuses.Active {
		if effect.HasDuration && !effect.ExpiresAt.IsZero() && now.After(effect.ExpiresAt) {
			delete(s.Statuses.Active, id)
		}
	}
}

func localActorHasStatus(ctx client.Context, statusID uint16) bool {
	if ctx.Session == nil || ctx.Session.Statuses.Active == nil {
		return false
	}
	_, ok := ctx.Session.Statuses.Active[statusID]
	return ok
}

func localActorHidden(ctx client.Context) bool {
	return localActorHasStatus(ctx, db.StatusHiding)
}
