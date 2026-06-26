package gamemode

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/kivutar/goro/internal/render"
	"github.com/kivutar/goro/internal/res"
	"github.com/kivutar/goro/internal/session"
)

const (
	spriteActionIdle = iota
	spriteActionWalk
	spriteActionSit
	spriteActionPickup
)

const (
	spriteActionNonPCAttack = 2
	spriteActionNonPCHurt   = 3
	spriteActionNonPCDeath  = 4
	spriteActionPCAttack1   = 5
	spriteActionPCHurt      = 6
	spriteActionPCDeath     = 8
	spriteActionPCAttack2   = 10
	spriteActionPCAttack3   = 11
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
	actSource     string
	source        string
	palette       *res.Palette
	paletteSource string
	images        map[spriteFrameKey]*render.Image
	billboards    map[singleSpriteBillboardKey]*spriteBillboard
	started       time.Time
}

type spriteFrameKey struct {
	index   int32
	sprType int32
}

type humanoidSpriteView struct {
	body            *playerSpriteView
	head            *playerSpriteView
	accessoryBottom *playerSpriteView
	accessoryMid    *playerSpriteView
	accessoryTop    *playerSpriteView
	weapon          *playerSpriteView
	weaponLight     *playerSpriteView
	shield          *playerSpriteView
	imf             *res.IMF
	imfSource       string
	billboards      map[humanoidBillboardKey]*spriteBillboard
	started         time.Time
}

type humanoidAppearance struct {
	job         int
	head        int
	sex         byte
	bodyPalette int
	headPalette int
	weapon      int
	shield      int
	headTop     int
	headMid     int
	headLow     int
}

type humanoidBillboardKey struct {
	actionFamily int
	direction    int
	bodyMotion   int
	headMotion   int
}

type singleSpriteBillboardKey struct {
	actionIndex int
	motion      int
}

type spriteBillboard struct {
	image   *render.Image
	anchorX float64
	anchorY float64
}

type spriteState struct {
	actionFamily int
	direction    int
	cameraYaw    float64
	moving       bool
	started      time.Time
	loop         bool
	loopIdle     bool
	moveSpeedMS  int
	walkDistance float64
}

func loadPlayerHumanoidSpriteView(manager *res.Manager, character session.Character, sex byte) (*humanoidSpriteView, string) {
	return loadHumanoidSpriteViewWithAppearance(manager, humanoidAppearance{
		job:         int(character.Job),
		head:        int(character.Hair),
		sex:         sex,
		bodyPalette: int(character.BodyPal),
		headPalette: characterHeadPalette(character),
		weapon:      int(character.Weapon),
		shield:      int(character.Shield),
		headTop:     int(character.HeadTop),
		headMid:     int(character.HeadMid),
		headLow:     int(character.HeadLow),
	}, "player")
}

func loadNonPCSpriteView(manager *res.Manager, job int, label string) (*playerSpriteView, string) {
	resourceName, ok := manager.JobResourceName(job)
	if !ok {
		return nil, fmt.Sprintf("%s job=%d resource-name=missing", label, job)
	}
	view, status := loadSpriteView(manager, res.NonPCSpriteResourceCandidates(job, resourceName, "act"), res.NonPCSpriteResourceCandidates(job, resourceName, "spr"), nil, label+" "+resourceName)
	if view == nil {
		return nil, status
	}
	if upgrade, ok := loadRicherNonPCSpritePair(manager, job, resourceName, len(view.act.Actions)); ok {
		view.act = upgrade.act
		view.actSource = upgrade.actSource
		view.spr = upgrade.spr
		view.source = upgrade.sprSource
		status += fmt.Sprintf(" sprite-upgraded act=%s spr=%s actions=%d frames=%d", upgrade.actSource, upgrade.sprSource, len(upgrade.act.Actions), len(upgrade.spr.Frames))
	}
	return view, status
}

func characterHeadPalette(character session.Character) int {
	if character.HeadPal > 0 {
		return int(character.HeadPal)
	}
	return int(character.HairColor)
}

func loadHumanoidSpriteView(manager *res.Manager, job int, head int, sex byte, bodyPalette int, headPalette int, label string) (*humanoidSpriteView, string) {
	return loadHumanoidSpriteViewWithAppearance(manager, humanoidAppearance{
		job:         job,
		head:        head,
		sex:         sex,
		bodyPalette: bodyPalette,
		headPalette: headPalette,
	}, label)
}

