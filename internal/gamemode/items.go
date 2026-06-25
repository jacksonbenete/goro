package gamemode

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kivutar/goro/internal/network"
	worldstate "github.com/kivutar/goro/internal/world"
)

type sceneItemDrawEntry struct {
	item    worldstate.FloorItem
	screenX float64
	screenY float64
	scale   float64
	depth   float64
}

func (m *WorldMode) applyFloorItemEntry(ctx Context, entry network.FloorItemEntry) {
	if ctx.World == nil {
		return
	}
	item := worldstate.FloorItem{
		ID:         entry.ID,
		ItemID:     entry.ItemID,
		Identified: entry.Identified,
		X:          entry.X,
		Y:          entry.Y,
		SubX:       entry.SubX,
		SubY:       entry.SubY,
		Amount:     entry.Amount,
		Falling:    entry.Falling,
		DroppedAt:  time.Now(),
	}
	ctx.World.UpsertItem(item)
	log.Printf("floor item entry id=%d item_id=%d identified=%t amount=%d x=%d y=%d sub=%d,%d falling=%t", item.ID, item.ItemID, item.Identified, item.Amount, item.X, item.Y, item.SubX, item.SubY, item.Falling)
}

func (m *WorldMode) applyFloorItemDisappear(ctx Context, disappear network.FloorItemDisappear) {
	if ctx.World == nil {
		return
	}
	ctx.World.RemoveItem(disappear.ID)
	if m.pendingPickup.itemID == disappear.ID {
		m.pendingPickup = pickupIntent{}
	}
	log.Printf("floor item disappear id=%d", disappear.ID)
}

func (m *WorldMode) applyItemPickupAck(ctx Context, ack network.ItemPickupAck) {
	if ack.Result == 0 {
		m.status = fmt.Sprintf("picked item %d x%d", ack.ItemID, ack.Amount)
		m.pendingPickup = pickupIntent{}
		log.Printf("item pickup ack success index=%d item_id=%d amount=%d identified=%t", ack.Index, ack.ItemID, ack.Amount, ack.Identified)
		return
	}
	m.status = fmt.Sprintf("pickup failed item %d result=%d", ack.ItemID, ack.Result)
	log.Printf("item pickup ack failed index=%d item_id=%d amount=%d result=%d", ack.Index, ack.ItemID, ack.Amount, ack.Result)
}

func (m *WorldMode) requestPickup(ctx Context, item worldstate.FloorItem, source string) {
	if ctx.Network == nil {
		m.status = "pickup request failed: not connected"
		m.walkCooldown = 30
		return
	}
	if itemWithinPickupRange(ctx.World.Player.X, ctx.World.Player.Y, item.X, item.Y) {
		m.sendPickupRequest(ctx, item, source)
		return
	}
	targetX, targetY, ok := pickupApproachCell(ctx, item)
	if !ok {
		m.status = fmt.Sprintf("%s pickup walk blocked: %d", source, item.ID)
		log.Printf("%s pickup walk blocked item=%d player=%d,%d item=%d,%d", source, item.ID, ctx.World.Player.X, ctx.World.Player.Y, item.X, item.Y)
		m.walkCooldown = 12
		return
	}
	m.pendingPickup = pickupIntent{
		itemID:  item.ID,
		expires: time.Now().Add(8 * time.Second),
	}
	log.Printf("%s pickup walk target item=%d player=%d,%d item=%d,%d walk=%d,%d", source, item.ID, ctx.World.Player.X, ctx.World.Player.Y, item.X, item.Y, targetX, targetY)
	m.requestWalk(ctx, targetX, targetY, source+" pickup")
}

func (m *WorldMode) continuePendingPickup(ctx Context, source string) {
	if m.pendingPickup.itemID == 0 || ctx.World == nil {
		return
	}
	now := time.Now()
	if now.After(m.pendingPickup.expires) {
		log.Printf("%s pending pickup expired item=%d", source, m.pendingPickup.itemID)
		m.pendingPickup = pickupIntent{}
		return
	}
	item, ok := ctx.World.Items[m.pendingPickup.itemID]
	if !ok {
		log.Printf("%s pending pickup item vanished id=%d", source, m.pendingPickup.itemID)
		m.pendingPickup = pickupIntent{}
		return
	}
	if !itemWithinPickupRange(ctx.World.Player.X, ctx.World.Player.Y, item.X, item.Y) {
		log.Printf("%s pending pickup still out of range item=%d player=%d,%d item=%d,%d", source, item.ID, ctx.World.Player.X, ctx.World.Player.Y, item.X, item.Y)
		return
	}
	readyAt := pendingPickupReadyAt(ctx.World.Player, now)
	if m.pendingPickup.readyAt.IsZero() || readyAt.After(m.pendingPickup.readyAt) {
		m.pendingPickup.readyAt = readyAt
	}
	log.Printf("%s pending pickup scheduled item=%d delay_ms=%d", source, item.ID, maxInt(0, int(m.pendingPickup.readyAt.Sub(now).Milliseconds())))
}

