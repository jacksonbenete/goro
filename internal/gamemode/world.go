package gamemode

import (
	"context"
	"fmt"
	"hash/fnv"
	"image/color"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kivutar/goro/internal/network"
	"github.com/kivutar/goro/internal/render"
	"github.com/kivutar/goro/internal/res"
	"github.com/kivutar/goro/internal/session"
	worldstate "github.com/kivutar/goro/internal/world"
)

type WorldMode struct {
	status        string
	walkCooldown  int
	tickCooldown  int
	camera        followCamera
	whitePixel    *ebiten.Image
	textures      map[string]*ebiten.Image
	textureMiss   map[string]struct{}
	rswMarkers    bool
	rsmRender     bool
	playerView    *humanoidSpriteView
	actorViews    map[actorSpriteKey]*humanoidSpriteView
	actorViewMiss map[actorSpriteKey]struct{}
	nonPCViews    map[int]*playerSpriteView
	nonPCViewMiss map[int]struct{}
	rsmDebugLog   map[string]struct{}
	pendingWarp   bool
}

type actorSpriteKey struct {
	job     int
	head    int
	sex     byte
	weapon  int
	shield  int
	headTop int
	headMid int
	headLow int
}

func NewWorldMode() *WorldMode {
	return &WorldMode{}
}

func (m *WorldMode) Name() string {
	return "world"
}

func (m *WorldMode) Enter(ctx Context) {
	m.status = "loading map"
	m.camera.Reset()
	ctx.World.GAT = nil
	ctx.World.GND = nil
	ctx.World.RSW = nil
	ctx.World.RSM = nil
	ctx.World.RSMFail = 0
	m.textures = make(map[string]*ebiten.Image)
	m.textureMiss = make(map[string]struct{})
	m.rswMarkers = os.Getenv("GORO_DEBUG_RSW_MARKERS") == "1"
	m.rsmRender = os.Getenv("GORO_RENDER_RSM") != "0"
	m.playerView = nil
	m.actorViews = make(map[actorSpriteKey]*humanoidSpriteView)
	m.actorViewMiss = make(map[actorSpriteKey]struct{})
	m.nonPCViews = make(map[int]*playerSpriteView)
	m.nonPCViewMiss = make(map[int]struct{})
	m.rsmDebugLog = make(map[string]struct{})
	m.pendingWarp = false
	playerStatus := ""
	character := selectedCharacter(ctx.Session)
	if view, status := loadPlayerHumanoidSpriteView(ctx.Resources, character, ctx.Session.Sex); view != nil {
		m.playerView = view
		playerStatus = status
	} else {
		playerStatus = status
	}
	log.Printf("player sprite resources char_id=%d name=%s job=%d hair=%d weapon=%d shield=%d head_top=%d head_mid=%d head_low=%d body_pal=%d head_pal=%d hair_color=%d account_sex=%d %s", character.ID, character.Name, character.Job, character.Hair, character.Weapon, character.Shield, character.HeadTop, character.HeadMid, character.HeadLow, character.BodyPal, character.HeadPal, character.HairColor, ctx.Session.Sex, playerStatus)
	if ctx.World.MapName == "" {
		m.status = "no map selected"
		return
	}

	gat, source, err := loadGAT(ctx.Resources, ctx.World.MapName)
	if err != nil {
		m.status = err.Error()
		return
	}
	ctx.World.GAT = gat
	m.status = fmt.Sprintf("loaded %s %dx%d", source, gat.Width, gat.Height)
	if gnd, gndSource, err := loadGND(ctx.Resources, ctx.World.MapName); err == nil {
		ctx.World.GND = gnd
		m.status = fmt.Sprintf("loaded %s %dx%d", gndSource, gnd.Width, gnd.Height)
	} else {
		ctx.World.GND = nil
		m.status += " gnd: " + err.Error()
	}
	if rsw, rswSource, err := loadRSW(ctx.Resources, ctx.World.MapName); err == nil {
		ctx.World.RSW = rsw
		ctx.World.RSM, ctx.World.RSMFail = loadRSMModels(ctx.Resources, rsw, rsmLoadLimit())
		m.status += fmt.Sprintf(" rsw=%s", rswSource)
	} else {
		ctx.World.RSW = nil
		ctx.World.RSM = nil
		ctx.World.RSMFail = 0
		m.status += " rsw: " + err.Error()
	}
	if err := ctx.Network.SendLoadEndAck(); err != nil {
		m.status += " load-ack failed: " + err.Error()
	} else {
		m.tickCooldown = 1
	}
	if playerStatus != "" {
		m.status += " " + playerStatus
	}
}

