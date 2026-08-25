package game

import (
	"fmt"
	"strings"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/network"
)

func (m *WorldMode) sendAdoptionRequest(ctx client.Context, targetAccountID uint32, name string) {
	if targetAccountID == 0 {
		return
	}
	if ctx.Network == nil {
		m.ui.console.AddErrorMessage("Adoption request failed: not connected.")
		return
	}
	if err := ctx.Network.SendAdoptionRequest(targetAccountID); err != nil {
		glog.Warnf("adoption request failed target=%d name=%q: %v", targetAccountID, name, err)
		m.ui.console.AddErrorMessage("Adoption request failed.")
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "That character"
	}
	m.ui.console.AddBlueMessage("%s has received your adoption request.", name)
}

func (m *WorldMode) openAdoptionRequest(ctx client.Context, request network.AdoptionRequest) {
	name := strings.TrimSpace(request.FatherName)
	if name == "" {
		name = "A married couple"
	}
	m.ui.adoptionRequest.Open(ctx, "Adoption", fmt.Sprintf("%s wishes to adopt you. Do you accept?", name), func() {
		m.sendAdoptionReply(ctx, request, true)
	}, func() {
		m.sendAdoptionReply(ctx, request, false)
	})
}

func (m *WorldMode) sendAdoptionReply(ctx client.Context, request network.AdoptionRequest, accepted bool) {
	if ctx.Network == nil {
		m.ui.console.AddErrorMessage("Adoption reply failed: not connected.")
		return
	}
	if err := ctx.Network.SendAdoptionReply(request.FatherAccountID, request.MotherAccountID, accepted); err != nil {
		glog.Warnf("adoption reply failed father=%d mother=%d accepted=%t: %v", request.FatherAccountID, request.MotherAccountID, accepted, err)
		m.ui.console.AddErrorMessage("Adoption reply failed.")
	}
}

func (m *WorldMode) handleAdoptionMessage(message network.AdoptionMessage) {
	switch message.Code {
	case 0:
		m.ui.console.AddErrorMessage("You cannot adopt more than one child.")
	case 1:
		m.ui.console.AddErrorMessage("Both parents must be at least Base Level 70.")
	case 2:
		m.ui.console.AddErrorMessage("You cannot adopt a married character.")
	default:
		m.ui.console.AddErrorMessage("Adoption failed.")
	}
}
