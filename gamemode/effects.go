package gamemode

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"strings"
	"time"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

const (
	effectFireBolt      = 10019
	effectMagicTarget   = 10020
	effectCastRing      = 10021
	effectProvoke       = 67
	effectEndure        = 11
	effectBeginSpell    = 12
	effectSafetyWall    = 315
	effectColdBolt      = 10014
	effectBashBegin     = 16
	effectBashHit       = 1
	effectMammonite     = 10
	effectSight         = 22
	effectSoulStrike    = 15
	effectMagnumBreak   = 17
	effectSteal         = 18
	effectPoisonAttack  = 20
	effectDetoxication  = 21
	effectStoneCurse    = 23
	effectFireBall      = 24
	effectFireWall      = 25
	effectFrostDiver    = 27
	effectFrostDiverHit = 28
	effectLightningBolt = 29
	effectThunderStorm  = 30
	effectIncAgility    = 37
	effectDecAgility    = 38
	effectAqua          = 39
	effectSignum        = 40
	effectAngelus       = 41
	effectBlessing      = 42
	effectFireHit       = 49
	effectFireSplashHit = 50
	effectColdHit       = 51
	effectWindHit       = 52
	effectBeginSpell2   = 54
	effectBeginSpell3   = 55
	effectBeginSpell4   = 56
	effectBeginSpell5   = 57
	effectBeginSpell6   = 58
	effectBeginSpell7   = 59
	effectFirefly       = 45
	effectTorch         = 47
	effectBubble        = 109
	effectCure          = 66
	effectPneuma        = 141
	effectConcentration = 153
	effectRefineOK      = 154
	effectRefineFail    = 155
	effectTeleportation = 304
	effectPharmacyOK    = 305
	effectPharmacyFail  = 306
	effectHeal          = 312
	effectPortal        = 317
	effectHealOffensive = 320
	effectBaseLevelUp   = 371
	effectJobLevelUp    = 158
	effectPotionRed     = 204
	effectPotionOrange  = 205
	effectPotionYellow  = 206
	effectPotionWhite   = 207
	effectPotionBlue    = 208
	effectPotionGreen   = 209
	effectFood          = 210
	effectFoodBlue      = 211
	effectEnergyCoat    = 169
)

type effectPrimitiveKind int

const (
	effectPrimitiveSTR effectPrimitiveKind = iota + 1
	effectPrimitiveCylinder
	effectPrimitiveBillboard
	effectPrimitiveBashHit
	effectPrimitive2D
	effectPrimitive3D
	effectPrimitiveSPR
	effectPrimitiveGroundPlane
	effectPrimitiveCastRing
)

type worldEffect struct {
	effectID int
	actorID  uint32
	targetID uint32
	x        int
	y        int
	starts   time.Time
	expires  time.Time
	duration time.Duration
}

type worldEffectSpec struct {
	duration   time.Duration
	sfx        []string
	components []worldEffectComponent
}

type worldEffectComponent struct {
	kind             effectPrimitiveKind
	color            color.RGBA
	duration         time.Duration
	delay            time.Duration
	duplicateDelay   time.Duration
	strFile          string
	strRandMin       int
	strRandMax       int
	attachedEntity   bool
	texturePath      string
	textureName      string
	textureFile      string
	textureFiles     []string
	spriteFile       string
	spriteHead       bool
	spriteDirection  bool
	spriteRepeat     bool
	spriteStopAtEnd  bool
	spriteFrame      int
	spriteDelay      time.Duration
	spriteXOffset    float64
	spriteYOffset    float64
	fromSrc          bool
	toSrc            bool
	arc              float64
	retreat          float64
	alphaMax         float64
	alphaMaxDelta    float64
	fade             bool
	fadeIn           bool
	fadeOut          bool
	rotate           bool
	rotateWithCamera bool
	fixedPerspective bool
	rotateToTarget   bool
	worldSizedSprite bool
	animation        int
	bottomSize       float64
	topSize          float64
	height           float64
	posX             float64
	posY             float64
	posZ             float64
	posXEnd          float64
	posYEnd          float64
	posZEnd          float64
	posXRand         float64
	posYRand         float64
	posXStartRand    float64
	posYStartRand    float64
	posZStartRand    float64
	posXStartMiddle  float64
	posYStartMiddle  float64
	posZStartMiddle  float64
	posXEndRand      float64
	posYEndRand      float64
	posZEndRand      float64
	posXEndMiddle    float64
	posYEndMiddle    float64
	posZEndMiddle    float64
	posXSmooth       bool
	posYSmooth       bool
	posZSmooth       bool
	sizeStart        float64
	sizeEnd          float64
	sizeRand         float64
	sizeStartX       float64
	sizeStartY       float64
	sizeEndX         float64
	sizeEndY         float64
	sizeRandX        float64
	sizeRandY        float64
	sizeDelta        float64
	sizeSmooth       bool
	angleStart       float64
	angleEnd         float64
	totalCircleSides int
	circleSides      int
	duplicate        int
	angleZRandom     float64
	blendAdditive    bool
	overlay          bool
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
	if effectID == effectTeleportation && isLocalActor(ctx, actorID) {
		actorID = 0
	}
	if m.addWorldEffect(ctx, effectID, actorID) {
		log.Printf("item effect item=%d actor=%d effect=%d", ack.ItemID, actorID, effectID)
	}
}

func (m *WorldMode) applySkillNoDamageNotify(ctx Context, notify network.SkillNoDamageNotify) {
	if notify.Result == 0 {
		return
	}
	now := time.Now()
	m.startSkillNoDamageSourceAnimation(ctx, notify, now)
	if effectID := skillEffectID(notify.SkillID); effectID > 0 {
		if m.addWorldEffectBetweenAt(ctx, effectID, notify.TargetID, notify.SourceID, now) {
			log.Printf("skill effect skill=%d src=%d target=%d effect=%d amount=%d", notify.SkillID, notify.SourceID, notify.TargetID, effectID, notify.Amount)
		}
	}
	effectID := skillSuccessEffectID(notify.SkillID)
	if effectID <= 0 {
		return
	}
	if m.addWorldEffectBetweenAt(ctx, effectID, notify.TargetID, notify.SourceID, now) {
		log.Printf("skill success effect skill=%d src=%d target=%d effect=%d amount=%d", notify.SkillID, notify.SourceID, notify.TargetID, effectID, notify.Amount)
	}
	if notify.SkillID == 28 && !isLocalActor(ctx, notify.TargetID) {
		m.addTargetRecoveryFloater(ctx, notify.TargetID, int(notify.Amount), recoveryHPColor, damageFloaterRecoveryHP, now)
	}
}

func (m *WorldMode) applySkillCastNotify(ctx Context, notify network.SkillCastNotify) {
	if notify.DelayTime == 0 {
		return
	}
	duration := time.Duration(notify.DelayTime) * time.Millisecond
	now := time.Now()
	m.startSkillCastSourceAnimation(ctx, notify, duration, now)
	m.addSkillCastEffects(ctx, notify.SkillID, notify.Property, notify.SourceID, notify.TargetID, duration, now, "server")
}

func (m *WorldMode) startSkillNoDamageSourceAnimation(ctx Context, notify network.SkillNoDamageNotify, now time.Time) {
	source, ok, _ := actorForCombatID(ctx, notify.SourceID)
	if !ok {
		return
	}
	m.faceSkillSource(ctx, notify.SourceID, notify.TargetID, 0, 0)
	m.startCombatAnimation(ctx, notify.SourceID, skillActionFamilyForActor(source, notify.SkillID), now, defaultAttackAnimationDuration)
}

func (m *WorldMode) startSkillCastSourceAnimation(ctx Context, notify network.SkillCastNotify, duration time.Duration, now time.Time) {
	m.faceSkillSource(ctx, notify.SourceID, notify.TargetID, int(notify.X), int(notify.Y))
	m.startSkillSourceCastAnimation(ctx, notify.SourceID, notify.SkillID, duration, now)
}

func (m *WorldMode) startSkillSourceCastAnimation(ctx Context, sourceID uint32, skillID uint16, duration time.Duration, now time.Time) {
	source, ok, _ := actorForCombatID(ctx, sourceID)
	if !ok {
		return
	}
	m.startFixedMotionCombatAnimation(ctx, sourceID, skillActionFamilyForActor(source, skillID), skillCastMotion, now, duration)
}

func (m *WorldMode) faceSkillSource(ctx Context, sourceID, targetID uint32, cellX, cellY int) {
	source, sourceOK, sourceLocal := actorForCombatID(ctx, sourceID)
	if !sourceOK {
		return
	}
	if target, targetOK, _ := actorForCombatID(ctx, targetID); targetOK {
		m.faceCombatSource(ctx, source, sourceLocal, target)
		return
	}
	if cellX == 0 && cellY == 0 {
		return
	}
	dir := directionFromDelta(source.X, source.Y, cellX, cellY, source.Dir)
	if sourceLocal {
		ctx.World.Player.Dir = dir
		ctx.World.Dir = dir
		return
	}
	source.Dir = dir
	ctx.World.UpsertActor(source)
}