func loadHumanoidSpriteViewWithAppearance(manager *res.Manager, appearance humanoidAppearance, label string) (*humanoidSpriteView, string) {
	body, bodyStatus := loadBodySpriteView(manager, appearance.job, appearance.sex, appearance.bodyPalette, label+" body")
	if body == nil {
		return nil, bodyStatus
	}
	headView, headStatus := loadHeadSpriteView(manager, appearance.job, appearance.head, appearance.sex, appearance.headPalette, label+" head")
	imf, imfSource, imfStatus := loadPlayerIMF(manager, appearance.job, appearance.sex)
	accessoryBottom, accessoryBottomStatus := loadAccessorySpriteView(manager, appearance.job, appearance.head, appearance.sex, appearance.headLow, "", label+" accessory-bottom")
	accessoryMid, accessoryMidStatus := loadAccessorySpriteView(manager, appearance.job, appearance.head, appearance.sex, appearance.headMid, "", label+" accessory-mid")
	accessoryTop, accessoryTopStatus := loadAccessorySpriteView(manager, appearance.job, appearance.head, appearance.sex, appearance.headTop, "", label+" accessory-top")
	weapon, weaponStatus := loadWeaponOverlaySpriteView(manager, appearance.job, appearance.sex, appearance.weapon, false, label+" weapon")
	weaponLight, weaponLightStatus := loadWeaponOverlaySpriteView(manager, appearance.job, appearance.sex, appearance.weapon, true, label+" weapon-light")
	shield, shieldStatus := loadShieldOverlaySpriteView(manager, appearance.job, appearance.sex, appearance.shield, label+" shield")
	view := &humanoidSpriteView{
		body:            body,
		head:            headView,
		accessoryBottom: accessoryBottom,
		accessoryMid:    accessoryMid,
		accessoryTop:    accessoryTop,
		weapon:          weapon,
		weaponLight:     weaponLight,
		shield:          shield,
		imf:             imf,
		imfSource:       imfSource,
		billboards:      make(map[humanoidBillboardKey]*spriteBillboard),
		started:         time.Now(),
	}
	status := bodyStatus + " " + headStatus + imfStatus
	for _, overlayStatus := range []string{accessoryBottomStatus, accessoryMidStatus, accessoryTopStatus, weaponStatus, weaponLightStatus, shieldStatus} {
		if overlayStatus != "" {
			status += " " + overlayStatus
		}
	}
	return view, status
}

func loadBodySpriteView(manager *res.Manager, job int, sex byte, palette int, label string) (*playerSpriteView, string) {
	return loadSpriteView(manager, res.PlayerBodyResourceCandidates(job, sex, "act"), res.PlayerBodyResourceCandidates(job, sex, "spr"), res.PlayerBodyPaletteResourceCandidates(job, sex, palette, "pal"), label)
}

func loadHeadSpriteView(manager *res.Manager, job int, head int, sex byte, palette int, label string) (*playerSpriteView, string) {
	return loadSpriteView(manager, res.PlayerHeadResourceCandidates(job, head, sex, "act"), res.PlayerHeadResourceCandidates(job, head, sex, "spr"), res.PlayerHeadPaletteResourceCandidates(job, head, sex, palette, "pal"), label)
}

func loadAccessorySpriteView(manager *res.Manager, job int, head int, sex byte, viewID int, resourceName string, label string) (*playerSpriteView, string) {
	if viewID <= 0 {
		return nil, ""
	}
	if resourceName == "" {
		if name, ok := manager.AccessoryResourceName(viewID); ok {
			resourceName = name
		}
	}
	if viewID != 185 && resourceName == "" {
		return nil, fmt.Sprintf("%s skipped: missing accessory resource table", label)
	}
	return loadSpriteView(manager, res.PlayerAccessoryResourceCandidates(job, head, sex, viewID, resourceName, "act"), res.PlayerAccessoryResourceCandidates(job, head, sex, viewID, resourceName, "spr"), nil, label)
}

func loadWeaponOverlaySpriteView(manager *res.Manager, job int, sex byte, weapon int, secondLayer bool, label string) (*playerSpriteView, string) {
	if weapon <= 0 {
		return nil, ""
	}
	return loadSpriteView(manager, res.PlayerWeaponOverlayResourceCandidates(job, sex, weapon, secondLayer, "act"), res.PlayerWeaponOverlayResourceCandidates(job, sex, weapon, secondLayer, "spr"), nil, label)
}

func loadShieldOverlaySpriteView(manager *res.Manager, job int, sex byte, shield int, label string) (*playerSpriteView, string) {
	if shield <= 0 {
		return nil, ""
	}
	return loadSpriteView(manager, res.PlayerShieldOverlayResourceCandidates(job, sex, shield, "act"), res.PlayerShieldOverlayResourceCandidates(job, sex, shield, "spr"), nil, label)
}

func loadActorShadowSpriteView(manager *res.Manager) (*playerSpriteView, string) {
	return loadSpriteView(manager,
		[]string{"data\\sprite\\shadow.act", "data/sprite/shadow.act"},
		[]string{"data\\sprite\\shadow.spr", "data/sprite/shadow.spr"},
		nil,
		"actor shadow",
	)
}

