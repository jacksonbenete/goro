package world

import (
	"container/heap"
	"math"
	"time"

	"github.com/kivutar/goro/internal/res"
)

type World struct {
	MapName string
	Player  Actor
	Actors  map[uint32]Actor
	Items   map[uint32]FloorItem
	Camera  Camera
	Dir     int
	GAT     *res.GAT
	GND     *res.GND
	RSW     *res.RSW
	RSM     map[string]*res.RSM
	RSMFail int
}

type FloorItem struct {
	ID         uint32
	ItemID     uint16
	Identified bool
	X          int
	Y          int
	SubX       uint8
	SubY       uint8
	Amount     uint16
	Falling    bool
	DroppedAt  time.Time
}

type Actor struct {
	ID            uint32
	Name          string
	X             int
	Y             int
	Dir           int
	Job           int16
	Head          int16
	Weapon        int16
	Shield        int16
	HeadTop       int16
	HeadMid       int16
	HeadLow       int16
	Sex           byte
	Appearance    bool
	Moving        bool
	FromX         int
	FromY         int
	ToX           int
	ToY           int
	MoveStarted   time.Time
	MoveDuration  time.Duration
	MovePath      []WalkStep
	WalkDistance  float64
	ObjectType    uint8
	HasObjectType bool
	Speed         int
}

type WalkStep struct {
	X int
	Y int
}

type Camera struct {
	X float64
	Y float64
}

func New() *World {
	return &World{
		MapName: "prontera",
		Player:  Actor{Name: "Player", X: 50, Y: 50},
		Actors:  make(map[uint32]Actor),
		Items:   make(map[uint32]FloorItem),
	}
}

func (w *World) MovePlayer(dx, dy int) {
	w.TryMovePlayer(dx, dy)
}

func (w *World) TryMovePlayer(dx, dy int) bool {
	if dx == 0 && dy == 0 {
		return false
	}
	nextX := w.Player.X + dx
	nextY := w.Player.Y + dy
	if w.GAT != nil && !w.GAT.Walkable(nextX, nextY) {
		return false
	}
	w.Player.X = nextX
	w.Player.Y = nextY
	return true
}

func (w *World) SetPlayerPosition(x, y, dir int) {
	w.Player.X = x
	w.Player.Y = y
	w.Player.Dir = dir
	w.Dir = dir
	w.Player.Moving = false
	w.Player.FromX = x
	w.Player.FromY = y
	w.Player.ToX = x
	w.Player.ToY = y
	w.Player.MovePath = nil
}

func (w *World) SetPlayerMovement(fromX, fromY, toX, toY, dir int) {
	w.Player.X = toX
	w.Player.Y = toY
	w.Player.Dir = dir
	w.Player.FromX = fromX
	w.Player.FromY = fromY
	w.Player.ToX = toX
	w.Player.ToY = toY
	w.Player.Moving = true
	w.Player.MoveStarted = time.Now()
	w.Player.MovePath = walkPath(w.GAT, fromX, fromY, toX, toY)
	w.Player.MoveDuration = actorMovementDuration(w.Player.MovePath, fromX, fromY, toX, toY)
	w.Dir = dir
}

func (w *World) UpsertActor(actor Actor) {
	if actor.ID == 0 {
		return
	}
	now := time.Now()
	if existing, ok := w.Actors[actor.ID]; ok {
		if actor.Name == "" {
			actor.Name = existing.Name
		}
		if !actor.Appearance {
			actor.Job = existing.Job
			actor.Head = existing.Head
			actor.Weapon = existing.Weapon
			actor.Shield = existing.Shield
			actor.HeadTop = existing.HeadTop
			actor.HeadMid = existing.HeadMid
			actor.HeadLow = existing.HeadLow
			actor.Sex = existing.Sex
			actor.Appearance = existing.Appearance
		}
		if !actor.HasObjectType {
			actor.ObjectType = existing.ObjectType
			actor.HasObjectType = existing.HasObjectType
		}
		if actor.Speed <= 0 {
			actor.Speed = existing.Speed
		}
		if actor.Moving && actor.FromX == 0 && actor.FromY == 0 {
			actor.FromX = existing.X
			actor.FromY = existing.Y
		}
		if actor.Moving && existing.IsMovingAt(now) {
			actor.WalkDistance = existing.RenderWalkDistance(now)
		}
	}
	if actor.Moving {
		if actor.ToX == 0 && actor.ToY == 0 {
			actor.ToX = actor.X
			actor.ToY = actor.Y
		}
		actor.MoveStarted = now
		actor.MovePath = walkPath(w.GAT, actor.FromX, actor.FromY, actor.ToX, actor.ToY)
		actor.MoveDuration = actorMovementDurationWithSpeed(actor.MovePath, actor.FromX, actor.FromY, actor.ToX, actor.ToY, actorMoveSpeed(actor))
	} else {
		actor.FromX = actor.X
		actor.FromY = actor.Y
		actor.ToX = actor.X
		actor.ToY = actor.Y
		actor.MovePath = nil
		actor.WalkDistance = 0
	}
	w.Actors[actor.ID] = actor
}

