package gamemode

import (
	"testing"

	"github.com/kivutar/goro/session"
)

func TestVisibleSkillsUsesScrollWindow(t *testing.T) {
	s := &session.Session{}
	for i := 0; i < 12; i++ {
		s.Skills.List = append(s.Skills.List, session.Skill{ID: uint16(i + 1)})
	}
	got := visibleSkills(s, 3, 4)
	if len(got) != 4 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0].ID != 4 || got[3].ID != 7 {
		t.Fatalf("visible = %+v", got)
	}
}

func TestCanIncreaseSkillRequiresPointsAndFlag(t *testing.T) {
	s := &session.Session{Skills: session.Skills{Points: 1}}
	if !canIncreaseSkill(s, session.Skill{ID: 1, Upgradable: true}) {
		t.Fatal("expected skill to be increasable")
	}
	if canIncreaseSkill(s, session.Skill{ID: 1}) {
		t.Fatal("skill without upgradable flag should not increase")
	}
	s.Skills.Points = 0
	if canIncreaseSkill(s, session.Skill{ID: 1, Upgradable: true}) {
		t.Fatal("skill without points should not increase")
	}
}
