package gamemode

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

const (
	minimapWidth   = 188
	minimapHeight  = 206
	minimapMargin  = 16
	minimapPad     = 10
	minimapTitleH  = 24
	minimapFooterH = 22
)

var (
	minimapTextColor   = uiTextColor
	minimapMutedColor  = uiMutedTextColor
	minimapTitleColor  = uiTitleTextColor
	minimapPlayerColor = color.RGBA{R: 255, G: 232, B: 96, A: 255}
	minimapMobColor    = color.RGBA{R: 255, G: 96, B: 96, A: 230}
	minimapNPCColor    = color.RGBA{R: 120, G: 190, B: 255, A: 220}
)

type minimapState struct {
	mapName string
	img     *render.Image
}

type minimapRect struct {
	x int
	y int
	w int
	h int
}

func (m *minimapState) draw(screen *render.Image, ctx Context) {
	if screen == nil || ctx.World == nil {
		return
	}
	width, height := ctx.ScreenSize()
	x, y, w, h := minimapBounds(width, height)
	drawUIPanelSurface(screen, x, y, w, h, uiWindowBodyColor)
	drawUITitleTextAt(screen, x+minimapPad, y, minimapTitleH, "Mini Map", minimapTitleColor)

	mapRect := minimapMapRect(x, y, w, h)
	m.ensureImage(ctx.Resources, ctx.World.MapName)
	if m.img != nil {
		drawMinimapImage(screen, m.img, mapRect)
	} else {
		drawMinimapFallback(screen, mapRect)
	}
	render.DrawRect(screen, float64(mapRect.x), float64(mapRect.y), float64(mapRect.w), 1, uiWindowBorderColor)
	render.DrawRect(screen, float64(mapRect.x), float64(mapRect.y+mapRect.h-1), float64(mapRect.w), 1, uiWindowBorderColor)
	render.DrawRect(screen, float64(mapRect.x), float64(mapRect.y), 1, float64(mapRect.h), uiWindowBorderColor)
	render.DrawRect(screen, float64(mapRect.x+mapRect.w-1), float64(mapRect.y), 1, float64(mapRect.h), uiWindowBorderColor)

	mapW, mapH := minimapWorldSize(ctx.World)
	if mapW > 0 && mapH > 0 {
		m.drawActorMarkers(screen, ctx.World, mapRect, mapW, mapH)
		drawMinimapMarker(screen, mapRect, mapW, mapH, ctx.World.Player.X, ctx.World.Player.Y, minimapPlayerColor, 4)
	}

	label := minimapDisplayName(ctx.World.MapName)
	render.DebugPrintAtColor(screen, trimRunes(label, 13), x+minimapPad, y+h-18, minimapTextColor)
	coords := fmt.Sprintf("X:%d Y:%d", ctx.World.Player.X, ctx.World.Player.Y)
	render.DebugPrintAtColor(screen, coords, x+w-minimapPad-len(coords)*7, y+h-18, minimapMutedColor)
}

func (m *minimapState) ensureImage(manager *res.Manager, mapName string) {
	normalized := normalizeMinimapMapName(mapName)
	if normalized == "" || manager == nil {
		return
	}
	if m.mapName == normalized {
		return
	}
	m.mapName = normalized
	m.img = nil
	img, _, err := res.LoadImage(manager, minimapImageCandidates(normalized))
	if err != nil {
		return
	}
	m.img = render.NewImageFromImage(img)
}

func minimapBounds(width, _ int) (int, int, int, int) {
	x := maxInt(minimapMargin, width-minimapWidth-minimapMargin)
	return x, minimapMargin, minimapWidth, minimapHeight
}

func minimapMapRect(x, y, w, h int) minimapRect {
	available := h - minimapTitleH - minimapFooterH - minimapPad
	size := minInt(w-2*minimapPad, available)
	if size < 32 {
		size = 32
	}
	return minimapRect{
		x: x + (w-size)/2,
		y: y + minimapTitleH + 4,
		w: size,
		h: size,
	}
}

func drawMinimapImage(screen, img *render.Image, dst minimapRect) {
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return
	}
	srcAspect := float64(bounds.Dx()) / float64(bounds.Dy())
	drawW, drawH := dst.w, dst.h
	if srcAspect > 1 {
		drawH = int(float64(drawW)/srcAspect + 0.5)
	} else if srcAspect < 1 {
		drawW = int(float64(drawH)*srcAspect + 0.5)
	}
	drawX := dst.x + (dst.w-drawW)/2
	drawY := dst.y + (dst.h-drawH)/2
	var opts render.DrawImageOptions
	opts.GeoM.Scale(float64(drawW)/float64(bounds.Dx()), float64(drawH)/float64(bounds.Dy()))
	opts.GeoM.Translate(float64(drawX), float64(drawY))
	opts.Filter = render.FilterLinear
	screen.DrawImage(img, &opts)
}

