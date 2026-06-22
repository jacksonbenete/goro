package gamemode

import (
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kivutar/goro/internal/res"
	"github.com/kivutar/goro/internal/session"
)

const (
	spriteActionIdle = iota
	spriteActionWalk
	spriteActionSit
)

const (
	humanoidBillboardWidth   = 160
	humanoidBillboardHeight  = 160
	humanoidBillboardAnchorX = 80
	humanoidBillboardAnchorY = 120
)

type playerSpriteView struct {
	spr           *res.SPR
	act           *res.ACT
	source        string
	palette       *res.Palette
	paletteSource string
	images        map[spriteFrameKey]*ebiten.Image
	started       time.Time
}

type spriteFrameKey struct {
	index   int32
	sprType int32
}

type humanoidSpriteView struct {
	body       *playerSpriteView
	head       *playerSpriteView
	imf        *res.IMF
	imfSource  string
	billboards map[humanoidBillboardKey]*spriteBillboard
	started    time.Time
}

type humanoidBillboardKey struct {
	actionFamily int
	direction    int
	bodyMotion   int
	headMotion   int
}

type spriteBillboard struct {
	image   *ebiten.Image
	anchorX float64
	anchorY float64
}

type spriteState struct {
	actionFamily int
	direction    int
	moving       bool
}

func loadPlayerSpriteView(manager *res.Manager, character session.Character, sex byte) (*playerSpriteView, string) {
	view, status := loadBodySpriteView(manager, int(character.Job), sex, int(character.BodyPal), "player body")
	return view, fmt.Sprintf("sprite-sex=%s(%d) %s", res.PlayerSexLabel(sex), sex, status)
}

func loadPlayerHeadSpriteView(manager *res.Manager, character session.Character, sex byte) (*playerSpriteView, string) {
	view, status := loadHeadSpriteView(manager, int(character.Job), int(character.Hair), sex, characterHeadPalette(character), "player head")
	return view, status
}

func loadPlayerHumanoidSpriteView(manager *res.Manager, character session.Character, sex byte) (*humanoidSpriteView, string) {
	return loadHumanoidSpriteView(manager, int(character.Job), int(character.Hair), sex, int(character.BodyPal), characterHeadPalette(character), "player")
}

func characterHeadPalette(character session.Character) int {
	if character.HeadPal > 0 {
		return int(character.HeadPal)
	}
	return int(character.HairColor)
}

func loadHumanoidSpriteView(manager *res.Manager, job int, head int, sex byte, bodyPalette int, headPalette int, label string) (*humanoidSpriteView, string) {
	body, bodyStatus := loadBodySpriteView(manager, job, sex, bodyPalette, label+" body")
	if body == nil {
		return nil, bodyStatus
	}
	headView, headStatus := loadHeadSpriteView(manager, job, head, sex, headPalette, label+" head")
	imf, imfSource, imfStatus := loadPlayerIMF(manager, job, sex)
	view := &humanoidSpriteView{
		body:       body,
		head:       headView,
		imf:        imf,
		imfSource:  imfSource,
		billboards: make(map[humanoidBillboardKey]*spriteBillboard),
		started:    time.Now(),
	}
	if headView == nil {
		return view, bodyStatus + " " + headStatus + imfStatus
	}
	return view, bodyStatus + " " + headStatus + imfStatus
}

func loadBodySpriteView(manager *res.Manager, job int, sex byte, palette int, label string) (*playerSpriteView, string) {
	return loadSpriteView(manager, res.PlayerBodyResourceCandidates(job, sex, "act"), res.PlayerBodyResourceCandidates(job, sex, "spr"), res.PlayerBodyPaletteResourceCandidates(job, sex, palette, "pal"), label)
}

func loadHeadSpriteView(manager *res.Manager, job int, head int, sex byte, palette int, label string) (*playerSpriteView, string) {
	return loadSpriteView(manager, res.PlayerHeadResourceCandidates(job, head, sex, "act"), res.PlayerHeadResourceCandidates(job, head, sex, "spr"), res.PlayerHeadPaletteResourceCandidates(job, head, sex, palette, "pal"), label)
}

func loadPlayerIMF(manager *res.Manager, job int, sex byte) (*res.IMF, string, string) {
	data, source, err := readFirstResource(manager, res.PlayerIMFResourceCandidates(job, sex))
	if err != nil {
		return nil, "", " imf=missing"
	}
	imf, err := res.ParseIMF(data)
	if err != nil {
		return nil, "", fmt.Sprintf(" imf=%s parse-error=%v", source, err)
	}
	return imf, source, fmt.Sprintf(" imf=%s", source)
}

