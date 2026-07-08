package session

import (
	"testing"
	"time"
)

func TestServerTickEstimation(t *testing.T) {
	s := New()
	start := time.Unix(100, 0)
	s.SyncServerTick(1000, start)

	tick, ok := s.EstimatedServerTick(start.Add(250 * time.Millisecond))
	if !ok {
		t.Fatal("server tick not estimated")
	}
	if tick != 1250 {
		t.Fatalf("tick = %d, want 1250", tick)
	}

	elapsed, ok := s.ElapsedSinceServerTick(1100, start.Add(250*time.Millisecond))
	if !ok {
		t.Fatal("server tick elapsed not estimated")
	}
	if elapsed != 150*time.Millisecond {
		t.Fatalf("elapsed = %s, want 150ms", elapsed)
	}
}
