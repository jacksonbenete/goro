package game

import (
	"context"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
	gameui "github.com/kivutar/goro/ui"
	worldstate "github.com/kivutar/goro/world"
)

type WorldMode struct {
	walkCooldownUntil time.Time
	nextHeldWalkAt    time.Time
	tickCooldown      int
	camera            followCamera
	cameraShakeStart  time.Time
	cameraShakeEnd    time.Time
	whitePixel        *render.Image
	tileCursor        *render.Image
	textures          map[string]*render.Image
	textureMiss       map[string]struct{}
	imageCache        map[string]image.Image
	imageMiss         map[string]struct{}
	strEffects        map[string]*res.STR
	strEffectMiss     map[string]struct{}
	playerView        *humanoidSpriteView
	shadowView        *spriteView
	shadowViewMiss    bool
	cartViews         map[int]*spriteView
	cartViewMiss      map[int]struct{}
	cursorView        *spriteView
	cursorViewMiss    bool
	cursorFallback    *render.Image
	cursorAction      int
	cursorStarted     time.Time
	damageNumberView  *spriteView
	damageNumberMiss  bool
	damageNumbers     map[string]*spriteBillboard
	damageMsgView     *spriteView
	damageMsgMiss     bool
	itemMarker        *render.Image
	itemViews         map[itemSpriteKey]*spriteView
	itemViewMiss      map[itemSpriteKey]struct{}
	effectViews       map[string]*spriteView
	effectViewMiss    map[string]struct{}
	actorViews        map[actorSpriteKey]*humanoidSpriteView
	actorViewMiss     map[actorSpriteKey]struct{}
	nonPCViews        map[int]*spriteView
	nonPCViewMiss     map[int]struct{}
	rsmMeshCache      map[int][]retainedWorldMesh
	rsmNodeMatrices   map[*res.RSM]map[string]mat4
	rsmAnimNodes      map[animatedRSMNodeKey]map[string]mat4
	rsmBoundsCache    map[rsmBoundsCacheKey]rsmBounds
	rsmFaceMetaCache  map[*res.RSM]map[*res.RSMNode][]rsmFaceMeta
	rsmPlacementGrid  *rsmPlacementGrid
	gndMeshCache      *gndRetainedMeshCache
	pendingWarp       bool
	pendingAttack     attackIntent
	pendingPickup     pickupIntent
	pendingSkill      pendingSkillTarget
	pickupReqItemID   uint32
	lockedAttackID    uint32
	attackFocusID     uint32
	attackFocusStart  time.Time
	lastAttackAt      time.Time
	lastChaseAt       time.Time
	actorAnims        map[uint32]actorAnimation
	damageFloaters    []damageFloater
	worldEffects      []worldEffect
	actorCastBars     map[uint32]actorCastBar
	scheduledSounds   []scheduledSound
	scheduledStops    []scheduledActorStop
	scheduledResumes  []scheduledWalkResume
	mapSoundNext      map[int]time.Time
	mapWeatherSounds  map[int]time.Time
	actorDeaths       map[uint32]time.Time
	actorSoundFrames  map[uint32]actorSoundFrame
	actorLife         map[uint32]actorLife
	actorNameReqAt    map[uint32]time.Time
	gndNormalSource   *res.GND
	gndTopNormals     [][4]modelPoint3
	minimap           gameui.Minimap
	statusIcons       gameui.StatusIcons
	console           gameui.ChatConsole
	npcDialog         gameui.NPCDialog
	escapeMenu        gameui.EscapeMenu
	teleportModal     gameui.TeleportModal
	deathModal        gameui.DeathModal
	friendRequest     gameui.ConfirmModal
	characterWindow   gameui.CharacterWindow
	basicMenu         gameui.BasicMenu
	inventoryBag      gameui.InventoryBagWindow
	equipmentWindow   gameui.EquipmentWindow
	storageWindow     gameui.StorageWindow
	cartWindow        gameui.CartWindow
	changeCartWindow  gameui.ChangeCartWindow
	shopWindow        gameui.ShopWindow
	vendingWindow     gameui.VendingWindow
	itemInfoWindow    gameui.ItemInfoWindow
	identifyWindow    gameui.IdentifyWindow
	statsWindow       gameui.StatsWindow
	skillWindow       gameui.SkillWindow
	friendsWindow     gameui.FriendsWindow
	playerContext     gameui.PlayerContextMenu
	settingsWindow    gameui.SettingsWindow
	shortcutBar       gameui.ShortcutBar
	mapFade           mapFadeState
	hoveredWalk       hoveredWalkCellCache
}

const (
	walkRequestCooldown    = 200 * time.Millisecond
	walkErrorCooldown      = 500 * time.Millisecond
	turnDirectionCooldown  = 100 * time.Millisecond
	heldWalkRepeatInterval = 500 * time.Millisecond
)

func (m *WorldMode) setWalkCooldown(duration time.Duration) {
	m.walkCooldownUntil = time.Now().Add(duration)
}

func (m *WorldMode) walkReady(now time.Time) bool {
	return m.walkCooldownUntil.IsZero() || !now.Before(m.walkCooldownUntil)
}

type hoveredWalkCellCache struct {
	valid bool
	key   hoveredWalkCellKey
	x     int
	y     int
	ok    bool
}

type hoveredWalkCellKey struct {
	gat        *res.GAT
	mouseX     int
	mouseY     int
	playerX    int
	playerY    int
	targetX    float64
	targetY    float64
	targetZ    float64
	cameraYaw  float64
	cameraZoom float64
	screenW    float64
	screenH    float64
}