func (m *WorldMode) Update(ctx Context) (Mode, error) {
	for _, pkt := range ctx.Network.DrainPackets() {
		if change, ok, err := network.ParseMapChange(pkt); err != nil {
			log.Printf("parse map change 0x%04X: %v", pkt.ID, err)
		} else if ok {
			return m.handleMapChange(ctx, change), nil
		}
		if enter, err := network.ParseMapAcceptEnter(pkt); err == nil {
			applyMapAcceptEnter(ctx, enter)
			m.status = fmt.Sprintf("entered map %s at %d,%d dir=%d tick=%d", ctx.World.MapName, enter.X, enter.Y, enter.Dir, enter.ServerTick)
			if m.pendingWarp {
				m.pendingWarp = false
				return NewWorldMode(), nil
			}
			continue
		}
		if ack, ok, err := network.ParseActorNameAck(pkt); err != nil {
			log.Printf("parse actor name ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyActorNameAck(ctx, ack)
			continue
		}
		if ack, ok, err := network.ParseSelfMoveAck(pkt); err != nil {
			log.Printf("parse self move ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applySelfMoveAck(ctx, ack)
			m.status = fmt.Sprintf("walk ack: %d,%d -> %d,%d", ack.FromX, ack.FromY, ack.ToX, ack.ToY)
			continue
		}
		if position, ok, err := network.ParseActorSetPosition(pkt); err != nil {
			log.Printf("parse actor set position 0x%04X: %v", pkt.ID, err)
		} else if ok {
			if isLocalActor(ctx, position.ID) {
				m.status = fmt.Sprintf("position fix: %d,%d", position.X, position.Y)
			}
			applyActorSetPosition(ctx, position)
			continue
		}
		if vanish, ok, err := network.ParseActorVanish(pkt); err != nil {
			log.Printf("parse actor vanish 0x%04X: %v", pkt.ID, err)
		} else if ok {
			removeNetworkActor(ctx, vanish)
			continue
		}
		if look, ok, err := network.ParseActorLookChange(pkt); err != nil {
			log.Printf("parse actor look change 0x%04X: %v", pkt.ID, err)
		} else if ok {
			if applyActorLookChange(ctx, look) {
				if view, status := loadPlayerHumanoidSpriteView(ctx.Resources, selectedCharacter(ctx.Session), ctx.Session.Sex); view != nil {
					m.playerView = view
					log.Printf("player sprite changed type=%d value=%d %s", look.Type, look.Value, status)
				} else {
					m.playerView = nil
					log.Printf("player sprite reload failed after look change type=%d value=%d: %s", look.Type, look.Value, status)
				}
			}
			continue
		}
		if entry, ok, err := network.ParseActorEntry(pkt); err != nil {
			log.Printf("parse actor entry 0x%04X: %v", pkt.ID, err)
		} else if ok {
			upsertNetworkActor(ctx, entry)
		}
	}
	for _, err := range ctx.Network.DrainErrors() {
		log.Printf("network frame error: %v", err)
	}

	if m.tickCooldown > 0 {
		m.tickCooldown--
	}
	if m.tickCooldown == 0 {
		if err := ctx.Network.SendTick(uint32(time.Now().UnixMilli())); err == nil {
			m.tickCooldown = 300
		} else {
			m.tickCooldown = 60
		}
	}

	now := time.Now()
	m.camera.Update(ctx, now)

	dx, dy := 0, 0
	if ctx.Input.Pressed(ebiten.KeyArrowLeft) {
		dx--
	}
	if ctx.Input.Pressed(ebiten.KeyArrowRight) {
		dx++
	}
	if ctx.Input.Pressed(ebiten.KeyArrowUp) {
		dy--
	}
	if ctx.Input.Pressed(ebiten.KeyArrowDown) {
		dy++
	}
	if m.walkCooldown > 0 {
		m.walkCooldown--
	}
	if ctx.Input.MouseJustPressed(ebiten.MouseButtonLeft) && m.walkCooldown == 0 {
		projection := m.sceneProjection(ctx, ctx.Config.Window.Width, ctx.Config.Window.Height, now)
		if targetX, targetY, ok := clickedWalkTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY); ok {
			m.requestWalk(ctx, targetX, targetY, "click")
		}
	}
	if (dx != 0 || dy != 0) && m.walkCooldown == 0 {
		targetX := ctx.World.Player.X + dx
		targetY := ctx.World.Player.Y + dy
		m.requestWalk(ctx, targetX, targetY, "key")
	}
	return nil, nil
}

func (m *WorldMode) handleMapChange(ctx Context, change network.MapChange) Mode {
	log.Printf("map change map=%s x=%d y=%d server_move=%t addr=%s port=%d", change.MapName, change.X, change.Y, change.ServerMove, change.Address, change.Port)
	ctx.World.MapName = change.MapName
	ctx.Session.Zone.MapName = change.MapName
	applyWarpPosition(ctx, change.X, change.Y)
	ctx.World.Actors = make(map[uint32]worldstate.Actor)
	if change.ServerMove {
		ctx.Session.Zone.Address = change.Address
		ctx.Session.Zone.Port = change.Port
		dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := ctx.Network.Connect(dialCtx, change.Address, int(change.Port))
		cancel()
		if err != nil {
			m.status = "map reconnect failed: " + err.Error()
			log.Printf("map reconnect failed map=%s addr=%s port=%d: %v", change.MapName, change.Address, change.Port, err)
			return nil
		}
		if err := ctx.Network.SendMapServerEnter(ctx.Session.AccountID, ctx.Session.CharID, ctx.Session.AuthCode, uint32(time.Now().UnixMilli()), ctx.Session.Sex); err != nil {
			m.status = "map re-enter failed: " + err.Error()
			log.Printf("map re-enter failed map=%s addr=%s port=%d: %v", change.MapName, change.Address, change.Port, err)
			return nil
		}
		m.pendingWarp = true
		m.status = fmt.Sprintf("waiting for map enter: %s %s:%d", change.MapName, change.Address, change.Port)
		return nil
	}
	return NewWorldMode()
}

func (m *WorldMode) requestWalk(ctx Context, targetX, targetY int, source string) {
	if !walkTargetInBounds(ctx, targetX, targetY) {
		m.status = fmt.Sprintf("%s walk blocked by map bounds: %d,%d", source, targetX, targetY)
		m.walkCooldown = 12
		return
	}
	if err := ctx.Network.SendWalkToXY(targetX, targetY); err == nil {
		m.status = fmt.Sprintf("%s walk request: %d,%d", source, targetX, targetY)
		m.walkCooldown = 12
	} else {
		m.status = source + " walk request failed: " + err.Error()
		m.walkCooldown = 30
	}
}

func (m *WorldMode) Draw(ctx Context, screen *ebiten.Image) {
	clear(screen)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
	now := time.Now()
	playerX, playerY := ctx.World.Player.RenderPosition(now)
	projection := m.sceneProjection(ctx, width, height, now)

	if ctx.World.GND != nil {
		m.drawGND(screen, ctx.Resources, ctx.World.GND, projection)
		if ctx.World.RSW != nil && len(ctx.World.RSM) > 0 && m.rsmRender {
			m.drawRSMModels(screen, ctx.Resources, ctx.World.RSW, ctx.World.RSM, ctx.World.GND, projection)
		}
		if ctx.World.RSW != nil && m.rswMarkers {
			drawRSWModelMarkers(screen, ctx.World.RSW, ctx.World.GND, projection)
		}
		m.drawSceneActors(screen, ctx, projection)
	} else if ctx.World.GAT != nil {
		drawGAT(screen, ctx.World.GAT, ctx.World.Player.X, ctx.World.Player.Y)
		m.drawSceneActors(screen, ctx, projection)
	} else {
		const tile = 32
		for x := 0; x < width; x += tile {
			drawLine(screen, float64(x), 0, float64(x), float64(height), render.ColorGrid)
		}
		for y := 0; y < height; y += tile {
			drawLine(screen, 0, float64(y), float64(width), float64(y), render.ColorGrid)
		}
	}

	debugText(screen, 24, 24, "map: %s player=(%d,%d) dir=%d", ctx.World.MapName, ctx.World.Player.X, ctx.World.Player.Y, ctx.World.Dir)
	debugText(screen, 24, 44, "%s", m.status)
	if ctx.World.GND != nil {
		debugText(screen, 24, 64, "gnd: %dx%d textures=%d surfaces=%d", ctx.World.GND.Width, ctx.World.GND.Height, len(ctx.World.GND.Textures), len(ctx.World.GND.Surfaces))
		debugText(screen, 24, 84, "textures: loaded=%d missing=%d", len(m.textures), len(m.textureMiss))
	}
	if ctx.World.RSW != nil {
		debugText(screen, 24, 104, "rsw: v%d.%d models=%d lights=%d sounds=%d effects=%d water=%.1f", ctx.World.RSW.VersionMajor, ctx.World.RSW.VersionMinor, len(ctx.World.RSW.Models), len(ctx.World.RSW.Lights), len(ctx.World.RSW.Sounds), len(ctx.World.RSW.Effects), ctx.World.RSW.Water.Level)
		debugText(screen, 24, 124, "rsm: parsed=%d failed=%d limit=%d", len(ctx.World.RSM), ctx.World.RSMFail, rsmLoadLimit())
		if ctx.World.GND != nil {
			debugText(screen, 24, 144, "%s", nearestRSWModelDebug(ctx.World.RSW, ctx.World.GND, ctx.World.RSM, playerX, playerY))
		}
	}
	if ctx.World.GAT != nil {
		y := 104
		if ctx.World.RSW != nil {
			y = 164
		}
		debugText(screen, 24, y, "gat: %dx%d", ctx.World.GAT.Width, ctx.World.GAT.Height)
		debugText(screen, 24, y+20, "actors: %d", len(ctx.World.Actors))
	}
}

type followCamera struct {
	initialized bool
	x           float64
	y           float64
	z           float64
}

func (c *followCamera) Reset() {
	*c = followCamera{}
}

func (c *followCamera) Update(ctx Context, now time.Time) {
	targetX, targetY, targetZ := playerCameraTarget(ctx.World, now)
	if !c.initialized {
		c.x = targetX
		c.y = targetY
		c.z = targetZ
		c.initialized = true
		c.store(ctx)
		return
	}
	factor := cameraFollowFactor()
	c.x += (targetX - c.x) * factor
	c.y += (targetY - c.y) * factor
	c.z += (targetZ - c.z) * factor
	c.store(ctx)
}

func (c *followCamera) Projection(ctx Context, width, height int, now time.Time) sceneProjection {
	if !c.initialized {
		c.Update(ctx, now)
	}
	c.store(ctx)
	return newSceneProjectionForTarget(width, height, c.x, c.y, c.z)
}

func (c *followCamera) store(ctx Context) {
	if ctx.World == nil {
		return
	}
	ctx.World.Camera.X = c.x
	ctx.World.Camera.Y = c.y
}

func (m *WorldMode) sceneProjection(ctx Context, width, height int, now time.Time) sceneProjection {
	return m.camera.Projection(ctx, width, height, now)
}

func playerCameraTarget(world *worldstate.World, now time.Time) (float64, float64, float64) {
	if world == nil {
		return 0.5, 0.5, 0
	}
	playerX, playerY := world.Player.RenderPosition(now)
	return cellCenter(playerX), cellCenter(playerY), cameraTargetHeightAt(world, playerX, playerY)
}

func cameraFollowFactor() float64 {
	value := sceneFloatEnv("GORO_CAMERA_FOLLOW_FACTOR", 0.1)
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func upsertNetworkActor(ctx Context, entry network.ActorEntry) {
	if isLocalActor(ctx, entry.ID) {
		return
	}
	dir := entry.Dir
	if entry.Moving {
		dir = directionFromDelta(entry.FromX, entry.FromY, entry.ToX, entry.ToY, dir)
	}
	ctx.World.UpsertActor(worldstate.Actor{
		ID:         entry.ID,
		X:          entry.X,
		Y:          entry.Y,
		Dir:        dir,
		Job:        entry.Job,
		Head:       entry.Head,
		Weapon:     entry.Weapon,
		Shield:     entry.Shield,
		HeadTop:    entry.HeadTop,
		HeadMid:    entry.HeadMid,
		HeadLow:    entry.HeadLow,
		Sex:        entry.Sex,
		Appearance: entry.Appearance,
		Moving:     entry.Moving,
		FromX:      entry.FromX,
		FromY:      entry.FromY,
		ToX:        entry.ToX,
		ToY:        entry.ToY,
	})
}

func removeNetworkActor(ctx Context, vanish network.ActorVanish) {
	if isLocalActor(ctx, vanish.ID) {
		return
	}
	ctx.World.RemoveActor(vanish.ID)
}

func applyActorLookChange(ctx Context, look network.ActorLookChange) bool {
	if look.ID == 0 {
		return false
	}
	if isLocalActor(ctx, look.ID) {
		applyCharacterLookChange(ctx.Session, look)
		applyWorldActorLookChange(&ctx.World.Player, look)
		return true
	}
	actor, ok := ctx.World.Actors[look.ID]
	if !ok {
		actor = worldstate.Actor{ID: look.ID, Appearance: true}
	}
	applyWorldActorLookChange(&actor, look)
	ctx.World.UpsertActor(actor)
	return false
}

func applyCharacterLookChange(sessionState *session.Session, look network.ActorLookChange) {
	update := func(character *session.Character) {
		switch look.Type {
		case 0:
			character.Job = int16(look.Value)
		case 1:
			character.Hair = int16(look.Value)
		case 2:
			character.Weapon = int16(look.Value & 0xFFFF)
			character.Shield = int16((look.Value >> 16) & 0xFFFF)
		case 3:
			character.HeadLow = int16(look.Value)
		case 4:
			character.HeadTop = int16(look.Value)
		case 5:
			character.HeadMid = int16(look.Value)
		case 6:
			character.HeadPal = int16(look.Value)
			if look.Value <= 255 {
				character.HairColor = uint8(look.Value)
			}
		case 7:
			character.BodyPal = int16(look.Value)
		case 8:
			character.Shield = int16(look.Value)
		}
	}
	update(&sessionState.Selected)
	for index := range sessionState.Characters {
		if sessionState.Characters[index].ID == sessionState.CharID || sessionState.Characters[index].ID == sessionState.Selected.ID {
			update(&sessionState.Characters[index])
		}
	}
}

func applyWorldActorLookChange(actor *worldstate.Actor, look network.ActorLookChange) {
	actor.Appearance = true
	switch look.Type {
	case 0:
		actor.Job = int16(look.Value)
	case 1:
		actor.Head = int16(look.Value)
	case 2:
		actor.Weapon = int16(look.Value & 0xFFFF)
		actor.Shield = int16((look.Value >> 16) & 0xFFFF)
	case 3:
		actor.HeadLow = int16(look.Value)
	case 4:
		actor.HeadTop = int16(look.Value)
	case 5:
		actor.HeadMid = int16(look.Value)
	case 8:
		actor.Shield = int16(look.Value)
	}
}

func isLocalActor(ctx Context, id uint32) bool {
	return id != 0 && (id == ctx.Session.AccountID || id == ctx.Session.CharID)
}

func applySelfMoveAck(ctx Context, ack network.SelfMoveAck) {
	dir := directionFromDelta(ack.FromX, ack.FromY, ack.ToX, ack.ToY, ctx.World.Dir)
	ctx.World.SetPlayerMovement(ack.FromX, ack.FromY, ack.ToX, ack.ToY, dir)
	ctx.Session.PlayerX = ack.ToX
	ctx.Session.PlayerY = ack.ToY
}

func applyMapAcceptEnter(ctx Context, enter network.MapAcceptEnter) {
	ctx.Session.PlayerX = enter.X
	ctx.Session.PlayerY = enter.Y
	ctx.Session.PlayerDir = enter.Dir
	ctx.Session.Playing = true
	ctx.World.SetPlayerPosition(enter.X, enter.Y, enter.Dir)
}

func applyWarpPosition(ctx Context, x, y int) {
	dir := ctx.World.Dir
	if ctx.Session.PlayerDir != 0 {
		dir = ctx.Session.PlayerDir
	}
	ctx.Session.PlayerX = x
	ctx.Session.PlayerY = y
	ctx.World.SetPlayerPosition(x, y, dir)
}

func applyActorSetPosition(ctx Context, position network.ActorSetPosition) {
	if isLocalActor(ctx, position.ID) {
		ctx.World.SetPlayerPosition(position.X, position.Y, ctx.World.Dir)
		ctx.Session.PlayerX = position.X
		ctx.Session.PlayerY = position.Y
		return
	}
	ctx.World.UpsertActor(worldstate.Actor{
		ID: position.ID,
		X:  position.X,
		Y:  position.Y,
	})
}

func applyActorNameAck(ctx Context, ack network.ActorNameAck) {
	name := sanitizeActorName(ack.Name)
	if name == "" || ctx.World == nil {
		return
	}
	if isLocalActor(ctx, ack.ID) {
		ctx.World.Player.Name = name
		return
	}
	actor, ok := ctx.World.Actors[ack.ID]
	if !ok {
		return
	}
	actor.Name = name
	ctx.World.Actors[ack.ID] = actor
}

func walkTargetInBounds(ctx Context, x, y int) bool {
	if x < 0 || y < 0 {
		return false
	}
	if ctx.World.GAT != nil {
		return x < ctx.World.GAT.Width && y < ctx.World.GAT.Height
	}
	if ctx.World.GND != nil {
		return x < ctx.World.GND.Width && y < ctx.World.GND.Height
	}
	return x <= 1023 && y <= 1023
}

func directionFromDelta(fromX, fromY, toX, toY int, fallback int) int {
	dx := toX - fromX
	dy := toY - fromY
	if dx == 0 && dy == 0 {
		return normalizeDirectionIndex(fallback)
	}
	actionDir := int(math.Round(-math.Atan2(float64(dy), float64(dx))/(math.Pi/4)+6)) & 7
	return (4 - actionDir) & 7
}

func cameraTargetHeightAt(world *worldstate.World, x, y float64) float64 {
	if raw := os.Getenv("GORO_CAMERA_TARGET_Z"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err == nil {
			return value
		}
	}
	if os.Getenv("GORO_CAMERA_FOLLOW_TERRAIN_HEIGHT") == "0" {
		return 0
	}
	return terrainHeightAt(world, x, y)
}

func clickedWalkTarget(ctx Context, projection sceneProjection, mouseX, mouseY int) (int, int, bool) {
	radius := clickWalkSearchRadius()
	minX := maxInt(0, ctx.World.Player.X-radius)
	maxX := ctx.World.Player.X + radius
	minY := maxInt(0, ctx.World.Player.Y-radius)
	maxY := ctx.World.Player.Y + radius
	if ctx.World.GAT != nil {
		maxX = minInt(maxX, ctx.World.GAT.Width-1)
		maxY = minInt(maxY, ctx.World.GAT.Height-1)
	} else if ctx.World.GND != nil {
		maxX = minInt(maxX, ctx.World.GND.Width*2-1)
		maxY = minInt(maxY, ctx.World.GND.Height*2-1)
	}

	bestX, bestY := 0, 0
	bestDistance := math.Inf(1)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			point := projection.Project(cellCenter(float64(x)), cellCenter(float64(y)), terrainHeightAt(ctx.World, float64(x), float64(y)))
			dx := float64(point.x) - float64(mouseX)
			dy := float64(point.y) - float64(mouseY)
			distance := dx*dx + dy*dy
			if distance < bestDistance {
				bestDistance = distance
				bestX = x
				bestY = y
			}
		}
	}
	return bestX, bestY, bestDistance < math.Inf(1)
}

