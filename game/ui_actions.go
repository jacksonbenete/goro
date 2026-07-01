package game

import "github.com/kivutar/goro/session"

func (m *WorldMode) UseShortcutSkill(ctx Context, skill session.Skill) error {
	return m.skills().Use(ctx, skill, "shortcut")
}

func (m *WorldMode) AddTeleportEffect(ctx Context) {
	m.addWorldEffect(ctx, effectTeleportation, localSkillTarget(ctx))
}
