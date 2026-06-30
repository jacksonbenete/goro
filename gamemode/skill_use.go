package gamemode

import (
	"fmt"
	"log"
	"time"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

type skillController struct {
	mode *WorldMode
}

func (m *WorldMode) skills() skillController {
	return skillController{mode: m}
}

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
	return skill.Type&skillTargetPlace != 0 || skillForcesGroundTarget(skill.ID)
}

func isSelfTargetSkill(skill session.Skill) bool {
	return skillForcesSelfTarget(skill.ID) || (skill.Type&skillTargetSelf != 0 && !isGroundTargetSkill(skill))
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

func (c skillController) Use(ctx Context, skill session.Skill, source string) error {
	if skill.ID == 0 || skill.Level <= 0 {
		return fmt.Errorf("skill is not learned")
	}
	if skill.Type == 0 || skillForcesPassive(skill.ID) {
		return fmt.Errorf("passive skill")
	}
	if isSelfTargetSkill(skill) {
		target := localSkillTarget(ctx)
		if target == 0 {
			return fmt.Errorf("missing skill target")
		}
		return c.SendToID(ctx, skill, target, source)
	}
	if skill.Range > 0 || isGroundTargetSkill(skill) {
		c.mode.pendingSkill = pendingSkillTarget{skill: skill, started: time.Now()}
		c.mode.status = fmt.Sprintf("select target: %s", skillDisplayName(ctx.Resources, skill))
		log.Printf("%s skill target pending skill=%d level=%d range=%d", source, skill.ID, skill.Level, skill.Range)
		return nil
	}
	target := localSkillTarget(ctx)
	if target == 0 {
		return fmt.Errorf("missing skill target")
	}
	return c.SendToID(ctx, skill, target, source)
}

func (c skillController) SendToID(ctx Context, skill session.Skill, target uint32, source string) error {
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
		if effectDetachesLocalActor(effectID) && isLocalActor(ctx, actorID) {
			actorID = 0
		}
		c.mode.addWorldEffect(ctx, effectID, actorID)
	}
	if property, duration := skillCastFallback(skill.ID, level); duration > 0 {
		c.mode.addLocalSkillCastFallback(ctx, skill.ID, property, localSkillTarget(ctx), target, 0, 0, duration, time.Now(), source)
	}
	if isLevelOneTeleportSkill(skill) {
		c.mode.addWorldEffect(ctx, effectTeleportation, localSkillTarget(ctx))
	}
	return nil
}

func (c skillController) SendToGround(ctx Context, skill session.Skill, x, y int, source string) error {
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
		c.mode.addLocalSkillCastFallback(ctx, skill.ID, property, localSkillTarget(ctx), 0, x, y, castDuration, time.Now(), source+"-ground")
	}
	if castDuration <= 0 {
		now := time.Now()
		for _, effectID := range skillGroundEffectIDs(skill.ID) {
			c.mode.addWorldEffectAtCellIfMissing(ctx, effectID, x, y, now)
		}
	}
	return nil
}

func (c skillController) CancelFromInput(ctx Context) bool {
	if c.mode.pendingSkill.skill.ID == 0 || ctx.Input == nil {
		return false
	}
	if !ctx.Input.JustPressed(render.KeyEscape) && !ctx.Input.MouseJustPressed(render.MouseButtonRight) {
		return false
	}
	c.Cancel("input")
	return true
}

func (c skillController) Cancel(source string) {
	if c.mode.pendingSkill.skill.ID == 0 {
		return
	}
	log.Printf("skill target canceled skill=%d source=%s", c.mode.pendingSkill.skill.ID, source)
	c.mode.pendingSkill = pendingSkillTarget{}
	c.mode.status = "skill canceled"
}

func (c skillController) HandleClick(ctx Context, projection sceneProjection, now time.Time) {
	skill := c.mode.pendingSkill.skill
	if skill.ID == 0 {
		return
	}
	if isGroundTargetSkill(skill) {
		targetX, targetY, ok := clickedWalkTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY)
		if !ok {
			c.mode.status = fmt.Sprintf("select target: %s", skillDisplayName(ctx.Resources, skill))
			log.Printf("skill ground target miss skill=%d mouse=%d,%d", skill.ID, ctx.Input.MouseX, ctx.Input.MouseY)
			return
		}
		if err := c.SendToGround(ctx, skill, targetX, targetY, "target"); err != nil {
			c.mode.status = "skill failed: " + err.Error()
			log.Printf("skill ground target failed skill=%d target=%d,%d: %v", skill.ID, targetX, targetY, err)
			return
		}
		c.mode.pendingSkill = pendingSkillTarget{}
		c.mode.status = fmt.Sprintf("%s: %d,%d", skillDisplayName(ctx.Resources, skill), targetX, targetY)
		log.Printf("skill ground target sent skill=%d target=%d,%d", skill.ID, targetX, targetY)
		return
	}
	actor, ok := clickedSkillTarget(ctx, projection, skill, ctx.Input.MouseX, ctx.Input.MouseY, now, c.mode.actorDeaths)
	if !ok {
		c.mode.status = fmt.Sprintf("select target: %s", skillDisplayName(ctx.Resources, skill))
		log.Printf("skill target miss skill=%d mouse=%d,%d", skill.ID, ctx.Input.MouseX, ctx.Input.MouseY)
		return
	}
	if err := c.SendToID(ctx, skill, actor.ID, "target"); err != nil {
		c.mode.status = "skill failed: " + err.Error()
		log.Printf("skill target failed skill=%d target=%d: %v", skill.ID, actor.ID, err)
		return
	}
	c.mode.pendingSkill = pendingSkillTarget{}
	c.mode.status = fmt.Sprintf("%s: %d", skillDisplayName(ctx.Resources, skill), actor.ID)
	log.Printf("skill target sent skill=%d target=%d name=%q job=%d object_type=%d", skill.ID, actor.ID, actor.Name, actor.Job, actor.ObjectType)
}

func (c skillController) ApplyAutoRun(ctx Context, auto network.AutoRunSkill) {
	skill := sessionSkillFromNetwork(auto.Skill)
	target := localSkillTarget(ctx)
	log.Printf("auto-run skill received skill=%d level=%d range=%d name=%q target=%d", skill.ID, skill.Level, skill.Range, skill.Name, target)
	if target == 0 {
		c.mode.status = "auto skill failed: missing player id"
		return
	}
	if err := c.SendToID(ctx, skill, target, "auto"); err != nil {
		c.mode.status = "auto skill failed: " + err.Error()
		log.Printf("auto-run skill use failed skill=%d target=%d: %v", skill.ID, target, err)
		return
	}
	c.mode.status = skillDisplayName(ctx.Resources, skill)
}

func skillTargetOverrideActive(ctx Context) bool {
	return (ctx.Input != nil && ctx.Input.Pressed(render.KeyShift)) || (ctx.Session != nil && ctx.Session.NoShift)
}
