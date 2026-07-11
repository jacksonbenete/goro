package game

import (
	"fmt"
	"log"
	"strings"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	gameui "github.com/kivutar/goro/ui"
	worldstate "github.com/kivutar/goro/world"
)

func (m *WorldMode) sendPartyInvite(ctx client.Context, actorID uint32, name string) {
	if actorID == 0 {
		return
	}
	if ctx.Network == nil {
		log.Printf("party invite failed target=%d name=%q: not connected", actorID, name)
		return
	}
	if err := ctx.Network.SendPartyInvite(actorID, name); err != nil {
		log.Printf("party invite failed target=%d name=%q: %v", actorID, name, err)
		return
	}
	m.console.AddBlueMessage("%s has received an invitation to join your party.", partyDisplayName(name))
}

func (m *WorldMode) openPartyInviteRequest(ctx client.Context, request network.PartyInviteRequest) {
	name := partyDisplayName(request.Name)
	m.partyRequest.Open(ctx, "Party Invitation", fmt.Sprintf("%s has invited you to join a party.", name), func() {
		if ctx.Network == nil {
			log.Printf("party invite accept failed: not connected")
			return
		}
		if err := ctx.Network.SendPartyInviteAck(request.RequestID, true); err != nil {
			log.Printf("party invite accept failed request=%d name=%q: %v", request.RequestID, request.Name, err)
		}
	}, func() {
		if ctx.Network == nil {
			log.Printf("party invite reject failed: not connected")
			return
		}
		if err := ctx.Network.SendPartyInviteAck(request.RequestID, false); err != nil {
			log.Printf("party invite reject failed request=%d name=%q: %v", request.RequestID, request.Name, err)
		}
	})
}

func (m *WorldMode) handlePartyCreateResult(ctx client.Context, result network.PartyCreateResult) {
	switch result.Result {
	case 0:
		m.console.AddBlueMessage("Party created.")
		syncLocalPartyVitals(ctx)
	case 1:
		clearPendingParty(ctx)
		m.console.AddErrorMessage("Party name already exists.")
	case 2:
		clearPendingParty(ctx)
		m.console.AddErrorMessage("You are already in a party.")
	case 3:
		clearPendingParty(ctx)
		m.console.AddErrorMessage("Cannot organize a party on this map.")
	default:
		clearPendingParty(ctx)
		m.console.AddErrorMessage("Cannot form a party.")
	}
}

func (m *WorldMode) handlePartyInviteAnswer(answer network.PartyInviteAnswer) {
	name := partyDisplayName(answer.Name)
	switch answer.Answer {
	case 0:
		m.console.AddErrorMessage("%s is already in a party.", name)
	case 1:
		m.console.AddErrorMessage("%s denied your party invitation.", name)
	case 2:
		m.console.AddBlueMessage("%s joined your party.", name)
	case 3:
		m.console.AddErrorMessage("The party is full.")
	case 4:
		m.console.AddErrorMessage("A character from the same account is already in the party.")
	case 5:
		m.console.AddErrorMessage("%s blocked party invitations.", name)
	case 7:
		m.console.AddErrorMessage("%s is not online.", name)
	case 8:
		m.console.AddErrorMessage("Cannot invite players on this map.")
	case 9:
		m.console.AddErrorMessage("Cannot join a party on this map.")
	default:
		m.console.AddErrorMessage("Party invitation failed.")
	}
}

func applyPartyList(ctx client.Context, list network.PartyList) {
	if ctx.Session == nil {
		return
	}
	old := partyMembersByAccount(ctx.Session.Party.Members)
	ctx.Session.Party.Name = list.Name
	ctx.Session.Party.Members = ctx.Session.Party.Members[:0]
	for _, member := range list.Members {
		next := sessionPartyMemberFromNetwork(member)
		if prev, ok := old[next.AccountID]; ok {
			mergePartyMemberVitals(&next, prev)
		}
		ctx.Session.Party.Members = append(ctx.Session.Party.Members, next)
	}
	syncLocalPartyVitals(ctx)
	log.Printf("party list received name=%q members=%d", list.Name, len(list.Members))
}

func applyPartyMemberJoin(ctx client.Context, member network.PartyMember) {
	if ctx.Session == nil {
		return
	}
	next := sessionPartyMemberFromNetwork(member)
	if member.AccountID == ctx.Session.AccountID {
		ctx.Session.Party.Name = member.GroupName
	}
	if next.Name == "" {
		next.Name = "Player"
	}
	if next.MapName == "" && ctx.World != nil {
		next.MapName = ctx.World.MapName
	}
	upsertPartyMember(&ctx.Session.Party, next)
	syncLocalPartyVitals(ctx)
	log.Printf("party member join aid=%d name=%q map=%q", next.AccountID, next.Name, next.MapName)
}

func applyPartyMemberLeave(ctx client.Context, left network.PartyMemberLeave) bool {
	if ctx.Session == nil {
		return false
	}
	switch left.Result {
	case 0, 1:
	case 2:
		return false
	case 3:
		return false
	}
	if left.AccountID == ctx.Session.AccountID {
		ctx.Session.Party = session.Party{}
		return true
	}
	for i := range ctx.Session.Party.Members {
		member := ctx.Session.Party.Members[i]
		if member.AccountID != left.AccountID && (left.Name == "" || member.Name != left.Name) {
			continue
		}
		ctx.Session.Party.Members = append(ctx.Session.Party.Members[:i], ctx.Session.Party.Members[i+1:]...)
		return true
	}
	return true
}