type actorSpriteKey struct {
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

type pickupIntent struct {
	itemID  uint32
	expires time.Time
	readyAt time.Time
}

type pendingSkillTarget struct {
	skill    session.Skill
	maxLevel int
	targetID uint32
	expires  time.Time
	readyAt  time.Time
	source   string
	started  time.Time
}

type scheduledActorStop struct {
	id         uint32
	at         time.Time
	resumeWalk bool
	resumeAt   time.Time
}

type scheduledWalkResume struct {
	id  uint32
	at  time.Time
	toX int
	toY int
}

type mapFadePhase int

const (
	mapFadeNone mapFadePhase = iota
	mapFadeOut
	mapFadeHold
	mapFadeIn
)

type mapFadeState struct {
	phase     mapFadePhase
	started   time.Time
	change    network.MapChange
	hasChange bool
}

const (
	mapFadeOutDuration       = 220 * time.Millisecond
	mapFadeInDuration        = 340 * time.Millisecond
	actorNameRequestCooldown = time.Second
	defaultRSMLoadLimit      = 128
)

var (
	quadIndices012023 = []uint16{0, 1, 2, 0, 2, 3}
	quadIndices012213 = []uint16{0, 1, 2, 2, 1, 3}
)

func NewWorldMode() *WorldMode {
	return &WorldMode{}
}

func (m *WorldMode) Name() string {
	return "world"
}

func (m *WorldMode) Enter(ctx client.Context) {
	now := time.Now()
	if m.mapFade.phase == mapFadeNone {
		m.startMapFadeIn(now)
	} else if m.mapFade.started.IsZero() {
		m.mapFade.started = now
	}
	zoom := m.camera.zoom
	m.camera.Reset()
	m.camera.zoom = zoom
	ctx.World.GAT = nil
	ctx.World.GND = nil
	ctx.World.RSW = nil
	ctx.World.RSM = nil
	ctx.World.RSMFail = 0
	m.textures = make(map[string]*render.Image)
	m.textureMiss = make(map[string]struct{})
	m.playerView = nil
	m.shadowView = nil
	m.shadowViewMiss = false
	m.cursorView = nil
	m.cursorViewMiss = false
	m.cursorFallback = nil
	m.cursorAction = cursorActionDefault
	m.cursorStarted = now
	m.damageNumberView = nil
	m.damageNumberMiss = false
	m.damageNumbers = make(map[string]*spriteBillboard)
	m.damageMsgView = nil
	m.damageMsgMiss = false
	m.itemMarker = nil
	m.itemViews = make(map[itemSpriteKey]*spriteView)
	m.itemViewMiss = make(map[itemSpriteKey]struct{})
	m.effectViews = make(map[string]*spriteView)
	m.effectViewMiss = make(map[string]struct{})
	m.actorViews = make(map[actorSpriteKey]*humanoidSpriteView)
	m.actorViewMiss = make(map[actorSpriteKey]struct{})
	m.nonPCViews = make(map[int]*spriteView)
	m.nonPCViewMiss = make(map[int]struct{})
	m.rsmMeshCache = make(map[int][]retainedWorldMesh)
	m.rsmNodeMatrices = make(map[*res.RSM]map[string]mat4)
	m.rsmAnimNodes = make(map[animatedRSMNodeKey]map[string]mat4)
	m.rsmBoundsCache = make(map[rsmBoundsCacheKey]rsmBounds)
	m.rsmFaceMetaCache = make(map[*res.RSM]map[*res.RSMNode][]rsmFaceMeta)
	m.rsmPlacementGrid = nil
	m.gndMeshCache = nil
	m.pendingWarp = false
	m.pendingAttack = attackIntent{}
	m.pendingPickup = pickupIntent{}
	m.pendingSkill = pendingSkillTarget{}
	m.pickupReqItemID = 0
	m.lockedAttackID = 0
	m.clearAttackFocus()
	m.lastAttackAt = time.Time{}
	m.lastChaseAt = time.Time{}
	m.actorAnims = make(map[uint32]actorAnimation)
	m.damageFloaters = nil
	m.scheduledSounds = nil
	m.scheduledStops = nil
	m.mapSoundNext = make(map[int]time.Time)
	m.mapWeatherSounds = make(map[int]time.Time)
	m.actorDeaths = make(map[uint32]time.Time)
	m.actorSoundFrames = make(map[uint32]actorSoundFrame)
	m.actorLife = make(map[uint32]actorLife)
	m.shortcutBar.Load(ctx)
	m.npcDialog.ResetPublished(ctx)
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
	m.cartViews = make(map[int]*spriteView)
	m.cartViewMiss = make(map[int]struct{})
	if view, status := loadCursorSpriteView(ctx.Resources); view != nil {
		m.cursorView = view
		if status != "" {
			playerStatus += " " + status
		}
	} else {
		m.cursorViewMiss = true
		log.Printf("cursor resources unavailable: %s", status)
	}
	render.SetCursorMode(render.CursorModeHidden)
	log.Printf("player sprite resources char_id=%d name=%s job=%d hair=%d weapon=%d shield=%d head_top=%d head_mid=%d head_low=%d body_pal=%d head_pal=%d hair_color=%d account_sex=%d %s", character.ID, character.Name, character.Job, character.Hair, character.Weapon, character.Shield, character.HeadTop, character.HeadMid, character.HeadLow, character.BodyPal, character.HeadPal, character.HairColor, ctx.Session.Sex, playerStatus)
	m.rebindPersistentUI(ctx)
	if ctx.World.MapName == "" {
		return
	}

	gat, _, err := loadGAT(ctx.Resources, ctx.World.MapName)
	if err != nil {
		return
	}
	ctx.World.GAT = gat
	if gnd, _, err := loadGND(ctx.Resources, ctx.World.MapName); err == nil {
		ctx.World.GND = gnd
	} else {
		ctx.World.GND = nil
	}
	if rsw, rswSource, err := loadRSW(ctx.Resources, ctx.World.MapName); err == nil {
		ctx.World.RSW = rsw
		ctx.World.RSM, ctx.World.RSMFail = loadRSMModels(ctx.Resources, rsw, defaultRSMLoadLimit)
		m.playMapBGM(ctx, rswSource)
	} else {
		ctx.World.RSW = nil
		ctx.World.RSM = nil
		ctx.World.RSMFail = 0
		m.playMapBGM(ctx, ctx.World.MapName)
	}
	if err := ctx.Network.SendLoadEndAck(); err == nil {
		m.tickCooldown = 1
	}
}

func (m *WorldMode) rebindPersistentUI(ctx client.Context) {
	m.basicMenu.Rebind(ctx, m.basicMenuCallbacks(ctx))
	m.inventoryBag.Rebind(ctx, &m.itemInfoWindow, &m.cartWindow)
	m.equipmentWindow.Rebind(ctx, &m.itemInfoWindow, &m.cartWindow, m)
	m.cartWindow.Rebind(ctx, &m.itemInfoWindow)
	m.itemInfoWindow.Rebind(ctx, m)
	m.statsWindow.Rebind(ctx)
	m.skillWindow.Rebind(ctx, m)
	m.friendsWindow.Rebind(ctx)
	m.settingsWindow.Rebind(ctx)
	m.shortcutBar.ResetOverlay(ctx)
}

func (m *WorldMode) playMapBGM(ctx client.Context, rswName string) {
	if ctx.Audio == nil {
		return
	}
	_, err := ctx.Audio.PlayMap(rswName)
	if err != nil {
		log.Printf("bgm failed map=%s: %v", rswName, err)
		return
	}
}

func (m *WorldMode) Update(ctx client.Context) (Mode, error) {
	now := time.Now()
	if m.mapFade.phase == mapFadeOut {
		if !m.mapFadeElapsed(now) {
			return nil, nil
		}
		if m.mapFade.hasChange {
			change := m.mapFade.change
			m.mapFade = mapFadeState{phase: mapFadeHold, started: now}
			next := m.handleMapChange(ctx, change)
			if next != nil {
				return next, nil
			}
			if !m.pendingWarp {
				m.startMapFadeIn(now)
			}
			return nil, nil
		}
		m.startMapFadeIn(now)
	}
	if m.mapFade.phase == mapFadeIn && m.mapFadeElapsed(now) {
		m.mapFade = mapFadeState{}
	}

	for _, pkt := range ctx.Network.DrainPackets() {
		if chat, ok, err := network.ParseChatMessage(pkt); err != nil {
			log.Printf("parse chat message 0x%04X: %v", pkt.ID, err)
		} else if ok {
			addConsoleMessage(&m.console, ctx.Resources, chat)
			continue
		}
		if change, ok, err := network.ParseMapChange(pkt); err != nil {
			log.Printf("parse map change 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.startMapFadeOut(change, time.Now())
			return nil, nil
		}
		if enter, err := network.ParseMapAcceptEnter(pkt); err == nil {
			applyMapAcceptEnter(ctx, enter)
			if m.pendingWarp {
				m.pendingWarp = false
				return m.nextWorldMode(), nil
			}
			continue
		}
		if ack, ok, err := network.ParseActorNameAck(pkt); err != nil {
			log.Printf("parse actor name ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyActorNameAck(ctx, ack)
			continue
		}
		if ack, ok, err := network.ParseRestartAck(pkt); err != nil {
			log.Printf("parse restart ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			if m.deathModal.ApplyRestartAck(ack) {
				return m.nextCharacterSelectMode(ctx), nil
			}
			if m.escapeMenu.ApplyRestartAck(ack) {
				return m.nextCharacterSelectMode(ctx), nil
			}
			continue
		}
		if dialog, ok, err := network.ParseNPCDialog(pkt); err != nil {
			log.Printf("parse npc dialog 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.npcDialog.Apply(dialog)
			continue
		}
		if ack, ok, err := network.ParseSelfMoveAck(pkt); err != nil {
			log.Printf("parse self move ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applySelfMoveAck(ctx, ack)
			m.clearLocalActorAction(ctx)
			log.Printf("walk ack from=%d,%d to=%d,%d tick=%d", ack.FromX, ack.FromY, ack.ToX, ack.ToY, ack.ServerTick)
			m.continuePendingAttack(ctx, "walk ack")
			m.continuePendingPickup(ctx, "walk ack")
			m.skills().ContinuePendingTarget(ctx, "walk ack")
			continue
		}
		if position, ok, err := network.ParseActorSetPosition(pkt); err != nil {
			log.Printf("parse actor set position 0x%04X: %v", pkt.ID, err)
		} else if ok {
			if isLocalActor(ctx, position.ID) {
				log.Printf("local position fix id=%d x=%d y=%d", position.ID, position.X, position.Y)
			}
			applyActorSetPosition(ctx, position)
			if isLocalActor(ctx, position.ID) {
				m.continuePendingAttack(ctx, "position fix")
				m.continuePendingPickup(ctx, "position fix")
				m.skills().ContinuePendingTarget(ctx, "position fix")
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
			if pickup.Result == 0 {
				message := formatPickupConsoleMessage(ctx.Resources, pickup)
				log.Printf("console pickup message item_id=%d amount=%d text=%q", pickup.ItemID, pickup.Amount, message)
				m.console.AddBlueMessage("%s", message)
			} else {
				m.console.AddErrorMessage("Pickup failed item %d result=%d", pickup.ItemID, pickup.Result)
			}
			continue
		}
		if useAck, ok, err := network.ParseUseItemAck(pkt); err != nil {
			log.Printf("parse use item ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			log.Printf("use item ack index=%d item=%d aid=%d amount=%d result=%d", useAck.Index, useAck.ItemID, useAck.AID, useAck.Amount, useAck.Result)
			m.addItemUseEffect(ctx, useAck)
			applyUseItemAck(ctx, useAck)
			if useAck.Result != 0 && useAck.Amount == 0 && m.shortcutBar.ClearDepletedItem(ctx, useAck.Index, useAck.ItemID) {
				log.Printf("shortcut item depleted index=%d item=%d", useAck.Index, useAck.ItemID)
			}
			m.inventoryBag.ClampScroll(ctx.Session)
			continue
		}
		if identifyList, ok, err := network.ParseItemIdentifyList(pkt); err != nil {
			log.Printf("parse item identify list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			log.Printf("item identify list indexes=%v", identifyList.Indexes)
			m.identifyWindow.OpenList(ctx, identifyList)
			continue
		}
		if identifyAck, ok, err := network.ParseItemIdentifyAck(pkt); err != nil {
			log.Printf("parse item identify ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			log.Printf("item identify ack index=%d success=%v", identifyAck.Index, identifyAck.Success)
			applyItemIdentifyAck(ctx, identifyAck)
			m.identifyWindow.ApplyAck(ctx, identifyAck)
			m.inventoryBag.ClampScroll(ctx.Session)
			continue
		}
		if items, ok, err := network.ParseInventoryItemList(pkt); err != nil {
			log.Printf("parse inventory item list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyInventoryItemList(ctx, items)
			m.inventoryBag.ClampScroll(ctx.Session)
			continue
		}
		if itemDelete, ok, err := network.ParseInventoryItemDelete(pkt); err != nil {
			log.Printf("parse inventory item delete 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyInventoryItemDelete(ctx, itemDelete)
			m.inventoryBag.ClampScroll(ctx.Session)
			continue
		}
		if equipAck, ok, err := network.ParseInventoryEquipAck(pkt); err != nil {
			log.Printf("parse inventory equip ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyInventoryEquipAck(ctx, equipAck)
			m.inventoryBag.ClampScroll(ctx.Session)
			continue
		}
		if arrow, ok, err := network.ParseEquippedArrow(pkt); err != nil {
			log.Printf("parse equipped arrow 0x%04X: %v", pkt.ID, err)
		} else if ok {
			log.Printf("equipped arrow index=%d", arrow.Index)
			applyEquippedArrow(ctx, arrow)
			m.inventoryBag.ClampScroll(ctx.Session)
			continue
		}
		if storageItems, ok, err := network.ParseStorageItemList(pkt); err != nil {
			log.Printf("parse storage item list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyStorageItemList(ctx, storageItems)
			m.storageWindow.OpenWindow(ctx)
			continue
		}
		if cartItems, ok, err := network.ParseCartItemList(pkt); err != nil {
			log.Printf("parse cart item list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyCartItemList(ctx, cartItems)
			m.cartWindow.ClampScroll(ctx.Session)
			m.cartWindow.Refresh(ctx, &m.itemInfoWindow)
			continue
		}
		if storageAmount, ok, err := network.ParseStorageAmount(pkt); err != nil {
			log.Printf("parse storage amount 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyStorageAmount(ctx, storageAmount)
			m.storageWindow.OpenWindow(ctx)
			continue
		}
		if cartAmount, ok, err := network.ParseCartAmount(pkt); err != nil {
			log.Printf("parse cart amount 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyCartAmount(ctx, cartAmount)
			m.cartWindow.ClampScroll(ctx.Session)
			m.cartWindow.Refresh(ctx, &m.itemInfoWindow)
			continue
		}
		if friends, ok, err := network.ParseFriendsList(pkt); err != nil {
			log.Printf("parse friends list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyFriendsList(ctx, friends)
			continue
		}
		if friendState, ok, err := network.ParseFriendState(pkt); err != nil {
			log.Printf("parse friend state 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyFriendState(ctx, friendState)
			continue
		}
		if friendRequest, ok, err := network.ParseFriendRequest(pkt); err != nil {
			log.Printf("parse friend request 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.openFriendRequest(ctx, friendRequest)
			continue
		}
		if friendAdded, ok, err := network.ParseFriendAddResult(pkt); err != nil {
			log.Printf("parse friend add result 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyFriendAddResult(ctx, friendAdded)
			m.addFriendResultMessage(friendAdded)
			continue
		}
		if friendDeleted, ok, err := network.ParseFriendDelete(pkt); err != nil {
			log.Printf("parse friend delete 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyFriendDelete(ctx, friendDeleted)
			continue
		}
		if storageItem, ok, err := network.ParseStorageItemAdded(pkt); err != nil {
			log.Printf("parse storage item added 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyStorageItemAdded(ctx, storageItem)
			m.storageWindow.OpenWindow(ctx)
			m.storageWindow.ClampScroll(ctx.Session)
			m.inventoryBag.ClampScroll(ctx.Session)
			continue
		}
		if cartItem, ok, err := network.ParseCartItemAdded(pkt); err != nil {
			log.Printf("parse cart item added 0x%04X: %v", pkt.ID, err)
		} else if ok {
			log.Printf("cart item added index=%d item=%d amount=%d", cartItem.Index, cartItem.ItemID, cartItem.Amount)
			applyCartItemAdded(ctx, cartItem)
			m.cartWindow.ClampScroll(ctx.Session)
			m.cartWindow.Refresh(ctx, &m.itemInfoWindow)
			m.inventoryBag.ClampScroll(ctx.Session)
			continue
		}
		if ack, ok, err := network.ParseCartAddAck(pkt); err != nil {
			log.Printf("parse cart add ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			switch ack.Result {
			case 0:
				m.console.AddErrorMessage("Cart is overweight.")
				log.Printf("cart add rejected result=%d reason=weight", ack.Result)
			case 1:
				m.console.AddErrorMessage("Cart has too many items.")
				log.Printf("cart add rejected result=%d reason=count", ack.Result)
			default:
				m.console.AddErrorMessage("Cart add failed.")
				log.Printf("cart add rejected result=%d", ack.Result)
			}
			continue
		}
		if storageItem, ok, err := network.ParseStorageItemRemoved(pkt); err != nil {
			log.Printf("parse storage item removed 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyStorageItemRemoved(ctx, storageItem)
			m.storageWindow.ClampScroll(ctx.Session)
			m.storageWindow.Refresh(ctx, &m.itemInfoWindow)
			continue
		}
		if cartItem, ok, err := network.ParseCartItemRemoved(pkt); err != nil {
			log.Printf("parse cart item removed 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyCartItemRemoved(ctx, cartItem)
			m.cartWindow.ClampScroll(ctx.Session)
			m.cartWindow.Refresh(ctx, &m.itemInfoWindow)
			continue
		}
		if vendOpen, ok, err := network.ParseVendingOpenRequest(pkt); err != nil {
			log.Printf("parse vending open request 0x%04X: %v", pkt.ID, err)
		} else if ok {
			log.Printf("vending open request max_items=%d", vendOpen.MaxItems)
			m.vendingWindow.OpenSetup(ctx, vendOpen)
			continue
		}
		if board, ok, err := network.ParseVendingBoard(pkt); err != nil {
			log.Printf("parse vending board 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyVendingBoard(ctx, board)
			continue
		}
		if board, ok, err := network.ParseVendingBoardDisappear(pkt); err != nil {
			log.Printf("parse vending board disappear 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyVendingBoardDisappear(ctx, board)
			continue
		}
		if vendList, ok, err := network.ParseVendingItemList(pkt); err != nil {
			log.Printf("parse vending item list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			if vendList.Own {
				m.vendingWindow.ApplyOwnList(ctx, vendList)
			} else {
				m.vendingWindow.OpenBuy(ctx, vendList)
			}
			continue
		}
		if vendResult, ok, err := network.ParseVendingPurchaseResult(pkt); err != nil {
			log.Printf("parse vending purchase result 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.vendingWindow.ApplyPurchaseResult(ctx, vendResult)
			continue
		}
		if sold, ok, err := network.ParseVendingSoldItem(pkt); err != nil {
			log.Printf("parse vending sold item 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.vendingWindow.ApplySoldItem(ctx, sold)
			continue
		}
		if network.ParseStorageClosed(pkt) {
			applyStorageClosed(ctx)
			m.storageWindow.SetOpen(false)
			continue
		}
		if network.ParseCartClosed(pkt) {
			applyCartClosed(ctx)
			m.cartWindow.SetOpen(false)
			continue
		}
		if deal, ok, err := network.ParseShopDealSelection(pkt); err != nil {
			log.Printf("parse shop deal selection 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.shopWindow.OpenDeal(deal, ctx)
			continue
		}
		if sellList, ok, err := network.ParseShopSellList(pkt); err != nil {
			log.Printf("parse shop sell list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.shopWindow.OpenSell(sellList, ctx)
			continue
		}
		if buyList, ok, err := network.ParseShopBuyList(pkt); err != nil {
			log.Printf("parse shop buy list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.shopWindow.OpenBuy(buyList, ctx)
			continue
		}
		if result, ok, err := network.ParseShopResult(pkt); err != nil {
			log.Printf("parse shop result 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.shopWindow.ApplyResult(ctx, result)
			if result.Sell && result.Result == 0 {
				m.console.AddBlueMessage("The deal has successfully completed.")
			}
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
			if m.attackFocusID == vanish.ID {
				m.clearAttackFocus()
			}
			continue
		}
		if look, ok, err := network.ParseActorLookChange(pkt); err != nil {
			log.Printf("parse actor look change 0x%04X: %v", pkt.ID, err)
		} else if ok {
			if m.applySkillUnitLookChange(ctx, look) {
				continue
			}
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
		if direction, ok, err := network.ParseActorDirectionChange(pkt); err != nil {
			log.Printf("parse actor direction change 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyActorDirectionChange(ctx, direction)
			continue
		}
		if state, ok, err := network.ParseActorStateChange(pkt); err != nil {
			log.Printf("parse actor state change 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyActorStateChange(ctx, state)
			continue
		}
		if bladeStop, ok, err := network.ParseActorBladeStop(pkt); err != nil {
			log.Printf("parse actor blade stop 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyActorBladeStop(ctx, bladeStop)
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
		if snapshot, ok, err := network.ParseStatusSnapshot(pkt); err != nil {
			log.Printf("parse status snapshot 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applyStatusSnapshot(ctx, snapshot)
			continue
		}
		if ack, ok, err := network.ParseStatusChangeAck(pkt); err != nil {
			log.Printf("parse status change ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.statsWindow.ApplyStatusChangeAck(ctx, ack)
			continue
		}
		if statusEffect, ok, err := network.ParseStatusEffectChange(pkt); err != nil {
			log.Printf("parse status effect change 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyStatusEffectChange(ctx, statusEffect)
			continue
		}
		if list, ok, err := network.ParseSkillInfoList(pkt); err != nil {
			log.Printf("parse skill list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applySkillInfoList(ctx, list)
			m.skillWindow.ClampScroll(ctx.Session)
			continue
		}
		if update, ok, err := network.ParseSkillInfoUpdate(pkt); err != nil {
			log.Printf("parse skill update 0x%04X: %v", pkt.ID, err)
		} else if ok {
			applySkillInfoUpdate(ctx, update)
			m.skillWindow.ClampScroll(ctx.Session)
			continue
		}
		if auto, ok, err := network.ParseAutoRunSkill(pkt); err != nil {
			log.Printf("parse auto-run skill 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.skills().ApplyAutoRun(ctx, auto)
			continue
		}
		if warpList, ok, err := network.ParseWarpPointList(pkt); err != nil {
			log.Printf("parse warp point list 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyWarpPointList(ctx, warpList)
			continue
		}
		if memo, ok, err := network.ParseRememberWarpPointAck(pkt); err != nil {
			log.Printf("parse remember warp point ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyRememberWarpPointAck(ctx, memo)
			continue
		}
		if fail, ok, err := network.ParseSkillFailAck(pkt); err != nil {
			log.Printf("parse skill fail ack 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applySkillFailAck(ctx, fail)
			continue
		}
		if cast, ok, err := network.ParseSkillCastNotify(pkt); err != nil {
			log.Printf("parse skill cast 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applySkillCastNotify(ctx, cast)
			continue
		}
		if groundSkill, ok, err := network.ParseGroundSkillNotify(pkt); err != nil {
			log.Printf("parse ground skill 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyGroundSkillNotify(ctx, groundSkill)
			continue
		}
		if skillUnit, ok, err := network.ParseSkillUnitEntry(pkt); err != nil {
			log.Printf("parse skill unit 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applySkillUnitEntry(ctx, skillUnit)
			continue
		}
		if skillUnit, ok, err := network.ParseSkillUnitDisappear(pkt); err != nil {
			log.Printf("parse skill unit disappear 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applySkillUnitDisappear(skillUnit)
			continue
		}
		if effect, ok, err := network.ParseSpecialEffectNotify(pkt); err != nil {
			log.Printf("parse special effect 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applySpecialEffectNotify(ctx, effect)
			continue
		}
		if skill, ok, err := network.ParseSkillNoDamageNotify(pkt); err != nil {
			log.Printf("parse skill nodamage 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applySkillNoDamageNotify(ctx, skill)
			continue
		}
		if failure, ok, err := network.ParseAttackFailureForDistance(pkt); err != nil {
			log.Printf("parse attack distance failure 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyAttackFailureForDistance(ctx, failure)
			continue
		}
		if recovery, ok, err := network.ParseRecovery(pkt); err != nil {
			log.Printf("parse recovery 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyRecovery(ctx, recovery)
			continue
		}
		if change, ok, err := network.ParseParameterChange(pkt); err != nil {
			log.Printf("parse parameter change 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.applyParameterChange(ctx, change)
			if change.VarID == network.StatusHP {
				m.clearLocalDeathStateIfAlive(ctx)
			}
			continue
		}
		if entry, ok, err := network.ParseActorEntry(pkt); err != nil {
			log.Printf("parse actor entry 0x%04X: %v", pkt.ID, err)
		} else if ok {
			m.clearActorDeath(entry.ID)
			upsertNetworkActor(ctx, entry)
			m.applyWarpPortalEntry(ctx, entry)
		}
	}
	for _, err := range ctx.Network.DrainErrors() {
		log.Printf("network frame error: %v", err)
	}

	m.updatePendingAttack(ctx, "update", false)
	m.processPendingAttack(ctx)
	m.updatePendingPickup(ctx, "update", false)
	m.processPendingPickup(ctx)
	m.skills().UpdatePendingTarget(ctx, "update", false)
	m.skills().ProcessPendingTarget(ctx)
	m.processLockedAttack(ctx)
	now = time.Now()
	m.cleanupDeadActors(ctx, now)
	m.processScheduledActorStops(ctx, now)
	m.processScheduledWalkResumes(ctx, now)
	m.processActorMotionSounds(ctx, now)
	m.processMapSounds(ctx, now)
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
	if m.mapFade.phase == mapFadeHold {
		return nil, nil
	}
	if m.skills().CancelFromInput(ctx) {
		return nil, nil
	}
	m.skills().AdjustPendingLevelFromWheel(ctx)
	playerContextConsumed := m.playerContext.Update(ctx)
	if name := m.playerContext.PopAddFriendName(); name != "" {
		m.sendAddFriend(ctx, name)
		return nil, nil
	}
	if playerContextConsumed {
		return nil, nil
	}
	if m.openPlayerContextFromInput(ctx, now) {
		return nil, nil
	}
	if !m.escapeMenu.IsOpen() && !m.teleportModal.IsOpen() && !m.deathModal.IsOpen() && !m.friendRequest.IsOpen() && !m.settingsWindow.IsOpen() && !m.identifyWindow.IsOpen() {
		m.updateCameraRotation(ctx)
	}
	if m.friendRequest.Update(ctx) {
		return nil, nil
	}
	if m.deathModal.Update(ctx) {
		return nil, nil
	}
	if m.teleportModal.Update(ctx, m) {
		return nil, nil
	}
	if m.npcDialog.Update(ctx) {
		return nil, nil
	}
	if m.console.Update(ctx) {
		return nil, nil
	}
	if m.settingsWindow.Update(ctx) {
		return nil, nil
	}
	if m.escapeMenu.Update(ctx) {
		switch m.escapeMenu.ConsumeAction() {
		case gameui.EscapeMenuActionCharacterSelect:
			m.escapeMenu.RequestCharacterSelect(ctx)
		case gameui.EscapeMenuActionSettings:
			m.settingsWindow.OpenWindow(ctx)
		}
		return nil, nil
	}
	if m.characterWindow.Update(ctx) {
		return nil, nil
	}
	if m.itemInfoWindow.Update(ctx, m) {
		return nil, nil
	}
	if m.identifyWindow.Update(ctx) {
		return nil, nil
	}
	if m.inventoryBag.UpdateDrag(ctx, &m.shortcutBar, &m.storageWindow, &m.cartWindow) {
		return nil, nil
	}
	if m.storageWindow.UpdateDrag(ctx, &m.inventoryBag) {
		return nil, nil
	}
	if m.cartWindow.UpdateDrag(ctx, &m.inventoryBag) {
		return nil, nil
	}
	if m.skillWindow.UpdateDrag(ctx, &m.shortcutBar) {
		return nil, nil
	}
	if m.shortcutBar.Update(ctx, m) {
		return nil, nil
	}
	if m.inventoryBag.Update(ctx, &m.shortcutBar, &m.storageWindow, &m.cartWindow, &m.itemInfoWindow) {
		return nil, nil
	}
	if m.equipmentWindow.Update(ctx, &m.itemInfoWindow, &m.cartWindow, m) {
		return nil, nil
	}
	if m.storageWindow.Update(ctx, &m.inventoryBag, &m.itemInfoWindow) {
		return nil, nil
	}
	if m.cartWindow.Update(ctx, &m.inventoryBag, &m.itemInfoWindow) {
		return nil, nil
	}
	if m.changeCartWindow.Update(ctx) {
		return nil, nil
	}
	if m.shopWindow.Update(ctx, &m.itemInfoWindow) {
		return nil, nil
	}
	if m.vendingWindow.Update(ctx, &m.itemInfoWindow) {
		return nil, nil
	}
	if m.skillWindow.Update(ctx, &m.shortcutBar, m) {
		return nil, nil
	}
	if m.friendsWindow.Update(ctx) {
		return nil, nil
	}
	if m.statsWindow.Update(ctx) {
		return nil, nil
	}
	if m.basicMenu.Update(ctx, m.basicMenuCallbacks(ctx)) {
		return nil, nil
	}
	m.minimap.Update(ctx)
	removeExpiredStatusEffects(ctx.Session, now)
	m.statusIcons.Update(ctx, now)
	pointerBlocked := uiPointerBlocked(ctx)
	if !pointerBlocked {
		m.updateCameraZoom(ctx)
	}

	if !pointerBlocked && ctx.Input.MouseJustPressed(render.MouseButtonLeft) && m.walkReady(now) {
		m.nextHeldWalkAt = now.Add(heldWalkRepeatInterval)
		screenW, screenH := ctx.ScreenSize()
		projection := m.sceneProjection(ctx, screenW, screenH, now)
		if m.pendingSkill.skill.ID != 0 {
			m.skills().HandleClick(ctx, projection, now)
			return nil, nil
		}
		playerX, playerY := currentPlayerCell(ctx, now)
		if actor, ok := m.hoveredVendingBoard(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY, now); ok {
			log.Printf("click vending target mouse=%d,%d id=%d name=%q shop=%q player=%d,%d target=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY, actor.ID, actor.Name, actor.VendingName, playerX, playerY, actor.X, actor.Y)
			m.requestVendingList(ctx, actor, "click")
			return nil, nil
		}
		if item, ok := clickedGroundItem(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY, now); ok {
			log.Printf("click pickup target mouse=%d,%d id=%d item_id=%d amount=%d player=%d,%d target=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY, item.ID, item.ItemID, item.Amount, playerX, playerY, item.X, item.Y)
			m.clearLockedAttack()
			m.clearAttackFocus()
			m.requestPickup(ctx, item, "click")
			return nil, nil
		}
		if actor, ok := clickedAttackTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY, now, m.actorDeaths); ok {
			log.Printf("click attack target mouse=%d,%d id=%d name=%q job=%d object_type=%d player=%d,%d target=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY, actor.ID, actor.Name, actor.Job, actor.ObjectType, playerX, playerY, actor.X, actor.Y)
			m.requestAttack(ctx, actor, "click")
			return nil, nil
		}
		if actor, ok := clickedTalkTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY, now, m.actorDeaths); ok {
			log.Printf("click npc talk target mouse=%d,%d id=%d name=%q job=%d object_type=%d player=%d,%d target=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY, actor.ID, actor.Name, actor.Job, actor.ObjectType, playerX, playerY, actor.X, actor.Y)
			m.clearAttackFocus()
			m.requestNPCTalk(ctx, actor, "click")
			return nil, nil
		}
		if targetX, targetY, ok := clickedWalkTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY); ok {
			log.Printf("click walk target mouse=%d,%d player=%d,%d target=%d,%d", ctx.Input.MouseX, ctx.Input.MouseY, playerX, playerY, targetX, targetY)
			m.clearLockedAttack()
			m.clearAttackFocus()
			if shouldUseTurnOnlyGroundClick(ctx) {
				m.requestChangeDirection(ctx, targetX, targetY, "click")
				return nil, nil
			}
			m.requestWalk(ctx, targetX, targetY, "click")
		}
	}
	if m.updateHeldWalk(ctx, pointerBlocked, now) {
		return nil, nil
	}
	return nil, nil
}

func (m *WorldMode) basicMenuCallbacks(ctx client.Context) gameui.BasicMenuCallbacks {
	return gameui.BasicMenuCallbacks{
		OnStatus: func() { m.statsWindow.Toggle(ctx) },
		OnOption: func() { m.escapeMenu.OpenMenu() },
		OnItems:  func() { m.inventoryBag.Toggle(ctx) },
		OnEquip:  func() { m.equipmentWindow.Toggle(ctx) },
		OnSkill:  func() { m.skillWindow.Toggle(ctx) },
		OnMap:    func() { m.minimap.Toggle(ctx) },
		OnFriend: func() { m.friendsWindow.Toggle(ctx) },
	}
}

func (m *WorldMode) updateHeldWalk(ctx client.Context, pointerBlocked bool, now time.Time) bool {
	if ctx.Input == nil || !ctx.Input.MousePressed(render.MouseButtonLeft) || pointerBlocked {
		m.nextHeldWalkAt = time.Time{}
		return false
	}
	if ctx.Input.MouseJustPressed(render.MouseButtonLeft) || m.pendingSkill.skill.ID != 0 || shouldUseTurnOnlyGroundClick(ctx) {
		return false
	}
	if !m.walkReady(now) || (!m.nextHeldWalkAt.IsZero() && now.Before(m.nextHeldWalkAt)) {
		return false
	}
	m.nextHeldWalkAt = now.Add(heldWalkRepeatInterval)

	screenW, screenH := ctx.ScreenSize()
	projection := m.sceneProjection(ctx, screenW, screenH, now)
	targetX, targetY, ok := clickedWalkTarget(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY)
	if !ok || playerAtWalkTarget(ctx.World.Player, targetX, targetY, now) {
		return false
	}
	m.clearLockedAttack()
	m.clearAttackFocus()
	m.requestWalk(ctx, targetX, targetY, "held click")
	return true
}

func playerAtWalkTarget(player worldstate.Actor, targetX, targetY int, now time.Time) bool {
	x, y := player.RenderPosition(now)
	return int(math.Round(x)) == targetX && int(math.Round(y)) == targetY
}

func currentPlayerCell(ctx client.Context, now time.Time) (int, int) {
	if ctx.World == nil {
		return 0, 0
	}
	x, y := ctx.World.Player.RenderPosition(now)
	return int(math.Round(x)), int(math.Round(y))
}

func (m *WorldMode) handleMapChange(ctx client.Context, change network.MapChange) Mode {
	m.pendingAttack = attackIntent{}
	m.clearLockedAttack()
	m.clearAttackFocus()
	m.clearLocalActorAction(ctx)
	m.scheduledStops = nil
	m.npcDialog.ResetPublished(ctx)
	m.teleportModal = gameui.TeleportModal{}
	m.clearLocalDeathState(ctx)
	currentMap := ctx.World.MapName
	reuseLoadedMap := !change.ServerMove && sameLoadedMap(ctx, change.MapName)
	log.Printf("map change current=%s target=%s x=%d y=%d server_move=%t addr=%s port=%d reuse_loaded=%t", currentMap, change.MapName, change.X, change.Y, change.ServerMove, change.Address, change.Port, reuseLoadedMap)
	ctx.World.MapName = change.MapName
	ctx.Session.Zone.MapName = change.MapName
	applyWarpPosition(ctx, change.X, change.Y)
	ctx.World.Actors = make(map[uint32]worldstate.Actor)
	if reuseLoadedMap {
		zoom := m.camera.zoom
		m.camera.Reset()
		m.camera.zoom = zoom
		m.camera.Update(ctx, time.Now())
		if ctx.Network != nil {
			if err := ctx.Network.SendLoadEndAck(); err != nil {
				log.Printf("same-map warp load ack failed map=%s x=%d y=%d: %v", change.MapName, change.X, change.Y, err)
			} else {
				m.tickCooldown = 1
			}
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
			log.Printf("map reconnect failed map=%s addr=%s port=%d: %v", change.MapName, change.Address, change.Port, err)
			return nil
		}
		if err := ctx.Network.SendMapServerEnter(ctx.Session.AccountID, ctx.Session.CharID, ctx.Session.AuthCode, uint32(time.Now().UnixMilli()), ctx.Session.Sex); err != nil {
			log.Printf("map re-enter failed map=%s addr=%s port=%d: %v", change.MapName, change.Address, change.Port, err)
			return nil
		}
		m.pendingWarp = true
		return nil
	}
	return m.nextWorldMode()
}

func (m *WorldMode) nextWorldMode() *WorldMode {
	next := NewWorldMode()
	next.camera.zoom = m.camera.zoom
	next.console = m.console
	next.characterWindow = m.characterWindow
	next.basicMenu = m.basicMenu
	next.inventoryBag = m.inventoryBag
	next.equipmentWindow = m.equipmentWindow
	next.cartWindow = m.cartWindow
	next.itemInfoWindow = m.itemInfoWindow
	next.statsWindow = m.statsWindow
	next.skillWindow = m.skillWindow
	next.friendsWindow = m.friendsWindow
	next.settingsWindow = m.settingsWindow
	next.shortcutBar = m.shortcutBar
	next.minimap = m.minimap
	next.startMapFadeIn(time.Now())
	return next
}

func (m *WorldMode) nextCharacterSelectMode(ctx client.Context) *LoginMode {
	if ctx.Network != nil {
		ctx.Network.Close()
	}
	if ctx.Session != nil {
		ctx.Session.Playing = false
		ctx.Session.Storage = session.Storage{}
		ctx.Session.Cart = session.Cart{}
	}
	next := NewCharacterSelectMode(ctx, m.console)
	return next
}

func (m *WorldMode) startMapFadeOut(change network.MapChange, now time.Time) {
	m.mapFade = mapFadeState{
		phase:     mapFadeOut,
		started:   now,
		change:    change,
		hasChange: true,
	}
}

func (m *WorldMode) startMapFadeIn(now time.Time) {
	m.mapFade = mapFadeState{phase: mapFadeIn, started: now}
}

func (m *WorldMode) mapFadeElapsed(now time.Time) bool {
	switch m.mapFade.phase {
	case mapFadeOut:
		return now.Sub(m.mapFade.started) >= mapFadeOutDuration
	case mapFadeIn:
		return now.Sub(m.mapFade.started) >= mapFadeInDuration
	default:
		return false
	}
}

func (m *WorldMode) mapFadeAlpha(now time.Time) uint8 {
	if m.mapFade.started.IsZero() {
		return 0
	}
	switch m.mapFade.phase {
	case mapFadeOut:
		return clampColor(255 * clampUnit(float64(now.Sub(m.mapFade.started))/float64(mapFadeOutDuration)))
	case mapFadeHold:
		return 255
	case mapFadeIn:
		return clampColor(255 * (1 - clampUnit(float64(now.Sub(m.mapFade.started))/float64(mapFadeInDuration))))
	default:
		return 0
	}
}

func (m *WorldMode) drawMapFade(screen *render.Image, now time.Time) {
	alpha := m.mapFadeAlpha(now)
	if alpha == 0 {
		return
	}
	bounds := screen.Bounds()
	render.DrawRect(screen, 0, 0, float64(bounds.Dx()), float64(bounds.Dy()), color.RGBA{A: alpha})
}

func formatConsoleMessage(manager *res.Manager, chat network.ChatMessage) string {
	if chat.Text != "" {
		return chat.Text
	}
	if chat.MessageID < 0 {
		return ""
	}
	text := ""
	if manager != nil {
		text, _ = manager.MsgString(chat.MessageID)
	}
	if text == "" {
		text = fmt.Sprintf("message #%d", chat.MessageID)
	}
	if chat.Value != 0 {
		if strings.Contains(text, "%") {
			text = fmt.Sprintf(text, chat.Value)
		} else {
			text = fmt.Sprintf("%s %d", text, chat.Value)
		}
	}
	if chat.SkillID != 0 {
		text = fmt.Sprintf("skill %d: %s", chat.SkillID, text)
	}
	return text
}

func addConsoleMessage(console *gameui.ChatConsole, manager *res.Manager, chat network.ChatMessage) {
	if console == nil {
		return
	}
	text := formatConsoleMessage(manager, chat)
	if text == "" {
		return
	}
	if chat.Text == "" || !strings.Contains(text, " : ") {
		console.AddSystemMessage("%s", text)
		return
	}
	console.AddMessage("%s", text)
}

func formatPickupConsoleMessage(manager *res.Manager, pickup network.ItemPickupAck) string {
	itemName := fmt.Sprintf("item %d", pickup.ItemID)
	if manager != nil {
		if name, ok := manager.ItemDisplayName(int(pickup.ItemID), pickup.Identified); ok && name != "" {
			itemName = name
		}
	}
	amount := int(pickup.Amount)
	if amount <= 0 {
		amount = 1
	}
	template := ""
	if manager != nil {
		template, _ = manager.MsgString(153)
	}
	if template == "" {
		template = "You got %s %d."
	}
	if strings.Contains(template, "%s") {
		template = strings.Replace(template, "%s", itemName, 1)
	} else {
		template = strings.TrimSpace(template + " " + itemName)
	}
	if strings.Contains(template, "%d") {
		template = strings.Replace(template, "%d", fmt.Sprintf("%d", amount), 1)
	} else if amount != 1 {
		template = strings.TrimSpace(fmt.Sprintf("%s %d", template, amount))
	}
	return template
}

func sameLoadedMap(ctx client.Context, mapName string) bool {
	if ctx.World == nil || ctx.World.MapName == "" || mapName == "" {
		return false
	}
	if !strings.EqualFold(ctx.World.MapName, mapName) {
		return false
	}
	return ctx.World.GND != nil || ctx.World.GAT != nil
}

func (m *WorldMode) requestWalk(ctx client.Context, targetX, targetY int, source string) {
	if !walkTargetInBounds(ctx, targetX, targetY) {
		m.setWalkCooldown(walkRequestCooldown)
		return
	}
	playerX, playerY := currentPlayerCell(ctx, time.Now())
	log.Printf("%s walk request from=%d,%d to=%d,%d", source, playerX, playerY, targetX, targetY)
	if err := ctx.Network.SendWalkToXY(targetX, targetY); err == nil {
		m.setWalkCooldown(walkRequestCooldown)
	} else {
		log.Printf("%s walk request failed from=%d,%d to=%d,%d: %v", source, playerX, playerY, targetX, targetY, err)
		m.setWalkCooldown(walkErrorCooldown)
	}
}

func shouldUseTurnOnlyGroundClick(ctx client.Context) bool {
	if ctx.World == nil || ctx.Input == nil {
		return false
	}
	return ctx.World.Player.Sitting || ctx.Input.Pressed(render.KeyShift)
}

func (m *WorldMode) requestChangeDirection(ctx client.Context, targetX, targetY int, source string) {
	if ctx.Network == nil {
		m.setWalkCooldown(walkErrorCooldown)
		return
	}
	if ctx.World == nil {
		return
	}
	playerX, playerY := currentPlayerCell(ctx, time.Now())
	targetDir := directionFromDelta(playerX, playerY, targetX, targetY, ctx.World.Player.Dir)
	headDir, bodyDir, ok := resolveTurnOnlyDirection(ctx.World.Player.Dir, int(ctx.World.Player.HeadDir), targetDir)
	if !ok {
		return
	}
	log.Printf("%s change direction request player=%d,%d target=%d,%d head_dir=%d dir=%d", source, playerX, playerY, targetX, targetY, headDir, bodyDir)
	if err := ctx.Network.SendChangeDirection(headDir, bodyDir); err != nil {
		log.Printf("%s change direction failed target=%d,%d head_dir=%d dir=%d: %v", source, targetX, targetY, headDir, bodyDir, err)
		m.setWalkCooldown(walkErrorCooldown)
		return
	}
	m.applyLocalDirection(ctx, headDir, bodyDir)
	m.setWalkCooldown(turnDirectionCooldown)
}

func resolveTurnOnlyDirection(currentBodyDir int, currentHeadDir int, targetDir int) (uint8, uint8, bool) {
	bodyDir := normalizeDirectionIndex(currentBodyDir)
	headDir := normalizeHeadDir(currentHeadDir)
	targetDir = normalizeDirectionIndex(targetDir)
	delta := normalizeDirectionIndex(bodyDir - targetDir)

	resolvedBodyDir := bodyDir
	resolvedHeadDir := 0
	switch delta {
	case 0, 4:
		resolvedBodyDir = targetDir
	case 1:
		if headDir != 1 {
			resolvedHeadDir = 1
		} else {
			resolvedBodyDir = targetDir
		}
	case 2, 3:
		resolvedBodyDir = normalizeDirectionIndex(targetDir + 1)
		resolvedHeadDir = 1
	case 7:
		if headDir != 2 {
			resolvedHeadDir = 2
		} else {
			resolvedBodyDir = targetDir
		}
	case 5, 6:
		resolvedBodyDir = normalizeDirectionIndex(targetDir - 1)
		resolvedHeadDir = 2
	default:
		return 0, 0, false
	}
	return uint8(normalizeHeadDir(resolvedHeadDir)), uint8(resolvedBodyDir), true
}

func normalizeHeadDir(headDir int) int {
	if headDir < 0 {
		return 0
	}
	if headDir > 2 {
		return 2
	}
	return headDir
}

func (m *WorldMode) applyLocalDirection(ctx client.Context, headDir, dir uint8) {
	if ctx.World == nil {
		return
	}
	ctx.World.Player.Dir = int(dir & 7)
	ctx.World.Player.HeadDir = uint8(normalizeHeadDir(int(headDir)))
	ctx.World.Dir = ctx.World.Player.Dir
	if ctx.Session != nil {
		ctx.Session.PlayerDir = ctx.World.Player.Dir
	}
}

func (m *WorldMode) requestNPCTalk(ctx client.Context, actor worldstate.Actor, source string) {
	if ctx.Network == nil {
		m.setWalkCooldown(walkErrorCooldown)
		return
	}
	m.clearLockedAttack()
	if err := ctx.Network.SendNPCContact(actor.ID); err == nil {
		m.setWalkCooldown(walkRequestCooldown)
	} else {
		playerX, playerY := currentPlayerCell(ctx, time.Now())
		log.Printf("%s npc talk failed target=%d player=%d,%d target=%d,%d: %v", source, actor.ID, playerX, playerY, actor.X, actor.Y, err)
		m.setWalkCooldown(walkErrorCooldown)
	}
}

func (m *WorldMode) humanoidSpriteViewForActor(ctx client.Context, actor worldstate.Actor) *humanoidSpriteView {
	if isLocalActor(ctx, actor.ID) {
		return m.playerView
	}
	weapon, shield := res.NormalizePlayerWeaponShield(int(actor.Weapon), int(actor.Shield))
	key := actorSpriteKey{
		job:         int(actor.Job),
		head:        int(actor.Head),
		sex:         actor.Sex,
		bodyPalette: int(actor.BodyPal),
		headPalette: int(actor.HeadPal),
		weapon:      weapon,
		shield:      shield,
		headTop:     int(actor.HeadTop),
		headMid:     int(actor.HeadMid),
		headLow:     int(actor.HeadLow),
	}
	return m.actorViews[key]
}

func (m *WorldMode) scheduleActorStop(id uint32, at time.Time, resumeWalk bool, resumeAt time.Time) {
	if id == 0 || at.IsZero() {
		return
	}
	m.scheduledStops = append(m.scheduledStops, scheduledActorStop{id: id, at: at, resumeWalk: resumeWalk, resumeAt: resumeAt})
}

func (m *WorldMode) processScheduledActorStops(ctx client.Context, now time.Time) {
	if len(m.scheduledStops) == 0 {
		return
	}
	active := m.scheduledStops[:0]
	for _, stop := range m.scheduledStops {
		if now.Before(stop.at) {
			active = append(active, stop)
			continue
		}
		if stop.resumeWalk {
			m.pauseActorMovementForResume(ctx, stop.id, stop.at, stop.resumeAt)
		} else {
			m.stopActorMovementAt(ctx, stop.id, stop.at)
		}
	}
	m.scheduledStops = active
}

func (m *WorldMode) pauseActorMovementForResume(ctx client.Context, id uint32, at time.Time, resumeAt time.Time) {
	if !m.canResumeActorWalk(ctx, id, at) {
		m.stopActorMovementAt(ctx, id, at)
		return
	}
	if ctx.World == nil || !isLocalActor(ctx, id) {
		m.stopActorMovementAt(ctx, id, at)
		return
	}
	toX := ctx.World.Player.ToX
	toY := ctx.World.Player.ToY
	stopLocalPlayerMovementAt(ctx, at)
	if ctx.World.Player.X == toX && ctx.World.Player.Y == toY {
		return
	}
	m.scheduledResumes = append(m.scheduledResumes, scheduledWalkResume{
		id:  id,
		at:  resumeAt,
		toX: toX,
		toY: toY,
	})
}

func (m *WorldMode) canResumeActorWalk(ctx client.Context, id uint32, at time.Time) bool {
	if ctx.World == nil || !isLocalActor(ctx, id) || !m.hasCombatFocus() {
		return false
	}
	actor := ctx.World.Player
	if !actor.IsMovingAt(at) || len(actor.MovePath) < 2 {
		return false
	}
	x, y := actor.RenderPosition(at)
	return math.Hypot(float64(actor.ToX)-x, float64(actor.ToY)-y) > 0.25
}

func (m *WorldMode) processScheduledWalkResumes(ctx client.Context, now time.Time) {
	if len(m.scheduledResumes) == 0 {
		return
	}
	active := m.scheduledResumes[:0]
	for _, resume := range m.scheduledResumes {
		if now.Before(resume.at) {
			active = append(active, resume)
			continue
		}
		m.resumeActorWalk(ctx, resume)
	}
	m.scheduledResumes = active
}

func (m *WorldMode) resumeActorWalk(ctx client.Context, resume scheduledWalkResume) {
	if ctx.World == nil || !isLocalActor(ctx, resume.id) || !m.hasCombatFocus() {
		return
	}
	player := ctx.World.Player
	if player.X == resume.toX && player.Y == resume.toY {
		return
	}
	ctx.World.SetPlayerMovementAt(player.X, player.Y, resume.toX, resume.toY, player.Dir, resume.at, 0)
}

func (m *WorldMode) stopActorMovementAt(ctx client.Context, id uint32, at time.Time) {
	if ctx.World == nil || id == 0 {
		return
	}
	if isLocalActor(ctx, id) {
		stopLocalPlayerMovementAt(ctx, at)
		return
	}
	actor, ok := ctx.World.Actors[id]
	if !ok || !actor.IsMovingAt(at) {
		return
	}
	x, y := actor.RenderPosition(at)
	actor.X = int(math.Round(x))
	actor.Y = int(math.Round(y))
	clearActorMovement(&actor)
	ctx.World.Actors[id] = actor
}

func stopLocalPlayerMovementAt(ctx client.Context, at time.Time) {
	if ctx.World == nil || !ctx.World.Player.IsMovingAt(at) {
		return
	}
	x, y := ctx.World.Player.RenderPosition(at)
	ctx.World.Player.X = int(math.Round(x))
	ctx.World.Player.Y = int(math.Round(y))
	clearActorMovement(&ctx.World.Player)
	if ctx.Session != nil {
		ctx.Session.PlayerX = ctx.World.Player.X
		ctx.Session.PlayerY = ctx.World.Player.Y
	}
}

func clearActorMovement(actor *worldstate.Actor) {
	actor.Moving = false
	actor.FromX = actor.X
	actor.FromY = actor.Y
	actor.ToX = actor.X
	actor.ToY = actor.Y
	actor.MovePath = nil
	actor.HasMoveStart = false
	actor.MoveStartX = 0
	actor.MoveStartY = 0
	actor.WalkDistance = 0
}

func applyParameterChange(ctx client.Context, change network.ParameterChange) {
	if ctx.Session == nil {
		return
	}
	value := int(change.Value)
	switch change.VarID {
	case network.StatusSpeed:
		ctx.Session.Movement.ServerSpeed = value
		ctx.Session.Movement.HasServerSpeed = value > 0
		if ctx.World != nil {
			refreshLocalPlayerMoveSpeed(ctx)
		}
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
	case network.StatusPoint:
		ctx.Session.Stats.Points = value
	case network.StatusBaseLevel:
		ctx.Session.Progress.BaseLevel = value
		ctx.Session.Selected.Level = clampInt16(value)
	case network.StatusSkillPoint:
		ctx.Session.Skills.Points = value
	case network.StatusStr, network.StatusAgi, network.StatusVit, network.StatusInt, network.StatusDex, network.StatusLuk:
		setSessionStat(ctx.Session, change.VarID, value)
	case network.StatusUStr, network.StatusUAgi, network.StatusUVit, network.StatusUInt, network.StatusUDex, network.StatusULuk:
		setSessionStatCost(ctx.Session, change.VarID, value)
	case network.StatusZeny:
		ctx.Session.Inventory.Zeny = change.Value
	case network.StatusNextBaseExp:
		ctx.Session.Progress.NextBaseExp = change.Value
	case network.StatusNextJobExp:
		ctx.Session.Progress.NextJobExp = change.Value
	case network.StatusWeight:
		ctx.Session.Inventory.Weight = value
	case network.StatusMaxWeight:
		ctx.Session.Inventory.MaxWeight = value
	case network.StatusJobLevel:
		ctx.Session.Progress.JobLevel = value
		ctx.Session.Selected.JobLevel = clampInt16(value)
	default:
		return
	}
	log.Printf("parameter change var=%d value=%d hp=%d/%d sp=%d/%d base_lv=%d job_lv=%d base_exp=%d/%d job_exp=%d/%d zeny=%d weight=%d/%d",
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
		ctx.Session.Progress.NextJobExp,
		ctx.Session.Inventory.Zeny,
		ctx.Session.Inventory.Weight,
		ctx.Session.Inventory.MaxWeight)
}

func (m *WorldMode) applyParameterChange(ctx client.Context, change network.ParameterChange) {
	if ctx.Session == nil {
		return
	}
	previousHP := ctx.Session.Vitals.HP
	previousSP := ctx.Session.Vitals.SP
	previousBaseLevel := ctx.Session.Progress.BaseLevel
	previousJobLevel := ctx.Session.Progress.JobLevel
	applyParameterChange(ctx, change)
	if change.Value <= 0 {
		return
	}
	previousValues := map[uint16]int{
		network.StatusHP:        previousHP,
		network.StatusSP:        previousSP,
		network.StatusBaseLevel: previousBaseLevel,
		network.StatusJobLevel:  previousJobLevel,
	}
	if visual, ok := statusVisualEffects[change.VarID]; ok {
		visual.applyParameterChange(ctx, m, previousValues[change.VarID])
	}
}

var (
	recoveryHPColor = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	recoverySPColor = color.RGBA{R: 0, G: 0, B: 255, A: 255}
)

const (
	recoveryHPSFX = "_heal_effect.wav"
	recoverySPSFX = "effect\\\xC8\xED\xB1\xE2.wav"
)

var recoverySFXFallbacks = []string{"effect\\priest_recovery.wav"}

type statusVisualEffect struct {
	current       func(*session.Session) int
	recover       func(*session.Session, int) bool
	recovery      bool
	recoveryColor color.RGBA
	recoveryKind  damageFloaterKind
	recoverySFX   []string
	clearsDeath   bool
	levelEffectID int
}

var statusVisualEffects = map[uint16]statusVisualEffect{
	network.StatusHP: {
		current:       func(s *session.Session) int { return s.Vitals.HP },
		recover:       recoverSessionHP,
		recovery:      true,
		recoveryColor: recoveryHPColor,
		recoveryKind:  damageFloaterRecoveryHP,
		recoverySFX:   []string{recoveryHPSFX},
		clearsDeath:   true,
	},
	network.StatusSP: {
		current:       func(s *session.Session) int { return s.Vitals.SP },
		recover:       recoverSessionSP,
		recovery:      true,
		recoveryColor: recoverySPColor,
		recoveryKind:  damageFloaterRecoverySP,
		recoverySFX:   []string{recoverySPSFX},
	},
	network.StatusBaseLevel: {
		current:       func(s *session.Session) int { return s.Progress.BaseLevel },
		levelEffectID: effectBaseLevelUp,
	},
	network.StatusJobLevel: {
		current:       func(s *session.Session) int { return s.Progress.JobLevel },
		levelEffectID: effectJobLevelUp,
	},
}

func (v statusVisualEffect) applyParameterChange(ctx client.Context, mode *WorldMode, previous int) {
	if v.current == nil || mode == nil || ctx.Session == nil {
		return
	}
	current := v.current(ctx.Session)
	if v.recovery {
		delta := current - previous
		if delta > 0 {
			mode.addLocalRecoveryFloater(ctx, delta, v.recoveryColor, v.recoveryKind)
			mode.scheduleSound(time.Now(), v.sfxCandidates()...)
		}
		return
	}
	if v.levelEffectID > 0 && current > previous {
		mode.addWorldEffectIfMissing(ctx, v.levelEffectID, localSkillTarget(ctx))
	}
}

func (v statusVisualEffect) sfxCandidates() []string {
	if len(v.recoverySFX) == 0 {
		return append([]string(nil), recoverySFXFallbacks...)
	}
	paths := append([]string(nil), v.recoverySFX...)
	return append(paths, recoverySFXFallbacks...)
}

func recoverSessionHP(s *session.Session, amount int) bool {
	maxHP := s.Vitals.MaxHP
	if maxHP <= 0 {
		maxHP = int(s.Selected.MaxHP)
	}
	next := s.Vitals.HP + amount
	if maxHP > 0 && next > maxHP {
		next = maxHP
	}
	s.Vitals.HP = next
	s.Selected.HP = clampInt16(next)
	return true
}

func recoverSessionSP(s *session.Session, amount int) bool {
	maxSP := s.Vitals.MaxSP
	if maxSP <= 0 {
		maxSP = int(s.Selected.MaxSP)
	}
	next := s.Vitals.SP + amount
	if maxSP > 0 && next > maxSP {
		next = maxSP
	}
	s.Vitals.SP = next
	s.Selected.SP = clampInt16(next)
	return true
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

var (
	damageFloaterWhite  = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	damageFloaterYellow = color.RGBA{R: 230, G: 230, B: 38, A: 255}
	damageFloaterRed    = color.RGBA{R: 255, G: 64, B: 64, A: 255}
)

func clearWorldScene(screen *render.Image, mapName string) {
	screen.Fill(worldSceneClearColor(mapName))
}

func worldSceneClearColor(mapName string) color.RGBA {
	if skyColor, ok := robrSkyClearColors[normalizeMapNameForSceneClear(mapName)]; ok {
		return skyColor
	}
	return color.RGBA{A: 255}
}

var robrSkyClearColors = map[string]color.RGBA{
	"airplane.rsw":    {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"airplane_01.rsw": {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"gonryun.rsw":     {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"gon_dun02.rsw":   {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"himinn.rsw":      {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"ra_temsky.rsw":   {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"rwc01.rsw":       {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"sch_gld.rsw":     {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"valkyrie.rsw":    {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"yuno.rsw":        {R: 0x66, G: 0x99, B: 0xcc, A: 255},
	"5@tower.rsw":     {R: 0x33, G: 0x00, B: 0x33, A: 255},
	"thana_boss.rsw":  {R: 0xe0, G: 0xd4, B: 0xc2, A: 255},
}

func normalizeMapNameForSceneClear(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if index := strings.LastIndexAny(name, `\/`); index >= 0 {
		name = name[index+1:]
	}
	switch {
	case strings.HasSuffix(name, ".gat"):
		return strings.TrimSuffix(name, ".gat") + ".rsw"
	case name != "" && !strings.Contains(name, "."):
		return name + ".rsw"
	default:
		return name
	}
}

func (m *WorldMode) Draw(ctx client.Context, screen *render.Image) {
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
	now := time.Now()
	projection := m.sceneProjection(ctx, width, height, now)
	fog := sceneFogFromMap(ctx.Resources, ctx.World.MapName, ctx.Config.Fog)
	clearWorldScene(screen, ctx.World.MapName)
	var actorOverlays []sceneActorDrawEntry
	screen.SetCamera3D(projection.RenderCameraWithFog(fog))
	vertexFog := sceneFog{}

	if ctx.World.GND != nil {
		m.drawGNDMeshes(screen, ctx.Resources, ctx.World.GND, ctx.World.RSW, projection)
		m.drawGNDWater(screen, ctx.Resources, ctx.World.GND, ctx.World.RSW, projection, now, vertexFog)
		if !ctx.Config.Render.NoUI {
			m.drawTileCursor(screen, ctx, projection, now)
		}
		if ctx.World.RSW != nil && len(ctx.World.RSM) > 0 {
			actorOverlays = m.drawSceneModelsAndActors(screen, ctx, projection, vertexFog, now)
		} else {
			m.drawGroundItems(screen, ctx, projection, now)
			actorOverlays = m.drawSceneActors(screen, ctx, projection)
		}
	} else if ctx.World.GAT != nil {
		drawGAT(screen, ctx.World.GAT, ctx.World.Player.X, ctx.World.Player.Y)
		m.drawGroundItems(screen, ctx, projection, now)
		actorOverlays = m.drawSceneActors(screen, ctx, projection)
	} else {
		const tile = 32
		for x := 0; x < width; x += tile {
			render.DrawLine(screen, float64(x), 0, float64(x), float64(height), render.ColorGrid)
		}
		for y := 0; y < height; y += tile {
			render.DrawLine(screen, 0, float64(y), float64(width), float64(y), render.ColorGrid)
		}
	}

	if !ctx.Config.Render.NoUI {
		m.drawSceneActorOverlays(screen, ctx, projection, now, actorOverlays)
	}
	m.drawRSWEffects(screen, ctx, projection, now)
	m.drawMapWeatherEffects(screen, ctx, projection, now)
	m.drawWorldEffects(screen, ctx, projection, now)
	m.drawDamageFloaters(screen, ctx, projection, now)

	if !ctx.Config.Render.NoUI {
		m.inventoryBag.Draw(screen, ctx, m)
		m.storageWindow.Draw(screen, ctx, m)
		m.cartWindow.Draw(screen, ctx, m)
		m.shopWindow.Draw(screen, ctx, m)
		m.vendingWindow.Draw(screen, ctx, m)
		m.skillWindow.Draw(screen, ctx, m)
		m.drawHoveredGroundItemLabel(screen, ctx, projection, now)
		m.deathModal.Draw(screen, ctx, width, height)
	}
}

func (m *WorldMode) DrawOverlay(ctx client.Context, screen *render.Image) {
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
	now := time.Now()
	projection := m.sceneProjection(ctx, width, height, now)
	m.drawMapFade(screen, now)
	if !ctx.Config.Render.NoUI {
		m.drawUIDragGhosts(screen, ctx)
		m.drawROCursor(screen, ctx, projection, now)
	}
}

func (m *WorldMode) drawUIDragGhosts(screen *render.Image, ctx client.Context) {
	m.inventoryBag.DrawDragGhost(screen, ctx, m)
	m.storageWindow.DrawDragGhost(screen, ctx, m)
	m.cartWindow.DrawDragGhost(screen, ctx, m)
	m.shopWindow.DrawDragGhost(screen, ctx, m)
	m.vendingWindow.DrawDragGhost(screen, ctx, m)
	m.skillWindow.DrawDragGhost(screen, ctx, m)
}

func walkTargetInBounds(ctx client.Context, x, y int) bool {
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
	return terrainHeightAt(world, x, y)
}

func clickedWalkTarget(ctx client.Context, projection sceneProjection, mouseX, mouseY int) (int, int, bool) {
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

func hoveredWalkCell(ctx client.Context, projection sceneProjection, mouseX, mouseY int) (int, int, bool) {
	minX, maxX, minY, maxY, ok := walkTargetSearchBounds(ctx)
	if !ok {
		return 0, 0, false
	}
	return clickedWalkCellByProjectedPolygon(ctx, projection, mouseX, mouseY, minX, maxX, minY, maxY)
}

func (m *WorldMode) hoveredWalkCell(ctx client.Context, projection sceneProjection, mouseX, mouseY int) (int, int, bool) {
	if ctx.World == nil || ctx.World.GAT == nil {
		m.hoveredWalk.valid = false
		return 0, 0, false
	}
	key := hoveredWalkCellKey{
		gat:        ctx.World.GAT,
		mouseX:     mouseX,
		mouseY:     mouseY,
		playerX:    ctx.World.Player.X,
		playerY:    ctx.World.Player.Y,
		targetX:    projection.playerX,
		targetY:    projection.playerY,
		targetZ:    projection.playerZ,
		cameraYaw:  projection.cameraYaw,
		cameraZoom: projection.cameraZoom,
		screenW:    projection.screenW,
		screenH:    projection.screenH,
	}
	if m.hoveredWalk.valid && m.hoveredWalk.key == key {
		return m.hoveredWalk.x, m.hoveredWalk.y, m.hoveredWalk.ok
	}
	x, y, ok := hoveredWalkCell(ctx, projection, mouseX, mouseY)
	m.hoveredWalk = hoveredWalkCellCache{
		valid: true,
		key:   key,
		x:     x,
		y:     y,
		ok:    ok,
	}
	return x, y, ok
}

func walkTargetSearchBounds(ctx client.Context) (int, int, int, int, bool) {
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

func clickedWalkCellByProjectedPolygon(ctx client.Context, projection sceneProjection, mouseX, mouseY, minX, maxX, minY, maxY int) (int, int, bool) {
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

func clickedAttackTarget(ctx client.Context, projection sceneProjection, mouseX, mouseY int, now time.Time, deadActors map[uint32]time.Time) (worldstate.Actor, bool) {
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

func clickedSkillTarget(ctx client.Context, projection sceneProjection, skill session.Skill, mouseX, mouseY int, now time.Time, deadActors map[uint32]time.Time) (worldstate.Actor, bool) {
	if ctx.World == nil {
		return worldstate.Actor{}, false
	}
	bestDistance := math.Inf(1)
	var best worldstate.Actor
	for _, actor := range skillTargetCandidates(ctx, skill) {
		if _, dead := deadActors[actor.ID]; dead {
			continue
		}
		if !actorCanBeSkillTargeted(ctx, skill, actor) {
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

func clickedPlayerTarget(ctx client.Context, projection sceneProjection, mouseX, mouseY int, now time.Time, deadActors map[uint32]time.Time) (worldstate.Actor, bool) {
	if ctx.World == nil {
		return worldstate.Actor{}, false
	}
	bestDistance := math.Inf(1)
	var best worldstate.Actor
	for _, actor := range ctx.World.Actors {
		if _, dead := deadActors[actor.ID]; dead {
			continue
		}
		if !actorCanOpenPlayerContext(ctx, actor) {
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

func skillTargetCandidates(ctx client.Context, skill session.Skill) []worldstate.Actor {
	if ctx.World == nil {
		return nil
	}
	candidates := make([]worldstate.Actor, 0, len(ctx.World.Actors)+1)
	if skillTargetFlagsIncludeSelfPick(skill.Type) {
		if actor, ok, _ := actorForCombatID(ctx, localSkillTarget(ctx)); ok {
			candidates = append(candidates, actor)
		}
	}
	for _, actor := range ctx.World.Actors {
		candidates = append(candidates, actor)
	}
	return candidates
}

func skillTargetFlagsIncludeSelfPick(flags uint32) bool {
	return flags&(skillTargetFriend|skillTargetSelf) != 0
}

func clickWalkSearchRadius() int {
	return 70
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

func clampGameInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

type sceneDrawEntry struct {
	depth       float64
	actorIndex  int
	shadowIndex int
	itemIndex   int
}

func (m *WorldMode) drawSceneModelsAndActors(screen *render.Image, ctx client.Context, projection sceneProjection, fog sceneFog, now time.Time) []sceneActorDrawEntry {
	m.drawRSMModels(screen, ctx.Resources, ctx.World.RSW, ctx.World.RSM, ctx.World.GND, projection, fog, now)
	actors := m.collectSceneActorEntries(screen, ctx, projection)
	items := m.collectSceneItemEntries(screen, ctx, projection, now)
	entries := make([]sceneDrawEntry, 0, len(actors)+len(items))
	for i, item := range items {
		entries = append(entries, sceneDrawEntry{depth: item.depth, actorIndex: -1, shadowIndex: -1, itemIndex: i})
	}
	for i, actor := range actors {
		if actor.castShadow {
			entries = append(entries, sceneDrawEntry{depth: actor.shadowDepth, actorIndex: -1, shadowIndex: i, itemIndex: -1})
		}
		entries = append(entries, sceneDrawEntry{depth: actor.depth, actorIndex: i, shadowIndex: -1, itemIndex: -1})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].depth > entries[j].depth
	})
	for _, entry := range entries {
		if entry.shadowIndex >= 0 {
			m.drawActorShadowEntry(screen, projection, actors[entry.shadowIndex])
			continue
		}
		if entry.itemIndex >= 0 {
			m.drawGroundItemEntry3D(screen, projection, items[entry.itemIndex])
			continue
		}
		m.drawSceneActorEntry(screen, ctx, projection, actors[entry.actorIndex])
	}
	return actors
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

type screenPoint struct {
	x float32
	y float32
}

type texturePoint struct {
	u float32
	v float32
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
		x: -math.Cos(longitude) * math.Sin(latitude),
		y: -math.Cos(latitude),
		z: -math.Sin(longitude) * math.Sin(latitude),
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
		x: clampUnit(l.ambient.x+l.diffuse.x*weight) * clampUnit(l.env.x),
		y: clampUnit(l.ambient.y+l.diffuse.y*weight) * clampUnit(l.env.y),
		z: clampUnit(l.ambient.z+l.diffuse.z*weight) * clampUnit(l.env.z),
	}
}

func (l sceneLighting) modelScale(normal modelPoint3) modelPoint3 {
	weight := math.Max(dot3(normalize3(normal), l.direction), 0.5)
	return l.modelScaleFromWeight(weight)
}

func (l sceneLighting) modelScaleNormalized(normal modelPoint3) modelPoint3 {
	weight := math.Max(dot3(normal, l.direction), 0.5)
	return l.modelScaleFromWeight(weight)
}

func (l sceneLighting) modelScaleFromWeight(weight float64) modelPoint3 {
	return modelPoint3{
		x: clampUnit(l.ambient.x+l.diffuse.x*weight) * clampUnit(l.env.x),
		y: clampUnit(l.ambient.y+l.diffuse.y*weight) * clampUnit(l.env.y),
		z: clampUnit(l.ambient.z+l.diffuse.z*weight) * clampUnit(l.env.z),
	}
}

func clampUnitPoint(point modelPoint3) modelPoint3 {
	return modelPoint3{x: clampUnit(point.x), y: clampUnit(point.y), z: clampUnit(point.z)}
}

func clampUnit(value float64) float64 {
	return math.Max(0, math.Min(1, value))
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

func drawColoredSurfaceTints3DAlpha(screen, white *render.Image, verts [4]modelPoint3, indices []uint16, colors [4]color.RGBA) {
	drawColoredSurfaceTints3DWithOptions(screen, white, verts, indices, colors, triangleDrawOptions(render.FilterNearest, render.AddressUnsafe))
}

func drawColoredSurfaceTints3DWithOptions(screen, white *render.Image, verts [4]modelPoint3, indices []uint16, colors [4]color.RGBA, options *render.DrawTrianglesOptions) {
	vertices := []render.Vertex3D{
		coloredSurfaceVertex3D(verts[0], 0, 0, colors[0]),
		coloredSurfaceVertex3D(verts[1], 1, 0, colors[1]),
		coloredSurfaceVertex3D(verts[2], 1, 1, colors[2]),
		coloredSurfaceVertex3D(verts[3], 0, 1, colors[3]),
	}
	screen.DrawTriangles3DOwned(vertices, indices, white, options)
}

func drawTexturedSurface3DAlpha(screen, texture *render.Image, verts [4]modelPoint3, uvs [4]texturePoint, indices []uint16, tints [4]color.RGBA) {
	drawTexturedSurface3DWithOptions(screen, texture, verts, uvs, indices, tints, triangleDrawOptions(render.FilterLinear, render.AddressRepeat))
}

func drawTexturedSurface3DWithOptions(screen, texture *render.Image, verts [4]modelPoint3, uvs [4]texturePoint, indices []uint16, tints [4]color.RGBA, options *render.DrawTrianglesOptions) {
	bounds := texture.Bounds()
	w := float32(bounds.Dx())
	h := float32(bounds.Dy())
	vertices := []render.Vertex3D{
		texturedSurfaceVertex3D(verts[0], uvs[0], tints[0], w, h),
		texturedSurfaceVertex3D(verts[1], uvs[1], tints[1], w, h),
		texturedSurfaceVertex3D(verts[2], uvs[2], tints[2], w, h),
		texturedSurfaceVertex3D(verts[3], uvs[3], tints[3], w, h),
	}
	screen.DrawTriangles3DOwned(vertices, indices, texture, options)
}

func coloredSurfaceVertex3D(point modelPoint3, u, v float32, tint color.RGBA) render.Vertex3D {
	return render.Vertex3D{
		X:      float32(point.x),
		Y:      float32(point.y),
		Z:      float32(point.z),
		SrcX:   u,
		SrcY:   v,
		ColorR: float32(tint.R) / 255,
		ColorG: float32(tint.G) / 255,
		ColorB: float32(tint.B) / 255,
		ColorA: float32(tint.A) / 255,
		DepthX: float32(point.x),
		DepthY: float32(point.y),
		DepthZ: float32(point.z),
	}
}

func texturedSurfaceVertex3D(point modelPoint3, uv texturePoint, tint color.RGBA, textureWidth, textureHeight float32) render.Vertex3D {
	return render.Vertex3D{
		X:      float32(point.x),
		Y:      float32(point.y),
		Z:      float32(point.z),
		SrcX:   uv.u * textureWidth,
		SrcY:   uv.v * textureHeight,
		ColorR: float32(tint.R) / 255,
		ColorG: float32(tint.G) / 255,
		ColorB: float32(tint.B) / 255,
		ColorA: float32(tint.A) / 255,
		DepthX: float32(point.x),
		DepthY: float32(point.y),
		DepthZ: float32(point.z),
	}
}

func drawGAT(screen *render.Image, gat *res.GAT, playerX, playerY int) {
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
			render.DrawRect(screen, float64(sx*tile), float64(sy*tile), tile-1, tile-1, c)
		}
	}
}