func clickWalkSearchRadius() int {
	raw := os.Getenv("GORO_CLICK_WALK_RADIUS")
	if raw == "" {
		return 70
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 70
	}
	return value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type sceneActorDrawEntry struct {
	actor    worldstate.Actor
	label    string
	screenX  float64
	screenY  float64
	scale    float64
	depth    float64
	isPlayer bool
}

const (
	actorBillboardCellWorldUnits  = 5.0
	actorBillboardWorldHeightUnit = 1.0 * actorBillboardCellWorldUnits
)

func (m *WorldMode) drawSceneActors(screen *ebiten.Image, ctx Context, projection sceneProjection) {
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
	now := time.Now()
	entries := make([]sceneActorDrawEntry, 0, len(ctx.World.Actors)+1)
	player := ctx.World.Player
	player.ID = ctx.Session.CharID
	character := selectedCharacter(ctx.Session)
	player.Job = character.Job
	player.Head = character.Hair
	player.Sex = ctx.Session.Sex
	if character.Name != "" {
		player.Name = character.Name
	}
	player.Dir = ctx.World.Dir
	entries = appendActorDrawEntry(entries, ctx.World, projection, player, actorDisplayName(ctx, player, true), true, now, width, height)
	for _, actor := range ctx.World.Actors {
		if actor.ID == ctx.Session.AccountID || actor.ID == ctx.Session.CharID {
			continue
		}
		entries = appendActorDrawEntry(entries, ctx.World, projection, actor, actorDisplayName(ctx, actor, false), false, now, width, height)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].depth > entries[j].depth
	})
	for _, entry := range entries {
		if entry.isPlayer {
			if m.drawPlayerSprite(ctx, screen, entry.screenX, entry.screenY, entry.scale) {
				continue
			}
			drawPanel(screen, entry.screenX-6, entry.screenY-6, 24, 24)
			continue
		}
		if m.drawActorSprite(screen, ctx, entry.actor, entry.screenX, entry.screenY, entry.scale) {
			continue
		}
		drawActorMarker(screen, entry.screenX-6, entry.screenY-20, entry.actor, now)
	}
	for _, entry := range entries {
		drawActorNameLabel(screen, entry.label, entry.screenX, entry.screenY, entry.scale)
	}
}