func loadCursorSpriteView(manager *res.Manager) (*playerSpriteView, string) {
	return loadSpriteView(manager,
		[]string{"data\\sprite\\cursors.act", "data/sprite/cursors.act", "data\\sprite\\interface\\cursors.act", "data/sprite/interface/cursors.act"},
		[]string{"data\\sprite\\cursors.spr", "data/sprite/cursors.spr", "data\\sprite\\interface\\cursors.spr", "data/sprite/interface/cursors.spr"},
		nil,
		"cursor",
	)
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
		actSource:     actSource,
		source:        sprSource,
		palette:       palette,
		paletteSource: paletteSource,
		images:        make(map[spriteFrameKey]*render.Image),
		billboards:    make(map[singleSpriteBillboardKey]*spriteBillboard),
		started:       time.Now(),
	}, fmt.Sprintf("%s: %s actions=%d frames=%d%s", label, sprSource, len(act.Actions), len(spr.Frames), paletteStatus)
}

type nonPCSpritePairUpgrade struct {
	act       *res.ACT
	actSource string
	spr       *res.SPR
	sprSource string
}

func loadRicherNonPCSpritePair(manager *res.Manager, job int, resourceName string, currentActions int) (nonPCSpritePairUpgrade, bool) {
	if manager == nil || job < 1000 || currentActions > 8 {
		return nonPCSpritePairUpgrade{}, false
	}
	var best nonPCSpritePairUpgrade
	for _, archive := range manager.Archives {
		for _, candidate := range res.NonPCSpriteResourceCandidates(job, resourceName, "act") {
			data, err := archive.ReadFile(candidate)
			if err != nil {
				continue
			}
			act, err := res.ParseACT(data)
			if err != nil {
				continue
			}
			if !preferNonPCActUpgrade(job, currentActions, len(act.Actions)) {
				continue
			}
			for _, sprCandidate := range res.NonPCSpriteResourceCandidates(job, resourceName, "spr") {
				sprData, err := archive.ReadFile(sprCandidate)
				if err != nil {
					continue
				}
				spr, err := res.ParseSPR(sprData)
				if err != nil || !actFitsSPR(act, spr) {
					continue
				}
				if best.act == nil || len(act.Actions) > len(best.act.Actions) || len(act.Actions) == len(best.act.Actions) && len(spr.Frames) > len(best.spr.Frames) {
					best = nonPCSpritePairUpgrade{
						act:       act,
						actSource: archive.Path() + ":" + candidate,
						spr:       spr,
						sprSource: archive.Path() + ":" + sprCandidate,
					}
				}
			}
		}
	}
	return best, best.act != nil
}

func preferNonPCActUpgrade(job int, currentActions, candidateActions int) bool {
	return job >= 1000 && currentActions > 0 && currentActions <= 8 && candidateActions >= 40
}

func actFitsSPR(act *res.ACT, spr *res.SPR) bool {
	if act == nil || spr == nil {
		return false
	}
	for _, action := range act.Actions {
		for _, anim := range action.Animations {
			for _, layer := range anim.Layers {
				if layer.Index < 0 {
					continue
				}
				index := int(layer.Index)
				if layer.SPRType == res.SPRFrameRGBA {
					index += spr.RGBAIndex
				}
				if index < 0 || index >= len(spr.Frames) {
					return false
				}
			}
		}
	}
	return true
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

func (m *WorldMode) drawPlayerSprite3D(ctx Context, screen *render.Image, projection sceneProjection, entry sceneActorDrawEntry, direction int, cameraYaw float64, shadow float64) bool {
	now := time.Now()
	moving := ctx.World.Player.IsMovingAt(now)
	state := spriteState{
		actionFamily: spriteActionIdle,
		direction:    direction,
		cameraYaw:    cameraYaw,
		moving:       moving,
		moveSpeedMS:  ctx.World.Player.Speed,
	}
	if moving {
		state.actionFamily = spriteActionWalk
		state.loop = true
		state.walkDistance = ctx.World.Player.RenderWalkDistance(now)
	}
	if ctx.Session != nil {
		if anim, ok := m.actorAnimation(ctx.Session.CharID, now); ok {
			state.actionFamily = anim.actionFamily
			state.started = anim.started
			state.loop = false
			state.moving = false
		} else if anim, ok := m.actorAnimation(ctx.Session.AccountID, now); ok {
			state.actionFamily = anim.actionFamily
			state.started = anim.started
			state.loop = false
			state.moving = false
		}
	}
	billboard, ok := humanoidBillboardForState(m.playerView, state, now)
	if !ok {
		return false
	}
	drawSpriteBillboardAlpha3D(screen, projection, billboard, entry.worldX, entry.worldY, entry.worldZ, entry.scale, 1, shadow)
	return true
}

func drawFixedSpriteBillboardAlphaFlat3D(screen *render.Image, projection sceneProjection, view *playerSpriteView, worldX, worldY, worldZ, scale float64, alpha float64, shadow float64) bool {
	billboard, ok := fixedSpriteBillboard(view)
	if !ok {
		return false
	}
	drawSpriteBillboardAlphaFlat3D(screen, projection, billboard, worldX, worldY, worldZ, scale, alpha, shadow)
	return true
}

func drawSpriteBillboardAlpha3D(screen *render.Image, projection sceneProjection, billboard *spriteBillboard, worldX, worldY, worldZ, scale float64, alpha float64, shadow float64) {
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	if alpha < 0 || math.IsNaN(alpha) {
		alpha = 0
	}
	if alpha > 1 || math.IsInf(alpha, 0) {
		alpha = 1
	}
	if shadow < 0 || math.IsNaN(shadow) {
		shadow = 0
	}
	if shadow > 1 || math.IsInf(shadow, 0) {
		shadow = 1
	}
	right, up, unitsPerPixel, ok := projection.BillboardBasis(worldX, worldY, worldZ)
	if !ok {
		return
	}
	bounds := billboard.image.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())
	center := modelPoint3{x: worldX, y: worldZ, z: worldY}
	corner := func(px, py float64) modelPoint3 {
		dx := (px - billboard.anchorX) * scale * unitsPerPixel
		dy := (py - billboard.anchorY) * scale * unitsPerPixel
		return add3(add3(center, mul3(right, dx)), mul3(up, -dy))
	}
	depthCorner := func(px, py float64) modelPoint3 {
		dx := (px - billboard.anchorX) * scale * unitsPerPixel
		dy := (py - billboard.anchorY) * scale * unitsPerPixel
		return add3(add3(center, mul3(right, dx)), mul3(modelPoint3{y: 1}, -dy))
	}
	tint := colorRGBAFromFloats(shadow, shadow, shadow, alpha)
	vertices := []render.Vertex3D{
		spriteBillboardVertex3D(corner(0, 0), depthCorner(0, 0), texturePoint{u: 0, v: 0}, tint, float32(bounds.Dx()), float32(bounds.Dy())),
		spriteBillboardVertex3D(corner(w, 0), depthCorner(w, 0), texturePoint{u: 1, v: 0}, tint, float32(bounds.Dx()), float32(bounds.Dy())),
		spriteBillboardVertex3D(corner(w, h), depthCorner(w, h), texturePoint{u: 1, v: 1}, tint, float32(bounds.Dx()), float32(bounds.Dy())),
		spriteBillboardVertex3D(corner(0, h), depthCorner(0, h), texturePoint{u: 0, v: 1}, tint, float32(bounds.Dx()), float32(bounds.Dy())),
	}
	screen.DrawTriangles3D(vertices, []uint16{0, 1, 2, 0, 2, 3}, billboard.image, spriteBillboardTriangleDrawOptions())
}

