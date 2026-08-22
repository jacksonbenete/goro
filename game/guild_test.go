package game

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	worldstate "github.com/kivutar/goro/world"
)

func TestApplyLocalGuildDetailsInfersMasterFromSelectedCharacter(t *testing.T) {
	s := &session.Session{
		Selected: session.Character{Name: "Arcer"},
		Guild:    session.Guild{IsMaster: false},
	}

	applyLocalGuildDetails(client.Context{Session: s}, network.GuildInfo{
		GuildID:    1,
		GuildName:  "Mandala",
		MasterName: "Arcer",
	})

	if !s.Guild.IsMaster {
		t.Fatal("selected guild master should get master access from guild info")
	}
}

func TestApplyLocalGuildDetailsClearsMasterWhenSelectedCharacterIsNotMaster(t *testing.T) {
	s := &session.Session{
		Selected: session.Character{Name: "Kivutar"},
		Guild:    session.Guild{IsMaster: true},
	}

	applyLocalGuildDetails(client.Context{Session: s}, network.GuildInfo{
		GuildID:    1,
		GuildName:  "Mandala",
		MasterName: "Arcer",
	})

	if s.Guild.IsMaster {
		t.Fatal("non-master selected character should lose master access from guild info")
	}
}

func TestApplyLocalGuildBelongingStoresInviteRight(t *testing.T) {
	s := &session.Session{}
	applyLocalGuildBelonging(client.Context{Session: s}, network.GuildBelonging{
		GuildID: 1,
		Mode:    guildPermissionInvite,
	})

	if s.Guild.Right != guildPermissionInvite {
		t.Fatalf("guild right = 0x%X, want invite right", s.Guild.Right)
	}
	applyLocalGuildDetails(client.Context{Session: s}, network.GuildInfo{GuildID: 1, GuildName: "Mandala"})
	if s.Guild.Right != guildPermissionInvite {
		t.Fatalf("guild details cleared invite right: 0x%X", s.Guild.Right)
	}
}

func TestGuildCanInvitePlayerMatchesRobrowserRequirements(t *testing.T) {
	tests := []struct {
		name          string
		session       *session.Session
		targetGuildID uint32
		want          bool
	}{
		{name: "no session"},
		{name: "not in guild", session: &session.Session{Guild: session.Guild{Right: guildPermissionInvite}}},
		{name: "no invite right", session: &session.Session{GuildID: 1}},
		{name: "target already in guild", session: &session.Session{GuildID: 1, Guild: session.Guild{Right: guildPermissionInvite}}, targetGuildID: 2},
		{name: "invite permitted", session: &session.Session{GuildID: 1, Guild: session.Guild{Right: guildPermissionInvite}}, want: true},
		{name: "nested guild id", session: &session.Session{Guild: session.Guild{ID: 1, Right: guildPermissionInvite}}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := guildCanInvitePlayer(test.session, test.targetGuildID); got != test.want {
				t.Fatalf("guildCanInvitePlayer() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestActorGuildEmblemRequestsFromUninitializedCache(t *testing.T) {
	mode := &WorldMode{}
	ctx := client.Context{
		Network: network.NewClient(20080910, false),
		Session: &session.Session{GuildID: 0x01020304, EmblemVersion: 7},
	}

	if emblem := mode.actorGuildEmblem(ctx, worldstate.Actor{}, true); emblem != nil {
		t.Fatalf("emblem = %v, want nil until image packet arrives", emblem)
	}
	if mode.guildEmblems == nil {
		t.Fatal("local guild emblem lookup should initialize request cache")
	}
}