func loadSpriteView(manager *res.Manager, actCandidates []string, sprCandidates []string, palCandidates []string, label string) (*playerSpriteView, string) {
	actData, actSource, err := readFirstResource(manager, actCandidates)
	if err != nil {
		return nil, fmt.Sprintf("%s act: %v", label, err)
	}
	sprData, sprSource, err := readFirstResource(manager, sprCandidates)
	if err != nil {
		return nil, fmt.Sprintf("%s spr: %v", label, err)
	}
	act, err := res.ParseACT(actData)
	if err != nil {
		return nil, fmt.Sprintf("%s act parse %s: %v", label, actSource, err)
	}
	spr, err := res.ParseSPR(sprData)
	if err != nil {
		return nil, fmt.Sprintf("%s spr parse %s: %v", label, sprSource, err)
	}
	palette, paletteSource, paletteStatus := loadSpritePalette(manager, palCandidates)
	return &playerSpriteView{
		spr:           spr,
		act:           act,
		source:        sprSource,
		palette:       palette,
		paletteSource: paletteSource,
		images:        make(map[spriteFrameKey]*ebiten.Image),
		started:       time.Now(),
	}, fmt.Sprintf("%s: %s actions=%d frames=%d%s", label, sprSource, len(act.Actions), len(spr.Frames), paletteStatus)
}

func loadSpritePalette(manager *res.Manager, candidates []string) (*res.Palette, string, string) {
	if len(candidates) == 0 {
		return nil, "", ""
	}
	data, source, err := readFirstResource(manager, candidates)
	if err != nil {
		return nil, "", " palette=default"
	}
	palette, err := res.ParsePAL(data)
	if err != nil {
		return nil, "", fmt.Sprintf(" palette=%s parse-error=%v", source, err)
	}
	return &palette, source, fmt.Sprintf(" palette=%s", source)
}

func readFirstResource(manager *res.Manager, candidates []string) ([]byte, string, error) {
	for _, candidate := range candidates {
		data, err := manager.ReadFile(candidate)
		if err == nil {
			return data, candidate, nil
		}
	}
	return nil, "", fmt.Errorf("not found")
}

func selectedCharacter(s *session.Session) session.Character {
	if s.Selected.ID != 0 {
		return s.Selected
	}
	for _, character := range s.Characters {
		if character.ID == s.CharID {
			return character
		}
	}
	if len(s.Characters) > 0 {
		return s.Characters[0]
	}
	return session.Character{ID: s.CharID, Name: "Player", Job: 0}
}

func (m *WorldMode) drawPlayerSprite(ctx Context, screen *ebiten.Image, centerX, centerY float64) bool {
	moving := ctx.World.Player.IsMovingAt(time.Now())
	state := spriteState{
		actionFamily: spriteActionIdle,
		direction:    normalizeSpriteDirection(ctx.World.Dir),
		moving:       moving,
	}
	if moving {
		state.actionFamily = spriteActionWalk
	}
	return drawHumanoidBillboard(screen, m.playerView, state, centerX, centerY)
}

func drawHumanoidBillboard(screen *ebiten.Image, view *humanoidSpriteView, state spriteState, centerX, centerY float64) bool {
	billboard, ok := humanoidBillboardForState(view, state, time.Now())
	if !ok {
		return false
	}
	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(-billboard.anchorX, -billboard.anchorY)
	opts.GeoM.Translate(centerX, centerY)
	opts.Filter = ebiten.FilterNearest
	screen.DrawImage(billboard.image, &opts)
	return true
}

func humanoidBillboardForState(view *humanoidSpriteView, state spriteState, now time.Time) (*spriteBillboard, bool) {
	if view == nil || view.body == nil || view.body.act == nil || view.body.spr == nil {
		return nil, false
	}
	state.direction = normalizeSpriteDirection(state.direction)
	bodyActionIndex, bodyAction, ok := resolveSpriteAction(view.body.act, state.actionFamily, state.direction)
	if !ok || len(bodyAction.Animations) == 0 {
		return nil, false
	}
	bodyMotion := spriteMotionIndex(bodyAction, view.started, now, state.moving)
	headMotion := 0
	if view.head != nil {
		if _, headAction, headOK := resolveSpriteAction(view.head.act, state.actionFamily, state.direction); headOK && len(headAction.Animations) > 0 {
			headMotion = selectHeadMotion(state.actionFamily, bodyMotion, headAction)
		}
	}
	key := humanoidBillboardKey{
		actionFamily: bodyActionIndex / 8,
		direction:    bodyActionIndex % 8,
		bodyMotion:   bodyMotion,
		headMotion:   headMotion,
	}
	if billboard, ok := view.billboards[key]; ok {
		return billboard, true
	}
	billboard, ok := composeHumanoidBillboard(view, key.actionFamily, key.direction, bodyAction, bodyMotion, headMotion)
	if !ok {
		return nil, false
	}
	view.billboards[key] = billboard
	return billboard, true
}

