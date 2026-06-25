package gamemode

import (
	"context"
	"fmt"
	"hash/fnv"
	"image"
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
	status           string
	walkCooldown     int
	tickCooldown     int
	camera           followCamera
	whitePixel       *ebiten.Image
	tileCursor       *ebiten.Image
	textures         map[string]*ebiten.Image
	textureMiss      map[string]struct{}
	rswMarkers       bool
	rsmRender        bool
	playerView       *humanoidSpriteView
	shadowView       *playerSpriteView
	shadowViewMiss   bool
	cursorView       *playerSpriteView
	cursorViewMiss   bool
	cursorFallback   *ebiten.Image
	cursorAction     int
	cursorStarted    time.Time
	itemMarker       *ebiten.Image
	itemViews        map[itemSpriteKey]*playerSpriteView
	itemViewMiss     map[itemSpriteKey]struct{}
	actorViews       map[actorSpriteKey]*humanoidSpriteView
	actorViewMiss    map[actorSpriteKey]struct{}
	nonPCViews       map[int]*playerSpriteView
	nonPCViewMiss    map[int]struct{}
	rsmDebugLog      map[string]struct{}
	pendingWarp      bool
	pendingAttack    attackIntent
	pendingPickup    pickupIntent
	pickupReqItemID  uint32
	lockedAttackID   uint32
	lastAttackAt     time.Time
	lastChaseAt      time.Time
	actorAnims       map[uint32]actorAnimation
	damageFloaters   []damageFloater
	scheduledSounds  []scheduledSound
	actorDeaths      map[uint32]time.Time
	actorSoundFrames map[uint32]actorSoundFrame
	actorLife        map[uint32]actorLife
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

type attackIntent struct {
	targetID uint32
	expires  time.Time
	readyAt  time.Time
}

type pickupIntent struct {
	itemID  uint32
	expires time.Time
	readyAt time.Time
}

type damageFloater struct {
	actorID uint32
	x       int
	y       int
	text    string
	starts  time.Time
	expires time.Time
}

type scheduledSound struct {
	at    time.Time
	paths []string
}

type actorSoundFrame struct {
	actionFamily int
	motion       int
	soundIndex   int
}

type actorAnimation struct {
	actionFamily int
	started      time.Time
	duration     time.Duration
}

type actorLife struct {
	hp        int
	maxHP     int
	fromTiny  bool
	updatedAt time.Time
}

const attackRetryInterval = 1200 * time.Millisecond

const (
	defaultAttackAnimationDuration = 600 * time.Millisecond
	defaultHitAnimationDuration    = 250 * time.Millisecond
	defaultDeathAnimationDuration  = 900 * time.Millisecond
	maxCombatAnimationDuration     = 5 * time.Second
	nonPCDeathFadeDuration         = 5 * time.Second
)

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
	m.shadowView = nil
	m.shadowViewMiss = false
	m.cursorView = nil
	m.cursorViewMiss = false
	m.cursorFallback = nil
	m.cursorAction = cursorActionDefault
	m.cursorStarted = time.Now()
	m.itemMarker = nil
	m.itemViews = make(map[itemSpriteKey]*playerSpriteView)
	m.itemViewMiss = make(map[itemSpriteKey]struct{})
	m.actorViews = make(map[actorSpriteKey]*humanoidSpriteView)
	m.actorViewMiss = make(map[actorSpriteKey]struct{})
	m.nonPCViews = make(map[int]*playerSpriteView)
	m.nonPCViewMiss = make(map[int]struct{})
	m.rsmDebugLog = make(map[string]struct{})
	m.pendingWarp = false
	m.pendingAttack = attackIntent{}
	m.pendingPickup = pickupIntent{}
	m.pickupReqItemID = 0
	m.lockedAttackID = 0
	m.lastAttackAt = time.Time{}
	m.lastChaseAt = time.Time{}
	m.actorAnims = make(map[uint32]actorAnimation)
	m.damageFloaters = nil
	m.scheduledSounds = nil
	m.actorDeaths = make(map[uint32]time.Time)
	m.actorSoundFrames = make(map[uint32]actorSoundFrame)
	m.actorLife = make(map[uint32]actorLife)
	ctx.World.Items = make(map[uint32]worldstate.FloorItem)
	playerStatus := ""
	character := selectedCharacter(ctx.Session)
	if view, status := loadPlayerHumanoidSpriteView(ctx.Resources, character, ctx.Session.Sex); view != nil {
		m.playerView = view
		playerStatus = status
	} else {
		playerStatus = status
	}
	if view, status := loadActorShadowSpriteView(ctx.Resources); view != nil {
		m.shadowView = view
		if status != "" {
			playerStatus += " " + status
		}
	} else {
		m.shadowViewMiss = true
		log.Printf("actor shadow resources unavailable: %s", status)
	}
	if view, status := loadCursorSpriteView(ctx.Resources); view != nil {
		m.cursorView = view
		if status != "" {
			playerStatus += " " + status
		}
	} else {
		m.cursorViewMiss = true
		log.Printf("cursor resources unavailable: %s", status)
	}
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
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
		m.playMapBGM(ctx, rswSource)
	} else {
		ctx.World.RSW = nil
		ctx.World.RSM = nil
		ctx.World.RSMFail = 0
		m.status += " rsw: " + err.Error()
		m.playMapBGM(ctx, ctx.World.MapName)
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

func (m *WorldMode) playMapBGM(ctx Context, rswName string) {
	if ctx.Audio == nil {
		return
	}
	path, err := ctx.Audio.PlayMap(rswName)
	if err != nil {
		m.status += " bgm: " + err.Error()
		log.Printf("bgm failed map=%s: %v", rswName, err)
		return
	}
	if path != "" {
		m.status += " bgm=" + path
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
			log.Printf("walk ack from=%d,%d to=%d,%d tick=%d", ack.FromX, ack.FromY, ack.ToX, ack.ToY, ack.ServerTick)
			m.continuePendingAttack(ctx, "walk ack")
			m.continuePendingPickup(ctx, "walk ack")
			continue
		}
		if position, ok, err := network.ParseActorSetPosition(pkt); err != nil {
			log.Printf("parse actor set position 0x%04X: %v", pkt.ID, err)
		} else if ok {
			if isLocalActor(ctx, position.ID) {
				m.status = fmt.Sprintf("position fix: %d,%d", position.X, position.Y)
				log.Printf("local position fix id=%d x=%d y=%d", position.ID, position.X, position.Y)
			}
			applyActorSetPosition(ctx, position)
			if isLocalActor(ctx, position.ID) {
				m.continuePendingAttack(ctx, "position fix")
				m.continuePendingPickup(ctx, "position fix")
			}
			continue
		}
		if item, ok, err := network.ParseFloorItemEntry(pkt); err != nil {
			log.Printf("parse floor item entry 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyFloorItemEntry(ctx, item)
			continue
		}
		if disappear, ok, err := network.ParseFloorItemDisappear(pkt); err != nil {
			log.Printf("parse floor item disappear 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyFloorItemDisappear(ctx, disappear)
			continue
		}
		if pickup, ok, err := network.ParseItemPickupAck(pkt); err != nil {
			log.Printf("parse item pickup ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyItemPickupAck(ctx, pickup)
			continue
		}
		if vanish, ok, err := network.ParseActorVanish(pkt); err != nil {
			log.Printf("parse actor vanish 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyActorVanish(ctx, vanish)
			if m.pendingAttack.targetID == vanish.ID {
				m.pendingAttack = attackIntent{}
			}
			if m.lockedAttackID == vanish.ID {
				m.clearLockedAttack()
			}
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
		if action, ok, err := network.ParseActorActionNotify(pkt); err != nil {
			log.Printf("parse actor action 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyActorActionNotify(ctx, action)
			continue
		}
		if life, ok, err := network.ParseActorHPUpdate(pkt); err != nil {
			log.Printf("parse actor hp 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyActorHPUpdate(life)
			continue
		}
		if failure, ok, err := network.ParseAttackFailureForDistance(pkt); err != nil {
			log.Printf("parse attack distance failure 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyAttackFailureForDistance(ctx, failure)
			continue
		}
		if change, ok, err := network.ParseParameterChange(pkt); err != nil {
			log.Printf("parse parameter change 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyParameterChange(ctx, change)
			continue
		}
		if entry, ok, err := network.ParseActorEntry(pkt); err != nil {
			log.Printf("parse actor entry 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.clearActorDeath(entry.ID)
			upsertNetworkActor(ctx, entry)
		}
	}
	for _, err := range ctx.Network.DrainErrors() {
		log.Printf("network frame error: %v", err)
	}

	m.processPendingAttack(ctx)
	m.processPendingPickup(ctx)
	m.processLockedAttack(ctx)
	now := time.Now()
	m.cleanupDeadActors(ctx, now)
	m.processActorMotionSounds(ctx, now)
	m.playDueScheduledSounds(ctx, now)

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

	m.camera.Update(ctx, now)
	m.updateCameraRotation(ctx)

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
		screenW, screenH := ctx.ScreenSize()
		projection := m.sceneProjection(ctx, screenW, screenH, now)
		if item, ok := clickedGroundItem(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY, now); ok {
			log.Printf("click pickup target mouse=%d,%d id=%d item_id=%d amount=%d player=%d,%d target=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY, item.ID, item.ItemID, item.Amount, ctx.World.Player.X, ctx.World.Player.Y, item.X, item.Y)
			m.clearLockedAttack()
			m.requestPickup(ctx, item, "click")
			return nil, nil
		}
		if actor, ok := clickedAttackTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY, now, m.actorDeaths); ok {
			log.Printf("click attack target mouse=%d,%d id=%d name=%q job=%d object_type=%d player=%d,%d target=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY, actor.ID, actor.Name, actor.Job, actor.ObjectType, ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y)
			m.requestAttack(ctx, actor, "click")
			return nil, nil
		}
		if targetX, targetY, ok := clickedWalkTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY); ok {
			log.Printf("click walk target mouse=%d,%d player=%d,%d target=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY, ctx.World.Player.X, ctx.World.Player.Y, targetX, targetY)
			m.clearLockedAttack()
			m.requestWalk(ctx, targetX, targetY, "click")
		}
	}
	if (dx != 0 || dy != 0) && m.walkCooldown == 0 {
		targetX := ctx.World.Player.X + dx
		targetY := ctx.World.Player.Y + dy
		m.clearLockedAttack()
		m.requestWalk(ctx, targetX, targetY, "key")
	}
	return nil, nil
}

func (m *WorldMode) handleMapChange(ctx Context, change network.MapChange) Mode {
	m.pendingAttack = attackIntent{}
	m.clearLockedAttack()
	currentMap := ctx.World.MapName
	reuseLoadedMap := !change.ServerMove && sameLoadedMap(ctx, change.MapName)
	log.Printf("map change current=%s target=%s x=%d y=%d server_move=%t addr=%s port=%d reuse_loaded=%t", currentMap, change.MapName, change.X, change.Y, change.ServerMove, change.Address, change.Port, reuseLoadedMap)
	ctx.World.MapName = change.MapName
	ctx.Session.Zone.MapName = change.MapName
	applyWarpPosition(ctx, change.X, change.Y)
	ctx.World.Actors = make(map[uint32]worldstate.Actor)
	if reuseLoadedMap {
		m.camera.Reset()
		m.camera.Update(ctx, time.Now())
		if ctx.Network != nil {
			if err := ctx.Network.SendLoadEndAck(); err != nil {
				m.status = "same-map warp load-ack failed: " + err.Error()
				log.Printf("same-map warp load ack failed map=%s x=%d y=%d: %v", change.MapName, change.X, change.Y, err)
			} else {
				m.tickCooldown = 1
				m.status = fmt.Sprintf("warped on %s at %d,%d", change.MapName, change.X, change.Y)
			}
		} else {
			m.status = fmt.Sprintf("warped on %s at %d,%d", change.MapName, change.X, change.Y)
		}
		return nil
	}
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

func sameLoadedMap(ctx Context, mapName string) bool {
	if ctx.World == nil || ctx.World.MapName == "" || mapName == "" {
		return false
	}
	if !strings.EqualFold(ctx.World.MapName, mapName) {
		return false
	}
	return ctx.World.GND != nil || ctx.World.GAT != nil
}

func (m *WorldMode) requestWalk(ctx Context, targetX, targetY int, source string) {
	if !walkTargetInBounds(ctx, targetX, targetY) {
		m.status = fmt.Sprintf("%s walk blocked by map bounds: %d,%d", source, targetX, targetY)
		m.walkCooldown = 12
		return
	}
	log.Printf("%s walk request from=%d,%d to=%d,%d", source, ctx.World.Player.X, ctx.World.Player.Y, targetX, targetY)
	if err := ctx.Network.SendWalkToXY(targetX, targetY); err == nil {
		m.status = fmt.Sprintf("%s walk request: %d,%d", source, targetX, targetY)
		m.walkCooldown = 12
	} else {
		m.status = source + " walk request failed: " + err.Error()
		log.Printf("%s walk request failed from=%d,%d to=%d,%d: %v", source, ctx.World.Player.X, ctx.World.Player.Y, targetX, targetY, err)
		m.walkCooldown = 30
	}
}

func (m *WorldMode) requestAttack(ctx Context, actor worldstate.Actor, source string) {
	if ctx.Network == nil {
		m.status = "attack request failed: not connected"
		m.walkCooldown = 30
		return
	}
	m.lockAttack(actor.ID)
	if attackTargetWithinRange(ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y) {
		m.sendAttackAction(ctx, actor, source)
		return
	}
	targetX, targetY, ok := attackApproachCell(ctx, actor)
	if !ok {
		m.status = fmt.Sprintf("%s attack chase blocked: %d", source, actor.ID)
		log.Printf("%s attack chase blocked target=%d player=%d,%d target=%d,%d", source, actor.ID, ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y)
		m.walkCooldown = 12
		return
	}
	m.pendingAttack = attackIntent{
		targetID: actor.ID,
		expires:  time.Now().Add(8 * time.Second),
	}
	log.Printf("%s attack chase target=%d player=%d,%d target=%d,%d chase=%d,%d", source, actor.ID, ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y, targetX, targetY)
	m.requestWalk(ctx, targetX, targetY, source+" attack chase")
}

func (m *WorldMode) lockAttack(targetID uint32) {
	if targetID == 0 || m.lockedAttackID == targetID {
		return
	}
	m.lockedAttackID = targetID
	m.lastAttackAt = time.Time{}
	m.lastChaseAt = time.Time{}
}

func (m *WorldMode) clearLockedAttack() {
	m.lockedAttackID = 0
	m.lastAttackAt = time.Time{}
	m.lastChaseAt = time.Time{}
}

func (m *WorldMode) continuePendingAttack(ctx Context, source string) {
	if m.pendingAttack.targetID == 0 {
		return
	}
	now := time.Now()
	if now.After(m.pendingAttack.expires) {
		log.Printf("%s pending attack expired target=%d", source, m.pendingAttack.targetID)
		m.pendingAttack = attackIntent{}
		return
	}
	actor, ok := ctx.World.Actors[m.pendingAttack.targetID]
	if !ok {
		log.Printf("%s pending attack target vanished id=%d", source, m.pendingAttack.targetID)
		m.pendingAttack = attackIntent{}
		return
	}
	if !attackTargetWithinRange(ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y) {
		log.Printf("%s pending attack still out of range target=%d player=%d,%d target=%d,%d", source, actor.ID, ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y)
		return
	}
	readyAt := pendingAttackReadyAt(ctx.World.Player, now)
	if m.pendingAttack.readyAt.IsZero() || readyAt.After(m.pendingAttack.readyAt) {
		m.pendingAttack.readyAt = readyAt
	}
	log.Printf("%s pending attack scheduled target=%d delay_ms=%d", source, actor.ID, maxInt(0, int(m.pendingAttack.readyAt.Sub(now).Milliseconds())))
}

func (m *WorldMode) processPendingAttack(ctx Context) {
	if m.pendingAttack.targetID == 0 || m.pendingAttack.readyAt.IsZero() {
		return
	}
	now := time.Now()
	if now.After(m.pendingAttack.expires) {
		log.Printf("pending attack expired target=%d", m.pendingAttack.targetID)
		m.pendingAttack = attackIntent{}
		return
	}
	if now.Before(m.pendingAttack.readyAt) {
		return
	}
	actor, ok := ctx.World.Actors[m.pendingAttack.targetID]
	if !ok {
		log.Printf("pending attack target vanished id=%d", m.pendingAttack.targetID)
		m.pendingAttack = attackIntent{}
		return
	}
	if !attackTargetWithinRange(ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y) {
		log.Printf("pending attack became out of range target=%d player=%d,%d target=%d,%d", actor.ID, ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y)
		m.pendingAttack.readyAt = time.Time{}
		m.requestAttack(ctx, actor, "pending")
		return
	}
	m.pendingAttack = attackIntent{}
	m.sendAttackAction(ctx, actor, "pending")
}

func (m *WorldMode) processLockedAttack(ctx Context) {
	if m.lockedAttackID == 0 || ctx.Network == nil {
		return
	}
	if m.pendingAttack.targetID == m.lockedAttackID {
		return
	}
	now := time.Now()
	if ctx.World.Player.IsMovingAt(now) {
		return
	}
	actor, ok := ctx.World.Actors[m.lockedAttackID]
	if !ok {
		log.Printf("locked attack target vanished id=%d", m.lockedAttackID)
		m.clearLockedAttack()
		return
	}
	if !actorCanBeAttackClicked(ctx, actor) {
		log.Printf("locked attack target no longer attackable id=%d object_type=%d", actor.ID, actor.ObjectType)
		m.clearLockedAttack()
		return
	}
	if attackTargetWithinRange(ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y) {
		if !attackRetryDue(m.lastAttackAt, now) {
			return
		}
		log.Printf("locked attack retry target=%d player=%d,%d target=%d,%d", actor.ID, ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y)
		m.sendAttackAction(ctx, actor, "locked")
		return
	}
	if !attackRetryDue(m.lastChaseAt, now) {
		return
	}
	m.lastChaseAt = now
	log.Printf("locked attack chase retry target=%d player=%d,%d target=%d,%d", actor.ID, ctx.World.Player.X, ctx.World.Player.Y, actor.X, actor.Y)
	m.requestAttack(ctx, actor, "locked")
}

func attackRetryDue(last time.Time, now time.Time) bool {
	return last.IsZero() || now.Sub(last) >= attackRetryInterval
}

func pendingAttackReadyAt(player worldstate.Actor, now time.Time) time.Time {
	readyAt := now.Add(60 * time.Millisecond)
	if player.IsMovingAt(now) && player.MoveDuration > 0 {
		walkReadyAt := player.MoveStarted.Add(player.MoveDuration).Add(60 * time.Millisecond)
		if walkReadyAt.After(readyAt) {
			readyAt = walkReadyAt
		}
	}
	return readyAt
}

func (m *WorldMode) sendAttackAction(ctx Context, actor worldstate.Actor, source string) {
	if err := ctx.Network.SendActionRequest(actor.ID, network.ActionAttack); err == nil {
		m.status = fmt.Sprintf("%s attack request: %d", source, actor.ID)
		m.lastAttackAt = time.Now()
		m.walkCooldown = 12
	} else {
		m.status = source + " attack request failed: " + err.Error()
		log.Printf("%s attack request failed target=%d action=%d: %v", source, actor.ID, network.ActionAttack, err)
		m.walkCooldown = 30
	}
}

func (m *WorldMode) applyActorActionNotify(ctx Context, action network.ActorActionNotify) {
	log.Printf("actor action src=%d dst=%d damage=%d left_damage=%d hits=%d action=%d src_speed=%d dst_speed=%d tick=%d", action.SourceID, action.TargetID, action.Damage, action.LeftDamage, action.HitCount, action.Action, action.SourceSpeed, action.TargetSpeed, action.ServerTick)
	now := time.Now()
	if action.Action == network.ActorActionPickupItem {
		m.applyActorPickupActionNotify(ctx, action, now)
		return
	}
	source, sourceOK, sourceLocal := actorForCombatID(ctx, action.SourceID)
	target, targetOK, targetLocal := actorForCombatID(ctx, action.TargetID)
	if sourceOK && targetOK {
		m.faceCombatSource(ctx, source, sourceLocal, target)
		source.Dir = directionFromDelta(source.X, source.Y, target.X, target.Y, source.Dir)
	}
	attackDuration := combatDuration(action.SourceSpeed, defaultAttackAnimationDuration)
	attackFamily := spriteActionNonPCAttack
	if sourceOK {
		attackFamily = attackActionFamilyForActor(source)
		m.startCombatAnimation(ctx, action.SourceID, attackFamily, now, attackDuration)
	}
	hitDelay := combatDuration(action.SourceSpeed, 0)
	if sourceOK && !res.HasPlayerJobToken(int(source.Job)) {
		if actionDef, ok := m.nonPCResolvedAction(ctx, source, attackFamily); ok {
			hitDelay = combatHitDelayFromAction(actionDef, attackDuration)
			if sound := actionSoundName(m.nonPCActionACT(ctx, source), actionDef, firstActionSoundMotion(actionDef)); sound != "" {
				m.scheduleSound(now.Add(hitDelay), sound)
			}
		}
	}
	hitAt := now.Add(hitDelay)
	if targetOK && actionHasHitReaction(action) {
		if hitAt.Before(now) {
			hitAt = now
		}
		m.startCombatAnimation(ctx, action.TargetID, hurtActionFamilyForActor(target), hitAt, combatDuration(action.TargetSpeed, defaultHitAnimationDuration))
		m.scheduleSound(hitAt, combatHitSFXCandidates(source, sourceOK, target, targetOK)...)
		m.applyCombatLifeFallback(ctx, target, targetLocal, action, hitAt)
		if targetLocal {
			ctx.World.Player.Moving = false
		}
	}
	x, y := ctx.World.Player.X, ctx.World.Player.Y
	if targetOK {
		x, y = target.X, target.Y
	} else if isLocalActor(ctx, action.TargetID) {
		x, y = ctx.World.Player.X, ctx.World.Player.Y
	}
	text := actionDamageText(action)
	if text == "" {
		return
	}
	m.damageFloaters = append(m.damageFloaters, damageFloater{
		actorID: action.TargetID,
		x:       x,
		y:       y,
		text:    text,
		starts:  hitAt,
		expires: hitAt.Add(900 * time.Millisecond),
	})
}

func (m *WorldMode) applyActorPickupActionNotify(ctx Context, action network.ActorActionNotify, now time.Time) {
	source, sourceOK, sourceLocal := actorForCombatID(ctx, action.SourceID)
	if !sourceOK {
		return
	}
	if ctx.World != nil {
		if item, ok := ctx.World.Items[action.TargetID]; ok {
			dir := directionFromDelta(source.X, source.Y, item.X, item.Y, source.Dir)
			if sourceLocal {
				ctx.World.Player.Dir = dir
				ctx.World.Dir = dir
				if ctx.Session != nil {
					ctx.Session.PlayerDir = dir
				}
			} else {
				source.Dir = dir
				ctx.World.UpsertActor(source)
			}
		}
	}
	m.startCombatAnimation(ctx, action.SourceID, spriteActionPickup, now, pickupAnimationDuration)
}

func (m *WorldMode) applyActorHPUpdate(update network.ActorHPUpdate) {
	if update.ID == 0 || update.MaxHP <= 0 {
		return
	}
	hp := update.HP
	if hp < 0 {
		hp = 0
	}
	if hp > update.MaxHP {
		hp = update.MaxHP
	}
	if m.actorLife == nil {
		m.actorLife = make(map[uint32]actorLife)
	}
	m.actorLife[update.ID] = actorLife{
		hp:        hp,
		maxHP:     update.MaxHP,
		fromTiny:  update.Tiny,
		updatedAt: time.Now(),
	}
	log.Printf("actor hp id=%d hp=%d max_hp=%d tiny=%t", update.ID, hp, update.MaxHP, update.Tiny)
}

func (m *WorldMode) applyCombatLifeFallback(ctx Context, target worldstate.Actor, targetLocal bool, action network.ActorActionNotify, hitAt time.Time) {
	if targetLocal || !actorCanBeAttackClicked(ctx, target) {
		return
	}
	damage := int(action.Damage + action.LeftDamage)
	if damage <= 0 {
		return
	}
	if m.actorLife == nil {
		m.actorLife = make(map[uint32]actorLife)
	}
	life, ok := m.actorLife[target.ID]
	if !ok || life.maxHP <= 0 {
		life = actorLife{hp: 100, maxHP: 100}
	}
	life.hp -= damage
	if life.hp < 0 {
		life.hp = 0
	}
	life.updatedAt = hitAt
	m.actorLife[target.ID] = life
}

func actorForCombatID(ctx Context, id uint32) (worldstate.Actor, bool, bool) {
	if ctx.World == nil || id == 0 {
		return worldstate.Actor{}, false, false
	}
	if isLocalActor(ctx, id) {
		actor := ctx.World.Player
		character := selectedCharacter(ctx.Session)
		actor.ID = id
		actor.Job = character.Job
		actor.Head = character.Hair
		actor.Weapon = character.Weapon
		actor.Shield = character.Shield
		actor.HeadTop = character.HeadTop
		actor.HeadMid = character.HeadMid
		actor.HeadLow = character.HeadLow
		actor.Sex = ctx.Session.Sex
		actor.Appearance = true
		return actor, true, true
	}
	actor, ok := ctx.World.Actors[id]
	return actor, ok, false
}

func (m *WorldMode) faceCombatSource(ctx Context, source worldstate.Actor, sourceLocal bool, target worldstate.Actor) {
	dir := directionFromDelta(source.X, source.Y, target.X, target.Y, source.Dir)
	if sourceLocal {
		ctx.World.Player.Dir = dir
		ctx.World.Dir = dir
		return
	}
	source.Dir = dir
	ctx.World.UpsertActor(source)
}

func (m *WorldMode) startActorAnimation(id uint32, actionFamily int, started time.Time, duration time.Duration) {
	if id == 0 || actionFamily < 0 {
		return
	}
	if m.actorAnims == nil {
		m.actorAnims = make(map[uint32]actorAnimation)
	}
	m.actorAnims[id] = actorAnimation{
		actionFamily: actionFamily,
		started:      started,
		duration:     duration,
	}
}

func (m *WorldMode) startCombatAnimation(ctx Context, id uint32, actionFamily int, started time.Time, duration time.Duration) {
	m.startActorAnimation(id, actionFamily, started, duration)
	if ctx.Session == nil || !isLocalActor(ctx, id) {
		return
	}
	m.startActorAnimation(ctx.Session.AccountID, actionFamily, started, duration)
	m.startActorAnimation(ctx.Session.CharID, actionFamily, started, duration)
}

func (m *WorldMode) actorAnimation(id uint32, now time.Time) (actorAnimation, bool) {
	if m.actorAnims == nil || id == 0 {
		return actorAnimation{}, false
	}
	anim, ok := m.actorAnims[id]
	if !ok {
		return actorAnimation{}, false
	}
	if anim.duration <= 0 {
		anim.duration = defaultAttackAnimationDuration
	}
	if now.Before(anim.started) {
		return actorAnimation{}, false
	}
	if !now.Before(anim.started.Add(anim.duration)) {
		delete(m.actorAnims, id)
		return actorAnimation{}, false
	}
	return anim, true
}

func attackActionFamilyForActor(actor worldstate.Actor) int {
	if res.HasPlayerJobToken(int(actor.Job)) {
		if isSecondPCAttack(int(actor.Job), actor.Sex, int(actor.Weapon)) {
			return spriteActionPCAttack3
		}
		return spriteActionPCAttack2
	}
	return spriteActionNonPCAttack
}

func hurtActionFamilyForActor(actor worldstate.Actor) int {
	if res.HasPlayerJobToken(int(actor.Job)) {
		return spriteActionPCHurt
	}
	return spriteActionNonPCHurt
}

func deathActionFamilyForActor(actor worldstate.Actor) int {
	if res.HasPlayerJobToken(int(actor.Job)) {
		return spriteActionPCDeath
	}
	return spriteActionNonPCDeath
}

func isSecondPCAttack(job int, sex byte, weaponValue int) bool {
	weaponType := res.PlayerWeaponType(weaponValue)
	switch job {
	case 0, 23, 4001, 4045:
		if sex != 0 {
			return weaponType == 2 || weaponType == 3 || (weaponType >= 6 && weaponType <= 10) || weaponType == 23
		}
		return weaponType == 1
	case 1, 7, 13, 14, 21:
		return weaponType >= 4 && weaponType <= 5
	case 2, 5:
		return weaponType == 1
	case 3:
		return weaponType != 11
	case 6, 11, 17, 19, 20:
		return weaponType == 11
	case 8:
		return weaponType == 15
	case 10, 18:
		return weaponType == 2 || (weaponType > 5 && weaponType <= 8)
	case 12:
		return weaponType == 16 || (weaponType > 24 && weaponType <= 30)
	case 15:
		return weaponType == 0 || weaponType == 12
	case 16:
		return weaponType == 5 || weaponType == 10 || weaponType == 15 || weaponType == 23
	case 24:
		return weaponType >= 18 && weaponType <= 21
	case 25:
		return weaponType == 22
	default:
		return false
	}
}

func combatDuration(speed int32, fallback time.Duration) time.Duration {
	if speed <= 0 {
		return fallback
	}
	duration := time.Duration(speed) * time.Millisecond
	if duration > maxCombatAnimationDuration {
		return maxCombatAnimationDuration
	}
	return duration
}

func actionAnimationDuration(action res.ACTAction, fallback time.Duration) time.Duration {
	if len(action.Animations) == 0 {
		return fallback
	}
	delayMS := float64(action.DelayMS)
	if delayMS <= 0 {
		delayMS = 150
	}
	duration := time.Duration(delayMS * float64(time.Millisecond) * float64(len(action.Animations)))
	if duration <= 0 {
		return fallback
	}
	if duration > maxCombatAnimationDuration {
		return maxCombatAnimationDuration
	}
	return duration
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func combatHitDelayFromAction(action res.ACTAction, duration time.Duration) time.Duration {
	if duration <= 0 {
		return 0
	}
	motion := firstActionSoundMotion(action)
	if motion >= 0 && len(action.Animations) > 0 {
		return duration * time.Duration(motion) / time.Duration(len(action.Animations))
	}
	return duration / 2
}

func firstActionSoundMotion(action res.ACTAction) int {
	for index, animation := range action.Animations {
		if animation.Sound >= 0 {
			return index
		}
	}
	return -1
}

func actionSoundName(act *res.ACT, action res.ACTAction, motion int) string {
	if act == nil || motion < 0 || motion >= len(action.Animations) {
		return ""
	}
	soundIndex := action.Animations[motion].Sound
	if soundIndex < 0 || soundIndex >= len(act.Sounds) {
		return ""
	}
	return strings.TrimSpace(act.Sounds[soundIndex])
}

func (m *WorldMode) nonPCResolvedAction(ctx Context, actor worldstate.Actor, actionFamily int) (res.ACTAction, bool) {
	view := m.nonPCSpriteView(ctx, actor)
	if view == nil {
		return res.ACTAction{}, false
	}
	_, action, ok := resolveSpriteAction(view.act, actionFamily, actor.Dir)
	return action, ok
}

func (m *WorldMode) nonPCActionACT(ctx Context, actor worldstate.Actor) *res.ACT {
	view := m.nonPCSpriteView(ctx, actor)
	if view == nil {
		return nil
	}
	return view.act
}

func (m *WorldMode) actorActionDuration(ctx Context, actor worldstate.Actor, actionFamily int, fallback time.Duration) time.Duration {
	if res.HasPlayerJobToken(int(actor.Job)) {
		return fallback
	}
	if action, ok := m.nonPCResolvedAction(ctx, actor, actionFamily); ok {
		return actionAnimationDuration(action, fallback)
	}
	return fallback
}

func combatHitSFXCandidates(source worldstate.Actor, sourceOK bool, target worldstate.Actor, targetOK bool) []string {
	if targetOK && res.HasPlayerJobToken(int(target.Job)) {
		return []string{"player_clothes.wav", "player_wooden_male.wav", "player_metal.wav"}
	}
	if sourceOK && res.HasPlayerJobToken(int(source.Job)) {
		return weaponHitSFXCandidates(res.PlayerWeaponType(int(source.Weapon)))
	}
	return []string{"_enemy_hit_normal1.wav", "_enemy_hit_normal2.wav", "_enemy_hit_normal3.wav", "_enemy_hit_normal4.wav"}
}

func weaponHitSFXCandidates(weaponType int) []string {
	switch weaponType {
	case 1, 2, 3:
		return []string{"_hit_sword.wav", "_enemy_hit_normal1.wav"}
	case 4, 5:
		return []string{"_hit_spear.wav", "_enemy_hit_normal1.wav"}
	case 6, 7:
		return []string{"_hit_axe.wav", "_enemy_hit_normal1.wav"}
	case 11:
		return []string{"_hit_arrow.wav", "_enemy_hit_normal1.wav"}
	case 0, 8, 9, 10, 15, 23:
		return []string{"_hit_mace.wav", "_enemy_hit_normal1.wav"}
	default:
		return []string{"_enemy_hit_normal1.wav", "_enemy_hit_normal2.wav", "_enemy_hit_normal3.wav", "_enemy_hit_normal4.wav"}
	}
}

func (m *WorldMode) scheduleSound(at time.Time, paths ...string) {
	clean := paths[:0]
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path != "" {
			clean = append(clean, path)
		}
	}
	if len(clean) == 0 {
		return
	}
	m.scheduledSounds = append(m.scheduledSounds, scheduledSound{
		at:    at,
		paths: append([]string(nil), clean...),
	})
}

func (m *WorldMode) playDueScheduledSounds(ctx Context, now time.Time) {
	if len(m.scheduledSounds) == 0 {
		return
	}
	active := m.scheduledSounds[:0]
	for _, sound := range m.scheduledSounds {
		if now.Before(sound.at) {
			active = append(active, sound)
			continue
		}
		m.playSFXFirst(ctx, sound.paths...)
	}
	m.scheduledSounds = active
}

func (m *WorldMode) processActorMotionSounds(ctx Context, now time.Time) {
	if ctx.World == nil || len(ctx.World.Actors) == 0 {
		return
	}
	for _, actor := range ctx.World.Actors {
		m.processNonPCMotionSound(ctx, actor, now)
	}
}

func (m *WorldMode) processNonPCMotionSound(ctx Context, actor worldstate.Actor, now time.Time) {
	if actor.ID == 0 || res.HasPlayerJobToken(int(actor.Job)) || !actorWithinSoundRange(ctx, actor, now) {
		return
	}
	view := m.nonPCSpriteView(ctx, actor)
	if view == nil || view.act == nil {
		return
	}
	state := m.nonPCSpriteState(actor, now)
	switch state.actionFamily {
	case spriteActionNonPCAttack, spriteActionNonPCHurt:
		return
	}
	_, action, ok := resolveSpriteAction(view.act, state.actionFamily, state.direction)
	if !ok || len(action.Animations) == 0 {
		return
	}
	motion := bodyMotionForState(action, state, view.started, now)
	if motion < 0 || motion >= len(action.Animations) {
		return
	}
	soundIndex := action.Animations[motion].Sound
	current := actorSoundFrame{actionFamily: state.actionFamily, motion: motion, soundIndex: soundIndex}
	if m.actorSoundFrames == nil {
		m.actorSoundFrames = make(map[uint32]actorSoundFrame)
	}
	if previous, ok := m.actorSoundFrames[actor.ID]; ok && previous == current {
		return
	}
	m.actorSoundFrames[actor.ID] = current
	if soundIndex < 0 {
		return
	}
	if sound := actionSoundName(view.act, action, motion); sound != "" {
		m.scheduleSound(now, sound)
	}
}

func actorWithinSoundRange(ctx Context, actor worldstate.Actor, now time.Time) bool {
	if ctx.World == nil {
		return false
	}
	actorX, actorY := actor.RenderPosition(now)
	playerX, playerY := ctx.World.Player.RenderPosition(now)
	const soundRangeCells = 25
	return math.Hypot(actorX-playerX, actorY-playerY) <= soundRangeCells
}

func (m *WorldMode) playSFXFirst(ctx Context, paths ...string) {
	if ctx.Audio == nil {
		return
	}
	var lastErr error
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		source, err := ctx.Audio.PlaySFX(path)
		if err == nil {
			if source != "" {
				log.Printf("sfx playing path=%s source=%s", path, source)
			}
			return
		}
		lastErr = err
	}
	if lastErr != nil {
		log.Printf("sfx failed paths=%v: %v", paths, lastErr)
	}
}

func actionHasHitReaction(action network.ActorActionNotify) bool {
	if action.Action == 4 || action.Action == 9 || action.Action == 11 {
		return false
	}
	return action.Damage > 0 || action.LeftDamage > 0
}

func (m *WorldMode) applyAttackFailureForDistance(ctx Context, failure network.AttackFailureForDistance) {
	attackRange := maxInt(1, failure.AttackRange)
	log.Printf("attack distance failure target=%d server_player=%d,%d server_target=%d,%d range=%d client_player=%d,%d", failure.TargetID, failure.SourceX, failure.SourceY, failure.TargetX, failure.TargetY, attackRange, ctx.World.Player.X, ctx.World.Player.Y)
	ctx.World.SetPlayerPosition(failure.SourceX, failure.SourceY, ctx.World.Player.Dir)
	if actor, ok := ctx.World.Actors[failure.TargetID]; ok {
		actor.X = failure.TargetX
		actor.Y = failure.TargetY
		actor.Moving = false
		actor.FromX = failure.TargetX
		actor.FromY = failure.TargetY
		actor.ToX = failure.TargetX
		actor.ToY = failure.TargetY
		actor.MovePath = nil
		ctx.World.UpsertActor(actor)
	}
	if m.lockedAttackID != failure.TargetID && m.pendingAttack.targetID != failure.TargetID {
		return
	}
	m.pendingAttack = attackIntent{}
	m.lastAttackAt = time.Now()
	if !attackTargetWithinRange(failure.SourceX, failure.SourceY, failure.TargetX, failure.TargetY) {
		if actor, ok := ctx.World.Actors[failure.TargetID]; ok {
			m.requestAttack(ctx, actor, "attack failure")
		}
	}
}

func applyParameterChange(ctx Context, change network.ParameterChange) {
	if ctx.Session == nil {
		return
	}
	value := int(change.Value)
	switch change.VarID {
	case network.StatusBaseExp:
		ctx.Session.Progress.BaseExp = change.Value
	case network.StatusJobExp:
		ctx.Session.Progress.JobExp = change.Value
	case network.StatusHP:
		ctx.Session.Vitals.HP = value
		ctx.Session.Selected.HP = clampInt16(value)
	case network.StatusMaxHP:
		ctx.Session.Vitals.MaxHP = value
		ctx.Session.Selected.MaxHP = clampInt16(value)
	case network.StatusSP:
		ctx.Session.Vitals.SP = value
		ctx.Session.Selected.SP = clampInt16(value)
	case network.StatusMaxSP:
		ctx.Session.Vitals.MaxSP = value
		ctx.Session.Selected.MaxSP = clampInt16(value)
	case network.StatusBaseLevel:
		ctx.Session.Progress.BaseLevel = value
		ctx.Session.Selected.Level = clampInt16(value)
	case network.StatusNextBaseExp:
		ctx.Session.Progress.NextBaseExp = change.Value
	case network.StatusNextJobExp:
		ctx.Session.Progress.NextJobExp = change.Value
	case network.StatusJobLevel:
		ctx.Session.Progress.JobLevel = value
	default:
		return
	}
	log.Printf("parameter change var=%d value=%d hp=%d/%d sp=%d/%d base_lv=%d job_lv=%d base_exp=%d/%d job_exp=%d/%d",
		change.VarID,
		change.Value,
		ctx.Session.Vitals.HP,
		ctx.Session.Vitals.MaxHP,
		ctx.Session.Vitals.SP,
		ctx.Session.Vitals.MaxSP,
		ctx.Session.Progress.BaseLevel,
		ctx.Session.Progress.JobLevel,
		ctx.Session.Progress.BaseExp,
		ctx.Session.Progress.NextBaseExp,
		ctx.Session.Progress.JobExp,
		ctx.Session.Progress.NextJobExp)
}

func clampInt16(value int) int16 {
	if value < -32768 {
		return -32768
	}
	if value > 32767 {
		return 32767
	}
	return int16(value)
}

func actionDamageText(action network.ActorActionNotify) string {
	total := action.Damage + action.LeftDamage
	if total > 0 {
		return strconv.Itoa(int(total))
	}
	if action.Action == 10 {
		return "crit"
	}
	if action.Action == 11 {
		return "miss"
	}
	if action.Action == 0 || action.Action == 7 {
		return "miss"
	}
	return ""
}

func attackTargetWithinRange(playerX, playerY, targetX, targetY int) bool {
	return maxInt(absInt(playerX-targetX), absInt(playerY-targetY)) <= 1
}

func attackApproachCell(ctx Context, actor worldstate.Actor) (int, int, bool) {
	bestX, bestY := 0, 0
	bestDistance := math.Inf(1)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			x := actor.X + dx
			y := actor.Y + dy
			if !walkTargetInBounds(ctx, x, y) {
				continue
			}
			if ctx.World.GAT != nil && !ctx.World.GAT.Walkable(x, y) {
				continue
			}
			distance := math.Hypot(float64(x-ctx.World.Player.X), float64(y-ctx.World.Player.Y))
			if distance < bestDistance {
				bestDistance = distance
				bestX = x
				bestY = y
			}
		}
	}
	return bestX, bestY, bestDistance < math.Inf(1)
}

func (m *WorldMode) drawDamageFloaters(screen *ebiten.Image, ctx Context, projection sceneProjection, now time.Time) {
	if len(m.damageFloaters) == 0 {
		return
	}
	active := m.damageFloaters[:0]
	for _, floater := range m.damageFloaters {
		if now.After(floater.expires) {
			continue
		}
		active = append(active, floater)
		if now.Before(floater.starts) {
			continue
		}
		x, y := float64(floater.x), float64(floater.y)
		if actor, ok := ctx.World.Actors[floater.actorID]; ok {
			x, y = actor.RenderPosition(now)
		} else if isLocalActor(ctx, floater.actorID) {
			x, y = ctx.World.Player.RenderPosition(now)
		}
		terrainZ := terrainHeightAt(ctx.World, x, y)
		point := projection.Project(cellCenter(x), cellCenter(y), terrainZ)
		remaining := floater.expires.Sub(now)
		rise := float64(900*time.Millisecond-remaining) / float64(900*time.Millisecond) * 28
		debugText(screen, int(point.x)-8, int(point.y)-90-int(rise), "%s", floater.text)
	}
	m.damageFloaters = active
}

func drawVitalsHUD(screen *ebiten.Image, ctx Context) {
	if ctx.Session == nil {
		return
	}
	vitals := ctx.Session.Vitals
	if vitals.HP == 0 && vitals.MaxHP == 0 && vitals.SP == 0 && vitals.MaxSP == 0 {
		vitals = sessionVitalsFromCharacter(ctx.Session.Selected)
	}
	progress := ctx.Session.Progress
	if progress.BaseLevel == 0 {
		progress = sessionProgressFromCharacter(ctx.Session.Selected)
		progress.JobLevel = ctx.Session.Progress.JobLevel
		progress.BaseExp = ctx.Session.Progress.BaseExp
		progress.NextBaseExp = ctx.Session.Progress.NextBaseExp
		progress.JobExp = ctx.Session.Progress.JobExp
		progress.NextJobExp = ctx.Session.Progress.NextJobExp
	}
	width := screen.Bounds().Dx()
	x := maxInt(24, width-300)
	debugText(screen, x, 24, "HP %d / %d", vitals.HP, vitals.MaxHP)
	debugText(screen, x, 44, "SP %d / %d", vitals.SP, vitals.MaxSP)
	debugText(screen, x, 64, "Base Lv %d  EXP %s", progress.BaseLevel, formatProgressValue(progress.BaseExp, progress.NextBaseExp))
	debugText(screen, x, 84, "Job  Lv %d  EXP %s", progress.JobLevel, formatProgressValue(progress.JobExp, progress.NextJobExp))
}

func sessionVitalsFromCharacter(character session.Character) session.Vitals {
	return session.Vitals{
		HP:    int(character.HP),
		MaxHP: int(character.MaxHP),
		SP:    int(character.SP),
		MaxSP: int(character.MaxSP),
	}
}

func sessionProgressFromCharacter(character session.Character) session.Progress {
	return session.Progress{
		BaseLevel: int(character.Level),
	}
}

func formatProgressValue(current, next int64) string {
	if next > 0 {
		return fmt.Sprintf("%d / %d", current, next)
	}
	return fmt.Sprintf("%d", current)
}

func (m *WorldMode) Draw(ctx Context, screen *ebiten.Image) {
	clear(screen)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
	now := time.Now()
	playerX, playerY := ctx.World.Player.RenderPosition(now)
	projection := m.sceneProjection(ctx, width, height, now)

	if ctx.World.GND != nil {
		m.drawGND(screen, ctx.Resources, ctx.World.GND, ctx.World.RSW, projection, now)
		m.drawTileCursor(screen, ctx, projection, now)
		if ctx.World.RSW != nil && len(ctx.World.RSM) > 0 && m.rsmRender {
			m.drawSceneModelsAndActors(screen, ctx, projection)
		} else {
			m.drawGroundItems(screen, ctx, projection, now)
			m.drawSceneActors(screen, ctx, projection)
		}
		if ctx.World.RSW != nil && m.rswMarkers {
			drawRSWModelMarkers(screen, ctx.World.RSW, ctx.World.GND, projection)
		}
	} else if ctx.World.GAT != nil {
		drawGAT(screen, ctx.World.GAT, ctx.World.Player.X, ctx.World.Player.Y)
		m.drawGroundItems(screen, ctx, projection, now)
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

	m.drawDamageFloaters(screen, ctx, projection, now)

	debugText(screen, 24, 24, "map: %s player=(%d,%d) dir=%d yaw=%.1f", ctx.World.MapName, ctx.World.Player.X, ctx.World.Player.Y, ctx.World.Dir, projection.cameraYaw)
	debugText(screen, 24, 44, "%s", m.status)
	drawVitalsHUD(screen, ctx)
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
		debugText(screen, 24, y+20, "actors: %d items: %d", len(ctx.World.Actors), len(ctx.World.Items))
	}
	m.drawHoveredGroundItemLabel(screen, ctx, projection, now)
	m.drawROCursor(screen, ctx, projection, now)
}

type followCamera struct {
	initialized bool
	x           float64
	y           float64
	z           float64
	yawOffset   float64
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

func (c *followCamera) Rotate(delta float64) {
	c.yawOffset = normalizeCameraYaw(c.yawOffset + delta)
}

func (c *followCamera) ResetRotation() {
	c.yawOffset = 0
}

func (c *followCamera) Projection(ctx Context, width, height int, now time.Time) sceneProjection {
	if !c.initialized {
		c.Update(ctx, now)
	}
	c.store(ctx)
	return newSceneProjectionForTargetYaw(width, height, c.x, c.y, c.z, cameraYawForMap(ctx)+c.yawOffset)
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

func (m *WorldMode) updateCameraRotation(ctx Context) {
	delta := 0.0
	if ctx.Input.MousePressed(ebiten.MouseButtonRight) {
		screenW, _ := ctx.ScreenSize()
		delta += cameraDragYawDelta(ctx.Input.MouseDX, screenW)
	}
	if ctx.Input.Pressed(ebiten.KeyQ) {
		delta -= cameraRotateStep()
	}
	if ctx.Input.Pressed(ebiten.KeyE) {
		delta += cameraRotateStep()
	}
	if delta != 0 {
		m.camera.Rotate(delta)
	}
	if ctx.Input.JustPressed(ebiten.KeyR) {
		m.camera.ResetRotation()
	}
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

func cameraYawForMap(ctx Context) float64 {
	if ctx.Resources != nil && ctx.World != nil && ctx.Resources.IsIndoorMap(ctx.World.MapName) {
		return -45
	}
	return sceneCameraYaw()
}

func cameraRotateStep() float64 {
	return sceneFloatEnv("GORO_CAMERA_ROTATE_SPEED", 90) / 60
}

func cameraDragYawDelta(mouseDX, screenWidth int) float64 {
	if mouseDX == 0 || screenWidth <= 0 {
		return 0
	}
	return -(float64(mouseDX) / float64(screenWidth)) * 720
}

func normalizeCameraYaw(yaw float64) float64 {
	yaw = math.Mod(yaw, 360)
	if yaw <= -180 {
		yaw += 360
	}
	if yaw > 180 {
		yaw -= 360
	}
	return yaw
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
		ID:            entry.ID,
		X:             entry.X,
		Y:             entry.Y,
		Dir:           dir,
		Job:           entry.Job,
		Head:          entry.Head,
		Weapon:        entry.Weapon,
		Shield:        entry.Shield,
		HeadTop:       entry.HeadTop,
		HeadMid:       entry.HeadMid,
		HeadLow:       entry.HeadLow,
		Sex:           entry.Sex,
		Appearance:    entry.Appearance,
		Moving:        entry.Moving,
		FromX:         entry.FromX,
		FromY:         entry.FromY,
		ToX:           entry.ToX,
		ToY:           entry.ToY,
		ObjectType:    entry.ObjectType,
		HasObjectType: entry.HasObjectType,
		Speed:         entry.Speed,
	})
}

func (m *WorldMode) applyActorVanish(ctx Context, vanish network.ActorVanish) {
	log.Printf("actor vanish id=%d reason=%d", vanish.ID, vanish.Reason)
	if vanish.Reason == 1 {
		m.startActorDeath(ctx, vanish.ID)
		return
	}
	ctx.World.RemoveActor(vanish.ID)
	delete(m.actorAnims, vanish.ID)
	delete(m.actorDeaths, vanish.ID)
	delete(m.actorSoundFrames, vanish.ID)
	delete(m.actorLife, vanish.ID)
}

func (m *WorldMode) startActorDeath(ctx Context, id uint32) {
	actor, ok, local := actorForCombatID(ctx, id)
	if !ok {
		if !local {
			ctx.World.RemoveActor(id)
		}
		return
	}
	now := time.Now()
	actor.Moving = false
	actor.FromX = actor.X
	actor.FromY = actor.Y
	actor.ToX = actor.X
	actor.ToY = actor.Y
	actor.MovePath = nil
	actor.WalkDistance = 0
	if local {
		ctx.World.Player.Moving = false
	} else {
		ctx.World.UpsertActor(actor)
	}
	actionFamily := deathActionFamilyForActor(actor)
	deathDuration := m.actorActionDuration(ctx, actor, actionFamily, defaultDeathAnimationDuration)
	visibleDuration := deathDuration
	if !local {
		visibleDuration = maxDuration(deathDuration, nonPCDeathFadeDuration)
	}
	m.startCombatAnimation(ctx, id, actionFamily, now, visibleDuration)
	if !local {
		if m.actorDeaths == nil {
			m.actorDeaths = make(map[uint32]time.Time)
		}
		m.actorDeaths[id] = now.Add(visibleDuration)
		if m.actorLife != nil {
			if life, ok := m.actorLife[id]; ok {
				life.hp = 0
				life.updatedAt = now
				m.actorLife[id] = life
			}
		}
	}
	log.Printf("actor death id=%d job=%d local=%t action=%d death_ms=%d remove_ms=%d", id, actor.Job, local, actionFamily, deathDuration.Milliseconds(), visibleDuration.Milliseconds())
}

func (m *WorldMode) cleanupDeadActors(ctx Context, now time.Time) {
	if len(m.actorDeaths) == 0 || ctx.World == nil {
		return
	}
	for id, removeAt := range m.actorDeaths {
		if now.Before(removeAt) {
			continue
		}
		ctx.World.RemoveActor(id)
		delete(m.actorDeaths, id)
		delete(m.actorAnims, id)
		delete(m.actorSoundFrames, id)
		delete(m.actorLife, id)
		if m.pendingAttack.targetID == id {
			m.pendingAttack = attackIntent{}
		}
		if m.lockedAttackID == id {
			m.clearLockedAttack()
		}
		log.Printf("actor death removed id=%d", id)
	}
}

func (m *WorldMode) clearActorDeath(id uint32) {
	delete(m.actorDeaths, id)
	delete(m.actorAnims, id)
	delete(m.actorSoundFrames, id)
}

func (m *WorldMode) actorDeathAlpha(id uint32, now time.Time) float64 {
	removeAt, ok := m.actorDeaths[id]
	if !ok {
		return 1
	}
	started := now
	if anim, ok := m.actorAnims[id]; ok && !anim.started.IsZero() {
		started = anim.started
	}
	total := removeAt.Sub(started)
	if total <= 0 {
		return 0
	}
	elapsed := now.Sub(started)
	if elapsed <= 0 {
		return 1
	}
	alpha := 1 - float64(elapsed)/float64(total)
	if alpha < 0 {
		return 0
	}
	if alpha > 1 {
		return 1
	}
	return alpha
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
	return ctx.Session != nil && id != 0 && (id == ctx.Session.AccountID || id == ctx.Session.CharID)
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
	return worldstate.DirectionFromDelta(fromX, fromY, toX, toY, normalizeDirectionIndex(fallback))
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
	minX, maxX, minY, maxY, ok := walkTargetSearchBounds(ctx)
	if !ok {
		return 0, 0, false
	}

	if x, y, ok := clickedWalkCellByProjectedPolygon(ctx, projection, mouseX, mouseY, minX, maxX, minY, maxY); ok {
		return x, y, true
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

func hoveredWalkCell(ctx Context, projection sceneProjection, mouseX, mouseY int) (int, int, bool) {
	minX, maxX, minY, maxY, ok := walkTargetSearchBounds(ctx)
	if !ok {
		return 0, 0, false
	}
	return clickedWalkCellByProjectedPolygon(ctx, projection, mouseX, mouseY, minX, maxX, minY, maxY)
}

func walkTargetSearchBounds(ctx Context) (int, int, int, int, bool) {
	if ctx.World == nil {
		return 0, 0, 0, 0, false
	}
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
	return minX, maxX, minY, maxY, minX <= maxX && minY <= maxY
}

func clickedWalkCellByProjectedPolygon(ctx Context, projection sceneProjection, mouseX, mouseY, minX, maxX, minY, maxY int) (int, int, bool) {
	if ctx.World == nil || ctx.World.GAT == nil {
		return 0, 0, false
	}
	bestX, bestY := 0, 0
	bestDepth := math.Inf(1)
	found := false
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if !ctx.World.GAT.Walkable(x, y) {
				continue
			}
			points, depth, ok := projectedGATCell(projection, ctx.World.GAT, x, y)
			if !ok {
				continue
			}
			if !pointInProjectedGATCell(float64(mouseX), float64(mouseY), points) {
				continue
			}
			if !found || depth < bestDepth {
				found = true
				bestDepth = depth
				bestX = x
				bestY = y
			}
		}
	}
	return bestX, bestY, found
}

func (m *WorldMode) drawTileCursor(screen *ebiten.Image, ctx Context, projection sceneProjection, now time.Time) {
	if ctx.Input == nil || ctx.World == nil || ctx.World.GAT == nil {
		return
	}
	x, y, ok := hoveredWalkCell(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY)
	if !ok {
		return
	}
	points, ok := projectedTileCursorCell(projection, ctx.World.GAT, x, y, now)
	if !ok || quadHasInvalidPoint(points) || quadOutside(points, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())) {
		return
	}
	drawTileCursorSurface(screen, m.tileCursorTexture(), points)
}

func projectedTileCursorCell(projection sceneProjection, gat *res.GAT, x, y int, now time.Time) ([4]screenPoint, bool) {
	cell, ok := gat.Cell(x, y)
	if !ok {
		return [4]screenPoint{}, false
	}
	lift := tileCursorLift(now)
	verts := [4]modelPoint3{
		{x: float64(x), y: float64(cell.Heights[0]) + lift, z: float64(y)},
		{x: float64(x + 1), y: float64(cell.Heights[1]) + lift, z: float64(y)},
		{x: float64(x), y: float64(cell.Heights[2]) + lift, z: float64(y + 1)},
		{x: float64(x + 1), y: float64(cell.Heights[3]) + lift, z: float64(y + 1)},
	}
	return [4]screenPoint{
		projection.Project(verts[0].x, verts[0].z, verts[0].y),
		projection.Project(verts[1].x, verts[1].z, verts[1].y),
		projection.Project(verts[2].x, verts[2].z, verts[2].y),
		projection.Project(verts[3].x, verts[3].z, verts[3].y),
	}, true
}

func tileCursorLift(now time.Time) float64 {
	seconds := float64(now.UnixNano()) / float64(time.Second)
	return 0.06 + 0.025*math.Sin(seconds*math.Pi*2/1.2)
}

func (m *WorldMode) tileCursorTexture() *ebiten.Image {
	if m.tileCursor != nil {
		return m.tileCursor
	}
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dist := minInt(minInt(x, y), minInt(size-1-x, size-1-y))
			alpha := uint8(0)
			switch {
			case dist < 3:
				alpha = 190
			case dist < 6:
				alpha = 100
			case dist < 11:
				alpha = 32
			}
			if x == y || x == size-1-y {
				alpha = maxUint8(alpha, 34)
			}
			if alpha > 0 {
				img.SetRGBA(x, y, color.RGBA{R: 180, G: 230, B: 255, A: alpha})
			}
		}
	}
	m.tileCursor = ebiten.NewImageFromImage(img)
	return m.tileCursor
}

func drawTileCursorSurface(screen, texture *ebiten.Image, points [4]screenPoint) {
	if texture == nil {
		return
	}
	bounds := texture.Bounds()
	w := float32(bounds.Dx())
	h := float32(bounds.Dy())
	tint := color.RGBA{R: 255, G: 255, B: 255, A: 210}
	vertices := []ebiten.Vertex{
		texturedSurfaceVertex(points[0], texturePoint{u: 0, v: 0}, tint, w, h),
		texturedSurfaceVertex(points[1], texturePoint{u: 1, v: 0}, tint, w, h),
		texturedSurfaceVertex(points[2], texturePoint{u: 0, v: 1}, tint, w, h),
		texturedSurfaceVertex(points[3], texturePoint{u: 1, v: 1}, tint, w, h),
	}
	op := &ebiten.DrawTrianglesOptions{
		Filter:  ebiten.FilterLinear,
		Address: ebiten.AddressClampToZero,
	}
	screen.DrawTriangles(vertices, []uint16{0, 1, 2, 2, 1, 3}, texture, op)
}

func projectedGATCell(projection sceneProjection, gat *res.GAT, x, y int) ([4]screenPoint, float64, bool) {
	cell, ok := gat.Cell(x, y)
	if !ok {
		return [4]screenPoint{}, 0, false
	}
	verts := [4]modelPoint3{
		{x: float64(x), y: float64(cell.Heights[0]), z: float64(y)},
		{x: float64(x + 1), y: float64(cell.Heights[1]), z: float64(y)},
		{x: float64(x), y: float64(cell.Heights[2]), z: float64(y + 1)},
		{x: float64(x + 1), y: float64(cell.Heights[3]), z: float64(y + 1)},
	}
	points := [4]screenPoint{
		projection.Project(verts[0].x, verts[0].z, verts[0].y),
		projection.Project(verts[1].x, verts[1].z, verts[1].y),
		projection.Project(verts[2].x, verts[2].z, verts[2].y),
		projection.Project(verts[3].x, verts[3].z, verts[3].y),
	}
	depth := math.Inf(1)
	for _, vert := range verts {
		depth = math.Min(depth, projection.Depth(vert.x, vert.z, vert.y))
	}
	return points, depth, true
}

func pointInProjectedGATCell(x, y float64, points [4]screenPoint) bool {
	return pointInScreenTriangle(x, y, points[0], points[1], points[2]) ||
		pointInScreenTriangle(x, y, points[2], points[1], points[3])
}

func pointInScreenTriangle(x, y float64, a, b, c screenPoint) bool {
	d1 := screenTriangleSign(x, y, a, b)
	d2 := screenTriangleSign(x, y, b, c)
	d3 := screenTriangleSign(x, y, c, a)
	hasNegative := d1 < 0 || d2 < 0 || d3 < 0
	hasPositive := d1 > 0 || d2 > 0 || d3 > 0
	return !(hasNegative && hasPositive)
}

func screenTriangleSign(x, y float64, a, b screenPoint) float64 {
	return (x-float64(b.x))*(float64(a.y)-float64(b.y)) - (float64(a.x)-float64(b.x))*(y-float64(b.y))
}

func clickedAttackTarget(ctx Context, projection sceneProjection, mouseX, mouseY int, now time.Time, deadActors map[uint32]time.Time) (worldstate.Actor, bool) {
	if ctx.World == nil {
		return worldstate.Actor{}, false
	}
	bestDistance := math.Inf(1)
	var best worldstate.Actor
	for _, actor := range ctx.World.Actors {
		if _, dead := deadActors[actor.ID]; dead {
			continue
		}
		if !actorCanBeAttackClicked(ctx, actor) {
			continue
		}
		actorX, actorY := actor.RenderPosition(now)
		terrainZ := terrainHeightAt(ctx.World, actorX, actorY)
		point := projection.Project(cellCenter(actorX), cellCenter(actorY), terrainZ)
		scale := actorBillboardScreenScale(projection, cellCenter(actorX), cellCenter(actorY), terrainZ)
		if !pointInActorPickBounds(float64(mouseX), float64(mouseY), float64(point.x), float64(point.y), scale) {
			continue
		}
		dx := float64(point.x) - float64(mouseX)
		dy := float64(point.y) - float64(mouseY)
		distance := dx*dx + dy*dy
		if distance < bestDistance {
			bestDistance = distance
			best = actor
		}
	}
	return best, bestDistance < math.Inf(1)
}

func actorCanBeAttackClicked(ctx Context, actor worldstate.Actor) bool {
	if isLocalActor(ctx, actor.ID) {
		return false
	}
	if actor.ID == 0 || !actor.HasObjectType {
		return false
	}
	switch actor.ObjectType {
	case actorObjectTypeMob, actorObjectTypeNPCABR, actorObjectTypeNPCBionic:
		return true
	default:
		return false
	}
}

func pointInActorPickBounds(mouseX, mouseY, centerX, centerY, scale float64) bool {
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	left := centerX - 44*scale
	right := centerX + 44*scale
	top := centerY - float64(humanoidBillboardAnchorY)*scale
	bottom := centerY + 20*scale
	return mouseX >= left && mouseX <= right && mouseY >= top && mouseY <= bottom
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

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

type sceneActorDrawEntry struct {
	actor       worldstate.Actor
	label       string
	screenX     float64
	screenY     float64
	worldX      float64
	worldY      float64
	worldZ      float64
	scale       float64
	shadow      float64
	castShadow  bool
	shadowX     float64
	shadowY     float64
	shadowScale float64
	shadowDepth float64
	depth       float64
	isPlayer    bool
}

const (
	actorBillboardCellWorldUnits  = 5.0
	actorBillboardWorldHeightUnit = 1.0 * actorBillboardCellWorldUnits
	actorJobWarpPortal            = 45
	actorJobClearNPC              = 844
	actorObjectTypeMob            = 5
	actorObjectTypeNPCABR         = 13
	actorObjectTypeNPCBionic      = 14
)

var monsterShadowSize = map[int]float64{
	111:  0.0,
	139:  0.0,
	1004: 0.5,
	1005: 0.5,
	1007: 0.5,
	1008: 0.3,
	1009: 0.7,
	1011: 0.5,
	1013: 1.2,
	1018: 0.7,
	1019: 1.2,
	1020: 0.0,
	1025: 0.0,
	1030: 0.0,
	1035: 0.5,
	1037: 0.0,
	1039: 1.2,
	1040: 2.0,
	1042: 0.5,
	1046: 0.0,
	1047: 0.2,
	1048: 0.2,
	1049: 0.3,
	1050: 0.3,
	1051: 0.3,
	1056: 0.7,
	1057: 0.7,
	1061: 1.5,
	1063: 0.5,
	1069: 1.2,
	1070: 0.3,
	1072: 0.5,
	1074: 0.5,
	1078: 0.0,
	1079: 0.0,
	1080: 0.0,
	1081: 0.0,
	1082: 0.0,
	1083: 0.0,
	1084: 0.0,
	1085: 0.0,
	1087: 1.2,
	1089: 1.5,
	1090: 1.0,
	1091: 0.5,
	1092: 1.2,
	1094: 0.7,
	1095: 0.5,
	1097: 0.2,
	1098: 2.0,
	1101: 0.5,
	1102: 1.2,
	1103: 0.3,
	1104: 0.7,
	1105: 0.7,
	1106: 1.2,
	1107: 0.7,
	1108: 0.7,
	1109: 0.7,
	1110: 0.7,
	1111: 0.5,
	1114: 0.5,
	1115: 1.2,
	1121: 0.7,
	1127: 0.0,
	1129: 0.5,
	1131: 0.0,
	1138: 0.0,
	1139: 0.5,
	1140: 1.2,
	1141: 0.5,
	1142: 0.5,
	1143: 0.5,
	1145: 0.5,
	1147: 1.5,
	1149: 1.5,
	1155: 0.5,
	1156: 0.5,
	1158: 0.7,
	1159: 1.2,
	1160: 0.7,
	1161: 0.5,
	1162: 0.5,
	1167: 0.5,
	1170: 0.7,
	1174: 0.5,
	1175: 0.5,
	1176: 0.7,
	1182: 0.0,
	1183: 0.5,
	1184: 0.5,
	1186: 2.0,
	1190: 1.2,
	1192: 1.5,
	1193: 2.0,
	1194: 0.5,
	1195: 0.5,
	1199: 0.5,
	1201: 1.2,
	1202: 1.5,
	1203: 0.5,
	1204: 0.5,
	1208: 1.2,
	1209: 0.7,
	1211: 0.5,
	1214: 0.7,
	1219: 5.0,
}

func (m *WorldMode) drawSceneActors(screen *ebiten.Image, ctx Context, projection sceneProjection) {
	entries := m.collectSceneActorEntries(screen, ctx, projection)
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].depth > entries[j].depth
	})
	for _, entry := range entries {
		m.drawActorShadowEntry(screen, entry)
	}
	for _, entry := range entries {
		m.drawSceneActorEntry(screen, ctx, projection, entry)
	}
	for _, entry := range entries {
		m.drawActorLifeBar(screen, ctx, entry)
	}
	for _, entry := range entries {
		drawActorNameLabel(screen, entry.label, entry.screenX, entry.screenY, entry.scale)
	}
}

type sceneDrawEntry struct {
	depth       float64
	modelIndex  int
	actorIndex  int
	shadowIndex int
	itemIndex   int
}

func (m *WorldMode) drawSceneModelsAndActors(screen *ebiten.Image, ctx Context, projection sceneProjection) {
	models := m.collectRSMModelTriangles(screen, ctx.Resources, ctx.World.RSW, ctx.World.RSM, ctx.World.GND, projection)
	actors := m.collectSceneActorEntries(screen, ctx, projection)
	items := m.collectSceneItemEntries(screen, ctx, projection, time.Now())
	entries := make([]sceneDrawEntry, 0, len(models)+len(actors)+len(items))
	for i, tri := range models {
		entries = append(entries, sceneDrawEntry{depth: tri.depth, modelIndex: i, actorIndex: -1, shadowIndex: -1, itemIndex: -1})
	}
	for i, item := range items {
		entries = append(entries, sceneDrawEntry{depth: item.depth, modelIndex: -1, actorIndex: -1, shadowIndex: -1, itemIndex: i})
	}
	for i, actor := range actors {
		if actor.castShadow {
			entries = append(entries, sceneDrawEntry{depth: actor.shadowDepth, modelIndex: -1, actorIndex: -1, shadowIndex: i, itemIndex: -1})
		}
		entries = append(entries, sceneDrawEntry{depth: actor.depth, modelIndex: -1, actorIndex: i, shadowIndex: -1, itemIndex: -1})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].depth > entries[j].depth
	})
	for _, entry := range entries {
		if entry.modelIndex >= 0 {
			m.drawModelTriangle(screen, ctx.Resources, models[entry.modelIndex])
			continue
		}
		if entry.shadowIndex >= 0 {
			m.drawActorShadowEntry(screen, actors[entry.shadowIndex])
			continue
		}
		if entry.itemIndex >= 0 {
			m.drawGroundItemEntry(screen, items[entry.itemIndex])
			continue
		}
		m.drawSceneActorEntry(screen, ctx, projection, actors[entry.actorIndex])
	}
	for _, actor := range actors {
		m.drawActorLifeBar(screen, ctx, actor)
	}
	for _, actor := range actors {
		drawActorNameLabel(screen, actor.label, actor.screenX, actor.screenY, actor.scale)
	}
}

func (m *WorldMode) collectSceneActorEntries(screen *ebiten.Image, ctx Context, projection sceneProjection) []sceneActorDrawEntry {
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
	return entries
}

func (m *WorldMode) drawSceneActorEntry(screen *ebiten.Image, ctx Context, projection sceneProjection, entry sceneActorDrawEntry) {
	cameraYaw := projection.cameraYaw
	if entry.isPlayer {
		if m.drawPlayerSprite(ctx, screen, entry.screenX, entry.screenY, entry.scale, entry.actor.Dir, cameraYaw, entry.shadow) {
			return
		}
		drawPanel(screen, entry.screenX-6, entry.screenY-6, 24, 24)
		return
	}
	if isWarpActor(entry.actor) {
		if m.whitePixel == nil {
			m.whitePixel = ebiten.NewImage(1, 1)
			m.whitePixel.Fill(color.White)
		}
		drawWarpZoneEffect(screen, m.whitePixel, m.effectTexture(ctx.Resources, "ring_blue"), projection, entry.worldX, entry.worldY, entry.worldZ, time.Now())
		return
	}
	if m.drawActorSprite(screen, ctx, entry.actor, entry.screenX, entry.screenY, entry.scale, cameraYaw, entry.shadow) {
		return
	}
	drawActorMarker(screen, entry.screenX-6, entry.screenY-20, entry.actor, time.Now())
}

func (m *WorldMode) drawActorShadowEntry(screen *ebiten.Image, entry sceneActorDrawEntry) {
	if !entry.castShadow || m.shadowView == nil || m.shadowViewMiss {
		return
	}
	now := time.Now()
	if m.actorShadowSuppressed(entry.actor, now) {
		return
	}
	scale := entry.scale * entry.shadowScale
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		return
	}
	drawFixedSpriteBillboardAlpha(screen, m.shadowView, entry.shadowX, entry.shadowY, scale, m.actorDeathAlpha(entry.actor.ID, now), entry.shadow)
}

func appendActorDrawEntry(entries []sceneActorDrawEntry, world *worldstate.World, projection sceneProjection, actor worldstate.Actor, label string, isPlayer bool, now time.Time, screenWidth, screenHeight int) []sceneActorDrawEntry {
	actorX, actorY := actor.RenderPosition(now)
	actor.Dir = actor.RenderDirection(now)
	terrainZ := terrainHeightAt(world, actorX, actorY)
	point := projection.Project(cellCenter(actorX), cellCenter(actorY), terrainZ)
	if point.x < -96 || point.y < -160 || point.x > float32(screenWidth+96) || point.y > float32(screenHeight+96) {
		return entries
	}
	worldX := cellCenter(actorX)
	worldY := cellCenter(actorY)
	depth := actorBillboardSortDepth(projection, worldX, worldY, terrainZ)
	shadowDepth := projection.Depth(worldX, worldY, terrainZ+0.05)
	shadowPoint := projection.Project(worldX, worldY, terrainZ+0.05)
	return append(entries, sceneActorDrawEntry{
		actor:       actor,
		label:       label,
		screenX:     float64(point.x),
		screenY:     float64(point.y),
		worldX:      worldX,
		worldY:      worldY,
		worldZ:      terrainZ,
		scale:       actorBillboardScreenScale(projection, worldX, worldY, terrainZ),
		shadow:      actorShadowFactor(world, actorX, actorY),
		castShadow:  actorCastsShadow(actor),
		shadowX:     float64(shadowPoint.x),
		shadowY:     float64(shadowPoint.y),
		shadowScale: actorShadowSize(actor),
		shadowDepth: shadowDepth,
		depth:       depth,
		isPlayer:    isPlayer,
	})
}

func actorCastsShadow(actor worldstate.Actor) bool {
	if isWarpActor(actor) || int(actor.Job) == actorJobClearNPC {
		return false
	}
	return actorShadowSize(actor) > 0
}

func (m *WorldMode) actorShadowSuppressed(actor worldstate.Actor, now time.Time) bool {
	if anim, ok := m.actorAnimation(actor.ID, now); ok {
		switch anim.actionFamily {
		case spriteActionSit, spriteActionPCDeath, spriteActionNonPCDeath:
			return true
		}
	}
	return false
}

func actorShadowSize(actor worldstate.Actor) float64 {
	if size, ok := monsterShadowSize[int(actor.Job)]; ok {
		return size
	}
	return 1
}

func actorShadowFactor(world *worldstate.World, x, y float64) float64 {
	if world == nil || world.GND == nil {
		return 1
	}
	shadowX, shadowY := gndShadowMapPoint(x, y)
	total := 0
	for dy := -3; dy < 3; dy++ {
		for dx := -3; dx < 3; dx++ {
			total += int(gndShadowMapAlpha(world.GND, shadowX+dx, shadowY+dy))
		}
	}
	return clampUnit(float64(total) / (6 * 6 * 255))
}

func gndShadowMapPoint(x, y float64) (int, int) {
	x += 0.5
	y += 0.5
	shadowX := int(math.Floor(x/2)) * 8
	shadowY := int(math.Floor(y/2)) * 8
	localX := 0
	if int(x)&1 != 0 {
		localX = 4
	}
	localY := 0
	if int(y)&1 != 0 {
		localY = 4
	}
	localX += int(math.Floor((x - math.Floor(x)) * 4))
	localY += int(math.Floor((y - math.Floor(y)) * 4))
	shadowX += minInt(localX, 6)
	shadowY += minInt(localY, 6)
	return shadowX, shadowY
}

func gndShadowMapAlpha(gnd *res.GND, shadowX, shadowY int) uint8 {
	if gnd == nil || shadowX < 0 || shadowY < 0 || shadowX >= gnd.Width*8 || shadowY >= gnd.Height*8 {
		return 255
	}
	cellX := shadowX / 8
	cellY := shadowY / 8
	localX := shadowX % 8
	localY := shadowY % 8
	cell, ok := gnd.Cell(cellX, cellY)
	if !ok || cell.Top < 0 {
		return 255
	}
	surface, ok := gnd.Surface(cell.Top)
	if !ok {
		return 255
	}
	lightmap, ok := gnd.Lightmap(surface.LightmapID)
	if !ok {
		return 255
	}
	return lightmap.Alpha[localY][localX]
}

func actorBillboardSortDepth(projection sceneProjection, x, y, z float64) float64 {
	footDepth := projection.Depth(x, y, z)
	if !projection.camera {
		return footDepth
	}
	topDepth := projection.Depth(x, y, z+actorBillboardWorldHeightUnit)
	if topDepth <= 0 || !isFinite(topDepth) {
		return footDepth
	}
	return math.Min(footDepth, topDepth)
}

func actorDisplayName(ctx Context, actor worldstate.Actor, isPlayer bool) string {
	if isPlayer {
		if name := sanitizeActorName(selectedCharacterName(ctx.Session)); name != "" {
			return name
		}
		return sanitizeActorName(actor.Name)
	}
	if isWarpActor(actor) {
		return ""
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

func isWarpActor(actor worldstate.Actor) bool {
	return actor.Job == actorJobWarpPortal
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

func (m *WorldMode) drawActorLifeBar(screen *ebiten.Image, ctx Context, entry sceneActorDrawEntry) {
	life, ok := m.actorLifeForDisplay(ctx, entry.actor)
	if !ok {
		return
	}
	ratio := float64(life.hp) / float64(life.maxHP)
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	const width = 60.0
	const height = 5.0
	x := math.Round(entry.screenX - width/2)
	y := math.Round(entry.screenY + 3*entry.scale)
	fillWidth := math.Round((width - 2) * ratio)
	fill := color.RGBA{R: 255, G: 0, B: 231, A: 255}
	if ratio < 0.25 {
		fill = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	}
	ebitenutil.DrawRect(screen, x, y, width, height, color.RGBA{R: 16, G: 24, B: 156, A: 255})
	ebitenutil.DrawRect(screen, x+1, y+1, width-2, height-2, color.RGBA{R: 66, G: 66, B: 66, A: 255})
	if fillWidth > 0 {
		ebitenutil.DrawRect(screen, x+1, y+1, fillWidth, 3, fill)
	}
}

func (m *WorldMode) actorLifeForDisplay(ctx Context, actor worldstate.Actor) (actorLife, bool) {
	if actor.ID == 0 || m.actorLife == nil {
		return actorLife{}, false
	}
	if !actorCanBeAttackClicked(ctx, actor) {
		return actorLife{}, false
	}
	life, ok := m.actorLife[actor.ID]
	if !ok || life.maxHP <= 0 || life.hp < 0 {
		return actorLife{}, false
	}
	return life, true
}

func (m *WorldMode) drawActorSprite(screen *ebiten.Image, ctx Context, actor worldstate.Actor, centerX, centerY, scale float64, cameraYaw float64, shadow float64) bool {
	if !res.HasPlayerJobToken(int(actor.Job)) {
		return m.drawNonPCSprite(screen, ctx, actor, centerX, centerY, scale, cameraYaw, shadow)
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
	now := time.Now()
	state := spriteState{
		actionFamily: spriteActionIdle,
		direction:    actor.Dir,
		cameraYaw:    cameraYaw,
		moving:       actor.IsMovingAt(now),
		moveSpeedMS:  actor.Speed,
	}
	if state.moving {
		state.actionFamily = spriteActionWalk
		state.loop = true
		state.walkDistance = actor.RenderWalkDistance(now)
	}
	if anim, ok := m.actorAnimation(actor.ID, now); ok {
		state.actionFamily = anim.actionFamily
		state.started = anim.started
		state.loop = false
		state.moving = false
	}
	return drawHumanoidBillboard(screen, view, state, centerX, centerY, scale, shadow)
}

func (m *WorldMode) drawNonPCSprite(screen *ebiten.Image, ctx Context, actor worldstate.Actor, centerX, centerY, scale float64, cameraYaw float64, shadow float64) bool {
	view := m.nonPCSpriteView(ctx, actor)
	if view == nil {
		return false
	}
	now := time.Now()
	state := m.nonPCSpriteState(actor, now)
	state.cameraYaw = cameraYaw
	return drawSingleSpriteBillboardAlpha(screen, view, state, centerX, centerY, scale, m.actorDeathAlpha(actor.ID, now), shadow)
}

func (m *WorldMode) nonPCSpriteState(actor worldstate.Actor, now time.Time) spriteState {
	state := spriteState{
		actionFamily: spriteActionIdle,
		direction:    actor.Dir,
		moving:       actor.IsMovingAt(now),
		loopIdle:     true,
		moveSpeedMS:  actor.Speed,
	}
	if state.moving {
		state.actionFamily = spriteActionWalk
		state.loop = true
		state.walkDistance = actor.RenderWalkDistance(now)
	}
	if anim, ok := m.actorAnimation(actor.ID, now); ok {
		state.actionFamily = anim.actionFamily
		state.started = anim.started
		state.loop = false
		state.moving = false
		state.loopIdle = false
	}
	return state
}

func (m *WorldMode) nonPCSpriteView(ctx Context, actor worldstate.Actor) *playerSpriteView {
	job := int(actor.Job)
	if _, ok := m.nonPCViewMiss[job]; ok {
		return nil
	}
	if m.nonPCViews == nil {
		m.nonPCViews = make(map[int]*playerSpriteView)
	}
	view, ok := m.nonPCViews[job]
	if ok {
		return view
	}
	if ctx.Resources == nil {
		return nil
	}
	loaded, status := loadNonPCSpriteView(ctx.Resources, job, "nonpc")
	if loaded == nil {
		if m.nonPCViewMiss == nil {
			m.nonPCViewMiss = make(map[int]struct{})
		}
		m.nonPCViewMiss[job] = struct{}{}
		log.Printf("nonpc sprite unavailable id=%d job=%d: %s", actor.ID, job, status)
		return nil
	}
	m.nonPCViews[job] = loaded
	log.Printf("nonpc sprite resources id=%d job=%d %s", actor.ID, job, status)
	return loaded
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

func drawWarpZoneEffect(screen, white, ringTexture *ebiten.Image, projection sceneProjection, x, y, z float64, now time.Time) {
	const (
		segments       = 64
		ringCount      = 4
		baseRadius     = 0.25
		radiusRange    = 1.18
		bandWidth      = 0.34
		cycleSeconds   = 4.0
		bottomBaseSize = 0.95
		topBaseSize    = 1.58
		heightBase     = 1.10
		groundLift     = 0.04
	)
	z += groundLift
	seconds := float64(now.UnixNano()) / float64(time.Second)

	for i := 0; i < ringCount; i++ {
		phase := math.Mod(seconds+float64(i), cycleSeconds) / cycleSeconds
		sizeFactor := 1 - phase
		heightFactor := phase * 2
		if phase > 0.5 {
			heightFactor = (1 - phase) * 2
		}
		alpha := uint8(102 * warpCycleFade(phase))
		drawProjectedCylinderBand(
			screen,
			white,
			ringTexture,
			projection,
			x,
			y,
			z,
			bottomBaseSize*sizeFactor,
			topBaseSize*sizeFactor,
			heightBase*heightFactor,
			color.RGBA{R: 155, G: 205, B: 255, A: alpha},
			segments,
		)
	}
	drawProjectedRadialGradient(screen, white, projection, x, y, z, 0.18, 0.85, color.RGBA{R: 170, G: 210, B: 255, A: 54}, segments)
	for i := 0; i < ringCount; i++ {
		phase := math.Mod(seconds*0.55+float64(i)/ringCount, 1)
		radius := baseRadius + phase*radiusRange
		alpha := uint8(155 * (1 - phase))
		if alpha < 28 {
			alpha = 28
		}
		drawProjectedSoftRing(screen, white, projection, x, y, z, radius, bandWidth, color.RGBA{R: 185, G: 215, B: 255, A: alpha}, segments)
	}
	pulse := 0.5 + 0.5*math.Sin(seconds*2.4)
	drawProjectedSoftRing(screen, white, projection, x, y, z, 0.35+pulse*0.06, 0.26, color.RGBA{R: 235, G: 245, B: 255, A: 150}, segments)
}

func warpCycleFade(phase float64) float64 {
	switch {
	case phase < 0.25:
		return phase / 0.25
	case phase > 0.75:
		return (1 - phase) / 0.25
	default:
		return 1
	}
}

func drawProjectedRadialGradient(screen, white *ebiten.Image, projection sceneProjection, x, y, z, innerRadius, outerRadius float64, c color.RGBA, segments int) {
	drawProjectedRingBand(screen, white, projection, x, y, z, innerRadius, outerRadius, c.A, 0, c, segments)
}

func drawProjectedSoftRing(screen, white *ebiten.Image, projection sceneProjection, x, y, z, radius, width float64, c color.RGBA, segments int) {
	inner := math.Max(0, radius-width*0.5)
	mid := math.Max(inner+0.01, radius)
	outer := math.Max(mid+0.01, radius+width*0.5)
	drawProjectedRingBand(screen, white, projection, x, y, z, inner, mid, 0, c.A, c, segments)
	drawProjectedRingBand(screen, white, projection, x, y, z, mid, outer, c.A, 0, c, segments)
}

func drawProjectedCylinderBand(screen, white, texture *ebiten.Image, projection sceneProjection, x, y, z, bottomRadius, topRadius, height float64, c color.RGBA, segments int) {
	if segments < 3 || bottomRadius <= 0.01 || topRadius <= 0.01 || height <= 0.01 || c.A == 0 {
		return
	}
	vertices := make([]ebiten.Vertex, 0, (segments+1)*2)
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
		angle := float64(i) * 2 * math.Pi / float64(segments)
		cosine := math.Cos(angle)
		sine := math.Sin(angle)
		vertices = append(vertices,
			warpEffectTexturedVertex(projection.Project(x+cosine*bottomRadius, y+sine*bottomRadius, z), u*srcW, srcH, tint),
			warpEffectTexturedVertex(projection.Project(x+cosine*topRadius, y+sine*topRadius, z+height), u*srcW, 0, tint),
		)
		if i == segments {
			continue
		}
		base := uint16(i * 2)
		indices = append(indices, base, base+1, base+3, base, base+3, base+2)
	}
	screen.DrawTriangles(vertices, indices, source, &ebiten.DrawTrianglesOptions{
		Blend:   ebiten.BlendLighter,
		Filter:  ebiten.FilterLinear,
		Address: ebiten.AddressRepeat,
	})
}

func drawProjectedRingBand(screen, white *ebiten.Image, projection sceneProjection, x, y, z, innerRadius, outerRadius float64, innerAlpha, outerAlpha uint8, c color.RGBA, segments int) {
	if segments < 3 || outerRadius <= innerRadius {
		return
	}
	vertices := make([]ebiten.Vertex, 0, (segments+1)*2)
	indices := make([]uint16, 0, segments*6)
	innerColor := c
	outerColor := c
	innerColor.A = innerAlpha
	outerColor.A = outerAlpha
	for i := 0; i <= segments; i++ {
		angle := float64(i) * 2 * math.Pi / float64(segments)
		cosine := math.Cos(angle)
		sine := math.Sin(angle)
		vertices = append(vertices,
			warpEffectVertex(projection.Project(x+cosine*innerRadius, y+sine*innerRadius, z), innerColor),
			warpEffectVertex(projection.Project(x+cosine*outerRadius, y+sine*outerRadius, z), outerColor),
		)
		if i == segments {
			continue
		}
		base := uint16(i * 2)
		indices = append(indices, base, base+1, base+3, base, base+3, base+2)
	}
	screen.DrawTriangles(vertices, indices, white, &ebiten.DrawTrianglesOptions{
		Blend: ebiten.BlendLighter,
	})
}

func warpEffectTexturedVertex(point screenPoint, srcX, srcY float32, c color.RGBA) ebiten.Vertex {
	return ebiten.Vertex{
		DstX:   point.x,
		DstY:   point.y,
		SrcX:   srcX,
		SrcY:   srcY,
		ColorR: float32(c.R) / 255,
		ColorG: float32(c.G) / 255,
		ColorB: float32(c.B) / 255,
		ColorA: float32(c.A) / 255,
	}
}

func warpEffectVertex(point screenPoint, c color.RGBA) ebiten.Vertex {
	return ebiten.Vertex{
		DstX:   point.x,
		DstY:   point.y,
		SrcX:   0,
		SrcY:   0,
		ColorR: float32(c.R) / 255,
		ColorG: float32(c.G) / 255,
		ColorB: float32(c.B) / 255,
		ColorA: float32(c.A) / 255,
	}
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

func (m *WorldMode) drawGND(screen *ebiten.Image, manager *res.Manager, gnd *res.GND, rsw *res.RSW, projection sceneProjection, now time.Time) {
	if m.whitePixel == nil {
		m.whitePixel = ebiten.NewImage(1, 1)
		m.whitePixel.Fill(color.White)
	}

	width := screen.Bounds().Dx()
	height := screen.Bounds().Dy()
	startX, endX, startY, endY, ok := gndDrawBounds(gnd, projection, width, height)
	if !ok {
		return
	}
	lighting := sceneLightingFromRSW(rsw)
	surfaces := make([]gndSurfaceDraw, 0, (endX-startX+1)*(endY-startY+1))

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
					verts := [4]modelPoint3{
						{x: float64(x) * 2, y: float64(cell.Heights[0]), z: float64(y) * 2},
						{x: float64(x+1) * 2, y: float64(cell.Heights[1]), z: float64(y) * 2},
						{x: float64(x+1) * 2, y: float64(cell.Heights[3]), z: float64(y+1) * 2},
						{x: float64(x) * 2, y: float64(cell.Heights[2]), z: float64(y+1) * 2},
					}
					surfaces = append(surfaces, newGNDSurfaceDraw(projection, verts, surfaceUVs(surface, vertexOrder), vertexOrder, []uint16{0, 1, 2, 0, 2, 3}, surface, cell.Heights, lighting))
					if waterDraw, ok := newGNDWaterDraw(projection, x, y, cell, gnd, rsw, now); ok {
						surfaces = append(surfaces, waterDraw)
					}
				}
			}

			if cell.Front >= 0 && y+1 < gnd.Height {
				neighbor, neighborOK := gnd.Cell(x, y+1)
				surface, surfaceOK := gnd.Surface(cell.Front)
				if neighborOK && surfaceOK {
					vertexOrder := [4]int{0, 1, 3, 2}
					verts := [4]modelPoint3{
						{x: float64(x) * 2, y: float64(cell.Heights[2]), z: float64(y+1) * 2},
						{x: float64(x+1) * 2, y: float64(cell.Heights[3]), z: float64(y+1) * 2},
						{x: float64(x+1) * 2, y: float64(neighbor.Heights[1]), z: float64(y+1) * 2},
						{x: float64(x) * 2, y: float64(neighbor.Heights[0]), z: float64(y+1) * 2},
					}
					heights := [4]float32{cell.Heights[2], cell.Heights[3], neighbor.Heights[1], neighbor.Heights[0]}
					surfaces = append(surfaces, newGNDSurfaceDraw(projection, verts, surfaceUVs(surface, vertexOrder), vertexOrder, []uint16{0, 1, 2, 0, 2, 3}, surface, heights, lighting))
				}
			}

			if cell.Right >= 0 && x+1 < gnd.Width {
				neighbor, neighborOK := gnd.Cell(x+1, y)
				surface, surfaceOK := gnd.Surface(cell.Right)
				if neighborOK && surfaceOK {
					vertexOrder := [4]int{0, 1, 3, 2}
					verts := [4]modelPoint3{
						{x: float64(x+1) * 2, y: float64(cell.Heights[3]), z: float64(y+1) * 2},
						{x: float64(x+1) * 2, y: float64(cell.Heights[1]), z: float64(y) * 2},
						{x: float64(x+1) * 2, y: float64(neighbor.Heights[0]), z: float64(y) * 2},
						{x: float64(x+1) * 2, y: float64(neighbor.Heights[2]), z: float64(y+1) * 2},
					}
					heights := [4]float32{cell.Heights[3], cell.Heights[1], neighbor.Heights[0], neighbor.Heights[2]}
					surfaces = append(surfaces, newGNDSurfaceDraw(projection, verts, surfaceUVs(surface, vertexOrder), vertexOrder, []uint16{0, 1, 2, 0, 2, 3}, surface, heights, lighting))
				}
			}
		}
	}
	sortGNDSurfaces(surfaces)
	for _, surface := range surfaces {
		m.drawGNDSurface(screen, manager, gnd, surface, float64(width), float64(height))
	}
}

func gndDrawBounds(gnd *res.GND, projection sceneProjection, screenWidth, screenHeight int) (int, int, int, int, bool) {
	if gnd == nil || gnd.Width <= 0 || gnd.Height <= 0 {
		return 0, 0, 0, 0, false
	}

	centerX := gndTileFromWorld(projection.playerX)
	centerY := gndTileFromWorld(projection.playerY)
	if projection.camera {
		if minWorldX, maxWorldX, minWorldY, maxWorldY, ok := cameraGroundFootprint(projection, screenWidth, screenHeight); ok {
			const margin = 24
			startX := minInt(gndTileFromWorld(minWorldX), centerX) - margin
			endX := maxInt(gndTileFromWorld(maxWorldX), centerX) + margin
			startY := minInt(gndTileFromWorld(minWorldY), centerY) - margin
			endY := maxInt(gndTileFromWorld(maxWorldY), centerY) + margin
			return clampGNDRange(gnd, startX, endX, startY, endY)
		}
	}

	radiusX := int(float64(screenWidth)/projection.tileW) + 12
	radiusY := int(float64(screenHeight)/projection.tileH) + 12
	return clampGNDRange(gnd, centerX-radiusX, centerX+radiusX, centerY-radiusY, centerY+radiusY)
}

func gndTileFromWorld(coord float64) int {
	return int(math.Floor(coord * 0.5))
}

func clampGNDRange(gnd *res.GND, startX, endX, startY, endY int) (int, int, int, int, bool) {
	startX = maxInt(0, startX)
	endX = minInt(gnd.Width-1, endX)
	startY = maxInt(0, startY)
	endY = minInt(gnd.Height-1, endY)
	return startX, endX, startY, endY, startX <= endX && startY <= endY
}

func cameraGroundFootprint(projection sceneProjection, screenWidth, screenHeight int) (float64, float64, float64, float64, bool) {
	aspect := 1.0
	if screenHeight > 0 {
		aspect = float64(screenWidth) / float64(screenHeight)
	}

	distance := sceneCameraZoom() * 0.5
	pitch := sceneCameraPitch()
	if pitch > 180 {
		pitch -= 180
	}
	pitch = degreesToRadians(pitch)
	yaw := degreesToRadians(projection.cameraYaw)
	horizontal := math.Cos(pitch) * distance
	eye := modelPoint3{
		x: projection.playerX + math.Sin(yaw)*horizontal,
		y: projection.playerZ + math.Sin(pitch)*distance,
		z: projection.playerY - math.Cos(yaw)*horizontal,
	}
	target := modelPoint3{x: projection.playerX, y: projection.playerZ, z: projection.playerY}
	forward := normalize3(sub3(target, eye))
	right := normalize3(cross3(modelPoint3{y: 1}, forward))
	if right == (modelPoint3{}) {
		right = modelPoint3{x: 1}
	}
	up := cross3(forward, right)
	tanHalfFOV := math.Tan(degreesToRadians(sceneCameraFOV()) * 0.5)

	samples := [][2]float64{
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1},
		{0, 0}, {-1, 0}, {1, 0}, {0, -1}, {0, 1},
	}
	minX, maxX := 0.0, 0.0
	minY, maxY := 0.0, 0.0
	found := false
	for _, sample := range samples {
		dir := normalize3(add3(add3(forward, mul3(right, sample[0]*tanHalfFOV*aspect)), mul3(up, sample[1]*tanHalfFOV)))
		if math.Abs(dir.y) < 0.000001 {
			continue
		}
		t := (projection.playerZ - eye.y) / dir.y
		if t <= 0 || !isFinite(t) {
			continue
		}
		hit := add3(eye, mul3(dir, t))
		if !isFinite(hit.x) || !isFinite(hit.z) {
			continue
		}
		if !found {
			minX, maxX = hit.x, hit.x
			minY, maxY = hit.z, hit.z
			found = true
			continue
		}
		minX = math.Min(minX, hit.x)
		maxX = math.Max(maxX, hit.x)
		minY = math.Min(minY, hit.z)
		maxY = math.Max(maxY, hit.z)
	}
	return minX, maxX, minY, maxY, found
}

type gndSurfaceDraw struct {
	points      [4]screenPoint
	uvs         [4]texturePoint
	vertexOrder [4]int
	indices     []uint16
	surface     res.GNDSurface
	heights     [4]float32
	normal      modelPoint3
	lighting    sceneLighting
	water       bool
	waterType   int
	waterFrame  int
	tint        color.RGBA
	depth       float64
}

func sortGNDSurfaces(surfaces []gndSurfaceDraw) {
	sort.SliceStable(surfaces, func(i, j int) bool {
		return surfaces[i].depth > surfaces[j].depth
	})
}

func newGNDWaterDraw(projection sceneProjection, x, y int, cell res.GNDCell, gnd *res.GND, rsw *res.RSW, now time.Time) (gndSurfaceDraw, bool) {
	water, ok := mapWater(gnd, rsw)
	if !ok {
		return gndSurfaceDraw{}, false
	}
	if !waterVisibleForCell(cell, water) {
		return gndSurfaceDraw{}, false
	}
	heights := waterHeightsForCell(water, x, y, now)
	verts := [4]modelPoint3{
		{x: float64(x) * 2, y: float64(heights[0]), z: float64(y) * 2},
		{x: float64(x+1) * 2, y: float64(heights[1]), z: float64(y) * 2},
		{x: float64(x+1) * 2, y: float64(heights[3]), z: float64(y+1) * 2},
		{x: float64(x) * 2, y: float64(heights[2]), z: float64(y+1) * 2},
	}
	draw := newGNDSurfaceDraw(
		projection,
		verts,
		waterUVs(x, y),
		[4]int{0, 1, 3, 2},
		[]uint16{0, 1, 2, 0, 2, 3},
		res.GNDSurface{},
		heights,
		sceneLighting{},
	)
	draw.water = true
	draw.waterType = int(water.Type)
	draw.waterFrame = waterFrameForTime(water, now)
	draw.tint = waterTint(water, rsw)
	return draw, true
}

func mapWater(gnd *res.GND, rsw *res.RSW) (res.RSWWater, bool) {
	if gnd != nil && gnd.Water.Present {
		return res.RSWWater{
			Level:      gnd.Water.Level,
			Type:       gnd.Water.Type,
			WaveHeight: gnd.Water.WaveHeight,
			WaveSpeed:  gnd.Water.WaveSpeed,
			WavePitch:  gnd.Water.WavePitch,
			AnimSpeed:  gnd.Water.AnimSpeed,
		}, true
	}
	if rsw != nil {
		return rsw.Water, true
	}
	return res.RSWWater{}, false
}

func newGNDSurfaceDraw(projection sceneProjection, verts [4]modelPoint3, uvs [4]texturePoint, vertexOrder [4]int, indices []uint16, surface res.GNDSurface, heights [4]float32, lighting sceneLighting) gndSurfaceDraw {
	depth := 0.0
	for _, vert := range verts {
		depth += projection.Depth(vert.x, vert.z, vert.y)
	}
	return gndSurfaceDraw{
		points:      projectGNDQuad(projection, verts),
		uvs:         uvs,
		vertexOrder: vertexOrder,
		indices:     indices,
		surface:     surface,
		heights:     heights,
		normal:      quadNormal(verts),
		lighting:    lighting,
		depth:       depth / float64(len(verts)),
	}
}

func (m *WorldMode) drawGNDSurface(screen *ebiten.Image, manager *res.Manager, gnd *res.GND, draw gndSurfaceDraw, screenWidth, screenHeight float64) {
	if quadHasInvalidPoint(draw.points) {
		return
	}
	if quadOutside(draw.points, screenWidth, screenHeight) {
		return
	}
	if draw.water {
		m.drawWaterSurface(screen, manager, draw)
		return
	}

	textureName := gndTextureName(gnd, draw.surface.TextureID)
	if texture := m.groundTexture(manager, textureName); texture != nil {
		if lightmap, ok := gnd.Lightmap(draw.surface.LightmapID); ok {
			drawTexturedLightmappedSurface(screen, texture, draw.points, draw.uvs, draw.surface.Color, lightmap, draw.lighting.groundScale(draw.normal))
			return
		}
		drawTexturedSurface(screen, texture, draw.points, draw.uvs, draw.indices, surfaceVertexTints(gnd, draw.surface, draw.vertexOrder, draw.heights, draw.normal, draw.lighting))
		return
	}
	drawColoredSurface(screen, m.whitePixel, draw.points, draw.indices, groundSurfaceColor(textureName, draw.surface.Color, draw.heights, draw.normal, draw.lighting))
}

func (m *WorldMode) drawWaterSurface(screen *ebiten.Image, manager *res.Manager, draw gndSurfaceDraw) {
	texture := m.waterTexture(manager, draw.waterType, draw.waterFrame)
	if texture == nil {
		drawColoredSurface(screen, m.whitePixel, draw.points, draw.indices, draw.tint)
		return
	}
	tints := [4]color.RGBA{draw.tint, draw.tint, draw.tint, draw.tint}
	drawTexturedSurface(screen, texture, draw.points, draw.uvs, draw.indices, tints)
}

func (m *WorldMode) waterTexture(manager *res.Manager, waterType, frame int) *ebiten.Image {
	frame = ((frame % 32) + 32) % 32
	key := fmt.Sprintf("__water_%d_%02d", waterType, frame)
	if texture, ok := m.textures[key]; ok {
		return texture
	}
	if _, ok := m.textureMiss[key]; ok {
		return nil
	}
	img, _, err := res.LoadImage(manager, res.WaterTextureCandidates(waterType, frame))
	if err != nil && waterType >= 0 {
		img, _, err = res.LoadImage(manager, res.WaterTextureCandidates(waterType%6, frame))
	}
	if err != nil {
		m.textureMiss[key] = struct{}{}
		return nil
	}
	texture := ebiten.NewImageFromImage(img)
	m.textures[key] = texture
	return texture
}

func (m *WorldMode) effectTexture(manager *res.Manager, name string) *ebiten.Image {
	if manager == nil || strings.TrimSpace(name) == "" {
		return nil
	}
	key := "__effect_" + strings.TrimSpace(name)
	if m.textures == nil {
		m.textures = make(map[string]*ebiten.Image)
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
	img, _, err := res.LoadImage(manager, res.EffectTextureCandidates(name))
	if err != nil {
		m.textureMiss[key] = struct{}{}
		return nil
	}
	texture := ebiten.NewImageFromImage(res.ApplyEffectTransparency(img))
	m.textures[key] = texture
	return texture
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

func quadHasInvalidPoint(points [4]screenPoint) bool {
	for _, point := range points {
		if !isFinite(float64(point.x)) || !isFinite(float64(point.y)) {
			return true
		}
		if point.x <= -1<<19 && point.y <= -1<<19 {
			return true
		}
	}
	return false
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

func waterUVs(x, y int) [4]texturePoint {
	const scale = 0.25
	baseU := float32(x&3) * scale
	baseV := float32(y&3) * scale
	return [4]texturePoint{
		{u: baseU, v: baseV},
		{u: baseU + scale, v: baseV},
		{u: baseU + scale, v: baseV + scale},
		{u: baseU, v: baseV + scale},
	}
}

func waterHeightsForCell(water res.RSWWater, x, y int, now time.Time) [4]float32 {
	level := water.Level
	if water.WaveHeight == 0 {
		return [4]float32{level, level, level, level}
	}
	offset := waterOffsetForTime(water, now)
	pitch := float64(water.WavePitch)
	diagonal := float64(x + y)
	h1 := waterSin(offset+pitch*diagonal)*water.WaveHeight + level
	h0 := waterSin(offset+pitch*(diagonal-1))*water.WaveHeight + level
	h3 := waterSin(offset+pitch*(diagonal+1))*water.WaveHeight + level
	return [4]float32{h0, h1, h1, h3}
}

func waterVisibleForCell(cell res.GNDCell, water res.RSWWater) bool {
	threshold := water.Level + water.WaveHeight
	return cell.Heights[0] < threshold ||
		cell.Heights[1] < threshold ||
		cell.Heights[2] < threshold ||
		cell.Heights[3] < threshold
}

func waterFrameForTime(water res.RSWWater, now time.Time) int {
	animSpeed := int(water.AnimSpeed)
	if animSpeed <= 0 {
		animSpeed = 1
	}
	frame := int(now.UnixMilli()*60/1000) / animSpeed
	return frame % 32
}

func waterOffsetForTime(water res.RSWWater, now time.Time) float64 {
	offset := math.Mod(float64(now.UnixMilli()*60/1000)*float64(water.WaveSpeed), 360)
	if offset > 180 {
		offset -= 360
	}
	return offset
}

func waterSin(degrees float64) float32 {
	return float32(math.Sin(degreesToRadians(degrees)))
}

func waterTint(water res.RSWWater, rsw *res.RSW) color.RGBA {
	alpha := uint8(204)
	if water.Type == 4 || water.Type == 6 {
		alpha = 255
	}
	if rsw != nil && water.Type == 4 {
		return color.RGBA{
			R: clampColor(float64(rsw.Light.Ambient[0]) * 255),
			G: clampColor(float64(rsw.Light.Ambient[1]) * 255),
			B: clampColor(float64(rsw.Light.Ambient[2]) * 255),
			A: alpha,
		}
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: alpha}
}

func projectGNDQuad(projection sceneProjection, verts [4]modelPoint3) [4]screenPoint {
	return [4]screenPoint{
		projection.Project(verts[0].x, verts[0].z, verts[0].y),
		projection.Project(verts[1].x, verts[1].z, verts[1].y),
		projection.Project(verts[2].x, verts[2].z, verts[2].y),
		projection.Project(verts[3].x, verts[3].z, verts[3].y),
	}
}

func quadNormal(verts [4]modelPoint3) modelPoint3 {
	normal := normalize3(cross3(sub3(verts[1], verts[0]), sub3(verts[2], verts[0])))
	if normal == (modelPoint3{}) {
		return modelPoint3{y: 1}
	}
	return normal
}

type sceneLighting struct {
	direction modelPoint3
	diffuse   modelPoint3
	ambient   modelPoint3
	env       modelPoint3
	opacity   float64
}

func sceneLightingFromRSW(rsw *res.RSW) sceneLighting {
	longitude := 45.0
	latitude := 45.0
	diffuse := modelPoint3{x: 1, y: 1, z: 1}
	ambient := modelPoint3{}
	opacity := 1.0
	if rsw != nil {
		longitude = float64(rsw.Light.Longitude)
		latitude = float64(rsw.Light.Latitude)
		diffuse = modelPoint3{x: float64(rsw.Light.Diffuse[0]), y: float64(rsw.Light.Diffuse[1]), z: float64(rsw.Light.Diffuse[2])}
		ambient = modelPoint3{x: float64(rsw.Light.Ambient[0]), y: float64(rsw.Light.Ambient[1]), z: float64(rsw.Light.Ambient[2])}
		opacity = float64(rsw.Light.Opacity)
	}
	longitude = degreesToRadians(longitude)
	latitude = degreesToRadians(latitude)
	dir := normalize3(modelPoint3{
		x: -math.Sin(longitude) * math.Sin(latitude),
		y: -math.Cos(latitude),
		z: -math.Cos(longitude) * math.Sin(latitude),
	})
	if dir == (modelPoint3{}) {
		dir = normalize3(modelPoint3{x: -0.5, y: -0.7, z: -0.5})
	}
	diffuse = clampUnitPoint(diffuse)
	ambient = clampUnitPoint(ambient)
	opacity = math.Max(0, math.Min(1, opacity))
	env := modelPoint3{
		x: 1 - (1-ambient.x)*(1-diffuse.x),
		y: 1 - (1-ambient.y)*(1-diffuse.y),
		z: 1 - (1-ambient.z)*(1-diffuse.z),
	}
	return sceneLighting{
		direction: dir,
		diffuse:   diffuse,
		ambient:   ambient,
		env:       clampUnitPoint(env),
		opacity:   opacity,
	}
}

func (l sceneLighting) groundScale(normal modelPoint3) modelPoint3 {
	weight := math.Max(dot3(normalize3(normal), l.direction), 0)
	return modelPoint3{
		x: clampUnit((l.ambient.x + l.diffuse.x*weight) * l.env.x),
		y: clampUnit((l.ambient.y + l.diffuse.y*weight) * l.env.y),
		z: clampUnit((l.ambient.z + l.diffuse.z*weight) * l.env.z),
	}
}

func (l sceneLighting) modelScale(normal modelPoint3) modelPoint3 {
	weight := math.Max(dot3(normalize3(normal), l.direction), 0)
	weight = (1 - l.opacity) + weight*l.opacity
	return modelPoint3{
		x: clampUnit((l.ambient.x + l.diffuse.x*weight) * l.env.x),
		y: clampUnit((l.ambient.y + l.diffuse.y*weight) * l.env.y),
		z: clampUnit((l.ambient.z + l.diffuse.z*weight) * l.env.z),
	}
}

func clampUnitPoint(point modelPoint3) modelPoint3 {
	return modelPoint3{x: clampUnit(point.x), y: clampUnit(point.y), z: clampUnit(point.z)}
}

func clampUnit(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

func surfaceTint(surfaceColor color.RGBA, heights [4]float32, normal modelPoint3, lighting sceneLighting) color.RGBA {
	base := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if surfaceColor.A != 0 {
		base = surfaceColor
	}
	scale := lighting.groundScale(normal)
	heightShade := groundHeightShade(heights)
	return color.RGBA{
		R: clampColor(float64(base.R) * scale.x * heightShade),
		G: clampColor(float64(base.G) * scale.y * heightShade),
		B: clampColor(float64(base.B) * scale.z * heightShade),
		A: 255,
	}
}

func surfaceVertexTints(gnd *res.GND, surface res.GNDSurface, vertexOrder [4]int, heights [4]float32, normal modelPoint3, lighting sceneLighting) [4]color.RGBA {
	if lightmap, ok := gnd.Lightmap(surface.LightmapID); ok {
		return lightmapSurfaceVertexTints(surface.Color, lightmap, vertexOrder, lighting.groundScale(normal))
	}
	tint := surfaceTint(surface.Color, heights, normal, lighting)
	return [4]color.RGBA{tint, tint, tint, tint}
}

func lightmapSurfaceVertexTints(surfaceColor color.RGBA, lightmap res.GNDLightmap, vertexOrder [4]int, lightScale modelPoint3) [4]color.RGBA {
	base := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if surfaceColor.A != 0 {
		base = surfaceColor
	}
	var tints [4]color.RGBA
	for i, sourceVertex := range vertexOrder {
		sample := gndLightmapRenderSample(sourceVertex)
		alpha := float64(lightmap.Alpha[sample.y][sample.x]) / 255
		lm := lightmap.Color[sample.y][sample.x]
		tints[i] = color.RGBA{
			R: clampColor(float64(base.R) * (lightScale.x*alpha + float64(lm.R)/255)),
			G: clampColor(float64(base.G) * (lightScale.y*alpha + float64(lm.G)/255)),
			B: clampColor(float64(base.B) * (lightScale.z*alpha + float64(lm.B)/255)),
			A: 255,
		}
	}
	return tints
}

func gndLightmapRenderSample(sourceVertex int) struct{ y, x int } {
	switch sourceVertex {
	case 0:
		return struct{ y, x int }{y: 1, x: 1}
	case 1:
		return struct{ y, x int }{y: 1, x: 7}
	case 2:
		return struct{ y, x int }{y: 7, x: 1}
	case 3:
		return struct{ y, x int }{y: 7, x: 7}
	default:
		return struct{ y, x int }{y: 1, x: 1}
	}
}

func groundSurfaceColor(textureName string, surfaceColor color.RGBA, heights [4]float32, normal modelPoint3, lighting sceneLighting) color.RGBA {
	base := textureColor(textureName)
	if surfaceColor.A != 0 && !(surfaceColor.R == 255 && surfaceColor.G == 255 && surfaceColor.B == 255) {
		base.R = uint8((uint16(base.R)*2 + uint16(surfaceColor.R)) / 3)
		base.G = uint8((uint16(base.G)*2 + uint16(surfaceColor.G)) / 3)
		base.B = uint8((uint16(base.B)*2 + uint16(surfaceColor.B)) / 3)
	}

	scale := lighting.groundScale(normal)
	heightShade := groundHeightShade(heights)
	return color.RGBA{
		R: clampColor(float64(base.R) * scale.x * heightShade),
		G: clampColor(float64(base.G) * scale.y * heightShade),
		B: clampColor(float64(base.B) * scale.z * heightShade),
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

func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
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

func drawTexturedLightmappedSurface(screen, texture *ebiten.Image, points [4]screenPoint, uvs [4]texturePoint, surfaceColor color.RGBA, lightmap res.GNDLightmap, lightScale modelPoint3) {
	const steps = 6
	bounds := texture.Bounds()
	textureWidth := float32(bounds.Dx())
	textureHeight := float32(bounds.Dy())
	base := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if surfaceColor.A != 0 {
		base = surfaceColor
	}

	vertices := make([]ebiten.Vertex, 0, (steps+1)*(steps+1))
	for y := 0; y <= steps; y++ {
		t := float64(y) / steps
		for x := 0; x <= steps; x++ {
			s := float64(x) / steps
			alpha := float64(res.GNDLightmapSampleAlpha(lightmap, s, t)) / 255
			lm := res.GNDLightmapSampleColor(lightmap, s, t)
			tint := color.RGBA{
				R: clampColor(float64(base.R) * (lightScale.x*alpha + float64(lm.R)/255)),
				G: clampColor(float64(base.G) * (lightScale.y*alpha + float64(lm.G)/255)),
				B: clampColor(float64(base.B) * (lightScale.z*alpha + float64(lm.B)/255)),
				A: 255,
			}
			vertices = append(vertices, texturedSurfaceVertex(
				bilerpScreenPoint(points, s, t),
				bilerpTexturePoint(uvs, s, t),
				tint,
				textureWidth,
				textureHeight,
			))
		}
	}

	indices := make([]uint16, 0, steps*steps*6)
	row := steps + 1
	for y := 0; y < steps; y++ {
		for x := 0; x < steps; x++ {
			topLeft := uint16(y*row + x)
			topRight := uint16(y*row + x + 1)
			bottomLeft := uint16((y+1)*row + x)
			bottomRight := uint16((y+1)*row + x + 1)
			indices = append(indices, topLeft, topRight, bottomRight, topLeft, bottomRight, bottomLeft)
		}
	}

	op := &ebiten.DrawTrianglesOptions{
		Filter:  ebiten.FilterLinear,
		Address: ebiten.AddressRepeat,
	}
	screen.DrawTriangles(vertices, indices, texture, op)
}

func bilerpScreenPoint(points [4]screenPoint, s, t float64) screenPoint {
	top := lerpScreenPoint(points[0], points[1], s)
	bottom := lerpScreenPoint(points[3], points[2], s)
	return lerpScreenPoint(top, bottom, t)
}

func lerpScreenPoint(a, b screenPoint, t float64) screenPoint {
	return screenPoint{
		x: float32(float64(a.x) + (float64(b.x)-float64(a.x))*t),
		y: float32(float64(a.y) + (float64(b.y)-float64(a.y))*t),
	}
}

func bilerpTexturePoint(points [4]texturePoint, s, t float64) texturePoint {
	top := lerpTexturePoint(points[0], points[1], s)
	bottom := lerpTexturePoint(points[3], points[2], s)
	return lerpTexturePoint(top, bottom, t)
}

func lerpTexturePoint(a, b texturePoint, t float64) texturePoint {
	return texturePoint{
		u: float32(float64(a.u) + (float64(b.u)-float64(a.u))*t),
		v: float32(float64(a.v) + (float64(b.v)-float64(a.v))*t),
	}
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
