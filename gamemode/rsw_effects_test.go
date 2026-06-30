package gamemode

import (
	"testing"
	"time"

	"github.com/kivutar/goro/res"
)

func TestRSWEffectWorldPositionUsesMapOffset(t *testing.T) {
	gnd := &res.GND{Width: 300, Height: 200}
	effect := res.RSWEffect{Position: res.RSWVector3{X: 12, Y: -3, Z: 34}}
	x, y, z := rswEffectWorldPosition(gnd, effect)
	if x != 312 || y != 234 || z != -2 {
		t.Fatalf("rswEffectWorldPosition = %.1f, %.1f, %.1f; want 312, 234, -2", x, y, z)
	}
}

func TestLoopingRSWEffectStartUsesDelayPhase(t *testing.T) {
	now := time.Unix(10, 250*int64(time.Millisecond))
	effect := res.RSWEffect{Delay: 100}
	start := loopingRSWEffectStart(effect, 0, 600*time.Millisecond, now)
	if !start.Equal(time.Unix(9, 700*int64(time.Millisecond))) {
		t.Fatalf("loopingRSWEffectStart = %s", start)
	}
}

func TestMapEffectComponentUsesWorldSizedSpriteWithoutChangingPlacement(t *testing.T) {
	component := worldEffectComponent{
		kind:       effectComponent3D,
		spriteFile: "torch_01",
		posX:       0.1,
		posY:       -0.2,
		posZ:       0.8,
	}
	got := mapEffectComponent(component)
	if !got.worldSizedSprite {
		t.Fatalf("mapEffectComponent did not mark sprite as world sized: %+v", got)
	}
	if got.posX != 0.1 || got.posY != -0.2 || got.posZ != 0.8 {
		t.Fatalf("mapEffectComponent = %+v", got)
	}
}
