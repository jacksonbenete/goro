package game

import (
	"testing"

	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/render"
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

func TestSkillWindowDoubleClickUsesSharedSkillController(t *testing.T) {
	inputState := input.NewState()
	s := &session.Session{
		Skills: session.Skills{
			List: []session.Skill{{ID: 6, Type: skillTargetEnemy, Level: 2, Range: 9}},
		},
	}
	ctx := Context{Input: inputState, Session: s, ScreenW: 800, ScreenH: 600}
	mode := &WorldMode{}
	window := &skillWindowState{open: true, x: 20, y: 30, positioned: true}
	mx, my := window.x+skillWindowPad+6, window.skillRowY(0)+6

	inputState.SetMousePosition(mx, my)
	inputState.SetMouseButton(render.MouseButtonLeft, true)
	if !window.update(ctx, nil, mode) {
		t.Fatal("first click was not handled")
	}
	if mode.pendingSkill.skill.ID != 0 {
		t.Fatalf("pending skill after first click = %+v, want none", mode.pendingSkill.skill)
	}

	inputState.SetMouseButton(render.MouseButtonLeft, false)
	inputState.EndFrame()
	inputState.SetMouseButton(render.MouseButtonLeft, true)
	if !window.update(ctx, nil, mode) {
		t.Fatal("second click was not handled")
	}
	if mode.pendingSkill.skill.ID != 6 || mode.pendingSkill.skill.Level != 2 {
		t.Fatalf("pending skill = %+v, want provoke level 2", mode.pendingSkill.skill)
	}
}