func (m *WorldMode) applyGroundSkillNotify(ctx Context, notify network.GroundSkillNotify) {
	effectID := skillGroundEffectID(notify.SkillID)
	if effectID <= 0 {
		return
	}
	if m.addWorldEffectAtCellIfMissing(ctx, effectID, int(notify.X), int(notify.Y), time.Now()) {
		log.Printf("ground skill effect skill=%d src=%d level=%d cell=%d,%d effect=%d", notify.SkillID, notify.SourceID, notify.Level, notify.X, notify.Y, effectID)
	}
}

func (m *WorldMode) applySkillUnitEntry(ctx Context, entry network.SkillUnitEntry) {
	if !entry.Visible {
		return
	}
	effectID := skillUnitEffectID(entry.UnitID)
	if effectID <= 0 {
		return
	}
	if m.addWorldEffectAtCellWithActor(ctx, effectID, entry.ID, int(entry.X), int(entry.Y), time.Now()) {
		log.Printf("skill unit effect unit=%d id=%d creator=%d cell=%d,%d effect=%d", entry.UnitID, entry.ID, entry.CreatorID, entry.X, entry.Y, effectID)
	}
}

func (m *WorldMode) applySkillUnitDisappear(disappear network.SkillUnitDisappear) {
	if disappear.ID == 0 {
		return
	}
	active := m.worldEffects[:0]
	removed := false
	for _, effect := range m.worldEffects {
		if effect.actorID == disappear.ID {
			removed = true
			continue
		}
		active = append(active, effect)
	}
	m.worldEffects = active
	if removed {
		log.Printf("skill unit effect removed id=%d", disappear.ID)
	}
}

func (m *WorldMode) applySpecialEffectNotify(ctx Context, notify network.SpecialEffectNotify) {
	effectID := specialEffectID(notify.EffectID)
	if effectID <= 0 {
		return
	}
	if m.addWorldEffectIfMissing(ctx, effectID, notify.AID) {
		log.Printf("special effect actor=%d special=%d effect=%d", notify.AID, notify.EffectID, effectID)
	}
}

func (m *WorldMode) applySkillFailAck(ctx Context, ack network.SkillFailAck) {
	if ack.Result != 0 {
		return
	}
	message := skillFailMessage(ack)
	log.Printf("skill fail ack skill=%d num=%d item=%d result=%d cause=%d msg=%q", ack.SkillID, ack.Number, ack.ItemID, ack.Result, ack.Cause, message)
	m.console.addErrorMessage("%s", message)
}

func specialEffectID(effectID uint32) int {
	switch effectID {
	case network.SpecialEffectBaseLevelUp:
		return effectBaseLevelUp
	case network.SpecialEffectJobLevelUp:
		return effectJobLevelUp
	default:
		if _, ok := worldEffectSpecForID(int(effectID)); ok {
			return int(effectID)
		}
		return 0
	}
}

func skillFailMessage(ack network.SkillFailAck) string {
	if ack.SkillID == 1 && ack.Cause == 0 {
		switch ack.Number {
		case 0:
			return "Basic skill failed."
		case 1:
			return "Cannot use emotions."
		case 2:
			return "Cannot sit."
		case 3:
			return "Cannot chat."
		case 4:
			return "Cannot form a party."
		case 5:
			return "Cannot shout."
		case 6:
			return "Cannot PK."
		case 7:
			return "Cannot align."
		}
	}
	switch ack.Cause {
	case 0:
		return "Action failed."
	case 1:
		return "Not enough SP."
	case 2:
		return "Not enough HP."
	case 4:
		return "Action is still on cooldown."
	case 5:
		return "Not enough Zeny."
	case 9:
		return "Too much weight."
	default:
		return "Action failed."
	}
}

func (m *WorldMode) addWorldEffect(ctx Context, effectID int, actorID uint32) bool {
	return m.addWorldEffectAt(ctx, effectID, actorID, time.Now())
}

func (m *WorldMode) addWorldEffectIfMissing(ctx Context, effectID int, actorID uint32) bool {
	if m.hasActiveWorldEffect(effectID, actorID, time.Now()) {
		return false
	}
	return m.addWorldEffect(ctx, effectID, actorID)
}

func (m *WorldMode) hasActiveWorldEffect(effectID int, actorID uint32, now time.Time) bool {
	for _, effect := range m.worldEffects {
		if effect.effectID == effectID && effect.actorID == actorID && now.Before(effect.expires) {
			return true
		}
	}
	return false
}

func (m *WorldMode) addWorldEffectAt(ctx Context, effectID int, actorID uint32, starts time.Time) bool {
	return m.addWorldEffectBetweenAt(ctx, effectID, actorID, 0, starts)
}

func (m *WorldMode) addWorldEffectBetweenAt(ctx Context, effectID int, actorID, targetID uint32, starts time.Time) bool {
	return m.addWorldEffectBetweenAtDuration(ctx, effectID, actorID, targetID, starts, 0)
}

func (m *WorldMode) addWorldEffectBetweenAtDuration(ctx Context, effectID int, actorID, targetID uint32, starts time.Time, durationOverride time.Duration) bool {
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
	duration := spec.duration
	for _, component := range spec.components {
		componentDuration := m.worldEffectResolvedComponentDuration(ctx.Resources, spec, component)
		if componentDuration > duration {
			duration = componentDuration
		}
	}
	if duration <= 0 {
		duration = 500 * time.Millisecond
	}
	if durationOverride > duration {
		duration = durationOverride
	}
	m.worldEffects = append(m.worldEffects, worldEffect{
		effectID: effectID,
		actorID:  actorID,
		targetID: targetID,
		x:        x,
		y:        y,
		starts:   starts,
		expires:  starts.Add(duration),
		duration: durationOverride,
	})
	if len(spec.sfx) > 0 {
		m.scheduleSound(starts, spec.sfx...)
	}
	if effectID == effectMagnumBreak {
		m.startCameraShake(starts, 50*time.Millisecond)
	}
	return true
}

func (m *WorldMode) addWorldEffectAtCell(ctx Context, effectID int, x, y int, starts time.Time) bool {
	return m.addWorldEffectAtCellWithActor(ctx, effectID, 0, x, y, starts)
}

func (m *WorldMode) addWorldEffectAtCellWithActor(ctx Context, effectID int, actorID uint32, x, y int, starts time.Time) bool {
	if ctx.World == nil {
		return false
	}
	spec, ok := worldEffectSpecForID(effectID)
	if !ok {
		return false
	}
	duration := spec.duration
	for _, component := range spec.components {
		componentDuration := m.worldEffectResolvedComponentDuration(ctx.Resources, spec, component)
		if componentDuration > duration {
			duration = componentDuration
		}
	}
	if duration <= 0 {
		duration = 500 * time.Millisecond
	}
	m.worldEffects = append(m.worldEffects, worldEffect{
		effectID: effectID,
		actorID:  actorID,
		x:        x,
		y:        y,
		starts:   starts,
		expires:  starts.Add(duration),
	})
	if len(spec.sfx) > 0 {
		m.scheduleSound(starts, spec.sfx...)
	}
	return true
}

func (m *WorldMode) addWorldEffectAtCellIfMissing(ctx Context, effectID int, x, y int, starts time.Time) bool {
	now := time.Now()
	for _, effect := range m.worldEffects {
		if effect.effectID == effectID && effect.actorID == 0 && effect.x == x && effect.y == y && now.Before(effect.expires) {
			return false
		}
	}
	return m.addWorldEffectAtCell(ctx, effectID, x, y, starts)
}

func (m *WorldMode) addWorldEffectBetweenAtDurationIfMissing(ctx Context, effectID int, actorID, targetID uint32, starts time.Time, durationOverride time.Duration) bool {
	now := time.Now()
	for _, effect := range m.worldEffects {
		if effect.effectID == effectID && effect.actorID == actorID && effect.targetID == targetID && now.Before(effect.expires) {
			return false
		}
	}
	return m.addWorldEffectBetweenAtDuration(ctx, effectID, actorID, targetID, starts, durationOverride)
}

func (m *WorldMode) addSkillCastEffects(ctx Context, skillID uint16, property uint32, sourceID, targetID uint32, duration time.Duration, starts time.Time, source string) {
	if duration <= 0 || sourceID == 0 {
		return
	}
	if m.addWorldEffectBetweenAtDurationIfMissing(ctx, effectCastRing, sourceID, 0, starts, duration) {
		log.Printf("skill cast circle source=%s skill=%d src=%d target=%d delay_ms=%d", source, skillID, sourceID, targetID, duration.Milliseconds())
	}
	effectID := skillCastAuraEffectID(property)
	if effectID <= 0 {
		return
	}
	if m.addWorldEffectBetweenAtDurationIfMissing(ctx, effectID, sourceID, targetID, starts, duration) {
		log.Printf("skill cast aura source=%s skill=%d src=%d target=%d property=%d effect=%d delay_ms=%d", source, skillID, sourceID, targetID, property, effectID, duration.Milliseconds())
	}
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
	case 602:
		return effectTeleportation
	default:
		return 0
	}
}

