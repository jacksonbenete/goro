package game

import (
	"github.com/kivutar/goro/client"
	"testing"

	"github.com/kivutar/goro/session"
)

func TestSkillTargetModes(t *testing.T) {
	if !isGroundTargetSkill(session.Skill{ID: 21, Type: 0x01}) {
		t.Fatal("Thunderstorm should be treated as a ground target skill")
	}
	if !isGroundTargetSkill(session.Skill{ID: 12, Type: 0x01}) {
		t.Fatal("Safety Wall should be treated as a ground target skill")
	}
	if !isGroundTargetSkill(session.Skill{ID: 18, Type: 0x01}) {
		t.Fatal("Fire Wall should be treated as a ground target skill")
	}
	if !isGroundTargetSkill(session.Skill{ID: 25, Type: 0x10}) {
		t.Fatal("Pneuma should be treated as a ground target skill")
	}
	if !isGroundTargetSkill(session.Skill{ID: 18, Type: 0x02}) {
		t.Fatal("ground skill type bit should request floor target")
	}
	if !isSelfTargetSkill(session.Skill{ID: 26, Type: 0x04}) {
		t.Fatal("self skill type bit should target the player")
	}
	if isSelfTargetSkill(session.Skill{ID: 21, Type: 0x06}) {
		t.Fatal("ground bit should win over self bit")
	}
	for _, skillID := range []uint16{10, 24, 26, 31, 32, 33} {
		if !isSelfTargetSkill(session.Skill{ID: skillID, Type: 0x01}) {
			t.Fatalf("skill %d should force self-targeting even with stale server flags", skillID)
		}
	}
	for _, skillID := range []uint16{9, 22, 23} {
		if !skillForcesPassive(skillID) {
			t.Fatalf("skill %d should be passive", skillID)
		}
	}
}

func TestPassiveAcolyteSkillsCannotBeUsed(t *testing.T) {
	controller := skillController{}
	for _, skillID := range []uint16{22, 23} {
		err := controller.Use(client.Context{}, session.Skill{ID: skillID, Level: 1, Type: skillTargetSelf}, "test")
		if err == nil || err.Error() != "passive skill" {
			t.Fatalf("skill %d use error = %v, want passive skill", skillID, err)
		}
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
