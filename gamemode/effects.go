package gamemode

import (
	"image/color"
	"log"
	"math"
	"time"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
)

const (
	effectProvoke      = 67
	effectPotionRed    = 204
	effectPotionOrange = 205
	effectPotionYellow = 206
	effectPotionWhite  = 207
	effectPotionBlue   = 208
	effectPotionGreen  = 209
	effectFood         = 210
	effectFoodBlue     = 211
	effectStylePotion  = 1
	effectStyleProvoke = 2
)

type worldEffect struct {
	effectID int
	actorID  uint32
	x        int
	y        int
	starts   time.Time
	expires  time.Time
}

type worldEffectSpec struct {
	style    int
	color    color.RGBA
	duration time.Duration
	sfx      []string
}

func (m *WorldMode) addItemUseEffect(ctx Context, ack network.UseItemAck) {
	if ack.Result == 0 {
		return
	}
	effectID := itemUseEffectID(ack.ItemID)
	if effectID <= 0 {
		return
	}
	actorID := ack.AID
	if actorID == 0 && ctx.Session != nil {
		actorID = ctx.Session.AccountID
		if actorID == 0 {
			actorID = ctx.Session.CharID
		}
	}
	if m.addWorldEffect(ctx, effectID, actorID) {
		log.Printf("item effect item=%d actor=%d effect=%d", ack.ItemID, actorID, effectID)
	}
}

func (m *WorldMode) applySkillNoDamageNotify(ctx Context, notify network.SkillNoDamageNotify) {
	if notify.Result == 0 {
		return
	}
	effectID := skillSuccessEffectID(notify.SkillID)
	if effectID <= 0 {
		return
	}
	if m.addWorldEffect(ctx, effectID, notify.TargetID) {
		log.Printf("skill success effect skill=%d src=%d target=%d effect=%d amount=%d", notify.SkillID, notify.SourceID, notify.TargetID, effectID, notify.Amount)
	}
}

func (m *WorldMode) addWorldEffect(ctx Context, effectID int, actorID uint32) bool {
	if ctx.World == nil {
		return false
	}
	spec, ok := worldEffectSpecForID(effectID)
	if !ok {
		return false
	}
	x, y, ok := effectAnchor(ctx, actorID)
	if !ok {
		return false
	}
	now := time.Now()
	m.worldEffects = append(m.worldEffects, worldEffect{
		effectID: effectID,
		actorID:  actorID,
		x:        x,
		y:        y,
		starts:   now,
		expires:  now.Add(spec.duration),
	})
	if len(spec.sfx) > 0 {
		m.scheduleSound(now, spec.sfx...)
	}
	return true
}

func effectAnchor(ctx Context, actorID uint32) (int, int, bool) {
	if ctx.World == nil {
		return 0, 0, false
	}
	if actorID == 0 || isLocalActor(ctx, actorID) {
		return ctx.World.Player.X, ctx.World.Player.Y, true
	}
	if actor, ok := ctx.World.Actors[actorID]; ok {
		return actor.X, actor.Y, true
	}
	return 0, 0, false
}

func itemUseEffectID(itemID uint16) int {
	switch itemID {
	case 501, 507, 512, 513, 515, 516, 545, 549, 557, 562, 563, 564, 565, 566, 567, 568, 569, 570, 571, 572, 574, 575, 576, 577, 578, 579, 580, 581, 583, 584, 585, 586, 587, 588, 589, 590, 591, 592, 593, 594, 595, 596, 597, 598, 607, 608, 663, 669, 680, 685:
		return effectPotionRed
	case 502, 582, 599:
		return effectPotionOrange
	case 503, 508, 546, 11500:
		return effectPotionYellow
	case 504, 509, 547, 11501, 11503:
		return effectPotionWhite
	case 505, 510, 514, 11502, 11504:
		return effectPotionBlue
	case 506, 511:
		return effectPotionGreen
	case 517, 518, 519, 520, 521, 522, 523, 525, 526, 528, 529, 530, 531, 532, 534, 535, 536, 537, 538, 539, 540, 541, 542, 543, 544, 548, 550, 551, 552, 553, 554, 555, 556:
		return effectFood
	case 533:
		return effectFoodBlue
	default:
		return 0
	}
}

func skillSuccessEffectID(skillID uint16) int {
	switch skillID {
	case 6:
		return effectProvoke
	default:
		return 0
	}
}

