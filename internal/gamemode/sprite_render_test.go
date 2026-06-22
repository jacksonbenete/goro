package gamemode

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/kivutar/goro/internal/res"
)

func TestNormalizeSpriteDirection(t *testing.T) {
	cases := map[int]int{
		0:  0,
		7:  7,
		8:  0,
		-1: 7,
		-9: 7,
	}
	for input, want := range cases {
		if got := normalizeSpriteDirection(input); got != want {
			t.Fatalf("normalizeSpriteDirection(%d) = %d, want %d", input, got, want)
		}
	}
}

func TestResolveSpriteActionPrefersFamilyAndDirection(t *testing.T) {
	act := &res.ACT{Actions: make([]res.ACTAction, 16)}
	act.Actions[11] = res.ACTAction{Animations: []res.ACTAnimation{{}}, DelayMS: 100}
	index, _, ok := resolveSpriteAction(act, spriteActionWalk, 3)
	if !ok {
		t.Fatal("expected action")
	}
	if index != 11 {
		t.Fatalf("index = %d, want 11", index)
	}
}

func TestResolveSpriteActionFallsBackToFamilyBase(t *testing.T) {
	act := &res.ACT{Actions: make([]res.ACTAction, 16)}
	act.Actions[8] = res.ACTAction{Animations: []res.ACTAnimation{{}}, DelayMS: 100}
	index, _, ok := resolveSpriteAction(act, spriteActionWalk, 4)
	if !ok {
		t.Fatal("expected action")
	}
	if index != 8 {
		t.Fatalf("index = %d, want 8", index)
	}
}

func TestSpriteMotionIndexUsesActionDelay(t *testing.T) {
	action := res.ACTAction{
		Animations: []res.ACTAnimation{{}, {}, {}},
		DelayMS:    100,
	}
	started := time.Unix(0, 0)
	got := spriteMotionIndex(action, started, started.Add(250*time.Millisecond), true)
	if got != 2 {
		t.Fatalf("motion index = %d, want 2", got)
	}
}

func TestAttachmentDeltaMatchesAttachPointAttribute(t *testing.T) {
	base := res.ACTAnimation{Pos: []res.ACTPosition{
		{X: 1, Y: 2, Attr: 7},
		{X: 30, Y: 40, Attr: 2},
	}}
	attached := res.ACTAnimation{Pos: []res.ACTPosition{
		{X: 10, Y: 15, Attr: 2},
	}}
	x, y := attachmentDelta(base, attached)
	if x != 20 || y != 25 {
		t.Fatalf("attachment delta = (%d,%d), want (20,25)", x, y)
	}
}

func TestHumanoidIdleHeadMotionDoesNotCycle(t *testing.T) {
	action := res.ACTAction{
		Animations: []res.ACTAnimation{{}, {}, {}},
		DelayMS:    100,
	}
	if got := selectHeadMotion(spriteActionIdle, 2, action); got != 0 {
		t.Fatalf("idle head motion = %d, want 0", got)
	}
}

func TestHumanoidWalkHeadMotionFollowsBodyMotion(t *testing.T) {
	action := res.ACTAction{
		Animations: []res.ACTAnimation{{}, {}, {}},
		DelayMS:    100,
	}
	if got := selectHeadMotion(spriteActionWalk, 2, action); got != 2 {
		t.Fatalf("walk head motion = %d, want 2", got)
	}
}

func TestSpriteLayerCenterTreatsPositiveYAsScreenDown(t *testing.T) {
	_, y := spriteLayerCenter(5, 5, res.ACTLayer{Y: 3})
	if y != 8 {
		t.Fatalf("layer center Y = %.1f, want 8.0", y)
	}
}

func TestDebugPlayerSpriteBillboard(t *testing.T) {
	if os.Getenv("GORO_DEBUG_PLAYER_SPRITE") != "1" {
		t.Skip("set GORO_DEBUG_PLAYER_SPRITE=1")
	}
	manager, err := res.NewManager("/home/kivutar/Téléchargements/OldRO")
	if err != nil {
		t.Fatal(err)
	}
	for _, sex := range []byte{0, 1} {
		view, status := loadHumanoidSpriteView(manager, 0, 1, sex, 0, 0, "debug player")
		if view == nil {
			t.Logf("sex=%d load failed: %s", sex, status)
			continue
		}
		billboard, ok := humanoidBillboardForState(view, spriteState{actionFamily: spriteActionIdle, direction: 0}, time.Now())
		if !ok {
			t.Logf("sex=%d billboard failed: %s", sex, status)
			continue
		}
		bodyIndex, bodyAction, ok := resolveSpriteAction(view.body.act, spriteActionIdle, 0)
		if !ok {
			t.Fatalf("sex=%d missing body action", sex)
		}
		minX, minY, maxX, maxY := spriteAnimationBounds(view.body, bodyAction.Animations[0], humanoidBillboardAnchorX, humanoidBillboardAnchorY, 0, 0)
		t.Logf("sex=%d %s action=%d bounds=(%.1f,%.1f)-(%.1f,%.1f) anchor=(%.0f,%.0f) image=%dx%d", sex, status, bodyIndex, minX, minY, maxX, maxY, billboard.anchorX, billboard.anchorY, humanoidBillboardWidth, humanoidBillboardHeight)
	}
}

func spriteAnimationBounds(view *playerSpriteView, anim res.ACTAnimation, anchorX, anchorY float64, posX, posY int32) (float64, float64, float64, float64) {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, layer := range anim.Layers {
		if layer.Index < 0 {
			continue
		}
		frameIndex := int(layer.Index)
		if layer.SPRType == res.SPRFrameRGBA {
			frameIndex += view.spr.RGBAIndex
		}
		if frameIndex < 0 || frameIndex >= len(view.spr.Frames) {
			continue
		}
		frame := view.spr.Frames[frameIndex]
		cx := anchorX + float64(posX) + float64(layer.X)
		cy := anchorY + float64(posY) + float64(layer.Y)
		w := float64(frame.Width)
		h := float64(frame.Height)
		minX = math.Min(minX, cx-w/2)
		minY = math.Min(minY, cy-h/2)
		maxX = math.Max(maxX, cx+w/2)
		maxY = math.Max(maxY, cy+h/2)
	}
	return minX, minY, maxX, maxY
}
