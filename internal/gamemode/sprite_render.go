package gamemode

import (
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kivutar/goro/internal/res"
	"github.com/kivutar/goro/internal/session"
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
	body *playerSpriteView
	head *playerSpriteView
}

func loadPlayerSpriteView(manager *res.Manager, character session.Character, sex byte) (*playerSpriteView, string) {
	view, status := loadBodySpriteView(manager, int(character.Job), sex, "player body")
	return view, fmt.Sprintf("sprite-sex=%s(%d) %s", res.PlayerSexLabel(sex), sex, status)
}

func loadPlayerHeadSpriteView(manager *res.Manager, character session.Character, sex byte) (*playerSpriteView, string) {
	view, status := loadHeadSpriteView(manager, int(character.Job), int(character.Hair), sex, "player head")
	return view, status
}

func loadHumanoidSpriteView(manager *res.Manager, job int, head int, sex byte, label string) (*humanoidSpriteView, string) {
	body, bodyStatus := loadBodySpriteView(manager, job, sex, label+" body")
	if body == nil {
		return nil, bodyStatus
	}
	headView, headStatus := loadHeadSpriteView(manager, job, head, sex, label+" head")
	if headView == nil {
		return &humanoidSpriteView{body: body}, bodyStatus + " " + headStatus
	}
	return &humanoidSpriteView{body: body, head: headView}, bodyStatus + " " + headStatus
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
	return drawHumanoidSprite(screen, m.playerView, m.playerHeadView, 0, ctx.World.Dir, centerX, centerY)
}

func drawHumanoidSprite(screen *ebiten.Image, bodyView, headView *playerSpriteView, actionID int, direction int, centerX, centerY float64) bool {
	if bodyView == nil || bodyView.act == nil || bodyView.spr == nil {
		return false
	}
	bodyPosX, bodyPosY, rendered := drawSpriteView(screen, bodyView, actionID, direction, centerX, centerY, 0, 0)
	if headView != nil {
		_, _, headRendered := drawAttachedSpriteView(screen, headView, actionID, direction, centerX, centerY, bodyPosX, bodyPosY)
		rendered = rendered || headRendered
	}
	return rendered
}

func drawAttachedSpriteView(screen *ebiten.Image, view *playerSpriteView, actionID int, direction int, centerX, centerY float64, parentPosX, parentPosY int32) (int32, int32, bool) {
	action, ok := view.act.ActionFor(actionID, direction)
	if !ok || len(action.Animations) == 0 {
		return 0, 0, false
	}
	delay := action.DelayMS
	if delay <= 0 {
		delay = 150
	}
	animIndex := int(time.Since(view.started).Milliseconds()/int64(delay)) % len(action.Animations)
	anim := action.Animations[animIndex]
	posX, posY := int32(0), int32(0)
	if len(anim.Pos) > 0 {
		posX = parentPosX - anim.Pos[0].X
		posY = parentPosY - anim.Pos[0].Y
	}
	return drawSpriteAnimation(screen, view, anim, centerX, centerY, posX, posY)
}

func drawSpriteView(screen *ebiten.Image, view *playerSpriteView, actionID int, direction int, centerX, centerY float64, offsetX, offsetY int32) (int32, int32, bool) {
	action, ok := view.act.ActionFor(actionID, direction)
	if !ok || len(action.Animations) == 0 {
		return 0, 0, false
	}
	delay := action.DelayMS
	if delay <= 0 {
		delay = 150
	}
	animIndex := int(time.Since(view.started).Milliseconds()/int64(delay)) % len(action.Animations)
	anim := action.Animations[animIndex]
	posX, posY := offsetX, offsetY
	if len(anim.Pos) > 0 {
		posX += anim.Pos[0].X
		posY += anim.Pos[0].Y
	}
	return drawSpriteAnimation(screen, view, anim, centerX, centerY, posX, posY)
}

func drawSpriteAnimation(screen *ebiten.Image, view *playerSpriteView, anim res.ACTAnimation, centerX, centerY float64, posX, posY int32) (int32, int32, bool) {
	rendered := false
	for _, layer := range anim.Layers {
		if layer.Index < 0 {
			continue
		}
		img, ok := spriteViewImage(view, layer.Index, layer.SPRType)
		if !ok {
			continue
		}
		drawSpriteLayer(screen, img, layer, centerX+float64(posX), centerY-float64(posY))
		rendered = true
	}
	return posX, posY, rendered
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

func drawSpriteLayer(screen *ebiten.Image, img *ebiten.Image, layer res.ACTLayer, centerX, centerY float64) {
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
	opts.GeoM.Translate(centerX+float64(layer.X), centerY-float64(layer.Y))
	opts.Filter = ebiten.FilterNearest
	opts.ColorScale.Scale(layer.Color[0], layer.Color[1], layer.Color[2], layer.Color[3])
	screen.DrawImage(img, &opts)
}
