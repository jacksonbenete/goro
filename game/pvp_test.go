package game

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
	worldstate "github.com/kivutar/goro/world"
)

func TestPvPMapAllowsPlayerAttackAndEnemySkillTargets(t *testing.T) {
	world := worldstate.New()
	world.MapProperty = worldstate.MapPropertyFreePvPZone
	actor := worldstate.Actor{
		ID:            300,
		ObjectType:    actorObjectTypePC,
		HasObjectType: true,
	}
	ctx := client.Context{
		Session: &session.Session{AccountID: 100, CharID: 200},
		World:   world,
	}

	if !actorCanBeAttackClicked(ctx, actor) {
		t.Fatal("player was not attackable on a PvP map")
	}
	if !actorCanBeSkillTargeted(ctx, session.Skill{ID: 13, Type: skillTargetEnemy}, actor) {
		t.Fatal("player was not an enemy-skill target on a PvP map")
	}

	world.MapProperty = worldstate.MapPropertyNothing
	if actorCanBeAttackClicked(ctx, actor) {
		t.Fatal("player remained attackable outside PvP")
	}
	if actorCanBeSkillTargeted(ctx, session.Skill{ID: 13, Type: skillTargetEnemy}, actor) {
		t.Fatal("player remained an enemy-skill target outside PvP")
	}
}

func TestPvPTargetsLegacyPlayerEntryWithoutObjectType(t *testing.T) {
	world := worldstate.New()
	world.MapProperty = worldstate.MapPropertyFreePvPZone
	ctx := client.Context{
		Session: &session.Session{AccountID: 100, CharID: 200},
		World:   world,
	}
	legacyPlayer := worldstate.Actor{ID: 300, Job: db.JobArcher, Appearance: true}

	if !actorCanBeAttackClicked(ctx, legacyPlayer) {
		t.Fatal("legacy player entry without object type was not attackable on a PvP map")
	}
	if !actorCanBeSkillTargeted(ctx, session.Skill{ID: 13, Type: skillTargetEnemy}, legacyPlayer) {
		t.Fatal("legacy player entry without object type was not an enemy-skill target on a PvP map")
	}

	explicitMob := legacyPlayer
	explicitMob.HasObjectType = true
	explicitMob.ObjectType = actorObjectTypeMob
	if !actorCanBeAttackClicked(ctx, explicitMob) {
		t.Fatal("explicit mob stopped being attackable")
	}
	if actorRepresentsPlayer(explicitMob) {
		t.Fatal("explicit mob was classified as a player from its job value")
	}

	moveOnly := worldstate.Actor{ID: 400, Job: db.JobNovice}
	if actorRepresentsPlayer(moveOnly) || actorCanBeAttackClicked(ctx, moveOnly) {
		t.Fatal("movement-only actor was classified as a player from its zero-value job")
	}
}

func TestApplyPvPRankingUpdatesMatchingActorsOnlyInRankedPvP(t *testing.T) {
	world := worldstate.New()
	world.Player = worldstate.Actor{ID: 200}
	world.Actors[300] = worldstate.Actor{ID: 300}
	ctx := client.Context{
		Session: &session.Session{AccountID: 100, CharID: 200},
		World:   world,
	}

	applyPvPRanking(ctx, network.PvPRanking{ActorID: 100, Rank: 2, Total: 8})
	if world.Player.HasPvPRanking {
		t.Fatal("ranking was applied before a PvP map property")
	}

	applyMapPropertyNotify(ctx, network.MapPropertyNotify{Property: uint16(worldstate.MapPropertyFreePvPZone)})
	applyPvPRanking(ctx, network.PvPRanking{ActorID: 100, Rank: 2, Total: 8})
	if !world.Player.HasPvPRanking || world.Player.PvPRank != 2 || world.Player.PvPTotal != 8 {
		t.Fatalf("local ranking = %+v, want 2/8", world.Player)
	}

	applyPvPRanking(ctx, network.PvPRanking{ActorID: 300, Rank: 5, Total: 8})
	remote := world.Actors[300]
	if !remote.HasPvPRanking || remote.PvPRank != 5 || remote.PvPTotal != 8 {
		t.Fatalf("remote ranking = %+v, want 5/8", remote)
	}

	applyPvPRanking(ctx, network.PvPRanking{ActorID: 999, Rank: 1, Total: 8})
	if _, ok := world.Actors[999]; ok {
		t.Fatal("ranking packet created an actor that was not in the world")
	}
}

func TestMapPropertyChangeClearsStalePvPRankings(t *testing.T) {
	world := worldstate.New()
	world.MapProperty = worldstate.MapPropertyFreePvPZone
	world.Player = worldstate.Actor{ID: 200, PvPRank: 1, PvPTotal: 4, HasPvPRanking: true}
	world.Actors[300] = worldstate.Actor{ID: 300, PvPRank: 2, PvPTotal: 4, HasPvPRanking: true}
	ctx := client.Context{World: world}

	applyMapPropertyNotify(ctx, network.MapPropertyNotify{Property: uint16(worldstate.MapPropertyNothing)})
	if world.MapProperty != worldstate.MapPropertyNothing {
		t.Fatalf("map property = %d, want nothing", world.MapProperty)
	}
	if world.Player.HasPvPRanking || world.Actors[300].HasPvPRanking {
		t.Fatal("rankings survived leaving the PvP map")
	}
}

func TestApplyPvPInfoAddsClassicConsoleSummary(t *testing.T) {
	mode := &WorldMode{}
	mode.applyPvPInfo(network.PvPInfo{Wins: 7, Losses: 3, Points: 42})

	messages := mode.ui.console.Messages()
	if len(messages) != 1 || messages[0].Text != "PvP: 7 wins, 3 losses, 42 points." {
		t.Fatalf("console messages = %+v", messages)
	}
}
