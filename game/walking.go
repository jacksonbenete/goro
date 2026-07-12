package game

import (
	"log"
	"math"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

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
			point, ok := projection.projectPoint(cellCenter(float64(x)), cellCenter(float64(y)), terrainHeightAt(ctx.World, float64(x), float64(y)))
			if !ok {
				continue
			}
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
	verts, ok := gatCellVerts(gat, x, y)
	if !ok {
		return [4]screenPoint{}, 0, false
	}
	points, ok := projectGATCellVerts(projection, verts)
	if !ok || !projectedGATCellHasArea(points) {
		return [4]screenPoint{}, 0, false
	}
	return points, projectedGATCellDepth(projection, verts), true
}

func gatCellVerts(gat *res.GAT, x, y int) ([4]modelPoint3, bool) {
	cell, ok := gat.Cell(x, y)
	if !ok {
		return [4]modelPoint3{}, false
	}
	verts := [4]modelPoint3{
		{x: float64(x), y: float64(cell.Heights[0]), z: float64(y)},
		{x: float64(x + 1), y: float64(cell.Heights[1]), z: float64(y)},
		{x: float64(x), y: float64(cell.Heights[2]), z: float64(y + 1)},
		{x: float64(x + 1), y: float64(cell.Heights[3]), z: float64(y + 1)},
	}
	return verts, true
}

func projectGATCellVerts(projection sceneProjection, verts [4]modelPoint3) ([4]screenPoint, bool) {
	var points [4]screenPoint
	for i, vert := range verts {
		point, ok := projection.projectPoint(vert.x, vert.z, vert.y)
		if !ok {
			return [4]screenPoint{}, false
		}
		points[i] = point
	}
	return points, true
}

func projectedGATCellDepth(projection sceneProjection, verts [4]modelPoint3) float64 {
	depth := math.Inf(1)
	for _, vert := range verts {
		depth = math.Min(depth, projection.Depth(vert.x, vert.z, vert.y))
	}
	return depth
}

func projectedGATCellHasArea(points [4]screenPoint) bool {
	const minArea = 0.001
	return math.Abs(screenTriangleArea(points[0], points[1], points[2])) >= minArea ||
		math.Abs(screenTriangleArea(points[2], points[1], points[3])) >= minArea
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

func screenTriangleArea(a, b, c screenPoint) float64 {
	return screenTriangleSign(float64(a.x), float64(a.y), b, c) * 0.5
}

func clickWalkSearchRadius() int {
	return 70
}