func composeHumanoidBillboard(view *humanoidSpriteView, actionFamily, direction int, bodyAction res.ACTAction, bodyMotion, headMotion int) (*spriteBillboard, bool) {
	if bodyMotion < 0 || bodyMotion >= len(bodyAction.Animations) {
		return nil, false
	}
	target := ebiten.NewImage(humanoidBillboardWidth, humanoidBillboardHeight)
	bodyActionIndex := actionFamily*8 + direction
	drawn := false
	if view.imf != nil {
		drawn = drawPlayerIMFLayers(target, view, bodyActionIndex, bodyMotion, headMotion)
	} else {
		drawn = drawFallbackHumanoidLayers(target, view, actionFamily, direction, bodyAction, bodyMotion, headMotion)
	}
	if !drawn {
		return nil, false
	}
	return &spriteBillboard{
		image:   target,
		anchorX: humanoidBillboardAnchorX,
		anchorY: humanoidBillboardAnchorY,
	}, true
}

func drawFallbackHumanoidLayers(target *ebiten.Image, view *humanoidSpriteView, actionFamily, direction int, bodyAction res.ACTAction, bodyMotion, headMotion int) bool {
	bodyAnim := bodyAction.Animations[bodyMotion]
	drawn := drawSpriteAnimation(target, view.body, bodyAnim, humanoidBillboardAnchorX, humanoidBillboardAnchorY, 0, 0)
	if view.head != nil && view.head.act != nil && view.head.spr != nil {
		if _, headAction, ok := resolveSpriteAction(view.head.act, actionFamily, direction); ok && len(headAction.Animations) > 0 {
			if headMotion < 0 || headMotion >= len(headAction.Animations) {
				headMotion = 0
			}
			headAnim := headAction.Animations[headMotion]
			headPosX, headPosY := attachmentDelta(bodyAnim, headAnim)
			drawn = drawSpriteAnimation(target, view.head, headAnim, humanoidBillboardAnchorX, humanoidBillboardAnchorY, headPosX, headPosY) || drawn
		}
	}
	return drawn
}

func drawPlayerIMFLayers(target *ebiten.Image, view *humanoidSpriteView, actionIndex, bodyMotion, headMotion int) bool {
	order := playerRenderLayerOrder(view.imf, actionIndex, bodyMotion)
	bodyAnim, bodyAnimOK := actionAnimation(view.body.act, actionIndex, bodyMotion)
	drawn := false
	for _, layer := range order {
		switch layer {
		case 0:
			drawn = drawPlayerIMFLayer(target, view.body, view.imf, 0, actionIndex, bodyMotion, nil) || drawn
		case 1:
			if view.head == nil || view.head.act == nil || view.head.spr == nil {
				continue
			}
			var attachBase *res.ACTAnimation
			if bodyAnimOK {
				attachBase = &bodyAnim
			}
			drawn = drawPlayerIMFLayer(target, view.head, view.imf, 1, actionIndex, headMotion, attachBase) || drawn
		}
	}
	return drawn
}

func drawPlayerIMFLayer(target *ebiten.Image, sprite *playerSpriteView, imf *res.IMF, layerPriority, actionIndex, motionIndex int, attachBase *res.ACTAnimation) bool {
	if sprite == nil || sprite.act == nil || sprite.spr == nil {
		return false
	}
	if actionIndex < 0 || actionIndex >= len(sprite.act.Actions) {
		return false
	}
	action := sprite.act.Actions[actionIndex]
	if motionIndex < 0 || motionIndex >= len(action.Animations) {
		return false
	}
	anim := action.Animations[motionIndex]
	resolvedLayer := layerPriority
	if imf != nil {
		if layer := imf.LayerForPriority(layerPriority, actionIndex, motionIndex); layer >= 0 {
			resolvedLayer = layer
		}
	}
	if resolvedLayer < 0 || resolvedLayer >= len(anim.Layers) {
		return false
	}
	pointX, pointY := int32(0), int32(0)
	if imf != nil {
		pointX, pointY = imf.Point(resolvedLayer, actionIndex, motionIndex)
	}
	layer := anim.Layers[resolvedLayer]
	if attachBase != nil {
		dx, dy := attachmentDelta(*attachBase, anim)
		pointX += dx
		pointY += dy
	}
	return drawSpriteLayerByValue(target, sprite, layer, humanoidBillboardAnchorX+float64(pointX), humanoidBillboardAnchorY+float64(pointY))
}