func drawMinimapFallback(screen *render.Image, rect minimapRect) {
	render.DrawRect(screen, float64(rect.x), float64(rect.y), float64(rect.w), float64(rect.h), color.RGBA{R: 212, G: 228, B: 202, A: 255})
	for i := 1; i < 8; i++ {
		x := rect.x + rect.w*i/8
		y := rect.y + rect.h*i/8
		render.DrawRect(screen, float64(x), float64(rect.y), 1, float64(rect.h), color.RGBA{R: 132, G: 164, B: 118, A: 105})
		render.DrawRect(screen, float64(rect.x), float64(y), float64(rect.w), 1, color.RGBA{R: 132, G: 164, B: 118, A: 105})
	}
}

func (m *minimapState) drawActorMarkers(screen *render.Image, world *worldstate.World, rect minimapRect, mapW, mapH int) {
	for _, actor := range world.Actors {
		if actor.ID == 0 || actor.X < 0 || actor.Y < 0 || actor.X >= mapW || actor.Y >= mapH {
			continue
		}
		switch {
		case actor.HasObjectType && actor.ObjectType == actorObjectTypeNPC:
			drawMinimapMarker(screen, rect, mapW, mapH, actor.X, actor.Y, minimapNPCColor, 2)
		case actor.HasObjectType && (actor.ObjectType == actorObjectTypeMob || actor.ObjectType == actorObjectTypeNPCABR || actor.ObjectType == actorObjectTypeNPCBionic):
			drawMinimapMarker(screen, rect, mapW, mapH, actor.X, actor.Y, minimapMobColor, 2)
		}
	}
}

func drawMinimapMarker(screen *render.Image, rect minimapRect, mapW, mapH, cellX, cellY int, fill color.RGBA, radius int) {
	x, y, ok := minimapCellToScreen(rect, mapW, mapH, cellX, cellY)
	if !ok {
		return
	}
	render.DrawRect(screen, float64(x-radius-1), float64(y-radius-1), float64(radius*2+3), float64(radius*2+3), color.RGBA{A: 190})
	render.DrawRect(screen, float64(x-radius), float64(y-radius), float64(radius*2+1), float64(radius*2+1), fill)
}

func minimapCellToScreen(rect minimapRect, mapW, mapH, cellX, cellY int) (int, int, bool) {
	if mapW <= 0 || mapH <= 0 || rect.w <= 0 || rect.h <= 0 {
		return 0, 0, false
	}
	nx := clampUnit((float64(cellX) + 0.5) / float64(mapW))
	ny := clampUnit((float64(cellY) + 0.5) / float64(mapH))
	x := rect.x + int(nx*float64(rect.w-1)+0.5)
	y := rect.y + int((1-ny)*float64(rect.h-1)+0.5)
	return x, y, true
}

func minimapWorldSize(world *worldstate.World) (int, int) {
	if world == nil {
		return 0, 0
	}
	if world.GAT != nil && world.GAT.Width > 0 && world.GAT.Height > 0 {
		return world.GAT.Width, world.GAT.Height
	}
	if world.GND != nil && world.GND.Width > 0 && world.GND.Height > 0 {
		return world.GND.Width, world.GND.Height
	}
	return 0, 0
}

func minimapImageCandidates(mapName string) []string {
	base := normalizeMinimapMapName(mapName)
	if base == "" {
		return nil
	}
	file := base + ".bmp"
	koreanInterface := "\xc0\xaf\xc0\xfa\xc0\xce\xc5\xcd\xc6\xe4\xc0\xcc\xbd\xba"
	return []string{
		"data\\texture\\" + koreanInterface + "\\map\\" + file,
		"data\\texture\\" + koreanInterface + "\\minimap\\" + file,
		"texture\\" + koreanInterface + "\\map\\" + file,
		"texture\\" + koreanInterface + "\\minimap\\" + file,
		"data\\texture\\interface\\map\\" + file,
		"data\\texture\\interface\\minimap\\" + file,
		"data\\texture\\map\\" + file,
		"data\\texture\\minimap\\" + file,
		"texture\\interface\\map\\" + file,
		"texture\\interface\\minimap\\" + file,
		"texture\\map\\" + file,
		"texture\\minimap\\" + file,
		file,
	}
}

func normalizeMinimapMapName(mapName string) string {
	name := strings.TrimSpace(mapName)
	if name == "" {
		return ""
	}
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	name = strings.TrimSuffix(strings.TrimSuffix(name, ".rsw"), ".gat")
	name = strings.TrimSuffix(name, ".gnd")
	name = strings.TrimSuffix(name, ".bmp")
	return strings.ToLower(name)
}

func minimapDisplayName(mapName string) string {
	name := normalizeMinimapMapName(mapName)
	if name == "" {
		return "unknown"
	}
	return name
}
