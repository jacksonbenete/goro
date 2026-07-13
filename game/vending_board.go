package game

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

func (m *WorldMode) drawVendingBoardLabels(screen *render.Image, ctx client.Context, entries []sceneActorDrawEntry) {
	icon := m.vendingShopIcon(ctx.Resources)
	for _, entry := range entries {
		if !actorHasVending(entry.actor) {
			continue
		}
		label := sanitizeActorName(entry.actor.VendingName)
		labelY := actorSpriteTopY(entry.screenY, entry.scale) - vendingBoardGap - vendingBoardLabelHeight(label, icon)
		drawVendingBoardLabel(screen, entry.actor.VendingName, entry.screenX, labelY, icon)
	}
}

func (m *WorldMode) drawChatRoomBoardLabels(screen *render.Image, ctx client.Context, entries []sceneActorDrawEntry) {
	for _, entry := range entries {
		if !actorHasChatRoom(entry.actor) {
			continue
		}
		label := chatRoomBoardLabel(entry.actor)
		icon := m.chatRoomBoardIcon(ctx.Resources, entry.actor.ChatRoomPublic)
		labelY := actorSpriteTopY(entry.screenY, entry.scale) - vendingBoardGap - vendingBoardLabelHeight(label, icon)
		if actorHasVending(entry.actor) {
			vendingLabel := sanitizeActorName(entry.actor.VendingName)
			labelY -= vendingBoardLabelHeight(vendingLabel, m.vendingShopIcon(ctx.Resources)) + vendingBoardGap
		}
		drawVendingBoardLabel(screen, label, entry.screenX, labelY, icon)
	}
}

func chatRoomBoardLabel(actor worldstate.Actor) string {
	title := sanitizeActorName(actor.ChatRoomTitle)
	if title == "" {
		return ""
	}
	if actor.ChatRoomLimit == 0 {
		return title
	}
	return fmt.Sprintf("%s (%d/%d)", title, actor.ChatRoomCount, actor.ChatRoomLimit)
}

const (
	vendingBoardW            = 140
	vendingBoardH            = 28
	vendingBoardPadX         = 3
	vendingBoardIconGap      = 5
	vendingBoardGap          = 4
	vendingBoardIcon         = 24
	vendingBoardTextOverlayH = 20
)

type vendingBoardBounds struct {
	x float64
	y float64
	w float64
	h float64
}

func (b vendingBoardBounds) contains(x, y float64) bool {
	return x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h
}

func drawVendingBoardLabel(screen *render.Image, label string, centerX, topY float64, icon *render.Image) {
	label = sanitizeActorName(label)
	if label == "" {
		return
	}
	bounds, ok := vendingBoardLabelBounds(label, centerX, topY, icon)
	if !ok {
		return
	}
	render.DrawRect(screen, bounds.x, bounds.y, bounds.w, bounds.h, color.RGBA{R: 255, G: 255, B: 255, A: 245})
	border := color.RGBA{R: 176, G: 184, B: 190, A: 245}
	render.DrawRect(screen, bounds.x, bounds.y, bounds.w, 1, border)
	render.DrawRect(screen, bounds.x, bounds.y+bounds.h-1, bounds.w, 1, border)
	render.DrawRect(screen, bounds.x, bounds.y, 1, bounds.h, border)
	render.DrawRect(screen, bounds.x+bounds.w-1, bounds.y, 1, bounds.h, border)

	contentX := int(bounds.x) + vendingBoardPadX
	textX := contentX
	if icon != nil && !icon.Bounds().Empty() {
		drawBoardIcon(screen, icon, contentX, int(bounds.y)+(vendingBoardH-vendingBoardIcon)/2)
		textX += vendingBoardIcon + vendingBoardIconGap
	}
	maxTextWidth := int(bounds.x+bounds.w) - textX - vendingBoardPadX
	text := trimBoardLabel(label, maxTextWidth)
	textY := bounds.y + float64(vendingBoardH-vendingBoardTextOverlayH)/2
	render.DrawUITextAt(screen, text, float64(textX), textY, color.RGBA{R: 30, G: 34, B: 40, A: 255})
}