func appendActorDrawEntry(entries []sceneActorDrawEntry, world *worldstate.World, projection sceneProjection, actor worldstate.Actor, label string, isPlayer bool, now time.Time, screenWidth, screenHeight int) []sceneActorDrawEntry {
	actorX, actorY := actor.RenderPosition(now)
	terrainZ := terrainHeightAt(world, actorX, actorY)
	point := projection.Project(cellCenter(actorX), cellCenter(actorY), terrainZ)
	if point.x < -96 || point.y < -160 || point.x > float32(screenWidth+96) || point.y > float32(screenHeight+96) {
		return entries
	}
	depth := projection.Depth(cellCenter(actorX), cellCenter(actorY), terrainZ)
	return append(entries, sceneActorDrawEntry{
		actor:    actor,
		label:    label,
		screenX:  float64(point.x),
		screenY:  float64(point.y),
		scale:    actorBillboardScreenScale(projection, cellCenter(actorX), cellCenter(actorY), terrainZ),
		depth:    depth,
		isPlayer: isPlayer,
	})
}

func actorDisplayName(ctx Context, actor worldstate.Actor, isPlayer bool) string {
	if isPlayer {
		if name := sanitizeActorName(selectedCharacterName(ctx.Session)); name != "" {
			return name
		}
		return sanitizeActorName(actor.Name)
	}
	if name := sanitizeActorName(actor.Name); name != "" {
		return name
	}
	if res.HasPlayerJobToken(int(actor.Job)) || ctx.Resources == nil {
		return ""
	}
	if resourceName, ok := ctx.Resources.JobResourceName(int(actor.Job)); ok {
		return displayNameFromResource(resourceName)
	}
	return ""
}

