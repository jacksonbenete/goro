package game

import (
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/session"
)

func (m *WorldMode) UseShortcutSkill(ctx client.Context, skill session.Skill) error {
	return m.skills().Use(ctx, skill, "shortcut")
}

func (m *WorldMode) AddTeleportEffect(ctx client.Context) {
	m.addWorldEffect(ctx, effectTeleportation, localSkillTarget(ctx))
}