func drawSpriteBillboardAlphaFlat3D(screen *render.Image, projection sceneProjection, billboard *spriteBillboard, worldX, worldY, worldZ, scale float64, alpha float64, shadow float64) {
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	if alpha < 0 || math.IsNaN(alpha) {
		alpha = 0
	}
	if alpha > 1 || math.IsInf(alpha, 0) {
		alpha = 1
	}
	if shadow < 0 || math.IsNaN(shadow) {
		shadow = 0
	}
	if shadow > 1 || math.IsInf(shadow, 0) {
		shadow = 1
	}
	_, _, unitsPerPixel, ok := projection.BillboardBasis(worldX, worldY, worldZ)
	if !ok {
		return
	}
	yaw := degreesToRadians(projection.cameraYaw)
	right := normalize3(modelPoint3{x: math.Cos(yaw), z: math.Sin(yaw)})
	down := normalize3(modelPoint3{x: math.Sin(yaw), z: -math.Cos(yaw)})
	if right == (modelPoint3{}) {
		right = modelPoint3{x: 1}
	}
	if down == (modelPoint3{}) {
		down = modelPoint3{z: -1}
	}
	bounds := billboard.image.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())
	center := modelPoint3{x: worldX, y: worldZ, z: worldY}
	corner := func(px, py float64) modelPoint3 {
		dx := (px - billboard.anchorX) * scale * unitsPerPixel
		dy := (py - billboard.anchorY) * scale * unitsPerPixel
		return add3(add3(center, mul3(right, dx)), mul3(down, dy))
	}
	tint := colorRGBAFromFloats(shadow, shadow, shadow, alpha)
	vertices := []render.Vertex3D{
		texturedSurfaceVertex3D(corner(0, 0), texturePoint{u: 0, v: 0}, tint, float32(bounds.Dx()), float32(bounds.Dy())),
		texturedSurfaceVertex3D(corner(w, 0), texturePoint{u: 1, v: 0}, tint, float32(bounds.Dx()), float32(bounds.Dy())),
		texturedSurfaceVertex3D(corner(w, h), texturePoint{u: 1, v: 1}, tint, float32(bounds.Dx()), float32(bounds.Dy())),
		texturedSurfaceVertex3D(corner(0, h), texturePoint{u: 0, v: 1}, tint, float32(bounds.Dx()), float32(bounds.Dy())),
	}
	screen.DrawTriangles3D(vertices, []uint16{0, 1, 2, 0, 2, 3}, billboard.image, triangleDrawOptions(spriteDrawFilter(), render.AddressClampToZero))
}

