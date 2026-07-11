package game

import (
	"log"

	"github.com/kivutar/goro/client"
)

func (m *WorldMode) sendViewEquipmentRequest(ctx client.Context, actorID uint32, name string) {
	if actorID == 0 {
		return
	}
	if ctx.Network == nil {
		log.Printf("view equipment request failed target=%d name=%q: not connected", actorID, name)
		return
	}
	if err := ctx.Network.SendViewPlayerEquipment(actorID); err != nil {
		log.Printf("view equipment request failed target=%d name=%q: %v", actorID, name, err)
	}
}