func worldEffectSpecForID(effectID int) (worldEffectSpec, bool) {
	switch effectID {
	case effectProvoke:
		return worldEffectSpec{
			style:    effectStyleProvoke,
			color:    color.RGBA{R: 255, G: 70, B: 42, A: 255},
			duration: 900 * time.Millisecond,
			sfx:      []string{"effect\\ef_provoke.wav", "effect\\provoke.wav", "provoke.wav"},
		}, true
	case effectPotionRed:
		return potionEffectSpec(color.RGBA{R: 255, G: 82, B: 70, A: 255}), true
	case effectPotionOrange:
		return potionEffectSpec(color.RGBA{R: 255, G: 145, B: 58, A: 255}), true
	case effectPotionYellow:
		return potionEffectSpec(color.RGBA{R: 255, G: 226, B: 76, A: 255}), true
	case effectPotionWhite:
		return potionEffectSpec(color.RGBA{R: 245, G: 245, B: 255, A: 255}), true
	case effectPotionBlue:
		return potionEffectSpec(color.RGBA{R: 92, G: 150, B: 255, A: 255}), true
	case effectPotionGreen:
		return potionEffectSpec(color.RGBA{R: 78, G: 225, B: 98, A: 255}), true
	case effectFood:
		return potionEffectSpec(color.RGBA{R: 255, G: 182, B: 86, A: 255}), true
	case effectFoodBlue:
		return potionEffectSpec(color.RGBA{R: 132, G: 112, B: 255, A: 255}), true
	default:
		return worldEffectSpec{}, false
	}
}

func potionEffectSpec(c color.RGBA) worldEffectSpec {
	return worldEffectSpec{
		style:    effectStylePotion,
		color:    c,
		duration: 850 * time.Millisecond,
		sfx:      []string{"effect\\ef_awakening.wav", "effect\\awakening.wav", "effect\\potion.wav", "potion.wav"},
	}
}

func (m *WorldMode) drawWorldEffects(screen *render.Image, ctx Context, projection sceneProjection, now time.Time) {
	if len(m.worldEffects) == 0 || screen == nil || ctx.World == nil {
		return
	}
	if m.whitePixel == nil {
		m.whitePixel = render.NewImage(1, 1)
		m.whitePixel.Fill(color.White)
	}
	active := m.worldEffects[:0]
	for _, effect := range m.worldEffects {
		if now.After(effect.expires) {
			continue
		}
		spec, ok := worldEffectSpecForID(effect.effectID)
		if !ok {
			continue
		}
		active = append(active, effect)
		x, y := float64(effect.x), float64(effect.y)
		if actor, ok := ctx.World.Actors[effect.actorID]; ok {
			x, y = actor.RenderPosition(now)
		} else if isLocalActor(ctx, effect.actorID) {
			x, y = ctx.World.Player.RenderPosition(now)
		}
		progress := worldEffectProgress(effect, now)
		worldX := cellCenter(x)
		worldY := cellCenter(y)
		worldZ := terrainHeightAt(ctx.World, x, y) + 0.07
		switch spec.style {
		case effectStyleProvoke:
			drawProvokeEffect(screen, m.whitePixel, worldX, worldY, worldZ, progress, spec.color)
		default:
			drawPotionEffect(screen, m.whitePixel, worldX, worldY, worldZ, progress, spec.color)
		}
	}
	m.worldEffects = active
}

func worldEffectProgress(effect worldEffect, now time.Time) float64 {
	duration := effect.expires.Sub(effect.starts)
	if duration <= 0 {
		return 1
	}
	return clampFloat(float64(now.Sub(effect.starts))/float64(duration), 0, 1)
}

func drawPotionEffect(screen, white *render.Image, x, y, z, progress float64, c color.RGBA) {
	alpha := 1 - progress
	drawWorldRadialGradient(screen, white, x, y, z, 0.02, 0.34+progress*0.18, withAlpha(c, alpha*0.30), 48)
	drawWorldSoftRing(screen, white, x, y, z+0.01, 0.24+progress*0.52, 0.22, withAlpha(c, alpha*0.75), 48)
	drawWorldCylinderBand(screen, white, nil, x, y, z+0.05+progress*0.55, 0.18+progress*0.08, 0.06, 0.22, withAlpha(c, alpha*0.40), 32)
}

func drawProvokeEffect(screen, white *render.Image, x, y, z, progress float64, c color.RGBA) {
	alpha := 1 - progress
	radius := 0.22 + 0.18*math.Sin(progress*math.Pi)
	drawWorldSoftRing(screen, white, x, y, z+0.08, radius, 0.12, withAlpha(c, alpha*0.80), 32)
	drawWorldCylinderBand(screen, white, nil, x, y, z+0.10, radius*0.9, radius*1.18, 0.92, withAlpha(c, alpha*0.28), 24)
	for i := 0; i < 3; i++ {
		step := math.Mod(progress+float64(i)*0.24, 1)
		drawWorldSoftRing(screen, white, x, y, z+0.18+step*0.72, 0.12+step*0.28, 0.08, withAlpha(c, (1-step)*alpha*0.72), 32)
	}
}