func (w *World) RemoveActor(id uint32) {
	delete(w.Actors, id)
}

func (w *World) UpsertItem(item FloorItem) {
	if item.ID == 0 {
		return
	}
	if item.DroppedAt.IsZero() {
		item.DroppedAt = time.Now()
	}
	if w.Items == nil {
		w.Items = make(map[uint32]FloorItem)
	}
	w.Items[item.ID] = item
}

func (w *World) RemoveItem(id uint32) {
	delete(w.Items, id)
}

func (a Actor) IsMovingAt(now time.Time) bool {
	return a.Moving && !a.MoveStarted.IsZero() && now.Before(a.MoveStarted.Add(a.MoveDuration))
}

func (a Actor) RenderPosition(now time.Time) (float64, float64) {
	if !a.IsMovingAt(now) || a.MoveDuration <= 0 {
		return float64(a.X), float64(a.Y)
	}
	elapsed := now.Sub(a.MoveStarted)
	if len(a.MovePath) >= 2 {
		return renderPathPositionWithSpeed(a.MovePath, elapsed, actorMoveSpeed(a))
	}
	t := float64(elapsed) / float64(a.MoveDuration)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	x := float64(a.FromX) + float64(a.ToX-a.FromX)*t
	y := float64(a.FromY) + float64(a.ToY-a.FromY)*t
	return x, y
}

func (a Actor) RenderDirection(now time.Time) int {
	if !a.IsMovingAt(now) || len(a.MovePath) < 2 {
		return a.Dir
	}
	elapsed := now.Sub(a.MoveStarted)
	from, to := renderPathSegmentWithSpeed(a.MovePath, elapsed, actorMoveSpeed(a))
	if from == to {
		return a.Dir
	}
	return DirectionFromDelta(from.X, from.Y, to.X, to.Y, a.Dir)
}

func (a Actor) RenderWalkDistance(now time.Time) float64 {
	if !a.IsMovingAt(now) || a.MoveDuration <= 0 {
		return 0
	}
	base := a.WalkDistance
	elapsed := now.Sub(a.MoveStarted)
	if elapsed < 0 {
		elapsed = 0
	}
	if len(a.MovePath) >= 2 {
		return base + pathWalkDistanceWithSpeed(a.MovePath, elapsed, actorMoveSpeed(a))
	}
	t := float64(elapsed) / float64(a.MoveDuration)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return base + math.Hypot(float64(a.ToX-a.FromX), float64(a.ToY-a.FromY))*t
}

func actorMovementDuration(path []WalkStep, fromX, fromY, toX, toY int) time.Duration {
	return actorMovementDurationWithSpeed(path, fromX, fromY, toX, toY, defaultMoveSpeedMS)
}

func actorMovementDurationWithSpeed(path []WalkStep, fromX, fromY, toX, toY int, speedMS int) time.Duration {
	if len(path) >= 2 {
		total := time.Duration(0)
		for i := 1; i < len(path); i++ {
			total += movementSegmentDurationWithSpeed(path[i].X-path[i-1].X, path[i].Y-path[i-1].Y, speedMS)
		}
		if total > 0 {
			return total
		}
	}
	dx := absInt(toX - fromX)
	dy := absInt(toY - fromY)
	if dx == 0 && dy == 0 {
		return movementSegmentDurationWithSpeed(0, 0, speedMS)
	}
	return movementSegmentDurationWithSpeed(dx, dy, speedMS)
}