func selectedCharacterName(s *session.Session) string {
	if s == nil {
		return ""
	}
	return selectedCharacter(s).Name
}

func sanitizeActorName(name string) string {
	name = strings.TrimSpace(name)
	if hash := strings.IndexByte(name, '#'); hash >= 0 {
		name = strings.TrimSpace(name[:hash])
	}
	if strings.EqualFold(name, "actor") {
		return ""
	}
	return name
}

func displayNameFromResource(name string) string {
	name = strings.TrimSpace(strings.TrimSuffix(name, ".spr"))
	name = strings.TrimSuffix(name, ".act")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ToLower(strings.TrimSpace(name))
	fields := strings.Fields(name)
	for i, field := range fields {
		fields[i] = titleASCIIWord(field)
	}
	return strings.Join(fields, " ")
}

func titleASCIIWord(word string) string {
	if word == "" {
		return ""
	}
	if word[0] < 'a' || word[0] > 'z' {
		return word
	}
	return strings.ToUpper(word[:1]) + word[1:]
}

func drawActorNameLabel(screen *ebiten.Image, label string, centerX, baseY, scale float64) {
	label = sanitizeActorName(label)
	if label == "" {
		return
	}
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	x := int(centerX) - len(label)*3
	y := int(baseY + 13*scale)
	debugText(screen, x+1, y+1, "%s", label)
	debugText(screen, x, y, "%s", label)
}

func (m *WorldMode) drawActorSprite(screen *ebiten.Image, ctx Context, actor worldstate.Actor, centerX, centerY, scale float64) bool {
	if !res.HasPlayerJobToken(int(actor.Job)) {
		return m.drawNonPCSprite(screen, ctx, actor, centerX, centerY, scale)
	}
	key := actorSpriteKey{
		job:     int(actor.Job),
		head:    int(actor.Head),
		sex:     actor.Sex,
		weapon:  int(actor.Weapon),
		shield:  int(actor.Shield),
		headTop: int(actor.HeadTop),
		headMid: int(actor.HeadMid),
		headLow: int(actor.HeadLow),
	}
	if _, ok := m.actorViewMiss[key]; ok {
		return false
	}
	view, ok := m.actorViews[key]
	if !ok {
		loaded, status := loadHumanoidSpriteViewWithAppearance(ctx.Resources, humanoidAppearance{
			job:     key.job,
			head:    key.head,
			sex:     key.sex,
			weapon:  key.weapon,
			shield:  key.shield,
			headTop: key.headTop,
			headMid: key.headMid,
			headLow: key.headLow,
		}, "actor")
		if loaded == nil {
			m.actorViewMiss[key] = struct{}{}
			log.Printf("actor sprite unavailable id=%d job=%d head=%d sex=%d: %s", actor.ID, key.job, key.head, key.sex, status)
			return false
		}
		m.actorViews[key] = loaded
		view = loaded
		log.Printf("actor sprite resources id=%d job=%d head=%d sex=%d %s", actor.ID, key.job, key.head, key.sex, status)
	}
	state := spriteState{
		actionFamily: spriteActionIdle,
		direction:    actor.Dir,
		moving:       actor.IsMovingAt(time.Now()),
	}
	if state.moving {
		state.actionFamily = spriteActionWalk
	}
	return drawHumanoidBillboard(screen, view, state, centerX, centerY, scale)
}

func (m *WorldMode) drawNonPCSprite(screen *ebiten.Image, ctx Context, actor worldstate.Actor, centerX, centerY, scale float64) bool {
	job := int(actor.Job)
	if _, ok := m.nonPCViewMiss[job]; ok {
		return false
	}
	view, ok := m.nonPCViews[job]
	if !ok {
		loaded, status := loadNonPCSpriteView(ctx.Resources, job, "nonpc")
		if loaded == nil {
			m.nonPCViewMiss[job] = struct{}{}
			log.Printf("nonpc sprite unavailable id=%d job=%d: %s", actor.ID, job, status)
			return false
		}
		m.nonPCViews[job] = loaded
		view = loaded
		log.Printf("nonpc sprite resources id=%d job=%d %s", actor.ID, job, status)
	}
	state := spriteState{
		actionFamily: spriteActionIdle,
		direction:    actor.Dir,
		moving:       actor.IsMovingAt(time.Now()),
	}
	if state.moving {
		state.actionFamily = spriteActionWalk
	}
	return drawSingleSpriteBillboard(screen, view, state, centerX, centerY, scale)
}

func actorBillboardScreenScale(projection sceneProjection, x, y, z float64) float64 {
	if !projection.camera {
		return 1
	}
	base := projection.Project(x, y, z)
	top := projection.Project(x, y, z+actorBillboardWorldHeightUnit)
	projectedHeight := math.Hypot(float64(top.x-base.x), float64(top.y-base.y))
	if projectedHeight <= 0 || math.IsNaN(projectedHeight) || math.IsInf(projectedHeight, 0) {
		return 1
	}
	return projectedHeight / float64(humanoidBillboardAnchorY)
}

