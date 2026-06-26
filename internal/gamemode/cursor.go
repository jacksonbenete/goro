package gamemode

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/kivutar/goro/internal/render"
	worldstate "github.com/kivutar/goro/internal/world"
)

const (
	cursorActionDefault = 0
	cursorActionTalk    = 1
	cursorActionClick   = 2
	cursorActionRotate  = 4
	cursorActionAttack  = 5
	cursorActionWarp    = 7
	cursorActionPick    = 9
	cursorActionNoWalk  = 13
)

type cursorActionInfo struct {
	drawX     float64
	drawY     float64
	delayMult float64
}

var cursorActionInfos = map[int]cursorActionInfo{
	cursorActionDefault: {drawX: 1, drawY: 19, delayMult: 2.0},
	cursorActionTalk:    {drawX: 20, drawY: 40, delayMult: 1.0},
	cursorActionRotate:  {drawX: 18, drawY: 26, delayMult: 1.0},
	cursorActionWarp:    {drawX: 10, drawY: 32, delayMult: 1.0},
	cursorActionPick:    {drawX: 20, drawY: 40, delayMult: 1.0},
	cursorActionNoWalk:  {drawX: 13, drawY: 25, delayMult: 1.0},
}

func (m *WorldMode) drawROCursor(screen *render.Image, ctx Context, projection sceneProjection, now time.Time) {
	if ctx.Input == nil {
		return
	}
	render.SetCursorMode(render.CursorModeHidden)
	action := m.cursorDesiredAction(ctx, projection, now)
	if action != m.cursorAction {
		m.cursorAction = action
		m.cursorStarted = now
	}
	if m.cursorStarted.IsZero() {
		m.cursorStarted = now
	}

	info := cursorInfo(action)
	if m.cursorView == nil || m.cursorViewMiss {
		drawFallbackROCursor(screen, m.cursorFallbackTexture(), ctx.Input.MouseX, ctx.Input.MouseY)
		return
	}
	frame, ok := m.cursorFrame(action, info, now)
	if !ok {
		drawFallbackROCursor(screen, m.cursorFallbackTexture(), ctx.Input.MouseX, ctx.Input.MouseY)
		return
	}
	var opts render.DrawImageOptions
	opts.GeoM.Translate(float64(ctx.Input.MouseX)-frame.anchorX, float64(ctx.Input.MouseY)-frame.anchorY)
	opts.Filter = spriteDrawFilter()
	screen.DrawImage(frame.image, &opts)
}

func (m *WorldMode) cursorFrame(action int, info cursorActionInfo, now time.Time) (*spriteBillboard, bool) {
	if m.cursorView == nil || m.cursorView.act == nil {
		return nil, false
	}
	if action < 0 || action >= len(m.cursorView.act.Actions) || len(m.cursorView.act.Actions[action].Animations) == 0 {
		action = cursorActionDefault
		info = cursorInfo(action)
	}
	actionDef := m.cursorView.act.Actions[action]
	delay := float64(actionDef.DelayMS) * info.delayMult
	motion := spriteMotionIndexWithDelay(actionDef, m.cursorStarted, now, true, delay)
	return cursorFrameBillboard(m.cursorView, action, motion, info.drawX, info.drawY)
}

func (m *WorldMode) cursorDesiredAction(ctx Context, projection sceneProjection, now time.Time) int {
	mouseX, mouseY := ctx.Input.MouseX, ctx.Input.MouseY
	if ctx.Input.MousePressed(render.MouseButtonRight) {
		return cursorActionRotate
	}
	if action, ok := m.escapeMenu.cursorAction(ctx); ok {
		return action
	}
	if action, ok := m.npcDialog.cursorAction(ctx); ok {
		return action
	}
	if action, ok := m.shopWindow.cursorAction(ctx); ok {
		return action
	}
	if action, ok := m.inventoryWindow.cursorAction(ctx); ok {
		return action
	}
	if action, ok := m.inventoryBag.cursorAction(ctx); ok {
		return action
	}
	if action, ok := m.skillWindow.cursorAction(ctx); ok {
		return action
	}
	if action, ok := m.statsWindow.cursorAction(ctx); ok {
		return action
	}
	if action, ok := m.basicMenu.cursorAction(ctx); ok {
		return action
	}
	if _, ok := clickedGroundItem(ctx, projection, mouseX, mouseY, now); ok {
		return cursorActionPick
	}
	if actor, ok := hoveredCursorActor(ctx, projection, mouseX, mouseY, now, m.actorDeaths); ok {
		switch {
		case isWarpActor(actor):
			return cursorActionWarp
		case actorCanBeAttackClicked(ctx, actor):
			return cursorActionAttack
		case cursorActorCanTalk(actor):
			return cursorActionTalk
		}
	}
	if ctx.Input.MousePressed(render.MouseButtonLeft) {
		return cursorActionClick
	}
	if ctx.World != nil && ctx.World.GAT != nil {
		if _, _, ok := hoveredWalkCell(ctx, projection, mouseX, mouseY); !ok {
			return cursorActionNoWalk
		}
	}
	return cursorActionDefault
}

func hoveredCursorActor(ctx Context, projection sceneProjection, mouseX, mouseY int, now time.Time, deadActors map[uint32]time.Time) (worldstate.Actor, bool) {
	if ctx.World == nil {
		return worldstate.Actor{}, false
	}
	bestDistance := math.Inf(1)
	var best worldstate.Actor
	for _, actor := range ctx.World.Actors {
		if _, dead := deadActors[actor.ID]; dead {
			continue
		}
		if isLocalActor(ctx, actor.ID) || actor.ID == 0 {
			continue
		}
		if int(actor.Job) == actorJobClearNPC {
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

func cursorActorCanTalk(actor worldstate.Actor) bool {
	if actor.ID == 0 || !actor.HasObjectType {
		return false
	}
	return actor.ObjectType != actorObjectTypeMob && !isWarpActor(actor)
}

func cursorInfo(action int) cursorActionInfo {
	if info, ok := cursorActionInfos[action]; ok {
		return info
	}
	return cursorActionInfos[cursorActionDefault]
}

func (m *WorldMode) cursorFallbackTexture() *render.Image {
	if m.cursorFallback != nil {
		return m.cursorFallback
	}
	const width = 18
	const height = 24
	mask := []string{
		"X.................",
		"XX................",
		"XOX...............",
		"XOOX..............",
		"XOOOX.............",
		"XOOOOX............",
		"XOOOOOX...........",
		"XOOOOOOX..........",
		"XOOOOOOOX.........",
		"XOOOOOOOOX........",
		"XOOOOOOOOOX.......",
		"XOOOOOOOOOOX......",
		"XOOOOXXXXXXXX.....",
		"XOOXOX............",
		"XOX..OX...........",
		"XX....OX..........",
		"X......OX.........",
		"........X.........",
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y, row := range mask {
		for x, ch := range row {
			switch ch {
			case 'X':
				img.SetRGBA(x, y, color.RGBA{A: 255})
			case 'O':
				img.SetRGBA(x, y, color.RGBA{R: 246, G: 246, B: 246, A: 255})
			}
		}
	}
	m.cursorFallback = render.NewImageFromImage(img)
	return m.cursorFallback
}

func drawFallbackROCursor(screen, img *render.Image, mouseX, mouseY int) {
	if img == nil {
		return
	}
	var opts render.DrawImageOptions
	opts.GeoM.Translate(float64(mouseX), float64(mouseY))
	opts.Filter = spriteDrawFilter()
	screen.DrawImage(img, &opts)
}
