package ui

import (
	"testing"

	"github.com/kivutar/goro/client"
	worldstate "github.com/kivutar/goro/world"
)

func TestPvPCounterPublishesOnlyForValidLocalPvPRanking(t *testing.T) {
	manager := &escapeMenuTestUIManager{}
	world := worldstate.New()
	ctx := client.Context{UIManager: manager, World: world, ScreenW: 800, ScreenH: 600}
	var counter PvPCounter

	counter.Update(ctx)
	if len(manager.overlays) != 0 {
		t.Fatalf("counter published outside PvP: %d overlays", len(manager.overlays))
	}

	world.MapProperty = worldstate.MapPropertyFreePvPZone
	world.Player.PvPRank = 3
	world.Player.PvPTotal = 12
	world.Player.HasPvPRanking = true
	counter.Update(ctx)
	if len(manager.overlays) != 1 {
		t.Fatalf("counter overlays = %d, want 1", len(manager.overlays))
	}
	if counter.widget == nil || counter.widget.text != "3/12" {
		t.Fatalf("counter text = %q, want 3/12", counter.widget.text)
	}
	root := counter.root
	counter.Update(ctx)
	if len(manager.overlays) != 1 || counter.root != root {
		t.Fatal("unchanged counter rebuilt its overlay")
	}

	world.MapProperty = worldstate.MapPropertyNothing
	counter.Update(ctx)
	if len(manager.overlays) != 0 {
		t.Fatalf("counter remained outside PvP: %d overlays", len(manager.overlays))
	}
}

func TestPvPCounterBoundsStayOnScreen(t *testing.T) {
	if x, y := pvpCounterBounds(800, 600); x != 632 || y != 544 {
		t.Fatalf("counter bounds = %d,%d, want 632,544", x, y)
	}
	if x, y := pvpCounterBounds(100, 30); x != 0 || y != 0 {
		t.Fatalf("small-screen counter bounds = %d,%d, want 0,0", x, y)
	}
}

func TestPvPCounterDoesNotBlockWorldPointerInput(t *testing.T) {
	manager := NewManager()
	world := worldstate.New()
	world.MapProperty = worldstate.MapPropertyFreePvPZone
	world.Player.PvPRank = 1
	world.Player.PvPTotal = 2
	world.Player.HasPvPRanking = true
	ctx := client.Context{UIManager: manager, World: world, ScreenW: 800, ScreenH: 600}
	var counter PvPCounter

	counter.Update(ctx)
	if manager.PointerBlocked(700, 570) {
		t.Fatal("PvP counter blocked pointer input through its transparent overlay")
	}
}