func movementSegmentDuration(dx, dy int) time.Duration {
	return movementSegmentDurationWithSpeed(dx, dy, defaultMoveSpeedMS)
}

const defaultMoveSpeedMS = 150

func actorMoveSpeed(actor Actor) int {
	if actor.Speed > 0 {
		return actor.Speed
	}
	return defaultMoveSpeedMS
}

func movementSegmentDurationWithSpeed(dx, dy int, speedMS int) time.Duration {
	if speedMS <= 0 {
		speedMS = defaultMoveSpeedMS
	}
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx == 0 && dy == 0 {
		return time.Duration(speedMS) * time.Millisecond
	}
	if dx != 0 && dy != 0 && dx == dy {
		return time.Duration(math.Round(float64(dx)*float64(speedMS)*math.Sqrt2)) * time.Millisecond
	}
	steps := dx
	if dy > steps {
		steps = dy
	}
	return time.Duration(steps*speedMS) * time.Millisecond
}

func renderPathPosition(path []WalkStep, elapsed time.Duration) (float64, float64) {
	return renderPathPositionWithSpeed(path, elapsed, defaultMoveSpeedMS)
}

func renderPathPositionWithSpeed(path []WalkStep, elapsed time.Duration, speedMS int) (float64, float64) {
	from, to := renderPathSegmentWithSpeed(path, elapsed, speedMS)
	segmentElapsed := elapsed - pathElapsedBeforeSegmentWithSpeed(path, from, to, speedMS)
	duration := movementSegmentDurationWithSpeed(to.X-from.X, to.Y-from.Y, speedMS)
	if duration <= 0 {
		return float64(to.X), float64(to.Y)
	}
	t := float64(segmentElapsed) / float64(duration)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	x := float64(from.X) + float64(to.X-from.X)*t
	y := float64(from.Y) + float64(to.Y-from.Y)*t
	return x, y
}

func renderPathSegment(path []WalkStep, elapsed time.Duration) (WalkStep, WalkStep) {
	return renderPathSegmentWithSpeed(path, elapsed, defaultMoveSpeedMS)
}

func renderPathSegmentWithSpeed(path []WalkStep, elapsed time.Duration, speedMS int) (WalkStep, WalkStep) {
	if len(path) == 0 {
		return WalkStep{}, WalkStep{}
	}
	if len(path) == 1 || elapsed <= 0 {
		return path[0], path[0]
	}
	remaining := elapsed
	for i := 1; i < len(path); i++ {
		duration := movementSegmentDurationWithSpeed(path[i].X-path[i-1].X, path[i].Y-path[i-1].Y, speedMS)
		if remaining < duration || i == len(path)-1 {
			return path[i-1], path[i]
		}
		remaining -= duration
	}
	return path[len(path)-1], path[len(path)-1]
}

func pathElapsedBeforeSegment(path []WalkStep, from, to WalkStep) time.Duration {
	return pathElapsedBeforeSegmentWithSpeed(path, from, to, defaultMoveSpeedMS)
}

func pathElapsedBeforeSegmentWithSpeed(path []WalkStep, from, to WalkStep, speedMS int) time.Duration {
	total := time.Duration(0)
	for i := 1; i < len(path); i++ {
		if path[i-1] == from && path[i] == to {
			return total
		}
		total += movementSegmentDurationWithSpeed(path[i].X-path[i-1].X, path[i].Y-path[i-1].Y, speedMS)
	}
	return total
}

func pathWalkDistanceWithSpeed(path []WalkStep, elapsed time.Duration, speedMS int) float64 {
	if len(path) < 2 || elapsed <= 0 {
		return 0
	}
	remaining := elapsed
	total := 0.0
	for i := 1; i < len(path); i++ {
		from := path[i-1]
		to := path[i]
		segmentDistance := math.Hypot(float64(to.X-from.X), float64(to.Y-from.Y))
		duration := movementSegmentDurationWithSpeed(to.X-from.X, to.Y-from.Y, speedMS)
		if duration <= 0 {
			total += segmentDistance
			continue
		}
		if remaining >= duration {
			total += segmentDistance
			remaining -= duration
			continue
		}
		t := float64(remaining) / float64(duration)
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		total += segmentDistance * t
		break
	}
	return total
}

