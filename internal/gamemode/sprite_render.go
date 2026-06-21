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
	spr     *res.SPR
	act     *res.ACT
	source  string
	images  map[spriteFrameKey]*ebiten.Image
	started time.Time
}

type spriteFrameKey struct {
	index   int32
	sprType int32
}

type humanoidSpriteView struct {
	body       *playerSpriteView
	head       *playerSpriteView
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
	view, status := loadBodySpriteView(manager, int(character.Job), sex, "player body")
	return view, fmt.Sprintf("sprite-sex=%s(%d) %s", res.PlayerSexLabel(sex), sex, status)
}

func loadPlayerHeadSpriteView(manager *res.Manager, character session.Character, sex byte) (*playerSpriteView, string) {
	view, status := loadHeadSpriteView(manager, int(character.Job), int(character.Hair), sex, "player head")
	return view, status
}

func loadPlayerHumanoidSpriteView(manager *res.Manager, character session.Character, sex byte) (*humanoidSpriteView, string) {
	return loadHumanoidSpriteView(manager, int(character.Job), int(character.Hair), sex, "player")
}

func loadHumanoidSpriteView(manager *res.Manager, job int, head int, sex byte, label string) (*humanoidSpriteView, string) {
	body, bodyStatus := loadBodySpriteView(manager, job, sex, label+" body")
	if body == nil {
		return nil, bodyStatus
	}
	headView, headStatus := loadHeadSpriteView(manager, job, head, sex, label+" head")
	view := &humanoidSpriteView{
		body:       body,
		head:       headView,
		billboards: make(map[humanoidBillboardKey]*spriteBillboard),
		started:    time.Now(),
	}
	if headView == nil {
		return view, bodyStatus + " " + headStatus
	}
	return view, bodyStatus + " " + headStatus
}

func loadBodySpriteView(manager *res.Manager, job int, sex byte, label string) (*playerSpriteView, string) {
	return loadSpriteView(manager, res.PlayerBodyResourceCandidates(job, sex, "act"), res.PlayerBodyResourceCandidates(job, sex, "spr"), label)
}

func loadHeadSpriteView(manager *res.Manager, job int, head int, sex byte, label string) (*playerSpriteView, string) {
	return loadSpriteView(manager, res.PlayerHeadResourceCandidates(job, head, sex, "act"), res.PlayerHeadResourceCandidates(job, head, sex, "spr"), label)
}

func loadSpriteView(manager *res.Manager, actCandidates []string, sprCandidates []string, label string) (*playerSpriteView, string) {
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
	return &playerSpriteView{
		spr:     spr,
		act:     act,
		source:  sprSource,
		images:  make(map[spriteFrameKey]*ebiten.Image),
		started: time.Now(),
	}, fmt.Sprintf("%s: %s actions=%d frames=%d", label, sprSource, len(act.Actions), len(spr.Frames))
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
			headMotion = spriteMotionIndex(headAction, view.started, now, state.moving)
			if headMotion >= len(headAction.Animations) {
				headMotion = len(headAction.Animations) - 1
			}
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
	bodyAnim := bodyAction.Animations[bodyMotion]
	bodyPosX, bodyPosY, bodyDrawn := drawSpriteAnimation(target, view.body, bodyAnim, humanoidBillboardAnchorX, humanoidBillboardAnchorY, 0, 0)
	drawn := bodyDrawn
	if view.head != nil && view.head.act != nil && view.head.spr != nil {
		if _, headAction, ok := resolveSpriteAction(view.head.act, actionFamily, direction); ok && len(headAction.Animations) > 0 {
			if headMotion < 0 || headMotion >= len(headAction.Animations) {
				headMotion = 0
			}
			headAnim := headAction.Animations[headMotion]
			headPosX, headPosY := int32(0), int32(0)
			if len(headAnim.Pos) > 0 {
				headPosX = bodyPosX - headAnim.Pos[0].X
				headPosY = bodyPosY - headAnim.Pos[0].Y
			}
			_, _, headDrawn := drawSpriteAnimation(target, view.head, headAnim, humanoidBillboardAnchorX, humanoidBillboardAnchorY, headPosX, headPosY)
			drawn = drawn || headDrawn
		}
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

func drawSpriteAnimation(target *ebiten.Image, view *playerSpriteView, anim res.ACTAnimation, anchorX, anchorY float64, posX, posY int32) (int32, int32, bool) {
	rendered := false
	baseX, baseY := posX, posY
	if len(anim.Pos) > 0 {
		baseX += anim.Pos[0].X
		baseY += anim.Pos[0].Y
	}
	for _, layer := range anim.Layers {
		if layer.Index < 0 {
			continue
		}
		img, ok := spriteViewImage(view, layer.Index, layer.SPRType)
		if !ok {
			continue
		}
		drawSpriteLayer(target, img, layer, anchorX+float64(baseX), anchorY+float64(baseY))
		rendered = true
	}
	return baseX, baseY, rendered
}

func spriteViewImage(view *playerSpriteView, index int32, sprType int32) (*ebiten.Image, bool) {
	key := spriteFrameKey{index: index, sprType: sprType}
	if img, ok := view.images[key]; ok {
		return img, true
	}
	frame, ok := view.spr.FrameImage(int(index), int(sprType))
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
