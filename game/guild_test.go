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
