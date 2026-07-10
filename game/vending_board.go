package game

import (
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

const (
	vendingBoardPadX    = 9
	vendingBoardPadY    = 5
	vendingBoardIconGap = 4
	vendingBoardGap     = 4
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
	render.DrawRect(screen, bounds.x, bounds.y, bounds.w, bounds.h, color.RGBA{R: 80, G: 88, B: 96, A: 230})
	render.DrawRect(screen, bounds.x+1, bounds.y+1, bounds.w-2, bounds.h-2, color.RGBA{R: 255, G: 255, B: 255, A: 245})

	_, textH := render.DebugTextSize(label)
	iconW, iconH := vendingBoardIconSize(icon)
	contentH := maxInt(textH, iconH)
	contentX := int(bounds.x) + vendingBoardPadX
	contentY := int(bounds.y) + vendingBoardPadY
	textX := contentX
	if iconW > 0 && iconH > 0 {
		iconY := contentY + (contentH-iconH)/2
		var opts render.DrawImageOptions
		opts.GeoM.Translate(float64(contentX), float64(iconY))
		screen.DrawImage(icon, &opts)
		textX += iconW + vendingBoardIconGap
	}
	textY := contentY + (contentH-textH)/2
	render.DebugPrintAtColor(screen, label, textX, textY, color.RGBA{R: 30, G: 34, B: 40, A: 255})
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

func vendingBoardLabelBounds(label string, centerX, topY float64, icon *render.Image) (vendingBoardBounds, bool) {
	label = sanitizeActorName(label)
	if label == "" {
		return vendingBoardBounds{}, false
	}
	textW, textH := render.DebugTextSize(label)
	iconW, iconH := vendingBoardIconSize(icon)
	contentW := textW
	if iconW > 0 && iconH > 0 {
		contentW += iconW + vendingBoardIconGap
	}
	contentH := maxInt(textH, iconH)
	w := float64(contentW + vendingBoardPadX*2)
	h := float64(contentH + vendingBoardPadY*2)
	return vendingBoardBounds{
		x: math.Round(centerX - w/2),
		y: math.Round(topY),
		w: w,
		h: h,
	}, true
}

func vendingBoardLabelHeight(label string, icon *render.Image) float64 {
	bounds, ok := vendingBoardLabelBounds(label, 0, 0, icon)
	if !ok {
		return 0
	}
	return bounds.h
}

func vendingBoardIconSize(icon *render.Image) (int, int) {
	if icon == nil || icon.Bounds().Empty() {
		return 0, 0
	}
	return icon.Bounds().Dx(), icon.Bounds().Dy()
}

func (m *WorldMode) vendingShopIcon(manager *res.Manager) *render.Image {
	if manager == nil {
		return nil
	}
	const key = "__interface_basic_interface_shop"
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
	img, _, err := res.LoadImage(manager, res.InterfaceTextureCandidates("basic_interface\\shop"))
	if err != nil {
		m.textureMiss[key] = struct{}{}
		return nil
	}
	texture := render.NewImageFromImage(img)
	m.textures[key] = texture
	return texture
}
