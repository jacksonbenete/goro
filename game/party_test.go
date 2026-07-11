package game

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestPartyVitalsDoNotCreateMembers(t *testing.T) {
	s := &session.Session{}
	ctx := client.Context{Session: s}

	applyPartyMemberHP(ctx, network.PartyMemberHP{AccountID: 123, HP: 10, MaxHP: 20})
	applyPartyMemberPosition(ctx, network.PartyMemberPosition{AccountID: 123, X: 5, Y: 6})

	if len(s.Party.Members) != 0 {
		t.Fatalf("party vitals created members: %+v", s.Party.Members)
	}
}

func TestPartyVitalsUpdateKnownMembers(t *testing.T) {
	s := &session.Session{
		Party: session.Party{
			Name:    "party",
			Members: []session.PartyMember{{AccountID: 123, Name: "Alice"}},
		},
	}
	ctx := client.Context{Session: s}

	applyPartyMemberHP(ctx, network.PartyMemberHP{AccountID: 123, HP: 10, MaxHP: 20})
	applyPartyMemberPosition(ctx, network.PartyMemberPosition{AccountID: 123, X: 5, Y: 6})

	member := s.Party.Members[0]
	if member.HP != 10 || member.MaxHP != 20 || member.X != 5 || member.Y != 6 {
		t.Fatalf("party vitals not applied: %+v", member)
	}
}