func actionAnimation(act *res.ACT, actionIndex, motionIndex int) (res.ACTAnimation, bool) {
	if act == nil || actionIndex < 0 || actionIndex >= len(act.Actions) {
		return res.ACTAnimation{}, false
	}
	action := act.Actions[actionIndex]
	if motionIndex < 0 || motionIndex >= len(action.Animations) {
		return res.ACTAnimation{}, false
	}
	return action.Animations[motionIndex], true
}

func playerRenderLayerOrder(imf *res.IMF, actionIndex, motionIndex int) [8]int {
	if imf == nil {
		return [8]int{7, 0, 1, 4, 3, 2, 5, 6}
	}
	resolveLayerPriority := func(priority int) int {
		layer := imf.LayerForPriority(priority, actionIndex, motionIndex)
		if layer < 0 {
			layer = priority
		}
		return layer
	}

	dir := actionIndex & 7
	headLayerPassed := false
	bodyAndAccessoryExchanged := 0
	var order [8]int
	outIndex := 0
	for pass := 7; pass >= 0; pass-- {
		layer := 0
		if dir >= 2 && dir <= 5 {
			if pass == 7 {
				layer = 7
			} else if pass >= 5 && pass <= 6 {
				layer = resolveLayerPriority(pass - 5)
			} else {
				layer = 6 - pass
			}
		} else if pass >= 6 && pass <= 7 {
			layer = resolveLayerPriority(pass - 6)
		} else {
			layer = 7 - pass
		}

		originalLayer := layer
		if (headLayerPassed || layer == 1) && layer == 0 {
			headLayerPassed = true
			layer = 2
			bodyAndAccessoryExchanged++
		}
		if !headLayerPassed && layer == 1 {
			headLayerPassed = true
		}
		if bodyAndAccessoryExchanged == 1 && originalLayer == 2 {
			bodyAndAccessoryExchanged = 2
			layer = 0
		}
		if layer >= 8 {
			layer = 0
		}
		if layer == 2 {
			layer = 4
		} else if layer == 4 {
			layer = 2
		}
		order[outIndex] = layer
		outIndex++
	}

	bodyIndex, headIndex := -1, -1
	for index, layer := range order {
		if layer == 0 && bodyIndex < 0 {
			bodyIndex = index
		} else if layer == 1 && headIndex < 0 {
			headIndex = index
		}
	}
	if bodyIndex >= 0 && headIndex >= 0 && headIndex < bodyIndex {
		order[bodyIndex], order[headIndex] = order[headIndex], order[bodyIndex]
	}

	var reordered [8]int
	var delayed [8]int
	delayedCount := 0
	reorderedCount := 0
	headDrawn := false
	for _, layer := range order {
		if !headDrawn && isHeadAccessoryLayer(layer) {
			delayed[delayedCount] = layer
			delayedCount++
			continue
		}
		reordered[reorderedCount] = layer
		reorderedCount++
		if layer == 1 {
			headDrawn = true
			for index := 0; index < delayedCount; index++ {
				reordered[reorderedCount] = delayed[index]
				reorderedCount++
			}
			delayedCount = 0
		}
	}
	for index := 0; index < delayedCount; index++ {
		reordered[reorderedCount] = delayed[index]
		reorderedCount++
	}
	return reordered
}

func isHeadAccessoryLayer(layer int) bool {
	return layer == 2 || layer == 3 || layer == 4
}

func resolveSpriteAction(act *res.ACT, actionFamily, direction int) (int, res.ACTAction, bool) {
	if act == nil || len(act.Actions) == 0 {
		return 0, res.ACTAction{}, false
	}
	direction = normalizeSpriteDirection(direction)
	preferred := actionFamily*8 + direction
	if preferred >= 0 && preferred < len(act.Actions) && len(act.Actions[preferred].Animations) > 0 {
		return preferred, act.Actions[preferred], true
	}
	base := actionFamily * 8
	if base >= 0 && base < len(act.Actions) && len(act.Actions[base].Animations) > 0 {
		return base, act.Actions[base], true
	}
	for index, action := range act.Actions {
		if len(action.Animations) > 0 {
			return index, action, true
		}
	}
	return 0, res.ACTAction{}, false
}