func skillSuccessEffectID(skillID uint16) int {
	switch skillID {
	case 6:
		return effectProvoke
	case 8:
		return effectEndure
	case 10:
		return effectSight
	case 28:
		return effectHeal
	case 29:
		return effectIncAgility
	case 30:
		return effectDecAgility
	case 31:
		return effectAqua
	case 32:
		return effectSignum
	case 33:
		return effectAngelus
	case 34:
		return effectBlessing
	case 35:
		return effectCure
	case 45:
		return effectConcentration
	case 50:
		return effectSteal
	case 53:
		return effectDetoxication
	default:
		return 0
	}
}

func skillEffectID(skillID uint16) int {
	switch skillID {
	case 15:
		return effectFrostDiver
	case 16:
		return effectStoneCurse
	case 20:
		return effectLightningBolt
	case 21:
		return effectThunderStorm
	case 157:
		return effectEnergyCoat
	default:
		return 0
	}
}

func skillBeginEffectID(skillID uint16) int {
	switch skillID {
	case 5:
		return effectBashBegin
	case 7:
		return effectMagnumBreak
	case 26:
		return effectTeleportation
	case 42:
		return effectMammonite
	case 46:
		return effectBashBegin
	default:
		return 0
	}
}

func skillBeforeHitEffectID(skillID uint16) int {
	switch skillID {
	case 13:
		return effectSoulStrike
	case 14:
		return effectColdBolt
	case 19:
		return effectFireBolt
	case 17:
		return effectFireBall
	default:
		return 0
	}
}

func skillGroundEffectID(skillID uint16) int {
	switch skillID {
	case 12:
		return effectSafetyWall
	case 18:
		return effectFireWall
	case 21:
		return effectThunderStorm
	case 25:
		return effectPneuma
	default:
		return 0
	}
}

func skillUnitEffectID(unitID uint16) int {
	switch unitID {
	case 126:
		return effectSafetyWall
	case 127:
		return effectFireWall
	case 133:
		return effectPneuma
	default:
		return 0
	}
}

func skillCastAuraEffectID(property uint32) int {
	switch property {
	case 1:
		return effectBeginSpell2
	case 2:
		return effectBeginSpell5
	case 3:
		return effectBeginSpell3
	case 4:
		return effectBeginSpell4
	case 5:
		return effectBeginSpell7
	case 6, 8:
		return effectBeginSpell6
	default:
		return effectBeginSpell
	}
}

func skillCastFallback(skillID uint16, level uint16) (uint32, time.Duration) {
	lv := maxInt(1, int(level))
	switch skillID {
	case 13:
		return 8, time.Duration(500*lv) * time.Millisecond
	case 14:
		return 1, time.Duration(700*lv) * time.Millisecond
	case 17:
		return 3, 1000 * time.Millisecond
	case 19:
		return 3, time.Duration(700*lv) * time.Millisecond
	case 20:
		return 4, time.Duration(700*lv) * time.Millisecond
	case 21:
		return 4, time.Duration(1000+200*lv) * time.Millisecond
	default:
		return 0, 0
	}
}

func skillHitEffectID(skillID uint16) int {
	switch skillID {
	case 5:
		return effectBashHit
	case 11:
		return effectBashHit
	case 13:
		return effectBashHit
	case 14:
		return effectColdHit
	case 15:
		return effectFrostDiverHit
	case 17, 18, 19:
		return effectFireHit
	case 20, 21:
		return effectWindHit
	case 24:
		return effectBashHit
	case 28:
		return effectHealOffensive
	case 46, 47:
		return effectBashHit
	case 52:
		return effectPoisonAttack
	default:
		return 0
	}
}

func worldEffectSpecForID(effectID int) (worldEffectSpec, bool) {
	spec, ok := worldEffectSpecs[effectID]
	if !ok {
		return worldEffectSpec{}, false
	}
	return cloneWorldEffectSpec(spec), true
}

func cloneWorldEffectSpec(spec worldEffectSpec) worldEffectSpec {
	if len(spec.sfx) > 0 {
		spec.sfx = append([]string(nil), spec.sfx...)
	}
	if len(spec.components) > 0 {
		spec.components = append([]worldEffectComponent(nil), spec.components...)
	}
	return spec
}

func teleportCylinderComponent(bottomSize, topSize, height float64) worldEffectComponent {
	return worldEffectComponent{
		kind:             effectPrimitiveCylinder,
		textureName:      "ring_blue",
		duration:         1500 * time.Millisecond,
		alphaMax:         0.5,
		fade:             true,
		rotate:           true,
		animation:        5,
		bottomSize:       bottomSize,
		topSize:          topSize,
		height:           height,
		totalCircleSides: 32,
		circleSides:      32,
	}
}

func portalCylinderComponent(bottomSize, topSize, height, posZ float64, textureName string, alphaMax float64) worldEffectComponent {
	return worldEffectComponent{
		kind:             effectPrimitiveCylinder,
		textureName:      textureName,
		duration:         25000 * time.Millisecond,
		alphaMax:         alphaMax,
		fade:             true,
		rotate:           true,
		animation:        0,
		bottomSize:       bottomSize,
		topSize:          topSize,
		height:           height,
		posZ:             posZ,
		totalCircleSides: 32,
		circleSides:      32,
	}
}

func healCylinderComponent(bottomSize, topSize, height float64) worldEffectComponent {
	return worldEffectComponent{
		kind:             effectPrimitiveCylinder,
		textureName:      "ring_white",
		duration:         1500 * time.Millisecond,
		alphaMax:         0.2,
		fade:             true,
		rotate:           true,
		animation:        1,
		bottomSize:       bottomSize,
		topSize:          topSize,
		height:           height,
		totalCircleSides: 32,
		circleSides:      32,
	}
}

func healOffensiveCylinderComponent(bottomSize, topSize, height float64) worldEffectComponent {
	component := healCylinderComponent(bottomSize, topSize, height)
	component.duration = time.Second
	component.color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	component.blendAdditive = true
	return component
}

func strEffectSpec(file, wav string) worldEffectSpec {
	return strEffectSpecRandom(file, wav, 0, 0)
}

func strEffectSpecRandom(file, wav string, randMin, randMax int) worldEffectSpec {
	return strEffectSpecRandomAttached(file, wav, randMin, randMax, false, false)
}

func strEffectSpecAttached(file, wav string, head bool) worldEffectSpec {
	return strEffectSpecRandomAttached(file, wav, 0, 0, true, head)
}

func strEffectSpecRandomAttached(file, wav string, randMin, randMax int, attached, head bool) worldEffectSpec {
	spec := worldEffectSpec{
		components: []worldEffectComponent{{
			kind:           effectPrimitiveSTR,
			strFile:        file,
			strRandMin:     randMin,
			strRandMax:     randMax,
			attachedEntity: attached,
			spriteHead:     head,
		}},
	}
	if wav != "" {
		spec.sfx = []string{wav}
	}
	return spec
}

func soundOnlyEffectSpec(paths ...string) worldEffectSpec {
	return worldEffectSpec{
		duration: 500 * time.Millisecond,
		sfx:      paths,
	}
}

func potionEffectSpec(file string, c color.RGBA) worldEffectSpec {
	return worldEffectSpec{
		duration: 850 * time.Millisecond,
		components: []worldEffectComponent{{
			kind:    effectPrimitiveSTR,
			color:   c,
			strFile: file,
		}},
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
		if now.Before(effect.starts) {
			continue
		}
		x, y := float64(effect.x), float64(effect.y)
		if actor, ok := ctx.World.Actors[effect.actorID]; ok {
			x, y = actor.RenderPosition(now)
		} else if isLocalActor(ctx, effect.actorID) {
			x, y = ctx.World.Player.RenderPosition(now)
		}
		worldX := cellCenter(x)
		worldY := cellCenter(y)
		worldZ := terrainHeightAt(ctx.World, x, y) + 0.07
		for index, component := range spec.components {
			componentDuration := m.worldEffectResolvedComponentDuration(ctx.Resources, spec, component)
			if effect.duration > componentDuration {
				componentDuration = effect.duration
			}
			progress := worldEffectComponentProgress(effect.starts, componentDuration, now)
			if progress >= 1 {
				continue
			}
			m.drawWorldEffectComponent(screen, ctx, projection, effect, component, index, worldX, worldY, worldZ, progress, now)
		}
	}
	m.worldEffects = active
}

func (m *WorldMode) worldEffectResolvedComponentDuration(manager *res.Manager, spec worldEffectSpec, component worldEffectComponent) time.Duration {
	duration := worldEffectComponentDuration(spec, component)
	if component.kind == effectPrimitiveSTR {
		if str := m.loadWorldEffectSTR(manager, resolveEffectSTRFile(component, worldEffect{}), component.texturePath); str != nil {
			duration = strEffectDuration(str, duration)
		}
	}
	if component.kind == effectPrimitiveSPR && component.duration <= 0 && !component.spriteRepeat {
		if view := m.effectSpriteView(manager, component.spriteFile); view != nil && len(view.act.Actions) > 0 {
			actionIndex := component.spriteFrame
			if actionIndex < 0 || actionIndex >= len(view.act.Actions) {
				actionIndex = 0
			}
			duration = actionAnimationDuration(view.act.Actions[actionIndex], duration)
		}
	}
	return duration
}