func (m *WorldMode) processPendingPickup(ctx Context) {
	if m.pendingPickup.itemID == 0 || m.pendingPickup.readyAt.IsZero() || ctx.World == nil {
		return
	}
	now := time.Now()
	if now.After(m.pendingPickup.expires) {
		log.Printf("pending pickup expired item=%d", m.pendingPickup.itemID)
		m.pendingPickup = pickupIntent{}
		return
	}
	if now.Before(m.pendingPickup.readyAt) {
		return
	}
	item, ok := ctx.World.Items[m.pendingPickup.itemID]
	if !ok {
		log.Printf("pending pickup item vanished id=%d", m.pendingPickup.itemID)
		m.pendingPickup = pickupIntent{}
		return
	}
	if !itemWithinPickupRange(ctx.World.Player.X, ctx.World.Player.Y, item.X, item.Y) {
		log.Printf("pending pickup became out of range item=%d player=%d,%d item=%d,%d", item.ID, ctx.World.Player.X, ctx.World.Player.Y, item.X, item.Y)
		m.pendingPickup.readyAt = time.Time{}
		m.requestPickup(ctx, item, "pending")
		return
	}
	m.pendingPickup = pickupIntent{}
	m.sendPickupRequest(ctx, item, "pending")
}

func pendingPickupReadyAt(player worldstate.Actor, now time.Time) time.Time {
	readyAt := now.Add(60 * time.Millisecond)
	if player.IsMovingAt(now) && player.MoveDuration > 0 {
		walkReadyAt := player.MoveStarted.Add(player.MoveDuration).Add(60 * time.Millisecond)
		if walkReadyAt.After(readyAt) {
			readyAt = walkReadyAt
		}
	}
	return readyAt
}

func (m *WorldMode) sendPickupRequest(ctx Context, item worldstate.FloorItem, source string) {
	if err := ctx.Network.SendItemPickup(item.ID); err == nil {
		m.status = fmt.Sprintf("%s pickup request: %d", source, item.ID)
		m.walkCooldown = 12
	} else {
		m.status = source + " pickup request failed: " + err.Error()
		log.Printf("%s pickup request failed item=%d: %v", source, item.ID, err)
		m.walkCooldown = 30
	}
}

func itemWithinPickupRange(playerX, playerY, itemX, itemY int) bool {
	return maxInt(absInt(playerX-itemX), absInt(playerY-itemY)) <= 1
}

func pickupApproachCell(ctx Context, item worldstate.FloorItem) (int, int, bool) {
	if walkTargetInBounds(ctx, item.X, item.Y) {
		return item.X, item.Y, true
	}
	bestX, bestY := 0, 0
	bestDistance := math.Inf(1)
	found := false
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			x, y := item.X+dx, item.Y+dy
			if !walkTargetInBounds(ctx, x, y) {
				continue
			}
			dist := float64((ctx.World.Player.X-x)*(ctx.World.Player.X-x) + (ctx.World.Player.Y-y)*(ctx.World.Player.Y-y))
			if !found || dist < bestDistance {
				found = true
				bestDistance = dist
				bestX = x
				bestY = y
			}
		}
	}
	return bestX, bestY, found
}

func (m *WorldMode) drawGroundItems(screen *ebiten.Image, ctx Context, projection sceneProjection, now time.Time) {
	for _, entry := range m.collectSceneItemEntries(screen, ctx, projection, now) {
		m.drawGroundItemEntry(screen, entry)
	}
}

func (m *WorldMode) collectSceneItemEntries(screen *ebiten.Image, ctx Context, projection sceneProjection, now time.Time) []sceneItemDrawEntry {
	if ctx.World == nil || len(ctx.World.Items) == 0 {
		return nil
	}
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
	entries := make([]sceneItemDrawEntry, 0, len(ctx.World.Items))
	for _, item := range ctx.World.Items {
		x, y := floorItemWorldPosition(item)
		z := floorItemRenderHeight(ctx.World, item, now)
		point := projection.Project(cellCenter(x), cellCenter(y), z)
		if point.x < -48 || point.y < -80 || point.x > float32(width+48) || point.y > float32(height+48) {
			continue
		}
		scale := actorBillboardScreenScale(projection, cellCenter(x), cellCenter(y), z) * 0.42
		entries = append(entries, sceneItemDrawEntry{
			item:    item,
			screenX: float64(point.x),
			screenY: float64(point.y),
			scale:   scale,
			depth:   projection.Depth(cellCenter(x), cellCenter(y), z),
		})
	}
	return entries
}