func drawBoardIcon(screen *render.Image, icon *render.Image, x, y int) {
	if screen == nil || icon == nil {
		return
	}
	src := visibleImageBounds(icon)
	if src.Empty() {
		return
	}
	srcW, srcH := float64(src.Dx()), float64(src.Dy())
	if srcW <= 0 || srcH <= 0 {
		return
	}
	scale := math.Min(float64(vendingBoardIcon)/srcW, float64(vendingBoardIcon)/srcH)
	dstW, dstH := srcW*scale, srcH*scale
	dstX := float64(x) + (float64(vendingBoardIcon)-dstW)/2
	dstY := float64(y) + (float64(vendingBoardIcon)-dstH)/2
	vertices := []render.Vertex{
		{DstX: float32(dstX), DstY: float32(dstY), SrcX: float32(src.Min.X), SrcY: float32(src.Min.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(dstX + dstW), DstY: float32(dstY), SrcX: float32(src.Max.X), SrcY: float32(src.Min.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(dstX), DstY: float32(dstY + dstH), SrcX: float32(src.Min.X), SrcY: float32(src.Max.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(dstX + dstW), DstY: float32(dstY + dstH), SrcX: float32(src.Max.X), SrcY: float32(src.Max.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
	screen.DrawTrianglesOwned(vertices, quadIndices012213, icon, &render.DrawTrianglesOptions{Filter: spriteDrawFilter(), Address: render.AddressClampToZero})
}

func (m *WorldMode) hoveredVendingBoard(ctx client.Context, projection sceneProjection, mouseX, mouseY int, now time.Time) (worldstate.Actor, bool) {
	if ctx.World == nil {
		return worldstate.Actor{}, false
	}
	icon := m.vendingShopIcon(ctx.Resources)
	bestDistance := math.Inf(1)
	var best worldstate.Actor
	for _, actor := range ctx.World.Actors {
		if _, dead := m.actorDeaths[actor.ID]; dead {
			continue
		}
		if actor.ID == 0 || isLocalActor(ctx, actor.ID) || !actorHasVending(actor) {
			continue
		}
		bounds, ok := vendingBoardActorBounds(ctx, projection, actor, now, icon)
		if !ok || !bounds.contains(float64(mouseX), float64(mouseY)) {
			continue
		}
		dx := bounds.x + bounds.w/2 - float64(mouseX)
		dy := bounds.y + bounds.h/2 - float64(mouseY)
		distance := dx*dx + dy*dy
		if distance < bestDistance {
			bestDistance = distance
			best = actor
		}
	}
	return best, bestDistance < math.Inf(1)
}

func (m *WorldMode) hoveredChatRoomBoard(ctx client.Context, projection sceneProjection, mouseX, mouseY int, now time.Time) (worldstate.Actor, bool) {
	if ctx.World == nil {
		return worldstate.Actor{}, false
	}
	bestDistance := math.Inf(1)
	var best worldstate.Actor
	for _, actor := range ctx.World.Actors {
		if _, dead := m.actorDeaths[actor.ID]; dead {
			continue
		}
		if actor.ID == 0 || isLocalActor(ctx, actor.ID) || !actorHasChatRoom(actor) {
			continue
		}
		bounds, ok := m.chatRoomBoardActorBounds(ctx, projection, actor, now)
		if !ok || !bounds.contains(float64(mouseX), float64(mouseY)) {
			continue
		}
		dx := bounds.x + bounds.w/2 - float64(mouseX)
		dy := bounds.y + bounds.h/2 - float64(mouseY)
		distance := dx*dx + dy*dy
		if distance < bestDistance {
			bestDistance = distance
			best = actor
		}
	}
	return best, bestDistance < math.Inf(1)
}

func vendingBoardActorBounds(ctx client.Context, projection sceneProjection, actor worldstate.Actor, now time.Time, icon *render.Image) (vendingBoardBounds, bool) {
	label := sanitizeActorName(actor.VendingName)
	if label == "" || ctx.World == nil {
		return vendingBoardBounds{}, false
	}
	actorX, actorY := actor.RenderPosition(now)
	terrainZ := terrainHeightAt(ctx.World, actorX, actorY)
	point := projection.Project(cellCenter(actorX), cellCenter(actorY), terrainZ)
	scale := actorBillboardScreenScale(projection, cellCenter(actorX), cellCenter(actorY), terrainZ)
	topY := actorSpriteTopY(float64(point.y), scale) - vendingBoardGap - vendingBoardLabelHeight(label, icon)
	return vendingBoardLabelBounds(label, float64(point.x), topY, icon)
}

func (m *WorldMode) chatRoomBoardActorBounds(ctx client.Context, projection sceneProjection, actor worldstate.Actor, now time.Time) (vendingBoardBounds, bool) {
	label := chatRoomBoardLabel(actor)
	if label == "" || ctx.World == nil {
		return vendingBoardBounds{}, false
	}
	icon := m.chatRoomBoardIcon(ctx.Resources, actor.ChatRoomPublic)
	actorX, actorY := actor.RenderPosition(now)
	terrainZ := terrainHeightAt(ctx.World, actorX, actorY)
	point := projection.Project(cellCenter(actorX), cellCenter(actorY), terrainZ)
	scale := actorBillboardScreenScale(projection, cellCenter(actorX), cellCenter(actorY), terrainZ)
	topY := actorSpriteTopY(float64(point.y), scale) - vendingBoardGap - vendingBoardLabelHeight(label, icon)
	if actorHasVending(actor) {
		topY -= vendingBoardLabelHeight(actor.VendingName, m.vendingShopIcon(ctx.Resources)) + vendingBoardGap
	}
	return vendingBoardLabelBounds(label, float64(point.x), topY, icon)
}

func vendingBoardLabelBounds(label string, centerX, topY float64, icon *render.Image) (vendingBoardBounds, bool) {
	label = sanitizeActorName(label)
	if label == "" {
		return vendingBoardBounds{}, false
	}
	return vendingBoardBounds{
		x: math.Round(centerX - float64(vendingBoardW)/2),
		y: math.Round(topY),
		w: vendingBoardW,
		h: vendingBoardH,
	}, true
}

func vendingBoardLabelHeight(label string, icon *render.Image) float64 {
	bounds, ok := vendingBoardLabelBounds(label, 0, 0, icon)
	if !ok {
		return 0
	}
	return bounds.h
}

func trimBoardLabel(label string, maxWidth int) string {
	label = sanitizeActorName(label)
	if label == "" || maxWidth <= 0 {
		return ""
	}
	maxWidth -= 14
	if maxWidth <= 0 {
		return "..."
	}
	if w, _ := render.DebugTextSize(label); w <= maxWidth {
		return label
	}
	runes := []rune(label)
	for len(runes) > 1 {
		runes = runes[:len(runes)-1]
		candidate := string(runes) + "..."
		if w, _ := render.DebugTextSize(candidate); w <= maxWidth {
			return candidate
		}
	}
	return "..."
}

func (m *WorldMode) vendingShopIcon(manager *res.Manager) *render.Image {
	return m.roomBoardIcon(manager, "__interface_basic_interface_shop", "basic_interface\\shop")
}

func (m *WorldMode) chatRoomBoardIcon(manager *res.Manager, public bool) *render.Image {
	if public {
		return m.roomBoardIcon(manager, "__interface_basic_interface_chat_open", "basic_interface\\chat_open")
	}
	return m.roomBoardIcon(manager, "__interface_basic_interface_chat_close", "basic_interface\\chat_close")
}

func (m *WorldMode) roomBoardIcon(manager *res.Manager, key, resource string) *render.Image {
	if manager == nil {
		return nil
	}
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
	img, _, err := res.LoadImage(manager, res.InterfaceTextureCandidates(resource))
	if err != nil {
		m.textureMiss[key] = struct{}{}
		return nil
	}
	texture := render.NewImageFromImage(img)
	m.textures[key] = texture
	return texture
}
