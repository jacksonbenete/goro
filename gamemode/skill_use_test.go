package gamemode

import (
	"testing"

	"github.com/kivutar/goro/session"
)

func TestSkillTargetModes(t *testing.T) {
	if !isGroundTargetSkill(session.Skill{ID: 21, Type: 0x01}) {
		t.Fatal("Thunderstorm should be treated as a ground target skill")
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
}