func (m *WorldMode) drawGroundItemEntry(screen *ebiten.Image, entry sceneItemDrawEntry) {
	img := m.itemMarkerTexture()
	if img == nil {
		return
	}
	scale := entry.scale
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(-16, -24)
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(entry.screenX, entry.screenY)
	opts.Filter = ebiten.FilterNearest
	screen.DrawImage(img, &opts)
}

func (m *WorldMode) drawHoveredGroundItemLabel(screen *ebiten.Image, ctx Context, projection sceneProjection, now time.Time) {
	if ctx.Input == nil {
		return
	}
	item, ok := clickedGroundItem(ctx, projection, ctx.Input.MouseX, ctx.Input.MouseY, now)
	if !ok {
		return
	}
	label := fmt.Sprintf("Item %d x%d", item.ItemID, item.Amount)
	debugText(screen, ctx.Input.MouseX+14, ctx.Input.MouseY+18, "%s", label)
}

func clickedGroundItem(ctx Context, projection sceneProjection, mouseX, mouseY int, now time.Time) (worldstate.FloorItem, bool) {
	if ctx.World == nil || len(ctx.World.Items) == 0 {
		return worldstate.FloorItem{}, false
	}
	bestDistance := math.Inf(1)
	var best worldstate.FloorItem
	for _, item := range ctx.World.Items {
		x, y := floorItemWorldPosition(item)
		z := floorItemRenderHeight(ctx.World, item, now)
		point := projection.Project(cellCenter(x), cellCenter(y), z)
		scale := actorBillboardScreenScale(projection, cellCenter(x), cellCenter(y), z) * 0.42
		if !pointInGroundItemPickBounds(float64(mouseX), float64(mouseY), float64(point.x), float64(point.y), scale) {
			continue
		}
		dx := float64(point.x) - float64(mouseX)
		dy := float64(point.y) - float64(mouseY)
		distance := dx*dx + dy*dy
		if distance < bestDistance {
			bestDistance = distance
			best = item
		}
	}
	return best, bestDistance < math.Inf(1)
}

func pointInGroundItemPickBounds(mouseX, mouseY, centerX, centerY, scale float64) bool {
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	left := centerX - 18*scale
	right := centerX + 18*scale
	top := centerY - 30*scale
	bottom := centerY + 10*scale
	return mouseX >= left && mouseX <= right && mouseY >= top && mouseY <= bottom
}

func floorItemWorldPosition(item worldstate.FloorItem) (float64, float64) {
	return float64(item.X) - 0.5 + float64(item.SubX)/12, float64(item.Y) - 0.5 + float64(item.SubY)/12
}

func floorItemRenderHeight(world *worldstate.World, item worldstate.FloorItem, now time.Time) float64 {
	x, y := floorItemWorldPosition(item)
	ground := terrainHeightAt(world, x, y)
	if !item.Falling || item.DroppedAt.IsZero() {
		return ground
	}
	start := ground + 5.0
	fall := now.Sub(item.DroppedAt).Seconds() * 2.5
	if fall < 0 {
		fall = 0
	}
	return math.Max(ground, start-fall)
}

func (m *WorldMode) itemMarkerTexture() *ebiten.Image {
	if m.itemMarker != nil {
		return m.itemMarker
	}
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - 16
			dy := float64(y) - 14
			r := math.Sqrt(dx*dx/1.8 + dy*dy)
			if r > 12 {
				continue
			}
			alpha := uint8(210)
			if r > 9 {
				alpha = uint8(180 - (r-9)*35)
			}
			img.SetRGBA(x, y, color.RGBA{R: 245, G: 220, B: 96, A: alpha})
		}
	}
	for y := 8; y < 22; y++ {
		for x := 10; x < 23; x++ {
			if x == 10 || x == 22 || y == 8 || y == 21 {
				img.SetRGBA(x, y, color.RGBA{R: 80, G: 48, B: 16, A: 230})
			}
		}
	}
	m.itemMarker = ebiten.NewImageFromImage(img)
	return m.itemMarker
}
