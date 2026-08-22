package ui

import (
	"os"
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/res"
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
	if x, y := pvpCounterBounds(800, 600); x != 560 || y != 504 {
		t.Fatalf("counter bounds = %d,%d, want 560,504", x, y)
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

func TestPvPCounterRankFontRealWhenConfigured(t *testing.T) {
	root := os.Getenv("GORO_DATA_DIR")
	if root == "" {
		t.Skip("set GORO_DATA_DIR to test the PvP rank font against real client data")
	}
	manager, err := res.NewManager(root)
	if err != nil {
		t.Fatal(err)
	}
	var sprites pvpRankSpriteSet
	img := sprites.image(manager, 3, 12)
	if img == nil {
		t.Fatal("PvP rank font image was not composed")
	}
	if got := img.Bounds().Size(); got.X != pvpCounterWidth || got.Y != pvpCounterHeight {
		t.Fatalf("PvP rank font size = %v, want %dx%d", got, pvpCounterWidth, pvpCounterHeight)
	}
	opaque := 0
	for y := 0; y < pvpCounterHeight; y++ {
		for x := 0; x < pvpCounterWidth; x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			if alpha != 0 {
				opaque++
			}
		}
	}
	if opaque == 0 {
		t.Fatal("PvP rank font image is fully transparent")
	}
	if cached := sprites.image(manager, 3, 12); cached != img {
		t.Fatal("unchanged PvP rank rebuilt its sprite image")
	}
}

func TestPvPRankFontRejectsIncompleteGlyphSet(t *testing.T) {
	act := &res.ACT{Actions: make([]res.ACTAction, 10)}
	spr := &res.SPR{Frames: make([]res.SPRFrame, 11)}
	if validPvPRankSpriteSet(act, spr) {
		t.Fatal("rank font without the slash action was accepted")
	}
}
