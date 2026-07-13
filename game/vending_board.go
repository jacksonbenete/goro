package game

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	gameui "github.com/kivutar/goro/ui"
	worldstate "github.com/kivutar/goro/world"
)

func (m *WorldMode) drawVendingBoardLabels(screen *render.Image, ctx client.Context, entries []sceneActorDrawEntry) {
	icon := m.vendingShopIcon(ctx.Resources)
	for _, entry := range entries {
		if !actorHasVending(entry.actor) {
			continue
		}
		label := sanitizeActorName(entry.actor.VendingName)
		labelY := actorSpriteTopY(entry.screenY, entry.scale) - boardLabelGap - vendingBoardLabelHeight(label, icon)
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
		labelY := actorSpriteTopY(entry.screenY, entry.scale) - boardLabelGap - chatRoomBoardLabelHeight(label, icon)
		if actorHasVending(entry.actor) {
			vendingLabel := sanitizeActorName(entry.actor.VendingName)
			labelY -= vendingBoardLabelHeight(vendingLabel, m.vendingShopIcon(ctx.Resources)) + boardLabelGap
		}
		drawChatRoomBoardLabel(screen, label, entry.screenX, labelY, icon)
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
	boardLabelW        = 158
	boardLabelH        = 36
	boardLabelPadX     = 5
	boardLabelIconGap  = 5
	boardLabelGap      = 4
	boardLabelIcon     = 24
	boardLabelTextH    = 20
	boardLabelRadius   = 4
	boardLabelOutlineW = 3
	boardLabelBorderW  = 1
)

type boardLabelStyle struct {
	width        int
	height       int
	padX         int
	iconGap      int
	iconSize     int
	outlineWidth int
	borderWidth  int
	radius       int
	fill         color.RGBA
	outline      color.RGBA
	border       color.RGBA
	text         color.RGBA
}

type boardSurfaceKey struct {
	width        int
	height       int
	outlineWidth int
	borderWidth  int
	radius       int
	fill         color.RGBA
	outline      color.RGBA
	border       color.RGBA
}

var boardSurfaceCache = map[boardSurfaceKey]*render.Image{}

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
	drawBoardLabel(screen, label, centerX, topY, icon, vendingBoardLabelStyle())
}

func drawChatRoomBoardLabel(screen *render.Image, label string, centerX, topY float64, icon *render.Image) {
	drawBoardLabel(screen, label, centerX, topY, icon, chatRoomBoardLabelStyle())
}

func drawBoardLabel(screen *render.Image, label string, centerX, topY float64, icon *render.Image, style boardLabelStyle) {
	label = sanitizeActorName(label)
	if label == "" {
		return
	}
	bounds, ok := boardLabelBounds(label, centerX, topY, icon, style)
	if !ok {
		return
	}
	bounds.x, bounds.y = render.SnapScreenPoint(screen, bounds.x, bounds.y)
	drawBoardSurface(screen, bounds, style)

	contentInset := style.outlineWidth + style.borderWidth
	contentX := bounds.x + float64(contentInset+style.padX)
	textX := contentX
	if icon != nil && !icon.Bounds().Empty() {
		drawBoardIcon(screen, icon, contentX, bounds.y+float64(style.height-style.iconSize)/2, style.iconSize)
		textX += float64(style.iconSize + style.iconGap)
	}
	maxTextWidth := int(bounds.x+bounds.w) - int(math.Round(textX)) - contentInset - style.padX
	text := trimBoardLabel(label, maxTextWidth)
	textY := bounds.y + float64(style.height-boardLabelTextH)/2 - 2
	render.DrawUITextAt(screen, text, textX, textY, style.text)
}

func drawBoardSurface(screen *render.Image, bounds vendingBoardBounds, style boardLabelStyle) {
	if screen == nil || bounds.w <= 0 || bounds.h <= 0 {
		return
	}
	surface := cachedBoardSurface(style)
	if surface == nil {
		return
	}
	var opts render.DrawImageOptions
	opts.GeoM.Translate(bounds.x, bounds.y)
	opts.Filter = render.FilterLinear
	screen.DrawImage(surface, &opts)
}

