package game

import (
	"log"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	worldstate "github.com/kivutar/goro/world"
)

func (m *WorldMode) applyVendingBoard(ctx client.Context, board network.VendingBoard) {
	if ctx.World == nil {
		return
	}
	if isLocalActor(ctx, board.OwnerAID) {
		ctx.World.Player.Vending = true
		ctx.World.Player.VendingName = board.Name
		return
	}
	actor, ok := ctx.World.Actors[board.OwnerAID]
	if !ok {
		actor = worldstate.Actor{ID: board.OwnerAID}
	}
	actor.Vending = true
	actor.VendingName = board.Name
	ctx.World.Actors[board.OwnerAID] = actor
	log.Printf("vending board actor=%d name=%q", board.OwnerAID, board.Name)
}

func (m *WorldMode) applyVendingBoardDisappear(ctx client.Context, board network.VendingBoardDisappear) {
	if ctx.World == nil {
		return
	}
	if isLocalActor(ctx, board.OwnerAID) {
		ctx.World.Player.Vending = false
		ctx.World.Player.VendingName = ""
		return
	}
	actor, ok := ctx.World.Actors[board.OwnerAID]
	if !ok {
		return
	}
	actor.Vending = false
	actor.VendingName = ""
	ctx.World.Actors[board.OwnerAID] = actor
	log.Printf("vending board removed actor=%d", board.OwnerAID)
}

func actorHasVending(actor worldstate.Actor) bool {
	return actor.Vending && actor.VendingName != ""
}

func (m *WorldMode) requestVendingList(ctx client.Context, actor worldstate.Actor, reason string) {
	if ctx.Network == nil {
		log.Printf("%s vending request failed actor=%d: not connected", reason, actor.ID)
		return
	}
	if err := ctx.Network.SendVendingListRequest(actor.ID); err != nil {
		log.Printf("%s vending request failed actor=%d: %v", reason, actor.ID, err)
	}
}