func worldEffectComponentDuration(spec worldEffectSpec, component worldEffectComponent) time.Duration {
	duration := spec.duration
	if component.duration > 0 {
		duration = component.duration
	}
	if component.duplicate > 1 && component.duplicateDelay > 0 {
		duration += time.Duration(component.duplicate-1) * component.duplicateDelay
	}
	duration += component.delay
	return duration
}

func worldEffectComponentProgress(starts time.Time, duration time.Duration, now time.Time) float64 {
	if duration <= 0 {
		return 1
	}
	return clampFloat(float64(now.Sub(starts))/float64(duration), 0, 1)
}

func (m *WorldMode) drawWorldEffectComponent(screen *render.Image, ctx Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, componentIndex int, worldX, worldY, worldZ, progress float64, now time.Time) {
	switch component.kind {
	case effectPrimitiveSTR:
		m.drawSTREffect(screen, ctx, projection, component, effect, worldX, worldY, worldZ, now)
	case effectPrimitiveCylinder:
		m.drawCylinderEffect(screen, ctx, projection, effect, component, componentIndex, worldX, worldY, worldZ, progress)
	case effectPrimitiveBillboard:
		m.drawBillboardEffect(screen, ctx, projection, component, worldX, worldY, worldZ, progress)
	case effectPrimitiveBashHit:
		drawBashHitEffect(screen, m.whitePixel, worldX, worldY, worldZ, progress, component.color)
	case effectPrimitive2D:
		m.draw2DEffect(screen, ctx, projection, component, worldX, worldY, worldZ, progress)
	case effectPrimitive3D:
		m.draw3DEffect(screen, ctx, projection, effect, component, componentIndex, worldX, worldY, worldZ, now)
	case effectPrimitiveSPR:
		m.drawSPREffect(screen, ctx, projection, effect, component, worldX, worldY, worldZ, now)
	case effectPrimitiveGroundPlane:
		m.drawGroundPlaneEffect(screen, ctx, component, effect, worldX, worldY, progress, now)
	case effectPrimitiveCastRing:
		m.drawCastRingEffect(screen, ctx, component, effect, componentIndex, worldX, worldY, worldZ, progress)
	default:
		drawPotionEffect(screen, m.whitePixel, worldX, worldY, worldZ, progress, component.color)
	}
}

func drawPotionEffect(screen, white *render.Image, x, y, z, progress float64, c color.RGBA) {
	alpha := 1 - progress
	drawWorldRadialGradient(screen, white, x, y, z, 0.02, 0.34+progress*0.18, withAlpha(c, alpha*0.30), 48)
	drawWorldSoftRing(screen, white, x, y, z+0.01, 0.24+progress*0.52, 0.22, withAlpha(c, alpha*0.75), 48)
	drawWorldCylinderBand(screen, white, nil, x, y, z+0.05+progress*0.55, 0.18+progress*0.08, 0.06, 0.22, withAlpha(c, alpha*0.40), 32)
}

func (m *WorldMode) drawCastRingEffect(screen *render.Image, ctx Context, component worldEffectComponent, effect worldEffect, componentIndex int, x, y, z, progress float64) {
	alpha := effectComponentAlpha(progress, component)
	if alpha <= 0 {
		return
	}
	texture := m.effectTexture(ctx.Resources, component.textureName)
	tint := effectComponentTint(component, alpha)
	z += component.posZ
	angleOffset := 0.0
	if component.rotate {
		angleOffset = progress * 2 * math.Pi
		angleOffset += deterministicAngle(effect, componentIndex+17) * 0.08
	}
	drawWorldSoftRing(screen, m.whitePixel, x, y, z+0.015, component.bottomSize*1.15, 0.16, withAlpha(component.color, alpha*0.75), maxInt(component.circleSides, 20))
	if texture != nil {
		drawWorldCylinderBandRotated(screen, m.whitePixel, texture, x, y, z+0.035, component.bottomSize, component.topSize, component.height, tint, maxInt(component.circleSides, component.totalCircleSides), angleOffset)
		return
	}
	drawWorldCylinderBandRotated(screen, m.whitePixel, nil, x, y, z+0.035, component.bottomSize, component.topSize, component.height, tint, maxInt(component.circleSides, component.totalCircleSides), angleOffset)
}

func drawWorldCylinderBandRotated(screen, white, texture *render.Image, x, y, z, bottomRadius, topRadius, height float64, c color.RGBA, segments int, angleOffset float64) {
	if segments < 3 || bottomRadius <= 0.01 || topRadius <= 0.01 || height <= 0.01 || c.A == 0 {
		return
	}
	vertices := make([]render.Vertex3D, 0, (segments+1)*2)
	indices := make([]uint16, 0, segments*6)
	tint := c
	srcW, srcH := float32(1), float32(1)
	source := white
	if texture != nil {
		source = texture
		bounds := texture.Bounds()
		srcW = float32(bounds.Dx())
		srcH = float32(bounds.Dy())
	}
	for i := 0; i <= segments; i++ {
		u := float32(i) / float32(segments)
		angle := angleOffset + float64(i)*2*math.Pi/float64(segments)
		cosine := math.Cos(angle)
		sine := math.Sin(angle)
		topAngle := angleOffset + (float64(i)+0.5)*2*math.Pi/float64(segments)
		topCosine := math.Cos(topAngle)
		topSine := math.Sin(topAngle)
		vertices = append(vertices,
			warpEffectTexturedVertex3D(x+cosine*bottomRadius, y+sine*bottomRadius, z, u*srcW, srcH, tint),
			warpEffectTexturedVertex3D(x+topCosine*topRadius, y+topSine*topRadius, z+height, u*srcW, 0, tint),
		)
		if i == segments {
			continue
		}
		base := uint16(i * 2)
		indices = append(indices, base, base+1, base+3, base, base+3, base+2)
	}
	options := triangleDrawOptions(render.FilterLinear, render.AddressRepeat)
	options.Blend = render.BlendLighter
	screen.DrawTriangles3D(vertices, indices, source, options)
}

func (m *WorldMode) drawCylinderEffect(screen *render.Image, ctx Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, componentIndex int, x, y, z, progress float64) {
	texture := m.effectTexture(ctx.Resources, component.textureName)
	if texture == nil {
		return
	}
	alpha := effectComponentAlpha(progress, component)
	if alpha <= 0 {
		return
	}
	topSize := component.topSize
	if component.animation == 2 {
		topSize *= progress
	}
	bottomSize := component.bottomSize
	height := component.height
	if component.animation == 4 {
		bottomSize *= progress
		topSize *= progress
	}
	if !component.fixedPerspective {
		drawWorldCylinderBand(screen, m.whitePixel, texture, x, y, z+component.posZ, bottomSize, topSize, height, effectComponentTint(component, alpha), maxInt(component.circleSides, component.totalCircleSides))
		return
	}
	duplicates := maxInt(component.duplicate, 1)
	for i := 0; i < duplicates; i++ {
		angle := 0.0
		if component.rotate {
			angle += progress * 2 * math.Pi
			angle += deterministicAngle(effect, componentIndex*101+i+31) * 0.08
		}
		if component.angleZRandom != 0 {
			angle += deterministicAngle(effect, componentIndex*101+i) * component.angleZRandom / 360
		}
		drawTexturedEffectCylinder(screen, projection, texture, x, y, z+component.posZ, effectCylinderDraw{
			bottomSize:       bottomSize,
			topSize:          topSize,
			totalCircleSides: component.totalCircleSides,
			circleSides:      component.circleSides,
			alpha:            alpha,
			angle:            angle,
		})
	}
}

func (m *WorldMode) drawBillboardEffect(screen *render.Image, ctx Context, projection sceneProjection, component worldEffectComponent, worldX, worldY, worldZ, progress float64) {
	texture := m.effectFileTexture(ctx.Resources, component.textureFile)
	if texture == nil {
		return
	}
	alpha := effectBillboardAlpha(progress, component)
	if alpha <= 0 {
		return
	}
	size := effectBillboardSize(progress, component)
	if size <= 0 {
		return
	}
	drawTexturedEffectBillboard(screen, projection, texture, worldX, worldY, worldZ+component.posZ, size, color.RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: uint8(clampFloat(alpha, 0, 1) * 255),
	})
}

func (m *WorldMode) draw2DEffect(screen *render.Image, ctx Context, projection sceneProjection, component worldEffectComponent, worldX, worldY, worldZ, progress float64) {
	texture := m.effectFileTexture(ctx.Resources, component.textureFile)
	if texture == nil {
		return
	}
	alpha := effectBillboardAlpha(progress, component)
	if alpha <= 0 {
		return
	}
	size := effectBillboardSize(progress, component)
	if size <= 0 {
		return
	}
	angle := worldEffectBillboardAngle(component, projection, progress)
	drawTexturedEffectBillboardRotated(screen, projection, texture, worldX, worldY, worldZ+component.posZ, size, angle, color.RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: uint8(clampFloat(alpha, 0, 1) * 255),
	})
}

