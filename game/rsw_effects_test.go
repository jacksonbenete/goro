package game

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

func TestSmokeEffectSpecMatchesReferenceMapEffect(t *testing.T) {
	spec, ok := worldEffectSpecForID(effectSmoke)
	if !ok {
		t.Fatal("smoke effect spec missing")
	}
	if spec.duration != 10*time.Second || len(spec.components) != 1 {
		t.Fatalf("smoke spec = %+v", spec)
	}
	component := spec.components[0]
	if component.kind != effectComponent3D || component.spriteFile != "\xb1\xbc\xb6\xd2\xbf\xac\xb1\xe2" {
		t.Fatalf("smoke component sprite = %+v", component)
	}
	if component.duplicate != 10 || component.duplicateDelay != time.Second || component.duration != 10*time.Second {
		t.Fatalf("smoke timing = duplicate %d delay %s duration %s", component.duplicate, component.duplicateDelay, component.duration)
	}
	if component.posZ != 0 || component.posZEnd != 20 || component.posXEndRand != 3 || !component.posXSmooth {
		t.Fatalf("smoke placement = %+v", component)
	}
	if component.sizeStart != effectTableSize(70) || component.sizeEnd != effectTableSize(300) || !component.sizeSmooth {
		t.Fatalf("smoke size = %.3f %.3f smooth=%t", component.sizeStart, component.sizeEnd, component.sizeSmooth)
	}
	if component.alphaMax != 0.8 || !component.fadeOut || !component.blendAdditive || !component.rotate || !component.rotateWithCamera {
		t.Fatalf("smoke alpha/rotation = %+v", component)
	}
}