func applyPartyOption(ctx client.Context, option network.PartyOption) {
	if ctx.Session == nil {
		return
	}
	ctx.Session.Party.ExpShare = option.ExpOption
	log.Printf("party option exp=%d", option.ExpOption)
}

func applyPartyMemberHP(ctx client.Context, hp network.PartyMemberHP) {
	if ctx.Session == nil || !ctx.Session.Party.Active() {
		return
	}
	member := findPartyMember(&ctx.Session.Party, hp.AccountID)
	if member == nil {
		return
	}
	member.HP = hp.HP
	member.MaxHP = hp.MaxHP
}

func applyPartyMemberPosition(ctx client.Context, pos network.PartyMemberPosition) {
	if ctx.Session == nil || !ctx.Session.Party.Active() {
		return
	}
	member := findPartyMember(&ctx.Session.Party, pos.AccountID)
	if member == nil {
		return
	}
	member.X = pos.X
	member.Y = pos.Y
}

func applyPartyChat(ctx client.Context, chat network.PartyChat, console *gameui.ChatConsole) {
	if console == nil {
		return
	}
	msg := strings.TrimSpace(chat.Message)
	if msg == "" {
		return
	}
	console.AddBlueMessage("%s", msg)
}

func upsertPartyMember(p *session.Party, member session.PartyMember) {
	for i := range p.Members {
		if p.Members[i].AccountID != member.AccountID {
			continue
		}
		mergePartyMemberVitals(&member, p.Members[i])
		p.Members[i] = member
		return
	}
	p.Members = append(p.Members, member)
}

func ensurePartyMember(p *session.Party, accountID uint32) *session.PartyMember {
	if member := findPartyMember(p, accountID); member != nil {
		return member
	}
	p.Members = append(p.Members, session.PartyMember{AccountID: accountID, Name: "Player"})
	return &p.Members[len(p.Members)-1]
}

func findPartyMember(p *session.Party, accountID uint32) *session.PartyMember {
	for i := range p.Members {
		if p.Members[i].AccountID == accountID {
			return &p.Members[i]
		}
	}
	return nil
}

func sessionPartyMemberFromNetwork(member network.PartyMember) session.PartyMember {
	return session.PartyMember{
		AccountID: member.AccountID,
		Name:      member.Name,
		MapName:   member.MapName,
		Role:      member.Role,
		State:     member.State,
		X:         member.X,
		Y:         member.Y,
	}
}

func partyMembersByAccount(members []session.PartyMember) map[uint32]session.PartyMember {
	out := make(map[uint32]session.PartyMember, len(members))
	for _, member := range members {
		out[member.AccountID] = member
	}
	return out
}

func mergePartyMemberVitals(dst *session.PartyMember, src session.PartyMember) {
	if dst.HP == 0 {
		dst.HP = src.HP
	}
	if dst.MaxHP == 0 {
		dst.MaxHP = src.MaxHP
	}
	if dst.X == 0 {
		dst.X = src.X
	}
	if dst.Y == 0 {
		dst.Y = src.Y
	}
	dst.Dead = src.Dead
}

func partyDisplayName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Player"
	}
	return name
}

func clearPendingParty(ctx client.Context) {
	if ctx.Session == nil || len(ctx.Session.Party.Members) > 0 {
		return
	}
	ctx.Session.Party = session.Party{}
}

func syncLocalPartyVitals(ctx client.Context) {
	if ctx.Session == nil || !ctx.Session.Party.Active() || ctx.Session.AccountID == 0 {
		return
	}
	member := ensurePartyMember(&ctx.Session.Party, ctx.Session.AccountID)
	if member.Name == "" {
		member.Name = partyDisplayName(ctx.Session.Selected.Name)
	}
	hp, maxHP := localPartyHP(ctx.Session)
	member.HP = hp
	member.MaxHP = maxHP
}

func localPartyHP(s *session.Session) (int, int) {
	if s == nil {
		return 0, 0
	}
	hp := s.Vitals.HP
	maxHP := s.Vitals.MaxHP
	if maxHP <= 0 {
		character := selectedCharacter(s)
		if character.MaxHP <= 0 && s.Selected.MaxHP > 0 {
			character = s.Selected
		}
		hp = int(character.HP)
		maxHP = int(character.MaxHP)
	}
	return hp, maxHP
}

func partyMemberLifeForDisplay(ctx client.Context, actor worldstate.Actor) (actorLife, bool) {
	if ctx.Session == nil || actor.ID == 0 || !actor.HasObjectType || actor.ObjectType != actorObjectTypePC {
		return actorLife{}, false
	}
	for _, member := range ctx.Session.Party.Members {
		if member.AccountID != actor.ID || member.MaxHP <= 0 {
			continue
		}
		return actorLife{
			hp:     clampGameInt(member.HP, 0, member.MaxHP),
			maxHP:  member.MaxHP,
			player: true,
		}, true
	}
	return actorLife{}, false
}

func partyCanManage(s *session.Session) bool {
	if s == nil || !s.Party.Active() {
		return false
	}
	if len(s.Party.Members) == 0 {
		return true
	}
	for _, member := range s.Party.Members {
		if member.AccountID == s.AccountID {
			return member.Leader()
		}
	}
	return false
}