func drawActorMarker(screen *ebiten.Image, x, y float64, actor worldstate.Actor, now time.Time) {
	col := color.RGBA{R: 82, G: 166, B: 255, A: 230}
	if actor.Job >= 1000 {
		col = color.RGBA{R: 229, G: 102, B: 72, A: 230}
	}
	if actor.IsMovingAt(now) {
		col = color.RGBA{R: 235, G: 190, B: 80, A: 230}
	}
	ebitenutil.DrawRect(screen, x, y, 12, 18, col)
	ebitenutil.DrawRect(screen, x+3, y-4, 6, 6, col)
	debugText(screen, int(x-12), int(y-16), "%d", actor.Job)
}

func loadGAT(manager *res.Manager, mapName string) (*res.GAT, string, error) {
	base := strings.TrimSuffix(strings.TrimSuffix(mapName, ".gat"), ".rsw")
	candidates := []string{
		"data\\" + base + ".gat",
		"data/" + base + ".gat",
		base + ".gat",
	}
	for _, candidate := range candidates {
		data, err := manager.ReadFile(candidate)
		if err != nil {
			continue
		}
		gat, err := res.ParseGAT(data)
		if err != nil {
			return nil, candidate, err
		}
		return gat, candidate, nil
	}
	return nil, "", fmt.Errorf("gat not found for map %s", mapName)
}

func loadGND(manager *res.Manager, mapName string) (*res.GND, string, error) {
	base := strings.TrimSuffix(strings.TrimSuffix(mapName, ".gat"), ".rsw")
	candidates := []string{
		"data\\" + base + ".gnd",
		"data/" + base + ".gnd",
		base + ".gnd",
	}
	for _, candidate := range candidates {
		data, err := manager.ReadFile(candidate)
		if err != nil {
			continue
		}
		gnd, err := res.ParseGND(data)
		if err != nil {
			return nil, candidate, err
		}
		return gnd, candidate, nil
	}
	return nil, "", fmt.Errorf("gnd not found for map %s", mapName)
}

func loadRSW(manager *res.Manager, mapName string) (*res.RSW, string, error) {
	base := strings.TrimSuffix(strings.TrimSuffix(mapName, ".gat"), ".rsw")
	candidates := []string{
		"data\\" + base + ".rsw",
		"data/" + base + ".rsw",
		base + ".rsw",
	}
	for _, candidate := range candidates {
		data, err := manager.ReadFile(candidate)
		if err != nil {
			continue
		}
		rsw, err := res.ParseRSW(data)
		if err != nil {
			return nil, candidate, err
		}
		return rsw, candidate, nil
	}
	return nil, "", fmt.Errorf("rsw not found for map %s", mapName)
}

func rsmLoadLimit() int {
	raw := os.Getenv("GORO_RSM_LOAD_LIMIT")
	if raw == "" {
		return 128
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 128
	}
	return limit
}

func loadRSMModels(manager *res.Manager, rsw *res.RSW, limit int) (map[string]*res.RSM, int) {
	if rsw == nil || limit == 0 {
		return nil, 0
	}
	models := make(map[string]*res.RSM)
	failures := 0
	for _, placement := range rsw.Models {
		if placement.Filename == "" {
			continue
		}
		if _, ok := models[placement.Filename]; ok {
			continue
		}
		if limit > 0 && len(models) >= limit {
			break
		}

		rsm, err := loadRSMModel(manager, placement.Filename)
		if err != nil {
			failures++
			continue
		}
		models[placement.Filename] = rsm
	}
	return models, failures
}

func loadRSMModel(manager *res.Manager, filename string) (*res.RSM, error) {
	var data []byte
	for _, candidate := range res.RSMModelCandidates(filename) {
		raw, err := manager.ReadFile(candidate)
		if err == nil {
			data = raw
			break
		}
	}
	if data == nil {
		return nil, fmt.Errorf("rsm not found: %s", filename)
	}
	return res.ParseRSM(data)
}

func (m *WorldMode) drawGND(screen *ebiten.Image, manager *res.Manager, gnd *res.GND, projection sceneProjection) {
	if m.whitePixel == nil {
		m.whitePixel = ebiten.NewImage(1, 1)
		m.whitePixel.Fill(color.White)
	}

	width := screen.Bounds().Dx()
	height := screen.Bounds().Dy()
	groundCenterX := int(math.Floor(projection.playerX * 0.5))
	groundCenterY := int(math.Floor(projection.playerY * 0.5))

	radiusX := int(float64(width)/projection.tileW) + 12
	radiusY := int(float64(height)/projection.tileH) + 12
	startX := max(0, groundCenterX-radiusX)
	endX := min(gnd.Width-1, groundCenterX+radiusX)
	startY := max(0, groundCenterY-radiusY)
	endY := min(gnd.Height-1, groundCenterY+radiusY)

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			cell, ok := gnd.Cell(x, y)
			if !ok {
				continue
			}

			if cell.Top >= 0 {
				surface, ok := gnd.Surface(cell.Top)
				if ok {
					vertexOrder := [4]int{0, 1, 3, 2}
					points := [4]screenPoint{
						projection.Project(float64(x)*2, float64(y)*2, float64(cell.Heights[0])),
						projection.Project(float64(x+1)*2, float64(y)*2, float64(cell.Heights[1])),
						projection.Project(float64(x+1)*2, float64(y+1)*2, float64(cell.Heights[3])),
						projection.Project(float64(x)*2, float64(y+1)*2, float64(cell.Heights[2])),
					}
					m.drawGNDSurface(screen, manager, gnd, points, surfaceUVs(surface, vertexOrder), vertexOrder, []uint16{0, 1, 2, 0, 2, 3}, surface, cell.Heights, 1.0, float64(width), float64(height))
				}
			}

			if cell.Front >= 0 && y+1 < gnd.Height {
				neighbor, neighborOK := gnd.Cell(x, y+1)
				surface, surfaceOK := gnd.Surface(cell.Front)
				if neighborOK && surfaceOK {
					vertexOrder := [4]int{2, 1, 3, 0}
					points := [4]screenPoint{
						projection.Project(float64(x)*2, float64(y+1)*2, float64(neighbor.Heights[0])),
						projection.Project(float64(x+1)*2, float64(y+1)*2, float64(cell.Heights[3])),
						projection.Project(float64(x+1)*2, float64(y+1)*2, float64(neighbor.Heights[1])),
						projection.Project(float64(x)*2, float64(y+1)*2, float64(cell.Heights[2])),
					}
					heights := [4]float32{neighbor.Heights[0], cell.Heights[3], neighbor.Heights[1], cell.Heights[2]}
					m.drawGNDSurface(screen, manager, gnd, points, surfaceUVs(surface, vertexOrder), vertexOrder, []uint16{0, 1, 2, 0, 1, 3}, surface, heights, 0.82, float64(width), float64(height))
				}
			}

			if cell.Right >= 0 && x+1 < gnd.Width {
				neighbor, neighborOK := gnd.Cell(x+1, y)
				surface, surfaceOK := gnd.Surface(cell.Right)
				if neighborOK && surfaceOK {
					vertexOrder := [4]int{1, 0, 3, 2}
					points := [4]screenPoint{
						projection.Project(float64(x+1)*2, float64(y)*2, float64(cell.Heights[1])),
						projection.Project(float64(x+1)*2, float64(y+1)*2, float64(cell.Heights[3])),
						projection.Project(float64(x+1)*2, float64(y)*2, float64(neighbor.Heights[0])),
						projection.Project(float64(x+1)*2, float64(y+1)*2, float64(neighbor.Heights[2])),
					}
					heights := [4]float32{cell.Heights[1], cell.Heights[3], neighbor.Heights[0], neighbor.Heights[2]}
					m.drawGNDSurface(screen, manager, gnd, points, surfaceUVs(surface, vertexOrder), vertexOrder, []uint16{0, 1, 2, 2, 3, 1}, surface, heights, 0.9, float64(width), float64(height))
				}
			}
		}
	}
}