func walkPath(gat *res.GAT, fromX, fromY, toX, toY int) []WalkStep {
	if fromX == toX && fromY == toY {
		return []WalkStep{{X: fromX, Y: fromY}}
	}
	if gat == nil || !gat.InBounds(fromX, fromY) || !gat.InBounds(toX, toY) {
		return []WalkStep{{X: fromX, Y: fromY}, {X: toX, Y: toY}}
	}
	if path, ok := findWalkPath(gat, fromX, fromY, toX, toY); ok {
		return path
	}
	return []WalkStep{{X: fromX, Y: fromY}, {X: toX, Y: toY}}
}

func findWalkPath(gat *res.GAT, fromX, fromY, toX, toY int) ([]WalkStep, bool) {
	start := pathPoint{x: fromX, y: fromY}
	goal := pathPoint{x: toX, y: toY}
	open := &pathHeap{}
	heap.Init(open)
	startNode := &pathNode{point: start, g: 0, f: pathHeuristic(start, goal)}
	heap.Push(open, startNode)
	nodes := map[pathPoint]*pathNode{start: startNode}
	closed := make(map[pathPoint]struct{})

	for open.Len() > 0 {
		current := heap.Pop(open).(*pathNode)
		if _, ok := closed[current.point]; ok {
			continue
		}
		if current.point == goal {
			return reconstructPath(current), true
		}
		closed[current.point] = struct{}{}
		for _, next := range pathNeighbors(gat, current.point, goal) {
			if _, ok := closed[next]; ok {
				continue
			}
			cost := current.g + pathStepCost(current.point, next)
			node, ok := nodes[next]
			if !ok {
				node = &pathNode{point: next, g: cost, f: cost + pathHeuristic(next, goal), parent: current}
				nodes[next] = node
				heap.Push(open, node)
				continue
			}
			if cost < node.g {
				node.g = cost
				node.f = cost + pathHeuristic(next, goal)
				node.parent = current
				heap.Push(open, node)
			}
		}
	}
	return nil, false
}

type pathPoint struct {
	x int
	y int
}

type pathNode struct {
	point  pathPoint
	g      int
	f      int
	parent *pathNode
}

type pathHeap []*pathNode

func (h pathHeap) Len() int { return len(h) }
func (h pathHeap) Less(i, j int) bool {
	if h[i].f == h[j].f {
		return h[i].g > h[j].g
	}
	return h[i].f < h[j].f
}
func (h pathHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *pathHeap) Push(x any)   { *h = append(*h, x.(*pathNode)) }
func (h *pathHeap) Pop() any {
	old := *h
	n := len(old)
	node := old[n-1]
	*h = old[:n-1]
	return node
}

func pathNeighbors(gat *res.GAT, point pathPoint, goal pathPoint) []pathPoint {
	var out []pathPoint
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			next := pathPoint{x: point.x + dx, y: point.y + dy}
			if !gat.InBounds(next.x, next.y) {
				continue
			}
			if !gat.Walkable(next.x, next.y) {
				continue
			}
			if dx != 0 && dy != 0 && (!gat.Walkable(point.x+dx, point.y) || !gat.Walkable(point.x, point.y+dy)) {
				continue
			}
			out = append(out, next)
		}
	}
	return out
}

func pathStepCost(from, to pathPoint) int {
	if from.x != to.x && from.y != to.y {
		return 14
	}
	return 10
}

func pathHeuristic(from, to pathPoint) int {
	dx := absInt(to.x - from.x)
	dy := absInt(to.y - from.y)
	if dx < dy {
		return 14*dx + 10*(dy-dx)
	}
	return 14*dy + 10*(dx-dy)
}

func reconstructPath(node *pathNode) []WalkStep {
	var reversed []WalkStep
	for node != nil {
		reversed = append(reversed, WalkStep{X: node.point.x, Y: node.point.y})
		node = node.parent
	}
	path := make([]WalkStep, len(reversed))
	for i := range reversed {
		path[i] = reversed[len(reversed)-1-i]
	}
	return path
}

func DirectionFromDelta(fromX, fromY, toX, toY int, fallback int) int {
	dx := toX - fromX
	dy := toY - fromY
	if dx == 0 && dy == 0 {
		if fallback < 0 {
			return 0
		}
		return fallback & 7
	}
	actionDir := int(math.Round(-math.Atan2(float64(dy), float64(dx))/(math.Pi/4)+6)) & 7
	return (4 - actionDir) & 7
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