func (m *WorldMode) draw3DEffect(screen *render.Image, ctx Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, componentIndex int, worldX, worldY, worldZ float64, now time.Time) {
	if component.textureFile == "" && len(component.textureFiles) == 0 && component.spriteFile == "" {
		return
	}
	duplicates := maxInt(component.duplicate, 1)
	componentDuration := component.duration
	if componentDuration <= 0 {
		componentDuration = 500 * time.Millisecond
	}
	for i := 0; i < duplicates; i++ {
		starts := effect.starts.Add(component.delay + time.Duration(i)*component.duplicateDelay)
		progress := worldEffectComponentProgress(starts, componentDuration, now)
		if now.Before(starts) || progress >= 1 {
			continue
		}
		alpha := effectBillboardAlphaForDuplicate(progress, component, i)
		if alpha <= 0 {
			continue
		}
		salt := componentIndex*1009 + i*37
		offsetX, offsetY, offsetZ := m.effect3DOffset(ctx, component, effect, salt, progress, worldX, worldY, worldZ)
		sizeX, sizeY := effect3DSize(component, effect, salt, progress, i)
		if sizeX <= 0 || sizeY <= 0 {
			continue
		}
		if component.textureFile != "" || len(component.textureFiles) > 0 {
			texture := m.effectTextureFrame(ctx.Resources, component, progress)
			if texture == nil {
				continue
			}
			angle := worldEffectBillboardAngle(component, projection, progress)
			drawTexturedEffectBillboardRotatedXY(screen, projection, texture, worldX+offsetX, worldY+offsetY, worldZ+offsetZ, sizeX, sizeY, angle, effectComponentTint(component, alpha), component.blendAdditive)
			continue
		}
		size := (sizeX + sizeY) * 0.5
		m.draw3DSpriteEffect(screen, ctx, projection, component, worldX+offsetX, worldY+offsetY, worldZ+offsetZ, size, alpha, starts, now)
	}
}

func (m *WorldMode) effectTextureFrame(manager *res.Manager, component worldEffectComponent, progress float64) *render.Image {
	if len(component.textureFiles) == 0 {
		return m.effectFileTexture(manager, component.textureFile)
	}
	index := int(clampFloat(progress, 0, 0.999999) * float64(len(component.textureFiles)))
	return m.effectFileTexture(manager, component.textureFiles[index])
}

func (m *WorldMode) drawGroundPlaneEffect(screen *render.Image, ctx Context, component worldEffectComponent, effect worldEffect, worldX, worldY, progress float64, now time.Time) {
	texture := m.effectFileTexture(ctx.Resources, component.textureFile)
	if texture == nil || ctx.World == nil {
		return
	}
	alpha := effectBillboardAlpha(progress, component)
	if alpha <= 0 {
		return
	}
	size := component.sizeStart
	if size <= 0 {
		size = 1
	}
	half := size * 0.5
	angle := now.Sub(effect.starts).Seconds() * 40 * math.Pi / 180
	uv := func(u, v float64) texturePoint {
		sinA, cosA := math.Sin(angle), math.Cos(angle)
		x, y := u-0.5, v-0.5
		return texturePoint{
			u: float32(x*cosA - y*sinA + 0.5),
			v: float32(x*sinA + y*cosA + 0.5),
		}
	}
	point := func(dx, dy float64) modelPoint3 {
		x := worldX + dx
		y := worldY + dy
		return modelPoint3{x: x, y: terrainHeightAt(ctx.World, x-0.5, y-0.5) + component.posZ, z: y}
	}
	verts := [4]modelPoint3{
		point(-half, -half),
		point(half, -half),
		point(half, half),
		point(-half, half),
	}
	uvs := [4]texturePoint{
		uv(0, 0),
		uv(1, 0),
		uv(1, 1),
		uv(0, 1),
	}
	tint := effectComponentTint(component, alpha)
	drawTexturedSurface3DAlpha(screen, texture, verts, uvs, quadIndices012023, [4]color.RGBA{tint, tint, tint, tint})
}

func (m *WorldMode) draw3DSpriteEffect(screen *render.Image, ctx Context, projection sceneProjection, component worldEffectComponent, worldX, worldY, worldZ float64, size float64, alpha float64, starts time.Time, now time.Time) {
	view := m.effectSpriteView(ctx.Resources, component.spriteFile)
	if view == nil || len(view.act.Actions) == 0 {
		return
	}
	actionIndex := 0
	action := view.act.Actions[actionIndex]
	if len(action.Animations) == 0 {
		return
	}
	delayMS := float64(action.DelayMS)
	if component.spriteDelay > 0 {
		delayMS = float64(component.spriteDelay / time.Millisecond)
	}
	motion := 0
	if component.spriteRepeat {
		motion = spriteMotionIndexWithDelay(action, starts, now, true, delayMS)
	} else {
		motion = spriteMotionIndexWithDelay(action, starts, now, false, delayMS)
	}
	if motion < 0 || motion >= len(action.Animations) {
		return
	}
	key := singleSpriteBillboardKey{actionIndex: actionIndex, motion: motion}
	billboard, ok := view.billboards[key]
	if !ok {
		var baseOK bool
		billboard, baseOK = composeSingleSpriteBillboard(view, action.Animations[motion])
		if !baseOK {
			return
		}
		view.billboards[key] = billboard
	}
	tint := effectComponentTint(component, 1)
	if component.worldSizedSprite {
		scale := size / 100
		angle := -worldEffectSpriteAngle(component) * math.Pi / 180
		drawSpriteBillboardTintAlphaWorld3D(screen, projection, billboard, worldX, worldY, worldZ, scale, angle, alpha, 1, tint)
		return
	}
	scale := size / (100 * roBrowserEffectPixelRatio)
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	drawSpriteBillboardTintAlpha3D(screen, projection, billboard, worldX, worldY, worldZ, scale, alpha, 1, tint)
}

func worldEffectSpriteAngle(component worldEffectComponent) float64 {
	angle := component.angleStart
	if !component.rotateToTarget {
		return angle
	}
	startX, startY := component.posX, component.posY
	endX, endY := component.posXEnd, component.posYEnd
	if component.posXEnd == 0 && component.posXEndRand == 0 {
		endX = startX
	}
	if component.posYEnd == 0 && component.posYEndRand == 0 {
		endY = startY
	}
	return angle + 90 - math.Atan2(endY-startY, endX-startX)*180/math.Pi
}