func (m *WorldMode) drawGNDSurface(screen *ebiten.Image, manager *res.Manager, gnd *res.GND, points [4]screenPoint, uvs [4]texturePoint, vertexOrder [4]int, indices []uint16, surface res.GNDSurface, heights [4]float32, brightness float64, screenWidth, screenHeight float64) {
	if quadOutside(points, screenWidth, screenHeight) {
		return
	}

	textureName := gndTextureName(gnd, surface.TextureID)
	if texture := m.groundTexture(manager, textureName); texture != nil {
		drawTexturedSurface(screen, texture, points, uvs, indices, surfaceVertexTints(gnd, surface, vertexOrder, heights, brightness))
		return
	}
	drawColoredSurface(screen, m.whitePixel, points, indices, groundSurfaceColor(textureName, surface.Color, heights, brightness))
}

func drawRSWModelMarkers(screen *ebiten.Image, rsw *res.RSW, gnd *res.GND, projection sceneProjection) {
	width := screen.Bounds().Dx()
	height := screen.Bounds().Dy()

	markerColor := color.RGBA{R: 238, G: 181, B: 73, A: 190}
	for _, model := range rsw.Models {
		x := float64(model.Position.X) + float64(gnd.Width)
		y := float64(model.Position.Z) + float64(gnd.Height)
		if math.Abs(x-projection.playerX) > float64(width)/projection.tileW*2+16 || math.Abs(y-projection.playerY) > float64(height)/projection.tileH*2+16 {
			continue
		}

		point := projection.Project(x, y, float64(model.Position.Y))
		if point.x < -8 || point.y < -8 || point.x > float32(width+8) || point.y > float32(height+8) {
			continue
		}
		ebitenutil.DrawRect(screen, float64(point.x)-2, float64(point.y)-6, 4, 8, markerColor)
	}
}

func nearestRSWModelDebug(rsw *res.RSW, gnd *res.GND, models map[string]*res.RSM, playerX, playerY float64) string {
	if rsw == nil || gnd == nil || len(rsw.Models) == 0 {
		return "rsw nearest: none"
	}
	bestIndex := -1
	bestDistance := math.Inf(1)
	bestX, bestY, bestZ := 0.0, 0.0, 0.0
	bestName := ""
	for i, model := range rsw.Models {
		x := float64(model.Position.X) + float64(gnd.Width)
		y := float64(model.Position.Z) + float64(gnd.Height)
		dx := x - cellCenter(playerX)
		dy := y - cellCenter(playerY)
		distance := math.Sqrt(dx*dx + dy*dy)
		if distance < bestDistance {
			bestIndex = i
			bestDistance = distance
			bestX = x
			bestY = y
			bestZ = float64(model.Position.Y)
			bestName = model.Filename
		}
	}
	if len(bestName) > 34 {
		bestName = bestName[:34]
	}
	modelState := "missing"
	if models != nil {
		if model, ok := models[rsw.Models[bestIndex].Filename]; ok {
			if model != nil {
				modelState = fmt.Sprintf("loaded faces=%d", rsmFaceCount(model))
			} else {
				modelState = "failed"
			}
		}
	}
	return fmt.Sprintf("rsw nearest: #%d d=%.1f pos=(%.1f,%.1f,%.1f) %s %s", bestIndex, bestDistance, bestX, bestY, bestZ, modelState, bestName)
}

func rsmFaceCount(model *res.RSM) int {
	faces := 0
	for _, node := range model.Nodes {
		faces += len(node.Faces)
	}
	return faces
}

func (m *WorldMode) groundTexture(manager *res.Manager, name string) *ebiten.Image {
	if name == "" {
		return nil
	}
	if texture, ok := m.textures[name]; ok {
		return texture
	}
	if _, ok := m.textureMiss[name]; ok {
		return nil
	}

	img, _, err := res.LoadImage(manager, res.GroundTextureCandidates(name))
	if err != nil {
		m.textureMiss[name] = struct{}{}
		return nil
	}
	texture := ebiten.NewImageFromImage(img)
	m.textures[name] = texture
	return texture
}

type screenPoint struct {
	x float32
	y float32
}

type texturePoint struct {
	u float32
	v float32
}

func quadOutside(points [4]screenPoint, width, height float64) bool {
	minX, minY := float64(points[0].x), float64(points[0].y)
	maxX, maxY := minX, minY
	for _, point := range points[1:] {
		minX = math.Min(minX, float64(point.x))
		minY = math.Min(minY, float64(point.y))
		maxX = math.Max(maxX, float64(point.x))
		maxY = math.Max(maxY, float64(point.y))
	}
	return maxX < -32 || maxY < -32 || minX > width+32 || minY > height+32
}

func gndTextureName(gnd *res.GND, textureID int) string {
	if gnd == nil || textureID < 0 || textureID >= len(gnd.Textures) {
		return ""
	}
	return gnd.Textures[textureID]
}

func surfaceUVs(surface res.GNDSurface, order [4]int) [4]texturePoint {
	return [4]texturePoint{
		{u: surface.U[order[0]], v: surface.V[order[0]]},
		{u: surface.U[order[1]], v: surface.V[order[1]]},
		{u: surface.U[order[2]], v: surface.V[order[2]]},
		{u: surface.U[order[3]], v: surface.V[order[3]]},
	}
}