func cachedBoardSurface(style boardLabelStyle) *render.Image {
	if style.width <= 0 || style.height <= 0 {
		return nil
	}
	key := boardSurfaceKey{
		width:        style.width,
		height:       style.height,
		outlineWidth: style.outlineWidth,
		borderWidth:  style.borderWidth,
		radius:       style.radius,
		fill:         style.fill,
		outline:      style.outline,
		border:       style.border,
	}
	if img, ok := boardSurfaceCache[key]; ok {
		return img
	}
	img := render.NewImage(style.width, style.height)
	gameui.DrawRoundedSurface(img, 0, 0, style.width, style.height, style.outline, color.RGBA{}, float32(style.radius))
	blueInset := style.outlineWidth
	gameui.DrawRoundedSurface(
		img,
		blueInset,
		blueInset,
		style.width-blueInset*2,
		style.height-blueInset*2,
		style.border,
		color.RGBA{},
		float32(maxInt(0, style.radius-blueInset)),
	)
	fillInset := style.outlineWidth + style.borderWidth
	gameui.DrawRoundedSurface(
		img,
		fillInset,
		fillInset,
		style.width-fillInset*2,
		style.height-fillInset*2,
		style.fill,
		color.RGBA{},
		float32(maxInt(0, style.radius-fillInset)),
	)
	if len(boardSurfaceCache) > 32 {
		boardSurfaceCache = map[boardSurfaceKey]*render.Image{}
	}
	boardSurfaceCache[key] = img
	return img
}

func drawBoardIcon(screen *render.Image, icon *render.Image, x, y float64, size int) {
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
	if size <= 0 {
		size = boardLabelIcon
	}
	scale := math.Min(float64(size)/srcW, float64(size)/srcH)
	dstW, dstH := srcW*scale, srcH*scale
	dstX := x + (float64(size)-dstW)/2
	dstY := y + (float64(size)-dstH)/2
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
	topY := actorSpriteTopY(float64(point.y), scale) - boardLabelGap - vendingBoardLabelHeight(label, icon)
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
	topY := actorSpriteTopY(float64(point.y), scale) - boardLabelGap - chatRoomBoardLabelHeight(label, icon)
	if actorHasVending(actor) {
		topY -= vendingBoardLabelHeight(actor.VendingName, m.vendingShopIcon(ctx.Resources)) + boardLabelGap
	}
	return chatRoomBoardLabelBounds(label, float64(point.x), topY, icon)
}

func vendingBoardLabelBounds(label string, centerX, topY float64, icon *render.Image) (vendingBoardBounds, bool) {
	return boardLabelBounds(label, centerX, topY, icon, vendingBoardLabelStyle())
}

func chatRoomBoardLabelBounds(label string, centerX, topY float64, icon *render.Image) (vendingBoardBounds, bool) {
	return boardLabelBounds(label, centerX, topY, icon, chatRoomBoardLabelStyle())
}

func boardLabelBounds(label string, centerX, topY float64, _ *render.Image, style boardLabelStyle) (vendingBoardBounds, bool) {
	label = sanitizeActorName(label)
	if label == "" {
		return vendingBoardBounds{}, false
	}
	return vendingBoardBounds{
		x: math.Round(centerX - float64(style.width)/2),
		y: math.Round(topY),
		w: float64(style.width),
		h: float64(style.height),
	}, true
}

func vendingBoardLabelHeight(label string, icon *render.Image) float64 {
	bounds, ok := vendingBoardLabelBounds(label, 0, 0, icon)
	if !ok {
		return 0
	}
	return bounds.h
}

func chatRoomBoardLabelHeight(label string, icon *render.Image) float64 {
	bounds, ok := chatRoomBoardLabelBounds(label, 0, 0, icon)
	if !ok {
		return 0
	}
	return bounds.h
}

func vendingBoardLabelStyle() boardLabelStyle {
	return defaultBoardLabelStyle()
}

func chatRoomBoardLabelStyle() boardLabelStyle {
	return defaultBoardLabelStyle()
}

func defaultBoardLabelStyle() boardLabelStyle {
	return boardLabelStyle{
		width:        boardLabelW,
		height:       boardLabelH,
		padX:         boardLabelPadX,
		iconGap:      boardLabelIconGap,
		iconSize:     boardLabelIcon,
		outlineWidth: boardLabelOutlineW,
		borderWidth:  boardLabelBorderW,
		radius:       boardLabelRadius,
		fill:         color.RGBA{R: 255, G: 255, B: 255, A: 245},
		outline:      color.RGBA{R: 255, G: 255, B: 255, A: 245},
		border:       color.RGBA{R: 74, G: 138, B: 202, A: 245},
		text:         color.RGBA{R: 30, G: 34, B: 40, A: 255},
	}
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
