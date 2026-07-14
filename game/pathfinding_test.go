package game

import (
	"testing"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

func TestWalkPathBuildsPathAroundObstacle(t *testing.T) {
	gat := testPathGAT(3, 3, map[worldstate.WalkStep]bool{
		{X: 1, Y: 0}: true,
	})

	path := walkPath(gat, 0, 0, 2, 0)
	if len(path) < 4 {
		t.Fatalf("path too short: %+v", path)
	}
	for _, step := range path {
		if step.X == 1 && step.Y == 0 {
			t.Fatalf("path crosses blocked cell: %+v", path)
		}
	}
	w := worldstate.New()
	setPlayerMovementPathAt(client.Context{World: w}, path, 0, 0, 2, 0, 0, time.Now(), 0)
	if got := w.Player.MoveDuration; got <= 300*time.Millisecond {
		t.Fatalf("duration = %s, want longer than direct two-cell move", got)
	}
}

func TestWalkPathStoresFinalSegmentDirection(t *testing.T) {
	gat := testPathGAT(3, 3, map[worldstate.WalkStep]bool{
		{X: 1, Y: 0}: true,
	})
	path := walkPath(gat, 0, 0, 2, 0)
	w := worldstate.New()
	setPlayerMovementPathAt(client.Context{World: w}, path, 0, 0, 2, 0, directionFromDelta(0, 0, 2, 0, 0), time.Now(), 0)

	if got, want := w.Player.Dir, directionFromDelta(2, 1, 2, 0, 0); got != want {
		t.Fatalf("player dir = %d, want final segment dir %d; path=%+v", got, want, w.Player.MovePath)
	}
	if w.Dir != w.Player.Dir {
		t.Fatalf("world dir = %d, want player dir %d", w.Dir, w.Player.Dir)
	}
}

func testPathGAT(width, height int, blocked map[worldstate.WalkStep]bool) *res.GAT {
	gat := &res.GAT{
		Width:  width,
		Height: height,
		Cells:  make([]res.GATCell, width*height),
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cellType := res.GATTypeWalkable | res.GATTypeSnipable
			if blocked[worldstate.WalkStep{X: x, Y: y}] {
				cellType = res.GATTypeNone
			}
			gat.Cells[y*width+x] = res.GATCell{Type: cellType}
		}
	}
	return gat
}