func (m *WorldMode) drawSPREffect(screen *render.Image, ctx Context, projection sceneProjection, effect worldEffect, component worldEffectComponent, worldX, worldY, worldZ float64, now time.Time) {
	view := m.effectSpriteView(ctx.Resources, component.spriteFile)
	if view == nil || len(view.act.Actions) == 0 {
		return
	}
	actionIndex := component.spriteFrame
	if component.spriteDirection {
		if actor, ok := ctx.World.Actors[effect.actorID]; ok {
			actionIndex = actor.RenderDirection(now) % len(view.act.Actions)
		} else if isLocalActor(ctx, effect.actorID) {
			actionIndex = ctx.World.Player.RenderDirection(now) % len(view.act.Actions)
		}
	}
	if actionIndex < 0 || actionIndex >= len(view.act.Actions) {
		actionIndex = 0
	}
	action := view.act.Actions[actionIndex]
	if len(action.Animations) == 0 {
		return
	}
	delayMS := float64(action.DelayMS)
	if component.spriteDelay > 0 {
		delayMS = float64(component.spriteDelay / time.Millisecond)
	}
	motion := 0
	if component.spriteRepeat {
		motion = spriteMotionIndexWithDelay(action, effect.starts, now, true, delayMS)
	} else {
		motion = spriteMotionIndexWithDelay(action, effect.starts, now, false, delayMS)
		if motion >= len(action.Animations)-1 && !component.spriteStopAtEnd && component.duration <= 0 {
			return
		}
	}
	if motion < 0 || motion >= len(action.Animations) {
		return
	}
	key := singleSpriteBillboardKey{
		actionIndex: actionIndex,
		motion:      motion,
		anchorX:     component.spriteXOffset,
		anchorY:     component.spriteYOffset,
	}
	billboard, ok := view.billboards[key]
	if !ok {
		base, baseOK := composeSingleSpriteBillboard(view, action.Animations[motion])
		if !baseOK {
			return
		}
		copy := *base
		copy.anchorX -= component.spriteXOffset
		copy.anchorY -= component.spriteYOffset
		billboard = &copy
		view.billboards[key] = billboard
	}
	z := worldZ + component.posZ
	if component.spriteHead {
		z += 2.0
	}
	drawSpriteBillboardTintAlpha3D(screen, projection, billboard, worldX, worldY, z, 1, 1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
}

func drawBashHitEffect(screen, white *render.Image, x, y, z, progress float64, c color.RGBA) {
	alpha := 1 - progress
	base := z + 0.65
	drawWorldCylinderBand(screen, white, nil, x, y, base-0.16, 0.08+progress*0.24, 0.56+progress*0.42, 0.08, withAlpha(c, alpha*0.36), 24)
	drawWorldCylinderBand(screen, white, nil, x, y, base, 0.48+progress*0.18, 0.02, 0.42, withAlpha(c, alpha*0.55), 4)
	drawWorldCylinderBand(screen, white, nil, x, y, base+0.12, 0.02, 0.58+progress*0.22, 0.28, withAlpha(c, alpha*0.45), 4)
	drawWorldSoftRing(screen, white, x, y, base, 0.20+progress*0.55, 0.10, withAlpha(c, alpha*0.48), 32)
}

type effectCylinderDraw struct {
	bottomSize       float64
	topSize          float64
	totalCircleSides int
	circleSides      int
	alpha            float64
	angle            float64
}

func drawTexturedEffectCylinder(screen *render.Image, projection sceneProjection, texture *render.Image, worldX, worldY, worldZ float64, draw effectCylinderDraw) {
	if screen == nil || texture == nil || draw.alpha <= 0 || draw.topSize <= 0 || draw.totalCircleSides <= 0 || draw.circleSides <= 0 {
		return
	}
	right, up, _, ok := projection.BillboardBasis(worldX, worldY, worldZ)
	if !ok {
		return
	}
	bounds := texture.Bounds()
	w, h := float32(bounds.Dx()), float32(bounds.Dy())
	center := modelPoint3{x: worldX, y: worldZ, z: worldY}
	tint := color.RGBA{R: 255, G: 255, B: 255, A: uint8(clampFloat(draw.alpha, 0, 1) * 255)}
	vertices := make([]render.Vertex3D, 0, (draw.circleSides+1)*2)
	indices := make([]uint16, 0, draw.circleSides*6)
	point := func(radius, angle float64) modelPoint3 {
		return add3(add3(center, mul3(right, math.Sin(angle)*radius)), mul3(up, math.Cos(angle)*radius))
	}
	for i := 0; i <= draw.circleSides; i++ {
		a := float64(i) / float64(draw.totalCircleSides)
		angle := draw.angle + a*2*math.Pi
		u := float32(a * float64(draw.totalCircleSides) / float64(draw.circleSides))
		vertices = append(vertices,
			texturedSurfaceVertex3D(point(draw.bottomSize, angle), texturePoint{u: u, v: 1}, tint, w, h),
			texturedSurfaceVertex3D(point(draw.topSize, angle), texturePoint{u: u, v: 0}, tint, w, h),
		)
		if i < draw.circleSides {
			base := uint16(i * 2)
			indices = append(indices, base, base+1, base+2, base+1, base+3, base+2)
		}
	}
	screen.DrawTriangles3DOwned(vertices, indices, texture, triangleDrawOptions(render.FilterLinear, render.AddressRepeat))
}

func effectComponentAlpha(progress float64, component worldEffectComponent) float64 {
	alphaMax := component.alphaMax
	if alphaMax <= 0 {
		alphaMax = 1
	}
	if component.fade {
		switch {
		case progress < 0.25:
			return progress / 0.25 * alphaMax
		case progress > 0.75:
			return (1 - progress) / 0.25 * alphaMax
		default:
			return alphaMax
		}
	}
	switch {
	case component.fadeIn && progress < 0.25:
		return progress / 0.25 * alphaMax
	case component.fadeOut && progress > 0.75:
		return (1 - progress) / 0.25 * alphaMax
	default:
		return alphaMax
	}
}

func effectBillboardAlpha(progress float64, component worldEffectComponent) float64 {
	return effectBillboardAlphaForDuplicate(progress, component, 0)
}

func effectBillboardAlphaForDuplicate(progress float64, component worldEffectComponent, duplicateIndex int) float64 {
	alphaMax := component.alphaMax
	if alphaMax <= 0 {
		alphaMax = 1
	}
	alphaMax += component.alphaMaxDelta * float64(duplicateIndex)
	alphaMax = clampFloat(alphaMax, 0, 1)
	switch {
	case component.fadeIn && progress < 0.25:
		return progress / 0.25 * alphaMax
	case component.fadeOut && progress > 0.75:
		return (1 - progress) / 0.25 * alphaMax
	default:
		return alphaMax
	}
}

func effectBillboardSize(progress float64, component worldEffectComponent) float64 {
	start := component.sizeStart
	end := component.sizeEnd
	if start <= 0 {
		start = 1
	}
	if end <= 0 {
		end = start
	}
	if !component.sizeSmooth {
		return start + (end-start)*progress
	}
	return start + (end-start)*math.Log10(progress*9+1)
}

func (m *WorldMode) effect3DOffset(ctx Context, component worldEffectComponent, effect worldEffect, salt int, progress float64, worldX, worldY, worldZ float64) (float64, float64, float64) {
	staticX := deterministicSigned(effect, salt+1) * component.posXRand
	staticY := deterministicSigned(effect, salt+2) * component.posYRand
	startX := component.posX + staticX + component.posXStartMiddle + deterministicSigned(effect, salt+11)*component.posXStartRand
	startY := component.posY + staticY + component.posYStartMiddle + deterministicSigned(effect, salt+12)*component.posYStartRand
	startZ := component.posZ + component.posZStartMiddle + deterministicSigned(effect, salt+3)*component.posZStartRand
	endX := component.posXEnd + staticX + component.posXEndMiddle + deterministicSigned(effect, salt+4)*component.posXEndRand
	endY := component.posYEnd + staticY + component.posYEndMiddle + deterministicSigned(effect, salt+5)*component.posYEndRand
	endZ := component.posZEnd + component.posZEndMiddle + deterministicSigned(effect, salt+6)*component.posZEndRand
	if component.posXEnd == 0 && component.posXEndRand == 0 && component.posXEndMiddle == 0 && component.posXStartRand == 0 && component.posXStartMiddle == 0 {
		endX = startX
	}
	if component.posYEnd == 0 && component.posYEndRand == 0 && component.posYEndMiddle == 0 && component.posYStartRand == 0 && component.posYStartMiddle == 0 {
		endY = startY
	}
	if component.posZEnd == 0 && component.posZEndRand == 0 && component.posZEndMiddle == 0 {
		endZ = startZ
	}
	if component.fromSrc || component.toSrc {
		otherX, otherY, otherZ, ok := effectOtherEndpoint(ctx, effect, worldX, worldY, worldZ)
		if ok {
			dx := otherX - worldX
			dy := otherY - worldY
			dz := otherZ - worldZ
			if component.fromSrc {
				endX += dx
				endY += dy
				endZ += dz
			}
			if component.toSrc {
				startX += dx
				startY += dy
				startZ += dz
			}
		}
	}
	x := effectPositionAxis(progress, startX, endX, component.posXSmooth)
	y := effectPositionAxis(progress, startY, endY, component.posYSmooth)
	z := effectPositionAxis(progress, startZ, endZ, component.posZSmooth)
	if component.retreat != 0 {
		dx := endX - startX
		dy := endY - startY
		dist := math.Hypot(dx, dy)
		if dist > 0.001 {
			factor := math.Sin(progress*math.Pi) * component.retreat
			x -= dx / dist * factor
			y -= dy / dist * factor
		}
	}
	if component.arc != 0 {
		z += math.Sin(progress*math.Pi) * component.arc
	}
	return x, y, z
}

func effectOtherEndpoint(ctx Context, effect worldEffect, fallbackX, fallbackY, fallbackZ float64) (float64, float64, float64, bool) {
	if effect.targetID == 0 || ctx.World == nil {
		return fallbackX, fallbackY, fallbackZ, false
	}
	if actor, ok := ctx.World.Actors[effect.targetID]; ok {
		x, y := actor.RenderPosition(time.Now())
		return cellCenter(x), cellCenter(y), terrainHeightAt(ctx.World, x, y) + 0.07, true
	}
	if isLocalActor(ctx, effect.targetID) {
		x, y := ctx.World.Player.RenderPosition(time.Now())
		return cellCenter(x), cellCenter(y), terrainHeightAt(ctx.World, x, y) + 0.07, true
	}
	return fallbackX, fallbackY, fallbackZ, false
}

func effectPositionAxis(progress, start, end float64, smooth bool) float64 {
	if smooth {
		return start + (end-start)*math.Log10(progress*9+1)
	}
	return start + (end-start)*progress
}

func worldEffectBillboardAngle(component worldEffectComponent, projection sceneProjection, progress float64) float64 {
	angle := (component.angleStart + (component.angleEnd-component.angleStart)*progress) * math.Pi / 180
	if component.rotateWithCamera {
		angle += degreesToRadians(projection.cameraYaw)
	}
	return angle
}

func effect3DSize(component worldEffectComponent, effect worldEffect, salt int, progress float64, duplicateIndex int) (float64, float64) {
	size := effectBillboardSize(progress, component)
	sizeX := size
	sizeY := size
	if component.sizeStartX > 0 || component.sizeEndX > 0 {
		sizeX = effectAxisSize(progress, component.sizeStartX, component.sizeEndX, component.sizeSmooth)
	}
	if component.sizeStartY > 0 || component.sizeEndY > 0 {
		sizeY = effectAxisSize(progress, component.sizeStartY, component.sizeEndY, component.sizeSmooth)
	}
	if component.sizeDelta != 0 {
		delta := component.sizeDelta * float64(duplicateIndex) * roBrowserEffectPixelRatio
		sizeX += delta
		sizeY += delta
	}
	if component.sizeRand != 0 {
		sizeX += deterministicSigned(effect, salt+7) * component.sizeRand
		sizeY = sizeX
	}
	if component.sizeRandX != 0 {
		sizeX += deterministicSigned(effect, salt+8) * component.sizeRandX
	}
	if component.sizeRandY != 0 {
		sizeY += deterministicSigned(effect, salt+9) * component.sizeRandY
	}
	return sizeX, sizeY
}

func effectAxisSize(progress, start, end float64, smooth bool) float64 {
	if start <= 0 && end > 0 {
		start = end
	}
	if end <= 0 && start > 0 {
		end = start
	}
	if smooth {
		factor := math.Log10(progress*9 + 1)
		return start + (end-start)*factor
	}
	return start + (end-start)*progress
}

func effectComponentTint(component worldEffectComponent, alpha float64) color.RGBA {
	tint := component.color
	if tint.A == 0 {
		tint = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	tint.A = uint8(clampFloat(alpha, 0, 1) * 255)
	return tint
}

func deterministicAngle(effect worldEffect, salt int) float64 {
	return deterministicUnit(effect, salt) * 2 * math.Pi
}

func deterministicSigned(effect worldEffect, salt int) float64 {
	return deterministicUnit(effect, salt)*2 - 1
}

func deterministicUnit(effect worldEffect, salt int) float64 {
	value := uint32(effect.effectID*1103515245) ^ effect.actorID ^ uint32(effect.starts.UnixNano()) ^ uint32(salt*2654435761)
	value ^= value >> 16
	value *= 2246822519
	value ^= value >> 13
	return float64(value&0xFFFFFF) / float64(0xFFFFFF)
}

func drawTexturedEffectBillboard(screen *render.Image, projection sceneProjection, texture *render.Image, worldX, worldY, worldZ, size float64, tint color.RGBA) {
	drawTexturedEffectBillboardRotated(screen, projection, texture, worldX, worldY, worldZ, size, 0, tint)
}

func drawTexturedEffectBillboardRotated(screen *render.Image, projection sceneProjection, texture *render.Image, worldX, worldY, worldZ, size, angle float64, tint color.RGBA) {
	drawTexturedEffectBillboardRotatedXY(screen, projection, texture, worldX, worldY, worldZ, size, size, angle, tint, true)
}

func drawTexturedEffectBillboardRotatedXY(screen *render.Image, projection sceneProjection, texture *render.Image, worldX, worldY, worldZ, sizeX, sizeY, angle float64, tint color.RGBA, additive bool) {
	if screen == nil || texture == nil || tint.A == 0 {
		return
	}
	right, up, _, ok := projection.BillboardBasis(worldX, worldY, worldZ)
	if !ok {
		return
	}
	center := modelPoint3{x: worldX, y: worldZ, z: worldY}
	bounds := texture.Bounds()
	w, h := float32(bounds.Dx()), float32(bounds.Dy())
	axisScaleX := sizeX / float64(w)
	axisScaleY := sizeY / float64(h)
	rightAxis := mul3(right, axisScaleX)
	upAxis := mul3(up, -axisScaleY)
	if angle != 0 {
		sinA, cosA := math.Sin(angle), math.Cos(angle)
		rightAxis = add3(mul3(right, cosA*axisScaleX), mul3(up, sinA*axisScaleX))
		upAxis = add3(mul3(right, sinA*axisScaleY), mul3(up, -cosA*axisScaleY))
	}
	options := triangleDrawOptions(render.FilterLinear, render.AddressClampToZero)
	if additive {
		options.Blend = render.BlendLighter
	}
	screen.DrawWorldBillboard(render.WorldBillboardCommand{
		Texture:     texture,
		Options:     *options,
		Center:      [3]float32{float32(center.x), float32(center.y), float32(center.z)},
		RightAxis:   [3]float32{float32(rightAxis.x), float32(rightAxis.y), float32(rightAxis.z)},
		UpAxis:      [3]float32{float32(upAxis.x), float32(upAxis.y), float32(upAxis.z)},
		DepthUpAxis: [3]float32{float32(upAxis.x), float32(upAxis.y), float32(upAxis.z)},
		Width:       w,
		Height:      h,
		AnchorX:     w * 0.5,
		AnchorY:     h * 0.5,
		ColorR:      float32(tint.R) / 255,
		ColorG:      float32(tint.G) / 255,
		ColorB:      float32(tint.B) / 255,
		ColorA:      float32(tint.A) / 255,
	})
}

func (m *WorldMode) drawSTREffect(screen *render.Image, ctx Context, projection sceneProjection, component worldEffectComponent, effect worldEffect, worldX, worldY, worldZ float64, now time.Time) bool {
	str := m.loadWorldEffectSTR(ctx.Resources, resolveEffectSTRFile(component, effect), component.texturePath)
	if str == nil {
		return false
	}
	fps := str.FPS
	if fps <= 0 {
		fps = 60
	}
	keyIndex := float64(now.Sub(effect.starts)) / float64(time.Second) * float64(fps)
	drawn := false
	for _, layer := range str.Layers {
		anim, ok := calculateSTRAnimation(layer, keyIndex)
		if !ok {
			continue
		}
		if math.IsNaN(float64(anim.AniFrame)) || math.IsInf(float64(anim.AniFrame), 0) {
			continue
		}
		textureIndex := int(anim.AniFrame)
		if textureIndex < 0 || textureIndex >= len(layer.Textures) {
			continue
		}
		texture := m.strEffectTexture(ctx.Resources, layer.Textures[textureIndex])
		if texture == nil {
			continue
		}
		drawSTRAnimation(screen, projection, texture, worldX, worldY, worldZ, anim, component.attachedEntity)
		drawn = true
	}
	return drawn
}

func resolveEffectSTRFile(component worldEffectComponent, effect worldEffect) string {
	if component.strFile == "" || !strings.Contains(component.strFile, "%d") || component.strRandMax < component.strRandMin || component.strRandMin <= 0 {
		return component.strFile
	}
	span := component.strRandMax - component.strRandMin + 1
	index := component.strRandMin
	if span > 1 {
		value := uint32(effect.effectID*1103515245) ^ effect.actorID ^ uint32(effect.starts.UnixNano())
		value ^= value >> 16
		index += int(value % uint32(span))
	}
	return fmt.Sprintf(component.strFile, index)
}

func (m *WorldMode) loadWorldEffectSTR(manager *res.Manager, strFile, texturePath string) *res.STR {
	if manager == nil || strFile == "" {
		return nil
	}
	if m.strEffects == nil {
		m.strEffects = make(map[string]*res.STR)
	}
	if m.strEffectMiss == nil {
		m.strEffectMiss = make(map[string]struct{})
	}
	normalized := strings.ReplaceAll(strFile, "/", "\\")
	paths := []string{"data\\texture\\effect\\" + normalized + ".str"}
	if strings.ContainsAny(strFile, `/\`) {
		paths = append([]string{"data\\texture\\" + normalized + ".str"}, paths...)
	}
	attempted := false
	for _, path := range paths {
		key := "__str_" + path + "|" + texturePath
		if str, ok := m.strEffects[key]; ok {
			return str
		}
		if _, ok := m.strEffectMiss[key]; ok {
			continue
		}
		attempted = true
		data, err := manager.ReadFileExact(path)
		if err != nil {
			m.strEffectMiss[key] = struct{}{}
			continue
		}
		str, err := res.ParseSTR(data, texturePath)
		if err != nil {
			m.strEffectMiss[key] = struct{}{}
			log.Printf("str effect parse failed path=%s: %v", path, err)
			return nil
		}
		m.strEffects[key] = str
		return str
	}
	if attempted {
		log.Printf("str effect missing file=%s", strFile)
	}
	return nil
}

func (m *WorldMode) strEffectTexture(manager *res.Manager, path string) *render.Image {
	path = strings.TrimSpace(path)
	if manager == nil || path == "" {
		return nil
	}
	key := "__strtex_" + path
	if m.textures == nil {
		m.textures = make(map[string]*render.Image)
	}
	if m.textureMiss == nil {
		m.textureMiss = make(map[string]struct{})
	}
	if texture, ok := m.textures[key]; ok {
		return texture
	}
	if _, ok := m.textureMiss[key]; ok {
		return nil
	}
	candidates := []string{path, strings.ReplaceAll(path, "\\", "/")}
	img, _, err := res.LoadImageExact(manager, candidates)
	if err != nil {
		m.textureMiss[key] = struct{}{}
		log.Printf("str effect texture missing path=%s: %v", path, err)
		return nil
	}
	texture := render.NewImageFromImage(res.ApplyEffectTransparency(img))
	m.textures[key] = texture
	return texture
}

func (m *WorldMode) effectFileTexture(manager *res.Manager, path string) *render.Image {
	path = strings.TrimSpace(path)
	if manager == nil || path == "" {
		return nil
	}
	key := "__effectfile_" + path
	if m.textures == nil {
		m.textures = make(map[string]*render.Image)
	}
	if m.textureMiss == nil {
		m.textureMiss = make(map[string]struct{})
	}
	if texture, ok := m.textures[key]; ok {
		return texture
	}
	if _, ok := m.textureMiss[key]; ok {
		return nil
	}
	normalized := strings.ReplaceAll(path, "/", "\\")
	candidates := []string{
		"data\\texture\\" + normalized,
		strings.ReplaceAll("data\\texture\\"+normalized, "\\", "/"),
		normalized,
		strings.ReplaceAll(normalized, "\\", "/"),
	}
	img, _, err := res.LoadImageExact(manager, candidates)
	if err != nil {
		m.textureMiss[key] = struct{}{}
		log.Printf("effect texture missing path=%s: %v", path, err)
		return nil
	}
	texture := render.NewImageFromImage(res.ApplyEffectTransparency(img))
	m.textures[key] = texture
	return texture
}

const effectSpriteRoot = "data\\sprite\\\xC0\xCC\xC6\xD1\xC6\xAE\\"

func (m *WorldMode) effectSpriteView(manager *res.Manager, file string) *playerSpriteView {
	file = strings.TrimSpace(file)
	if manager == nil || file == "" {
		return nil
	}
	if m.effectViews == nil {
		m.effectViews = make(map[string]*playerSpriteView)
	}
	if m.effectViewMiss == nil {
		m.effectViewMiss = make(map[string]struct{})
	}
	key := strings.ReplaceAll(file, "/", "\\")
	if view, ok := m.effectViews[key]; ok {
		return view
	}
	if _, ok := m.effectViewMiss[key]; ok {
		return nil
	}
	actCandidates := effectSpriteResourceCandidates(file, "act")
	sprCandidates := effectSpriteResourceCandidates(file, "spr")
	view, status := loadSpriteView(manager, actCandidates, sprCandidates, nil, "effect sprite "+file)
	if view == nil {
		m.effectViewMiss[key] = struct{}{}
		log.Printf("effect sprite unavailable file=%q: %s", file, status)
		return nil
	}
	m.effectViews[key] = view
	log.Printf("effect sprite resources file=%q %s", file, status)
	return view
}

func effectSpriteResourceCandidates(file, ext string) []string {
	normalized := strings.TrimSpace(strings.ReplaceAll(file, "/", "\\"))
	normalized = strings.TrimSuffix(normalized, ".spr")
	normalized = strings.TrimSuffix(normalized, ".act")
	if normalized == "" {
		return nil
	}
	base := normalized
	if !strings.HasPrefix(strings.ToLower(base), "data\\sprite\\") {
		base = effectSpriteRoot + base
	}
	path := base + "." + ext
	return []string{path, strings.ReplaceAll(path, "\\", "/")}
}

func strEffectDuration(str *res.STR, fallback time.Duration) time.Duration {
	if str == nil || str.FPS <= 0 || str.MaxKey <= 0 {
		return fallback
	}
	duration := time.Duration(float64(str.MaxKey) / float64(str.FPS) * float64(time.Second))
	if duration <= 0 {
		return fallback
	}
	return duration + 100*time.Millisecond
}

func calculateSTRAnimation(layer res.STRLayer, keyIndex float64) (res.STRAnimation, bool) {
	animations := layer.Animations
	lastFrame := 0
	lastSource := 0
	fromID := -1
	toID := -1
	for i, anim := range animations {
		if float64(anim.Frame) <= keyIndex {
			if anim.Type == 0 {
				fromID = i
			}
			if anim.Type == 1 {
				toID = i
			}
		}
		if anim.Frame > lastFrame {
			lastFrame = anim.Frame
		}
		if anim.Type == 0 && anim.Frame > lastSource {
			lastSource = anim.Frame
		}
	}
	if fromID < 0 || (toID < 0 && float64(lastFrame) < keyIndex) {
		return res.STRAnimation{}, false
	}
	from := animations[fromID]
	var to res.STRAnimation
	hasTo := toID >= 0 && toID < len(animations)
	if hasTo {
		to = animations[toID]
	}
	delta := float32(keyIndex - float64(from.Frame))
	out := res.STRAnimation{
		SrcAlpha:  from.SrcAlpha,
		DestAlpha: from.DestAlpha,
	}
	if !hasTo || toID != fromID+1 || to.Frame != from.Frame {
		if hasTo && lastSource <= from.Frame {
			return res.STRAnimation{}, false
		}
		return from, true
	}
	out.Angle = from.Angle + to.Angle*delta
	out.AniFrame = strAnimFrame(from, to, delta, len(layer.Textures))
	for i := range out.Color {
		out.Color[i] = from.Color[i] + to.Color[i]*delta
	}
	for i := range out.Pos {
		out.Pos[i] = from.Pos[i] + to.Pos[i]*delta
	}
	for i := range out.UV {
		out.UV[i] = from.UV[i] + to.UV[i]*delta
	}
	for i := range out.XY {
		out.XY[i] = from.XY[i] + to.XY[i]*delta
	}
	return out, true
}

func strAnimFrame(from, to res.STRAnimation, delta float32, texCount int) float32 {
	switch to.AniType {
	case 1:
		return from.AniFrame + to.AniFrame*delta
	case 2:
		return minFloat32(from.AniFrame+to.Delay*delta, float32(texCount-1))
	case 3:
		count := float32(maxInt(texCount, 1))
		return float32(math.Mod(float64(from.AniFrame+to.Delay*delta), float64(count)))
	case 4:
		count := float32(maxInt(texCount, 1))
		value := float32(math.Mod(float64(from.AniFrame-to.Delay*delta), float64(count)))
		if value < 0 {
			value += count
		}
		return value
	default:
		return 0
	}
}

func drawSTRAnimation(screen *render.Image, projection sceneProjection, texture *render.Image, worldX, worldY, worldZ float64, anim res.STRAnimation, attached bool) {
	right, up, _, ok := projection.BillboardBasis(worldX, worldY, worldZ)
	if !ok {
		return
	}
	const pixelRatio = 1.0 / 35.0
	offsetX, offsetY := strAnimationOffset(anim, attached)
	center := modelPoint3{x: worldX, y: worldZ, z: worldY}
	angle := -float64(anim.Angle) * math.Pi / 180
	sinA, cosA := math.Sin(angle), math.Cos(angle)
	vertexPoint := func(ix, iy int) modelPoint3 {
		x := float64(anim.XY[ix])
		y := float64(anim.XY[iy])
		rotX := x*cosA - y*sinA
		rotY := x*sinA + y*cosA
		dx := rotX*pixelRatio + offsetX
		dy := -rotY*pixelRatio + offsetY
		return add3(add3(center, mul3(right, dx)), mul3(up, dy))
	}
	tint := strAnimationTint(anim)
	bounds := texture.Bounds()
	w, h := float32(bounds.Dx()), float32(bounds.Dy())
	vertices := []render.Vertex3D{
		texturedSurfaceVertex3D(vertexPoint(0, 4), texturePoint{u: 0, v: 0}, tint, w, h),
		texturedSurfaceVertex3D(vertexPoint(1, 5), texturePoint{u: 1, v: 0}, tint, w, h),
		texturedSurfaceVertex3D(vertexPoint(3, 7), texturePoint{u: 0, v: 1}, tint, w, h),
		texturedSurfaceVertex3D(vertexPoint(2, 6), texturePoint{u: 1, v: 1}, tint, w, h),
	}
	options := triangleDrawOptions(render.FilterLinear, render.AddressClampToZero)
	options.Blend = strAnimationBlend(anim)
	screen.DrawTriangles3DOwned(vertices, quadIndices012213, texture, options)
}

func strAnimationBlend(anim res.STRAnimation) render.Blend {
	// roBrowser applies the Direct3D blend factors stored in each STR layer.
	// D3DBLEND_SRCALPHA + D3DBLEND_DESTALPHA is used by bright fog-like
	// effects such as Pneuma; on our opaque world target it matches src-alpha
	// additive blending.
	if anim.SrcAlpha == 5 && (anim.DestAlpha == 2 || anim.DestAlpha == 7) {
		return render.BlendLighter
	}
	if anim.DestAlpha == 2 {
		return render.BlendLighter
	}
	return render.BlendSourceOver
}

func strAnimationOffset(anim res.STRAnimation, attached bool) (float64, float64) {
	const pixelRatio = 1.0 / 35.0
	verticalBase := -0.5
	if attached {
		verticalBase = 0
	}
	return float64(anim.Pos[0]-320) * pixelRatio, -float64(anim.Pos[1]-320)*pixelRatio + verticalBase
}

func strAnimationTint(anim res.STRAnimation) color.RGBA {
	return color.RGBA{
		R: uint8(clampFloat(float64(anim.Color[0]), 0, 1) * 255),
		G: uint8(clampFloat(float64(anim.Color[1]), 0, 1) * 255),
		B: uint8(clampFloat(float64(anim.Color[2]), 0, 1) * 255),
		A: uint8(clampFloat(float64(anim.Color[3]), 0, 1) * 255),
	}
}

func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
