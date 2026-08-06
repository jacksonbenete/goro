package game

import (
	"image/color"
	"math"
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
	if component.kind != effectComponent3D || component.spriteFile != "굴뚝연기" {
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

func TestReferenceRSWMapEffectIDsHaveSpecs(t *testing.T) {
	for _, effectID := range []int{
		effectSmoke,
		effectFirefly,
		effectTorch,
		effectBubble,
		effectDragonSmoke,
		effectBanjjakii,
		effectMapPillar,
		effectMapPillar2,
		effectMapPillar3,
		effectMapPillar4,
		effectTorchRed,
		effectTorchGreen,
		effectTorchPurple,
		effectMapGhost,
		effectGlow1,
		effectGlow2,
		effectGlow4,
		effectBubbleDrop,
		effectRainbow,
	} {
		if _, ok := worldEffectSpecForID(effectID); !ok {
			t.Fatalf("missing RSW map effect spec %d", effectID)
		}
	}
}

func TestMapPillarEffectSpecsMatchReferenceVariants(t *testing.T) {
	type variant struct {
		name        string
		id          int
		texture     string
		radiusStart float64
		alpha       float64
		tint        color.RGBA
	}
	for _, tc := range []variant{
		{"EF_MAPPILLAR", effectMapPillar, "ring_blue", 2, 50.0 / 255.0, color.RGBA{R: 110, G: 175, B: 255, A: 255}},
		{"EF_MAPPILLAR2", effectMapPillar2, "ring_blue", 11, 70.0 / 255.0, color.RGBA{R: 110, G: 175, B: 255, A: 255}},
		{"EF_MAPPILLAR3", effectMapPillar3, "magic_green", 2, 50.0 / 255.0, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"EF_MAPPILLAR4", effectMapPillar4, "ring_red", 2, 50.0 / 255.0, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
	} {
		spec, ok := worldEffectSpecForID(tc.id)
		if !ok || spec.duration != 16*time.Second || len(spec.components) != 1 {
			t.Fatalf("%s spec = %+v ok=%t", tc.name, spec, ok)
		}
		component := spec.components[0]
		if component.kind != effectComponentFUNC || component.funcName != "MapPillar" || component.funcAdapter != effectFuncMapPillar || component.textureName != tc.texture || component.duration != 16*time.Second {
			t.Fatalf("%s component timing/texture = %+v", tc.name, component)
		}
		if component.alphaMax != tc.alpha || component.color != tc.tint || component.blendMode != 2 || !component.blendAdditive || component.attachedEntity {
			t.Fatalf("%s component alpha/color/blend = %+v", tc.name, component)
		}
		if component.bottomSize != tc.radiusStart*0.2 || component.height != 24 {
			t.Fatalf("%s component geometry = %+v", tc.name, component)
		}
	}
}

func TestMapPillarBandHeightMatchesReferenceLifecycle(t *testing.T) {
	maxHeight := 24.0
	if got := mapPillarBandHeight(199, maxHeight); got != 0 {
		t.Fatalf("height before grow = %.3f, want 0", got)
	}
	if got := mapPillarBandHeight(245, maxHeight); math.Abs(got-maxHeight*math.Sin(45*math.Pi/180)) > 0.0001 {
		t.Fatalf("height mid grow = %.3f", got)
	}
	if got := mapPillarBandHeight(291, maxHeight); got != maxHeight {
		t.Fatalf("height after grow = %.3f, want %.3f", got, maxHeight)
	}
	if got := mapPillarBandHeight(800, maxHeight); math.Abs(got-23.8) > 0.0001 {
		t.Fatalf("height shrink start = %.3f, want 23.8", got)
	}
	if got := mapPillarBandHeight(919, maxHeight); got != 0 {
		t.Fatalf("height after shrink = %.3f, want 0", got)
	}
}
