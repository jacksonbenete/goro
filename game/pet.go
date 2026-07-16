package game

import (
	"log"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

const petCaptureSkillID = 0xFFFF

type petCaptureState struct {
	active  bool
	started time.Time
}

func petCaptureTargetSkill() session.Skill {
	return session.Skill{
		ID:    petCaptureSkillID,
		Name:  "Capture Monster",
		Type:  skillTargetPet,
		Range: 9,
	}
}

func (m *WorldMode) startPetCapture(ctx client.Context) {
	m.pendingSkill = pendingSkillTarget{}
	m.pendingSkillText = pendingSkillTextTarget{}
	m.pendingPetCapture = petCaptureState{active: true, started: time.Now()}
	m.ui.console.AddSystemMessage("Select a monster to capture.")
	log.Printf("pet capture target pending")
}

func (m *WorldMode) cancelPetCapture(source string) {
	if !m.pendingPetCapture.active {
		return
	}
	m.pendingPetCapture = petCaptureState{}
	log.Printf("pet capture canceled source=%s", source)
}

func (m *WorldMode) cancelPetCaptureFromInput(ctx client.Context) bool {
	if !m.pendingPetCapture.active || ctx.Input == nil {
		return false
	}
	if !ctx.Input.JustPressed(input.KeyEscape) && !ctx.Input.MouseJustPressed(input.MouseButtonRight) {
		return false
	}
	m.cancelPetCapture("input")
	return true
}

func (m *WorldMode) handlePetCaptureClick(ctx client.Context, projection sceneProjection, now time.Time) bool {
	if !m.pendingPetCapture.active || ctx.Input == nil || !ctx.Input.MouseJustPressed(input.MouseButtonLeft) {
		return false
	}
	skill := petCaptureTargetSkill()
	actor, ok := clickedSkillTarget(ctx, projection, skill, ctx.Input.MouseX, ctx.Input.MouseY, now, m.actorDeaths)
	if !ok {
		if x, y, groundOK := clickedWalkTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY); groundOK {
			log.Printf("pet capture canceled by ground click mouse=%d,%d target=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY, x, y)
			m.cancelPetCapture("ground-click")
			return true
		}
		log.Printf("pet capture target miss mouse=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY)
		return true
	}
	if ctx.Network == nil {
		m.ui.console.AddErrorMessage("Capture failed: not connected.")
		return true
	}
	m.pendingPetCapture = petCaptureState{}
	m.openPetSlotMachine(ctx, actor.ID)
	log.Printf("pet capture target selected target=%d name=%q job=%d object_type=%d", actor.ID, actor.Name, actor.Job, actor.ObjectType)
	return true
}

func (m *WorldMode) applyPetCaptureResult(ctx client.Context, result network.PetCaptureResult) {
	if m.petSlotMachine.active {
		m.petSlotMachine.setResult(result.Success)
	} else if result.Success {
		m.ui.console.AddBlueMessage("Pet capture succeeded.")
	} else {
		m.ui.console.AddErrorMessage("Pet capture failed.")
	}
	log.Printf("pet capture result success=%t", result.Success)
}

func (m *WorldMode) applyPetEggList(ctx client.Context, list network.PetEggList) {
	log.Printf("pet egg list indexes=%v", list.Indexes)
	if len(list.Indexes) == 0 {
		m.ui.console.AddErrorMessage("No pet eggs available.")
		return
	}
	m.ui.petEggWindow.OpenList(ctx, list)
}
