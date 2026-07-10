package game

import (
	"github.com/kivutar/goro/client"
	"testing"

	"github.com/kivutar/goro/session"
)

func TestSkillTargetModes(t *testing.T) {
	if !isGroundTargetSkill(session.Skill{ID: 18, Type: 0x02}) {
		t.Fatal("ground skill type bit should request floor target")
	}
	if !isSelfTargetSkill(session.Skill{ID: 26, Type: 0x04}) {
		t.Fatal("self skill type bit should target the player")
	}
	if isSelfTargetSkill(session.Skill{ID: 21, Type: 0x06}) {
		t.Fatal("ground bit should win over self bit")
	}
}

func TestChangeCartSkillOpensSelector(t *testing.T) {
	mode := &WorldMode{}
	controller := skillController{mode: mode}
	ctx := client.Context{Session: &session.Session{AccountID: 2000000}}

	if err := controller.Use(ctx, session.Skill{ID: skillChangeCart, Level: 1, Type: skillTargetSelf}, "test"); err != nil {
		t.Fatalf("change cart use failed: %v", err)
	}
	if !mode.changeCartWindow.IsOpen() {
		t.Fatal("change cart window was not opened")
	}
	if mode.pendingSkill.skill.ID != 0 {
		t.Fatalf("pending skill = %+v, want none", mode.pendingSkill.skill)
	}
}