func spriteBillboardTriangleDrawOptions() *render.DrawTrianglesOptions {
	options := triangleDrawOptions(spriteDrawFilter(), render.AddressClampToZero)
	return options
}

func spriteBillboardVertex3D(point, depthPoint modelPoint3, uv texturePoint, tint color.RGBA, textureWidth, textureHeight float32) render.Vertex3D {
	vertex := texturedSurfaceVertex3D(point, uv, tint, textureWidth, textureHeight)
	vertex.DepthX = float32(depthPoint.x)
	vertex.DepthY = float32(depthPoint.y)
	vertex.DepthZ = float32(depthPoint.z)
	return vertex
}

func colorRGBAFromFloats(r, g, b, a float64) color.RGBA {
	return color.RGBA{
		R: clampColor(r * 255),
		G: clampColor(g * 255),
		B: clampColor(b * 255),
		A: clampColor(a * 255),
	}
}

func spriteDrawFilter() render.Filter {
	return render.FilterLinear
}

func spriteCompositionFilter() render.Filter {
	return render.FilterNearest
}

func humanoidBillboardForState(view *humanoidSpriteView, state spriteState, now time.Time) (*spriteBillboard, bool) {
	if view == nil || view.body == nil || view.body.act == nil || view.body.spr == nil {
		return nil, false
	}
	state.direction = spriteDirectionFromWorldDirForCamera(state.direction, state.cameraYaw)
	bodyActionIndex, bodyAction, ok := resolveSpriteAction(view.body.act, state.actionFamily, state.direction)
	if !ok || len(bodyAction.Animations) == 0 {
		return nil, false
	}
	bodyMotion := bodyMotionForState(bodyAction, state, view.started, now)
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

func singleSpriteBillboardForState(view *playerSpriteView, state spriteState, now time.Time) (*spriteBillboard, bool) {
	if view == nil || view.act == nil || view.spr == nil {
		return nil, false
	}
	state.direction = spriteDirectionFromWorldDirForCamera(state.direction, state.cameraYaw)
	actionIndex, action, ok := resolveSpriteAction(view.act, state.actionFamily, state.direction)
	if !ok || len(action.Animations) == 0 {
		return nil, false
	}
	motion := bodyMotionForState(action, state, view.started, now)
	key := singleSpriteBillboardKey{actionIndex: actionIndex, motion: motion}
	if billboard, ok := view.billboards[key]; ok {
		return billboard, true
	}
	if motion < 0 || motion >= len(action.Animations) {
		return nil, false
	}
	billboard, ok := composeSingleSpriteBillboard(view, action.Animations[motion])
	if !ok {
		return nil, false
	}
	view.billboards[key] = billboard
	return billboard, true
}

func fixedSpriteBillboard(view *playerSpriteView) (*spriteBillboard, bool) {
	if view == nil || view.act == nil || view.spr == nil || len(view.act.Actions) == 0 || len(view.act.Actions[0].Animations) == 0 {
		return nil, false
	}
	key := singleSpriteBillboardKey{actionIndex: 0, motion: 0}
	if billboard, ok := view.billboards[key]; ok {
		return billboard, true
	}
	billboard, ok := composeSingleSpriteBillboard(view, view.act.Actions[0].Animations[0])
	if !ok {
		return nil, false
	}
	view.billboards[key] = billboard
	return billboard, true
}

func cursorFrameBillboard(view *playerSpriteView, actionIndex, motion int, anchorX, anchorY float64) (*spriteBillboard, bool) {
	if view == nil || view.act == nil || view.spr == nil || len(view.act.Actions) == 0 {
		return nil, false
	}
	if actionIndex < 0 || actionIndex >= len(view.act.Actions) || len(view.act.Actions[actionIndex].Animations) == 0 {
		actionIndex = 0
	}
	action := view.act.Actions[actionIndex]
	if len(action.Animations) == 0 {
		return nil, false
	}
	if motion < 0 {
		motion = 0
	}
	motion %= len(action.Animations)
	key := singleSpriteBillboardKey{actionIndex: actionIndex, motion: motion}
	if billboard, ok := view.billboards[key]; ok {
		return billboard, true
	}
	target := render.NewImage(50, 50)
	if !drawSpriteAnimation(target, view, action.Animations[motion], anchorX, anchorY, 0, 0) {
		return nil, false
	}
	billboard := &spriteBillboard{
		image:   target,
		anchorX: anchorX,
		anchorY: anchorY,
	}
	view.billboards[key] = billboard
	return billboard, true
}

func composeHumanoidBillboard(view *humanoidSpriteView, actionFamily, direction int, bodyAction res.ACTAction, bodyMotion, headMotion int) (*spriteBillboard, bool) {
	if bodyMotion < 0 || bodyMotion >= len(bodyAction.Animations) {
		return nil, false
	}
	target := render.NewImage(humanoidBillboardWidth, humanoidBillboardHeight)
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

func composeSingleSpriteBillboard(view *playerSpriteView, anim res.ACTAnimation) (*spriteBillboard, bool) {
	minX, minY, maxX, maxY, ok := spriteAnimationLayerBounds(view, anim)
	if !ok {
		return nil, false
	}
	const padding = 4.0
	minX -= padding
	minY -= padding
	maxX += padding
	maxY += padding
	width := int(math.Ceil(maxX - minX))
	height := int(math.Ceil(maxY - minY))
	if width <= 0 || height <= 0 {
		return nil, false
	}
	target := render.NewImage(width, height)
	anchorX := -minX
	anchorY := -minY
	if !drawSpriteAnimation(target, view, anim, anchorX, anchorY, 0, 0) {
		return nil, false
	}
	return &spriteBillboard{
		image:   target,
		anchorX: anchorX,
		anchorY: anchorY,
	}, true
}

func spriteAnimationLayerBounds(view *playerSpriteView, anim res.ACTAnimation) (float64, float64, float64, float64, bool) {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	ok := false
	for _, layer := range anim.Layers {
		if layer.Index < 0 {
			continue
		}
		width, height, frameOK := spriteLayerFrameSize(view, layer.Index, layer.SPRType)
		if !frameOK {
			continue
		}
		scaleX := math.Abs(float64(layer.ScaleX))
		scaleY := math.Abs(float64(layer.ScaleY))
		if scaleX == 0 {
			scaleX = 1
		}
		if scaleY == 0 {
			scaleY = 1
		}
		width *= scaleX
		height *= scaleY
		centerX, centerY := spriteLayerCenter(0, 0, layer)
		minX = math.Min(minX, centerX-width*0.5)
		maxX = math.Max(maxX, centerX+width*0.5)
		minY = math.Min(minY, centerY-height*0.5)
		maxY = math.Max(maxY, centerY+height*0.5)
		ok = true
	}
	return minX, minY, maxX, maxY, ok
}

func spriteLayerFrameSize(view *playerSpriteView, index int32, sprType int32) (float64, float64, bool) {
	if view == nil || view.spr == nil {
		return 0, 0, false
	}
	frameIndex := int(index)
	if sprType == res.SPRFrameRGBA {
		frameIndex += view.spr.RGBAIndex
	}
	if frameIndex < 0 || frameIndex >= len(view.spr.Frames) {
		return 0, 0, false
	}
	frame := view.spr.Frames[frameIndex]
	if frame.Width <= 0 || frame.Height <= 0 {
		return 0, 0, false
	}
	return float64(frame.Width), float64(frame.Height), true
}

func drawFallbackHumanoidLayers(target *render.Image, view *humanoidSpriteView, actionFamily, direction int, bodyAction res.ACTAction, bodyMotion, headMotion int) bool {
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

func drawPlayerIMFLayers(target *render.Image, view *humanoidSpriteView, actionIndex, bodyMotion, headMotion int) bool {
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
		case 2:
			if bodyAnimOK {
				drawn = drawAttachedAccessoryMotion(target, view.accessoryBottom, view.head, bodyAnim, actionIndex, headMotion) || drawn
			}
		case 3:
			if bodyAnimOK {
				drawn = drawAttachedAccessoryMotion(target, view.accessoryMid, view.head, bodyAnim, actionIndex, headMotion) || drawn
			}
		case 4:
			if bodyAnimOK {
				drawn = drawAttachedAccessoryMotion(target, view.accessoryTop, view.head, bodyAnim, actionIndex, headMotion) || drawn
			}
		case 5:
			drawn = drawPlayerOverlayMotion(target, view.weapon, view.body, view.imf, 5, actionIndex, bodyMotion) || drawn
		case 6:
			drawn = drawPlayerOverlayMotion(target, view.weaponLight, view.body, view.imf, 6, actionIndex, bodyMotion) || drawn
		case 7:
			drawn = drawPlayerOverlayMotion(target, view.shield, view.body, view.imf, 7, actionIndex, bodyMotion) || drawn
		}
	}
	return drawn
}

func drawPlayerIMFLayer(target *render.Image, sprite *playerSpriteView, imf *res.IMF, layerPriority, actionIndex, motionIndex int, attachBase *res.ACTAnimation) bool {
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
	drawLayerIndex := visibleACTLayerIndex(anim, resolvedLayer, layerPriority)
	if drawLayerIndex < 0 {
		return false
	}
	pointX, pointY := int32(0), int32(0)
	if imf != nil {
		pointX, pointY = imf.Point(drawLayerIndex, actionIndex, motionIndex)
	}
	layer := anim.Layers[drawLayerIndex]
	if attachBase != nil {
		dx, dy := attachmentDelta(*attachBase, anim)
		pointX += dx
		pointY += dy
	}
	return drawSpriteLayerByValue(target, sprite, layer, humanoidBillboardAnchorX+float64(pointX), humanoidBillboardAnchorY+float64(pointY))
}

func visibleACTLayerIndex(anim res.ACTAnimation, candidates ...int) int {
	for _, index := range candidates {
		if index >= 0 && index < len(anim.Layers) && anim.Layers[index].Index >= 0 {
			return index
		}
	}
	for index, layer := range anim.Layers {
		if layer.Index >= 0 {
			return index
		}
	}
	return -1
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

func drawAttachedAccessoryMotion(target *render.Image, accessory *playerSpriteView, head *playerSpriteView, bodyAnim res.ACTAnimation, actionIndex, headMotionIndex int) bool {
	if accessory == nil || accessory.act == nil || accessory.spr == nil || head == nil || head.act == nil || head.spr == nil {
		return false
	}
	headAnim, ok := actionAnimation(head.act, actionIndex, headMotionIndex)
	if !ok {
		return false
	}
	accessoryMotion := headMotionIndex
	accessoryAnim, ok := actionAnimation(accessory.act, actionIndex, accessoryMotion)
	if !ok {
		accessoryAnim, ok = actionAnimation(accessory.act, actionIndex, 0)
	}
	if !ok {
		return false
	}
	headDX, headDY := attachmentDelta(bodyAnim, headAnim)
	accessoryDX, accessoryDY := attachmentDelta(headAnim, accessoryAnim)
	return drawSpriteAnimation(target, accessory, accessoryAnim, humanoidBillboardAnchorX, humanoidBillboardAnchorY, headDX+accessoryDX, headDY+accessoryDY)
}

func drawPlayerOverlayMotion(target *render.Image, overlay *playerSpriteView, body *playerSpriteView, imf *res.IMF, layerPriority, actionIndex, bodyMotion int) bool {
	if overlay == nil || overlay.act == nil || overlay.spr == nil || body == nil || body.act == nil || imf == nil {
		return false
	}
	motionIndex := resolveOverlayMotionIndex(overlay.act, body.act, actionIndex, bodyMotion)
	overlayAnim, ok := actionAnimation(overlay.act, actionIndex, motionIndex)
	if !ok {
		overlayAnim, ok = actionAnimation(overlay.act, actionIndex, 0)
	}
	if !ok {
		return false
	}
	pointX, pointY := imf.Point(layerPriority, actionIndex, bodyMotion)
	return drawSpriteAnimation(target, overlay, overlayAnim, humanoidBillboardAnchorX, humanoidBillboardAnchorY, pointX, pointY)
}

func resolveOverlayMotionIndex(overlayAct *res.ACT, bodyAct *res.ACT, actionIndex, bodyMotion int) int {
	if overlayAct == nil || actionIndex < 0 || actionIndex >= len(overlayAct.Actions) {
		return bodyMotion
	}
	overlayMotionCount := len(overlayAct.Actions[actionIndex].Animations)
	if overlayMotionCount <= 0 {
		return 0
	}
	motionIndex := bodyMotion
	if bodyAct != nil && actionIndex >= 0 && actionIndex < len(bodyAct.Actions) {
		bodyMotionCount := len(bodyAct.Actions[actionIndex].Animations)
		if bodyMotionCount > 0 && overlayMotionCount > bodyMotionCount && overlayMotionCount%bodyMotionCount == 0 {
			motionIndex = bodyMotion * (overlayMotionCount / bodyMotionCount)
		}
	}
	if motionIndex < 0 {
		return 0
	}
	if motionIndex >= overlayMotionCount {
		return overlayMotionCount - 1
	}
	return motionIndex
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
	reordered = ensureRenderLayerPresent(reordered, 0)
	reordered = ensureRenderLayerPresent(reordered, 1)
	return reordered
}

func isHeadAccessoryLayer(layer int) bool {
	return layer == 2 || layer == 3 || layer == 4
}

func ensureRenderLayerPresent(order [8]int, required int) [8]int {
	counts := make(map[int]int, len(order))
	for _, layer := range order {
		counts[layer]++
	}
	if counts[required] > 0 {
		return order
	}
	for index := len(order) - 1; index >= 0; index-- {
		layer := order[index]
		if counts[layer] > 1 {
			counts[layer]--
			order[index] = required
			return order
		}
	}
	order[len(order)-1] = required
	return order
}

func resolveSpriteAction(act *res.ACT, actionFamily, direction int) (int, res.ACTAction, bool) {
	if act == nil || len(act.Actions) == 0 {
		return 0, res.ACTAction{}, false
	}
	direction = normalizeDirectionIndex(direction)
	preferred := actionFamily*8 + direction
	if preferred >= 0 && preferred < len(act.Actions) && len(act.Actions[preferred].Animations) > 0 {
		return preferred, act.Actions[preferred], true
	}
	if actionFamily >= 0 && actionFamily < len(act.Actions) && len(act.Actions[actionFamily].Animations) > 0 {
		return actionFamily, act.Actions[actionFamily], true
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
	return spriteMotionIndexWithDelay(action, started, now, loop, float64(action.DelayMS))
}

func spriteMotionIndexWithDelay(action res.ACTAction, started time.Time, now time.Time, loop bool, delayMS float64) int {
	if len(action.Animations) == 0 {
		return 0
	}
	delay := delayMS
	if delay <= 0 {
		delay = 150
	}
	elapsed := now.Sub(started)
	if elapsed < 0 {
		elapsed = 0
	}
	index := int(float64(elapsed.Milliseconds()) / delay)
	if loop || len(action.Animations) == 1 {
		return index % len(action.Animations)
	}
	if index >= len(action.Animations) {
		return len(action.Animations) - 1
	}
	return index
}

func bodyMotionForState(action res.ACTAction, state spriteState, started time.Time, now time.Time) int {
	if !state.started.IsZero() {
		started = state.started
	}
	delayMS := float64(action.DelayMS)
	if state.actionFamily == spriteActionWalk && state.moveSpeedMS > 0 {
		if delayMS <= 0 {
			delayMS = 150
		}
		delayMS = delayMS * float64(state.moveSpeedMS) / 150
	}
	if state.actionFamily == spriteActionWalk && state.walkDistance > 0 {
		return walkMotionIndex(action, state.walkDistance)
	}
	if state.actionFamily == spriteActionWalk || state.loop {
		return spriteMotionIndexWithDelay(action, started, now, true, delayMS)
	}
	if state.actionFamily == spriteActionIdle && state.loopIdle {
		return spriteMotionIndexWithDelay(action, started, now, true, delayMS)
	}
	if !state.started.IsZero() {
		return spriteMotionIndexWithDelay(action, started, now, false, delayMS)
	}
	if state.actionFamily != spriteActionWalk {
		return 0
	}
	return spriteMotionIndexWithDelay(action, started, now, true, delayMS)
}

const walkDistanceToMotion = 4.6 * 0.37 * 4 * 25

func walkMotionIndex(action res.ACTAction, distance float64) int {
	if len(action.Animations) == 0 {
		return 0
	}
	delay := float64(action.DelayMS)
	if delay <= 0 {
		delay = 150
	}
	motion := int(math.Floor(distance * walkDistanceToMotion / delay))
	if motion < 0 {
		return 0
	}
	return motion % len(action.Animations)
}

func selectHeadMotion(actionFamily int, bodyMotion int, headAction res.ACTAction) int {
	if len(headAction.Animations) == 0 {
		return 0
	}
	if actionFamily == spriteActionWalk && bodyMotion >= 0 && bodyMotion < len(headAction.Animations) {
		return bodyMotion
	}
	if isTransientPCAction(actionFamily) && bodyMotion >= 0 && bodyMotion < len(headAction.Animations) {
		return bodyMotion
	}
	return 0
}

func isTransientPCAction(actionFamily int) bool {
	switch actionFamily {
	case spriteActionPickup, spriteActionPCAttack1, spriteActionPCHurt, spriteActionPCDeath, spriteActionPCAttack2, spriteActionPCAttack3:
		return true
	default:
		return false
	}
}

func drawSpriteAnimation(target *render.Image, view *playerSpriteView, anim res.ACTAnimation, anchorX, anchorY float64, posX, posY int32) bool {
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

func drawSpriteLayerByValue(target *render.Image, view *playerSpriteView, layer res.ACTLayer, centerX, centerY float64) bool {
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

func spriteViewImage(view *playerSpriteView, index int32, sprType int32) (*render.Image, bool) {
	key := spriteFrameKey{index: index, sprType: sprType}
	if img, ok := view.images[key]; ok {
		return img, true
	}
	frame, ok := view.spr.FrameImageWithPalette(int(index), int(sprType), view.palette)
	if !ok {
		return nil, false
	}
	img := render.NewImageFromImage(frame)
	view.images[key] = img
	return img, true
}

func drawSpriteLayer(target *render.Image, img *render.Image, layer res.ACTLayer, centerX, centerY float64) {
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

	var opts render.DrawImageOptions
	opts.GeoM.Translate(-width/2, -height/2)
	opts.GeoM.Scale(scaleX, scaleY)
	if layer.Angle != 0 {
		opts.GeoM.Rotate(float64(layer.Angle) * math.Pi / 180)
	}
	layerCenterX, layerCenterY := spriteLayerCenter(centerX, centerY, layer)
	opts.GeoM.Translate(layerCenterX, layerCenterY)
	opts.Filter = spriteCompositionFilter()
	opts.ColorScale.Scale(layer.Color[0], layer.Color[1], layer.Color[2], layer.Color[3])
	target.DrawImage(img, &opts)
}

func spriteLayerCenter(centerX, centerY float64, layer res.ACTLayer) (float64, float64) {
	return centerX + float64(layer.X), centerY + float64(layer.Y)
}

func normalizeDirectionIndex(direction int) int {
	direction %= 8
	if direction < 0 {
		direction += 8
	}
	return direction
}

func spriteDirectionFromWorldDir(direction int) int {
	return spriteDirectionFromWorldDirForCamera(direction, 0)
}

func spriteDirectionFromWorldDirForCamera(direction int, cameraYaw float64) int {
	yawSteps := int(math.Round(cameraYaw / 45))
	return (4 - normalizeDirectionIndex(direction) + yawSteps) & 7
}
