package gamemode

import (
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
	effectProvoke      = 67
	effectBashBegin    = 16
	effectBashHit      = 1
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
	effectStyleBash    = 3
	effectStyleHit     = 4
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
	style       int
	color       color.RGBA
	duration    time.Duration
	sfx         []string
	strFile     string
	texturePath string
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

func (m *WorldMode) applySkillFailAck(ctx Context, ack network.SkillFailAck) {
	if ack.Result != 0 {
		return
	}
	message := skillFailMessage(ack)
	log.Printf("skill fail ack skill=%d num=%d item=%d result=%d cause=%d msg=%q", ack.SkillID, ack.Number, ack.ItemID, ack.Result, ack.Cause, message)
	m.console.addErrorMessage("%s", message)
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

func (m *WorldMode) addWorldEffectAt(ctx Context, effectID int, actorID uint32, starts time.Time) bool {
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
	if spec.strFile != "" {
		if str := m.loadWorldEffectSTR(ctx.Resources, spec); str != nil {
			duration = strEffectDuration(str, duration)
		}
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

func skillBeginEffectID(skillID uint16) int {
	switch skillID {
	case 5:
		return effectBashBegin
	default:
		return 0
	}
}

func skillHitEffectID(skillID uint16) int {
	switch skillID {
	case 5:
		return effectBashHit
	default:
		return 0
	}
}

func worldEffectSpecForID(effectID int) (worldEffectSpec, bool) {
	switch effectID {
	case effectBashHit:
		return worldEffectSpec{
			style:    effectStyleHit,
			color:    color.RGBA{R: 255, G: 248, B: 220, A: 255},
			duration: 280 * time.Millisecond,
			sfx:      []string{"effect\\ef_hit2.wav"},
		}, true
	case effectBashBegin:
		return worldEffectSpec{
			style:    effectStyleBash,
			color:    color.RGBA{R: 245, G: 250, B: 255, A: 255},
			duration: 650 * time.Millisecond,
			sfx:      []string{"effect\\ef_bash.wav"},
		}, true
	case effectProvoke:
		return worldEffectSpec{
			style:    effectStyleProvoke,
			color:    color.RGBA{R: 255, G: 70, B: 42, A: 255},
			duration: 900 * time.Millisecond,
			sfx:      []string{"effect\\swordman_provoke.wav"},
			strFile:  "provoke",
		}, true
	case effectPotionRed:
		return potionEffectSpec("\xbb\xa1\xb0\xa3\xc6\xf7\xbc\xc7", color.RGBA{R: 255, G: 82, B: 70, A: 255}), true
	case effectPotionOrange:
		return potionEffectSpec("\xc1\xd6\xc8\xab\xc6\xf7\xbc\xc7", color.RGBA{R: 255, G: 145, B: 58, A: 255}), true
	case effectPotionYellow:
		return potionEffectSpec("\xb3\xeb\xb6\xf5\xc6\xf7\xbc\xc7", color.RGBA{R: 255, G: 226, B: 76, A: 255}), true
	case effectPotionWhite:
		return potionEffectSpec("\xc7\xcf\xbe\xe1\xc6\xf7\xbc\xc7", color.RGBA{R: 245, G: 245, B: 255, A: 255}), true
	case effectPotionBlue:
		spec := potionEffectSpec("\xc6\xc4\xb6\xf5\xc6\xf7\xbc\xc7", color.RGBA{R: 92, G: 150, B: 255, A: 255})
		spec.sfx = []string{"effect\\\xc8\xed\xb1\xe2.wav"}
		return spec, true
	case effectPotionGreen:
		return potionEffectSpec("\xc3\xca\xb7\xcf\xc6\xf7\xbc\xc7", color.RGBA{R: 78, G: 225, B: 98, A: 255}), true
	case effectFood:
		return worldEffectSpec{
			style:    effectStylePotion,
			color:    color.RGBA{R: 255, G: 182, B: 86, A: 255},
			duration: 850 * time.Millisecond,
			strFile:  "fruit",
		}, true
	case effectFoodBlue:
		return worldEffectSpec{
			style:    effectStylePotion,
			color:    color.RGBA{R: 132, G: 112, B: 255, A: 255},
			duration: 850 * time.Millisecond,
			strFile:  "fruit",
		}, true
	default:
		return worldEffectSpec{}, false
	}
}

func potionEffectSpec(file string, c color.RGBA) worldEffectSpec {
	return worldEffectSpec{
		style:    effectStylePotion,
		color:    c,
		duration: 850 * time.Millisecond,
		strFile:  file,
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
		progress := worldEffectProgress(effect, now)
		worldX := cellCenter(x)
		worldY := cellCenter(y)
		worldZ := terrainHeightAt(ctx.World, x, y) + 0.07
		if spec.strFile != "" {
			m.drawSTREffect(screen, ctx, projection, spec, effect, worldX, worldY, worldZ, now)
			continue
		}
		switch spec.style {
		case effectStyleBash:
			m.drawBashBeginEffect(screen, ctx, projection, worldX, worldY, worldZ, progress, effect)
		case effectStyleHit:
			drawBashHitEffect(screen, m.whitePixel, worldX, worldY, worldZ, progress, spec.color)
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

func (m *WorldMode) drawBashBeginEffect(screen *render.Image, ctx Context, projection sceneProjection, x, y, z, progress float64, effect worldEffect) {
	alphaDown := m.effectTexture(ctx.Resources, "alpha_down")
	alphaCenter := m.effectTexture(ctx.Resources, "alpha_center")
	if alphaDown == nil && alphaCenter == nil {
		return
	}
	alpha := bashCylinderAlpha(progress, 0.6)
	centerZ := z + 1.5
	if alphaDown != nil {
		drawTexturedEffectCylinder(screen, projection, alphaDown, x, y, centerZ, effectCylinderDraw{
			bottomSize:       0.1,
			topSize:          2.0 * progress,
			totalCircleSides: 20,
			circleSides:      20,
			alpha:            alpha,
			angle:            bashCylinderSpin(effect, progress, 0),
		})
	}
	if alphaCenter == nil {
		return
	}
	for i := 0; i < 10; i++ {
		drawTexturedEffectCylinder(screen, projection, alphaCenter, x, y, centerZ, effectCylinderDraw{
			bottomSize:       0.01,
			topSize:          2.5 * progress,
			totalCircleSides: 30,
			circleSides:      1,
			alpha:            alpha,
			angle:            bashCylinderSpin(effect, progress, i+1) + deterministicAngle(effect, i),
		})
	}
	for i := 0; i < 8; i++ {
		drawTexturedEffectCylinder(screen, projection, alphaCenter, x, y, centerZ, effectCylinderDraw{
			bottomSize:       0.01,
			topSize:          4.0 * progress,
			totalCircleSides: 30,
			circleSides:      1,
			alpha:            alpha,
			angle:            bashCylinderSpin(effect, progress, i+11) + deterministicAngle(effect, i+10),
		})
	}
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

func bashCylinderAlpha(progress, alphaMax float64) float64 {
	switch {
	case progress < 0.25:
		return progress / 0.25 * alphaMax
	case progress > 0.75:
		return (1 - progress) / 0.25 * alphaMax
	default:
		return alphaMax
	}
}

func bashCylinderSpin(effect worldEffect, progress float64, salt int) float64 {
	return progress*2*math.Pi + deterministicAngle(effect, salt+31)*0.08
}

func deterministicAngle(effect worldEffect, salt int) float64 {
	value := uint32(effect.effectID*1103515245) ^ effect.actorID ^ uint32(effect.starts.UnixNano()) ^ uint32(salt*2654435761)
	value ^= value >> 16
	value *= 2246822519
	value ^= value >> 13
	return float64(value%360) * math.Pi / 180
}

func (m *WorldMode) drawSTREffect(screen *render.Image, ctx Context, projection sceneProjection, spec worldEffectSpec, effect worldEffect, worldX, worldY, worldZ float64, now time.Time) bool {
	str := m.loadWorldEffectSTR(ctx.Resources, spec)
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
		drawSTRAnimation(screen, projection, texture, worldX, worldY, worldZ, anim)
		drawn = true
	}
	return drawn
}

func (m *WorldMode) loadWorldEffectSTR(manager *res.Manager, spec worldEffectSpec) *res.STR {
	if manager == nil || spec.strFile == "" {
		return nil
	}
	path := "data\\texture\\effect\\" + spec.strFile + ".str"
	key := "__str_" + path + "|" + spec.texturePath
	if m.strEffects == nil {
		m.strEffects = make(map[string]*res.STR)
	}
	if m.strEffectMiss == nil {
		m.strEffectMiss = make(map[string]struct{})
	}
	if str, ok := m.strEffects[key]; ok {
		return str
	}
	if _, ok := m.strEffectMiss[key]; ok {
		return nil
	}
	data, err := manager.ReadFileExact(path)
	if err != nil {
		m.strEffectMiss[key] = struct{}{}
		log.Printf("str effect missing path=%s: %v", path, err)
		return nil
	}
	str, err := res.ParseSTR(data, spec.texturePath)
	if err != nil {
		m.strEffectMiss[key] = struct{}{}
		log.Printf("str effect parse failed path=%s: %v", path, err)
		return nil
	}
	m.strEffects[key] = str
	return str
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

func drawSTRAnimation(screen *render.Image, projection sceneProjection, texture *render.Image, worldX, worldY, worldZ float64, anim res.STRAnimation) {
	right, up, _, ok := projection.BillboardBasis(worldX, worldY, worldZ)
	if !ok {
		return
	}
	const pixelRatio = 1.0 / 35.0
	offsetX := float64(anim.Pos[0]-320) * pixelRatio
	offsetY := -float64(anim.Pos[1]-320)*pixelRatio - 0.5
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
	if anim.DestAlpha == 2 {
		options.Blend = render.BlendLighter
	}
	screen.DrawTriangles3DOwned(vertices, quadIndices012213, texture, options)
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