func spriteMotionIndex(action res.ACTAction, started time.Time, now time.Time, loop bool) int {
	if len(action.Animations) == 0 {
		return 0
	}
	delay := action.DelayMS
	if delay <= 0 {
		delay = 150
	}
	elapsed := now.Sub(started)
	if elapsed < 0 {
		elapsed = 0
	}
	index := int(float64(elapsed.Milliseconds()) / float64(delay))
	if loop || len(action.Animations) == 1 {
		return index % len(action.Animations)
	}
	return index % len(action.Animations)
}

func selectHeadMotion(actionFamily int, bodyMotion int, headAction res.ACTAction) int {
	if len(headAction.Animations) == 0 {
		return 0
	}
	if actionFamily == spriteActionWalk && bodyMotion >= 0 && bodyMotion < len(headAction.Animations) {
		return bodyMotion
	}
	return 0
}

func drawSpriteAnimation(target *ebiten.Image, view *playerSpriteView, anim res.ACTAnimation, anchorX, anchorY float64, posX, posY int32) bool {
	rendered := false
	for _, layer := range anim.Layers {
		if layer.Index < 0 {
			continue
		}
		img, ok := spriteViewImage(view, layer.Index, layer.SPRType)
		if !ok {
			continue
		}
		drawSpriteLayer(target, img, layer, anchorX+float64(posX), anchorY+float64(posY))
		rendered = true
	}
	return rendered
}

func drawSpriteLayerByValue(target *ebiten.Image, view *playerSpriteView, layer res.ACTLayer, centerX, centerY float64) bool {
	if layer.Index < 0 {
		return false
	}
	img, ok := spriteViewImage(view, layer.Index, layer.SPRType)
	if !ok {
		return false
	}
	drawSpriteLayer(target, img, layer, centerX, centerY)
	return true
}

func attachmentDelta(baseAnim, attachedAnim res.ACTAnimation) (int32, int32) {
	if len(baseAnim.Pos) == 0 || len(attachedAnim.Pos) == 0 {
		return 0, 0
	}
	attached := attachedAnim.Pos[0]
	for _, base := range baseAnim.Pos {
		if base.Attr == attached.Attr {
			return base.X - attached.X, base.Y - attached.Y
		}
	}
	base := baseAnim.Pos[0]
	return base.X - attached.X, base.Y - attached.Y
}

func spriteViewImage(view *playerSpriteView, index int32, sprType int32) (*ebiten.Image, bool) {
	key := spriteFrameKey{index: index, sprType: sprType}
	if img, ok := view.images[key]; ok {
		return img, true
	}
	frame, ok := view.spr.FrameImageWithPalette(int(index), int(sprType), view.palette)
	if !ok {
		return nil, false
	}
	img := ebiten.NewImageFromImage(frame)
	view.images[key] = img
	return img, true
}

func drawSpriteLayer(target *ebiten.Image, img *ebiten.Image, layer res.ACTLayer, centerX, centerY float64) {
	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	scaleX := float64(layer.ScaleX)
	scaleY := float64(layer.ScaleY)
	if scaleX == 0 {
		scaleX = 1
	}
	if scaleY == 0 {
		scaleY = 1
	}
	if layer.Mirror {
		scaleX = -scaleX
	}

	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(-width/2, -height/2)
	opts.GeoM.Scale(scaleX, scaleY)
	if layer.Angle != 0 {
		opts.GeoM.Rotate(float64(layer.Angle) * math.Pi / 180)
	}
	layerCenterX, layerCenterY := spriteLayerCenter(centerX, centerY, layer)
	opts.GeoM.Translate(layerCenterX, layerCenterY)
	opts.Filter = ebiten.FilterNearest
	opts.ColorScale.Scale(layer.Color[0], layer.Color[1], layer.Color[2], layer.Color[3])
	target.DrawImage(img, &opts)
}

func spriteLayerCenter(centerX, centerY float64, layer res.ACTLayer) (float64, float64) {
	return centerX + float64(layer.X), centerY + float64(layer.Y)
}

func normalizeSpriteDirection(direction int) int {
	direction %= 8
	if direction < 0 {
		direction += 8
	}
	return direction
}