func surfaceTint(surfaceColor color.RGBA, heights [4]float32, brightness float64) color.RGBA {
	base := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if surfaceColor.A != 0 {
		base = surfaceColor
	}
	shade := groundHeightShade(heights) * brightness
	return color.RGBA{
		R: clampColor(float64(base.R) * shade),
		G: clampColor(float64(base.G) * shade),
		B: clampColor(float64(base.B) * shade),
		A: 255,
	}
}

func surfaceVertexTints(gnd *res.GND, surface res.GNDSurface, vertexOrder [4]int, heights [4]float32, brightness float64) [4]color.RGBA {
	if lightmap, ok := gnd.Lightmap(surface.LightmapID); ok {
		return lightmapSurfaceVertexTints(surface.Color, lightmap, vertexOrder, brightness)
	}
	tint := surfaceTint(surface.Color, heights, brightness)
	return [4]color.RGBA{tint, tint, tint, tint}
}

func lightmapSurfaceVertexTints(surfaceColor color.RGBA, lightmap res.GNDLightmap, vertexOrder [4]int, brightness float64) [4]color.RGBA {
	base := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if surfaceColor.A != 0 {
		base = surfaceColor
	}
	var tints [4]color.RGBA
	for i, sourceVertex := range vertexOrder {
		light := float64(lightmap.Alpha[sourceVertex]) / 255 * brightness
		tints[i] = color.RGBA{
			R: clampColor(float64(base.R) * light),
			G: clampColor(float64(base.G) * light),
			B: clampColor(float64(base.B) * light),
			A: 255,
		}
	}
	return tints
}

func groundSurfaceColor(textureName string, surfaceColor color.RGBA, heights [4]float32, brightness float64) color.RGBA {
	base := textureColor(textureName)
	if surfaceColor.A != 0 && !(surfaceColor.R == 255 && surfaceColor.G == 255 && surfaceColor.B == 255) {
		base.R = uint8((uint16(base.R)*2 + uint16(surfaceColor.R)) / 3)
		base.G = uint8((uint16(base.G)*2 + uint16(surfaceColor.G)) / 3)
		base.B = uint8((uint16(base.B)*2 + uint16(surfaceColor.B)) / 3)
	}

	shade := groundHeightShade(heights) * brightness
	return color.RGBA{
		R: clampColor(float64(base.R) * shade),
		G: clampColor(float64(base.G) * shade),
		B: clampColor(float64(base.B) * shade),
		A: 255,
	}
}

func groundHeightShade(heights [4]float32) float64 {
	avgHeight := float32(0)
	for _, h := range heights {
		avgHeight += h
	}
	avgHeight /= 4
	return 0.88 + math.Sin(float64(avgHeight)*0.08)*0.06
}

func textureColor(name string) color.RGBA {
	if name == "" {
		return color.RGBA{R: 78, G: 86, B: 78, A: 255}
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(strings.ToLower(name)))
	value := hash.Sum32()
	return color.RGBA{
		R: 70 + uint8(value&0x3f),
		G: 80 + uint8((value>>8)&0x4f),
		B: 68 + uint8((value>>16)&0x4f),
		A: 255,
	}
}

func clampColor(value float64) uint8 {
	return uint8(min(255, max(0, int(value))))
}

func drawColoredSurface(screen, white *ebiten.Image, points [4]screenPoint, indices []uint16, c color.RGBA) {
	r := float32(c.R) / 255
	g := float32(c.G) / 255
	b := float32(c.B) / 255
	a := float32(c.A) / 255
	vertices := []ebiten.Vertex{
		{DstX: points[0].x, DstY: points[0].y, SrcX: 0, SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: points[1].x, DstY: points[1].y, SrcX: 1, SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: points[2].x, DstY: points[2].y, SrcX: 1, SrcY: 1, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: points[3].x, DstY: points[3].y, SrcX: 0, SrcY: 1, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
	}
	screen.DrawTriangles(vertices, indices, white, nil)
}

func drawTexturedSurface(screen, texture *ebiten.Image, points [4]screenPoint, uvs [4]texturePoint, indices []uint16, tints [4]color.RGBA) {
	bounds := texture.Bounds()
	w := float32(bounds.Dx())
	h := float32(bounds.Dy())
	vertices := []ebiten.Vertex{
		texturedSurfaceVertex(points[0], uvs[0], tints[0], w, h),
		texturedSurfaceVertex(points[1], uvs[1], tints[1], w, h),
		texturedSurfaceVertex(points[2], uvs[2], tints[2], w, h),
		texturedSurfaceVertex(points[3], uvs[3], tints[3], w, h),
	}
	op := &ebiten.DrawTrianglesOptions{
		Filter:  ebiten.FilterLinear,
		Address: ebiten.AddressRepeat,
	}
	screen.DrawTriangles(vertices, indices, texture, op)
}

func texturedSurfaceVertex(point screenPoint, uv texturePoint, tint color.RGBA, textureWidth, textureHeight float32) ebiten.Vertex {
	return ebiten.Vertex{
		DstX:   point.x,
		DstY:   point.y,
		SrcX:   uv.u * textureWidth,
		SrcY:   uv.v * textureHeight,
		ColorR: float32(tint.R) / 255,
		ColorG: float32(tint.G) / 255,
		ColorB: float32(tint.B) / 255,
		ColorA: float32(tint.A) / 255,
	}
}

func drawGAT(screen *ebiten.Image, gat *res.GAT, playerX, playerY int) {
	const tile = 10
	width := screen.Bounds().Dx()
	height := screen.Bounds().Dy()
	tilesX := width/tile + 2
	tilesY := height/tile + 2
	startX := playerX - tilesX/2
	startY := playerY - tilesY/2

	for sy := 0; sy < tilesY; sy++ {
		mapY := startY + sy
		for sx := 0; sx < tilesX; sx++ {
			mapX := startX + sx
			cell, ok := gat.Cell(mapX, mapY)
			c := color.RGBA{R: 22, G: 25, B: 32, A: 255}
			if ok {
				switch {
				case cell.Type&res.GATTypeWater != 0:
					c = color.RGBA{R: 38, G: 84, B: 112, A: 255}
				case cell.Type&res.GATTypeWalkable != 0:
					c = color.RGBA{R: 54, G: 75, B: 54, A: 255}
				case cell.Type&res.GATTypeSnipable != 0:
					c = color.RGBA{R: 87, G: 77, B: 42, A: 255}
				default:
					c = color.RGBA{R: 54, G: 45, B: 48, A: 255}
				}
			}
			ebitenutil.DrawRect(screen, float64(sx*tile), float64(sy*tile), tile-1, tile-1, c)
		}
	}
}
