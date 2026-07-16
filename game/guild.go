package game

import (
	"fmt"
	"log"
	"strings"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
)

func (m *WorldMode) sendGuildInvite(ctx client.Context, actorID uint32, name string) {
	name = strings.TrimSpace(name)
	if actorID == 0 {
		return
	}
	if ctx.Network == nil {
		log.Printf("guild invite failed target=%d name=%q: not connected", actorID, name)
		m.ui.console.AddErrorMessage("Guild invitation failed: not connected.")
		return
	}
	accountID, charID := uint32(0), uint32(0)
	if ctx.Session != nil {
		accountID = ctx.Session.AccountID
		charID = ctx.Session.CharID
	}
	if err := ctx.Network.SendGuildInvite(actorID, accountID, charID); err != nil {
		log.Printf("guild invite failed target=%d name=%q: %v", actorID, name, err)
		m.ui.console.AddErrorMessage("Guild invitation failed.")
		return
	}
	m.ui.console.AddBlueMessage("%s has received an invitation to join your guild.", guildDisplayName(name))
}

func (m *WorldMode) openGuildInviteRequest(ctx client.Context, request network.GuildInviteRequest) {
	name := guildDisplayName(request.GuildName)
	m.ui.guildRequest.Open(ctx, "Guild Invitation", fmt.Sprintf("Would you like to join %s?", name), func() {
		if ctx.Network == nil {
			log.Printf("guild invite accept failed: not connected")
			return
		}
		if err := ctx.Network.SendGuildInviteReply(request.GuildID, true); err != nil {
			log.Printf("guild invite accept failed guild=%d name=%q: %v", request.GuildID, request.GuildName, err)
		}
	}, func() {
		if ctx.Network == nil {
			log.Printf("guild invite reject failed: not connected")
			return
		}
		if err := ctx.Network.SendGuildInviteReply(request.GuildID, false); err != nil {
			log.Printf("guild invite reject failed guild=%d name=%q: %v", request.GuildID, request.GuildName, err)
		}
	})
}

func (m *WorldMode) handleGuildCreationResult(result network.GuildCreationResult) {
	switch result.Result {
	case 0:
		m.ui.console.AddBlueMessage("Guild created.")
	case 1:
		m.ui.console.AddErrorMessage("You are already in a guild.")
	case 2:
		m.ui.console.AddErrorMessage("Guild name already exists.")
	case 3:
		m.ui.console.AddErrorMessage("You need the required item to create a guild.")
	default:
		m.ui.console.AddErrorMessage("Guild creation failed.")
	}
}

func (m *WorldMode) handleGuildInviteAck(ack network.GuildInviteAck) {
	switch ack.Result {
	case 0:
		m.ui.console.AddErrorMessage("That character is already in a guild.")
	case 1:
		m.ui.console.AddErrorMessage("Guild invitation was rejected.")
	case 2:
		m.ui.console.AddBlueMessage("Guild invitation accepted.")
	case 3:
		m.ui.console.AddErrorMessage("The guild is full.")
	default:
		m.ui.console.AddErrorMessage("Guild invitation failed.")
	}
}

func guildDisplayName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Guild"
	}
	return name
}
