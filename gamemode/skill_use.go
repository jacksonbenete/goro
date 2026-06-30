package gamemode

import (
	"fmt"
	"log"
	"time"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

func skillByID(s *session.Session, skillID uint16) (session.Skill, bool) {
	if s == nil || skillID == 0 {
		return session.Skill{}, false
	}
	for _, skill := range s.Skills.List {
		if skill.ID == skillID {
			return skill, true
		}
	}
	return session.Skill{}, false
}

func localSkillTarget(ctx Context) uint32 {
	if ctx.Session == nil {
		return 0
	}
	if ctx.Session.AccountID != 0 {
		return ctx.Session.AccountID
	}
	return ctx.Session.CharID
}

func isGroundTargetSkill(skill session.Skill) bool {
	return skill.Type&skillTargetPlace != 0 || skill.ID == 21 || skill.ID == 25
}

func isSelfTargetSkill(skill session.Skill) bool {
	return skill.Type&skillTargetSelf != 0 && !isGroundTargetSkill(skill)
}

const (
	skillTargetEnemy  = 1
	skillTargetPlace  = 2
	skillTargetSelf   = 4
	skillTargetFriend = 16
	skillTargetTrap   = 32
	skillTargetPet    = 64
	skillTargetHomun  = 128
)

func (m *WorldMode) useSkill(ctx Context, skill session.Skill, source string) error {
	if skill.ID == 0 || skill.Level <= 0 {
		return fmt.Errorf("skill is not learned")
	}
	if skill.Type == 0 {
		return fmt.Errorf("passive skill")
	}
	if isSelfTargetSkill(skill) {
		target := localSkillTarget(ctx)
		if target == 0 {
			return fmt.Errorf("missing skill target")
		}
		return m.sendSkillToID(ctx, skill, target, source)
	}
	if skill.Range > 0 || isGroundTargetSkill(skill) {
		m.pendingSkill = pendingSkillTarget{skill: skill, started: time.Now()}
		m.status = fmt.Sprintf("select target: %s", skillDisplayName(ctx.Resources, skill))
		log.Printf("%s skill target pending skill=%d level=%d range=%d", source, skill.ID, skill.Level, skill.Range)
		return nil
	}
	target := localSkillTarget(ctx)
	if target == 0 {
		return fmt.Errorf("missing skill target")
	}
	return m.sendSkillToID(ctx, skill, target, source)
}

func (m *WorldMode) sendSkillToID(ctx Context, skill session.Skill, target uint32, source string) error {
	if ctx.Network == nil {
		return fmt.Errorf("not connected")
	}
	if skill.ID == 0 || skill.Level <= 0 {
		return fmt.Errorf("skill is not learned")
	}
	if target == 0 {
		return fmt.Errorf("missing skill target")
	}
	level := uint16(maxInt(1, skill.Level))
	log.Printf("%s skill use skill=%d level=%d target=%d", source, skill.ID, level, target)
	if err := ctx.Network.SendUseSkillToID(skill.ID, level, target); err != nil {
		return err
	}
	for _, effectID := range skillBeginEffectIDs(skill.ID) {
		actorID := localSkillTarget(ctx)
		if effectID == effectTeleportation && isLocalActor(ctx, actorID) {
			actorID = 0
		}
		m.addWorldEffect(ctx, effectID, actorID)
	}
	if property, duration := skillCastFallback(skill.ID, level); duration > 0 {
		m.addSkillCastEffects(ctx, skill.ID, property, localSkillTarget(ctx), target, 0, 0, duration, time.Now(), source)
	}
	return nil
}

func (m *WorldMode) sendSkillToGround(ctx Context, skill session.Skill, x, y int, source string) error {
	if ctx.Network == nil {
		return fmt.Errorf("not connected")
	}
	if skill.ID == 0 || skill.Level <= 0 {
		return fmt.Errorf("skill is not learned")
	}
	if !walkTargetInBounds(ctx, x, y) {
		return fmt.Errorf("invalid ground target %d,%d", x, y)
	}
	level := uint16(maxInt(1, skill.Level))
	log.Printf("%s ground skill use skill=%d level=%d target=%d,%d", source, skill.ID, level, x, y)
	if err := ctx.Network.SendUseSkillToGround(skill.ID, level, x, y); err != nil {
		return err
	}
	property, castDuration := skillCastFallback(skill.ID, level)
	if castDuration > 0 {
		m.addSkillCastEffects(ctx, skill.ID, property, localSkillTarget(ctx), 0, x, y, castDuration, time.Now(), source+"-ground")
	}
	if castDuration <= 0 {
		now := time.Now()
		for _, effectID := range skillGroundEffectIDs(skill.ID) {
			m.addWorldEffectAtCellIfMissing(ctx, effectID, x, y, now)
		}
	}
	return nil
}

func skillTargetOverrideActive(ctx Context) bool {
	return (ctx.Input != nil && ctx.Input.Pressed(render.KeyShift)) || (ctx.Session != nil && ctx.Session.NoShift)
}
