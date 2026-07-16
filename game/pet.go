package game

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	gameui "github.com/kivutar/goro/ui"
	worldstate "github.com/kivutar/goro/world"
)

const petCaptureSkillID = 0xFFFF

const (
	petStateInit        uint8 = 0
	petStateAccessory   uint8 = 3
	petStatePerformance uint8 = 4
)

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

func (m *WorldMode) applyPetProperty(ctx client.Context, property network.PetProperty) {
	log.Printf("pet property name=%q level=%d fullness=%d relationship=%d accessory=%d job=%d modified=%t", property.Name, property.Level, property.Fullness, property.Relationship, property.AccessoryID, property.Job, property.Modified)
	if strings.TrimSpace(property.Name) != "" {
		m.ui.console.AddSystemMessage("Pet: %s", property.Name)
	}
}

func (m *WorldMode) applyPetFeedResult(ctx client.Context, result network.PetFeedResult) {
	name := fmt.Sprintf("item %d", result.ItemID)
	if ctx.Resources != nil {
		if resolved, ok := ctx.Resources.ItemDisplayName(int(result.ItemID), true); ok && strings.TrimSpace(resolved) != "" {
			name = resolved
		}
	}
	if result.Success {
		m.ui.console.AddBlueMessage("Fed pet with %s.", name)
	} else {
		m.ui.console.AddErrorMessage("Failed to feed pet with %s.", name)
	}
	log.Printf("pet feed result success=%t item=%d", result.Success, result.ItemID)
}

func (m *WorldMode) applyPetStateChange(ctx client.Context, change network.PetStateChange) {
	log.Printf("pet state type=%d id=%d data=%d", change.Type, change.ID, change.Data)
	switch change.Type {
	case petStateInit:
		m.petID = change.ID
	case petStateAccessory:
		m.applyPetAccessoryChange(ctx, change.ID, change.Data)
	case petStatePerformance:
		m.startPetPerformance(ctx, change.ID, change.Data)
	}
}

func (m *WorldMode) applyPetAccessoryChange(ctx client.Context, id uint32, accessoryID uint32) {
	if accessoryID == 0 {
		delete(m.petAccessoryIDs, id)
		return
	}
	if m.petAccessoryIDs == nil {
		m.petAccessoryIDs = make(map[uint32]uint32)
	}
	m.petAccessoryIDs[id] = accessoryID
	if ctx.World != nil {
		if actor, ok := ctx.World.Actors[id]; ok {
			_ = m.nonPCSpriteView(ctx, actor)
		}
	}
}

func (m *WorldMode) applyPetAction(ctx client.Context, action network.PetAction) {
	if action.Data < 5000 {
		emotionType := uint8(action.Data / 10)
		m.applyEmotionNotify(ctx, network.EmotionNotify{GID: action.ID, Type: emotionType})
		log.Printf("pet emotion actor=%d data=%d type=%d", action.ID, action.Data, emotionType)
		return
	}
	log.Printf("pet talk actor=%d data=%d ignored=no_pet_talk_table", action.ID, action.Data)
}

func (m *WorldMode) startPetPerformance(ctx client.Context, id uint32, data uint32) {
	if ctx.World == nil {
		return
	}
	actor, ok := ctx.World.Actors[id]
	if !ok {
		return
	}
	action := spriteActionNonPCPerf1
	switch data {
	case 2:
		action = spriteActionNonPCPerf2
	case 3:
		action = spriteActionNonPCPerf3
	case 4:
		action = spriteActionNonPCSpecial
	}
	now := time.Now()
	duration := m.actorActionDuration(ctx, actor, action, defaultAttackAnimationDuration)
	m.startActorAnimation(ctx, id, action, now, duration)
}

func (m *WorldMode) openPetContextFromInput(ctx client.Context, now time.Time) bool {
	if ctx.Input == nil || !ctx.Input.MouseJustPressed(input.MouseButtonRight) || uiPointerBlocked(ctx) {
		return false
	}
	screenW, screenH := ctx.ScreenSize()
	projection := m.sceneProjection(ctx, screenW, screenH, now)
	actor, ok := clickedPetTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY, now, m.actorDeaths, m.petID)
	if !ok {
		return false
	}
	log.Printf("pet context target id=%d name=%q job=%d", actor.ID, actor.Name, actor.Job)
	m.ui.petContext.Open(ctx, ctx.Input.MouseX, ctx.Input.MouseY)
	return true
}

func (m *WorldMode) handlePetContextAction(ctx client.Context, action gameui.PetContextAction) {
	switch action.Kind {
	case gameui.PetContextActionFeed:
		m.openPetFeedConfirm(ctx)
	case gameui.PetContextActionPerformance:
		m.sendPetCommand(ctx, network.PetCommandPerformance)
	case gameui.PetContextActionBackToEgg:
		m.sendPetCommand(ctx, network.PetCommandBackToEgg)
	case gameui.PetContextActionUnequipAccessory:
		m.sendPetCommand(ctx, network.PetCommandUnequipAccessory)
	}
}

func (m *WorldMode) openPetFeedConfirm(ctx client.Context) {
	m.ui.petConfirm.Open(ctx, "Feed Pet", "Are you sure you want to feed your pet?", func() {
		m.sendPetCommand(ctx, network.PetCommandFeed)
	}, nil)
}

func (m *WorldMode) sendPetCommand(ctx client.Context, command uint8) {
	if ctx.Network == nil {
		m.ui.console.AddErrorMessage("Pet command failed: not connected.")
		return
	}
	if err := ctx.Network.SendPetCommand(command); err != nil {
		m.ui.console.AddErrorMessage("Pet command failed.")
		log.Printf("pet command failed command=%d: %v", command, err)
	}
}

func clickedPetTarget(ctx client.Context, projection sceneProjection, mouseX, mouseY int, now time.Time, deadActors map[uint32]time.Time, petID uint32) (worldstate.Actor, bool) {
	if petID == 0 {
		return worldstate.Actor{}, false
	}
	actor, ok := hoveredCursorActor(ctx, projection, mouseX, mouseY, now, deadActors)
	if !ok || actor.ID != petID {
		return worldstate.Actor{}, false
	}
	return actor, true
}
