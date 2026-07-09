package world

import (
	"container/heap"
	"math"
	"time"

	"github.com/kivutar/goro/res"
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
	HeadDir       uint8
	Appearance    bool
	Moving        bool
	FromX         int
	FromY         int
	ToX           int
	ToY           int
	MoveStarted   time.Time
	MoveDuration  time.Duration
	MovePath      []WalkStep
	MoveStartX    float64
	MoveStartY    float64
	HasMoveStart  bool
	WalkDistance  float64
	ObjectType    uint8
	HasObjectType bool
	Speed         int
	Sitting       bool
	BodyState     uint16
	HealthState   uint16
	EffectState   uint32
	HasState      bool
	HasCart       bool
	CartNum       int
	HasCartState  bool
	Vending       bool
	VendingName   string
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

func (w *World) SetPlayerPosition(x, y, dir int) {
	w.Player.X = x
	w.Player.Y = y
	w.Player.Dir = dir
	w.Player.HeadDir = 0
	w.Dir = dir
	w.Player.Moving = false
	w.Player.FromX = x
	w.Player.FromY = y
	w.Player.ToX = x
	w.Player.ToY = y
	w.Player.MovePath = nil
	w.Player.HasMoveStart = false
}

func (w *World) SetPlayerMovement(fromX, fromY, toX, toY, dir int) {
	w.SetPlayerMovementAt(fromX, fromY, toX, toY, dir, time.Now(), 0)
}

func (w *World) SetPlayerMovementAt(fromX, fromY, toX, toY, dir int, now time.Time, fastForward time.Duration) {
	if now.IsZero() {
		now = time.Now()
	}
	oldPlayer := w.Player
	path := walkPath(w.GAT, fromX, fromY, toX, toY)
	finalDir := movementFinalDirection(path, fromX, fromY, toX, toY, dir)
	speed := actorMoveSpeed(w.Player)
	startX, startY, hasMoveStart, offset := movementStart(path, oldPlayer, now, fastForward, speed, fromX, fromY)
	duration := actorMovementDurationFromWithSpeed(path, fromX, fromY, toX, toY, speed, startX, startY, hasMoveStart)
	w.Player.X = toX
	w.Player.Y = toY
	w.Player.Dir = finalDir
	w.Player.HeadDir = 0
	w.Player.FromX = fromX
	w.Player.FromY = fromY
	w.Player.ToX = toX
	w.Player.ToY = toY
	w.Player.Moving = true
	w.Player.Sitting = false
	w.Player.MoveStarted = now.Add(-offset)
	w.Player.MovePath = path
	w.Player.MoveDuration = duration
	w.Player.MoveStartX = startX
	w.Player.MoveStartY = startY
	w.Player.HasMoveStart = hasMoveStart
	w.Player.WalkDistance = movementWalkDistanceOffset(path, oldPlayer, now, offset, speed, startX, startY, hasMoveStart)
	w.Dir = finalDir
}

func movementFinalDirection(path []WalkStep, fromX, fromY, toX, toY int, fallback int) int {
	if len(path) >= 2 {
		from := path[len(path)-2]
		to := path[len(path)-1]
		return DirectionFromDelta(from.X, from.Y, to.X, to.Y, fallback)
	}
	return DirectionFromDelta(fromX, fromY, toX, toY, fallback)
}

func movementStart(path []WalkStep, oldPlayer Actor, now time.Time, fastForward time.Duration, speedMS, fromX, fromY int) (float64, float64, bool, time.Duration) {
	startX := float64(fromX)
	startY := float64(fromY)
	duration := time.Duration(0)
	if len(path) > 0 {
		duration = actorMovementDurationWithSpeed(path, path[0].X, path[0].Y, path[len(path)-1].X, path[len(path)-1].Y, speedMS)
	}
	if duration > 0 && fastForward > duration*4 {
		fastForward = 0
	}
	offset := clampMovementOffset(fastForward, duration)
	if oldPlayer.IsMovingAt(now) && len(path) >= 2 {
		x, y := oldPlayer.RenderPosition(now)
		if currentOffset, ok := pathElapsedAtPositionWithSpeed(path, x, y, speedMS); ok {
			if currentOffset > offset {
				offset = clampMovementOffset(currentOffset, duration)
			}
			return startX, startY, false, offset
		}
		if math.Hypot(x-float64(fromX), y-float64(fromY)) <= 1.25 {
			return x, y, true, 0
		}
		if nearMovementStartLine(path, x, y, 3.0, 0.25) {
			return x, y, true, 0
		}
	}
	return startX, startY, false, offset
}

func nearMovementStartLine(path []WalkStep, x, y float64, maxDistance, maxLateral float64) bool {
	if len(path) < 2 {
		return false
	}
	from := path[0]
	to := path[1]
	fromX := float64(from.X)
	fromY := float64(from.Y)
	dx := float64(to.X - from.X)
	dy := float64(to.Y - from.Y)
	lengthSq := dx*dx + dy*dy
	if lengthSq == 0 {
		return false
	}
	if math.Hypot(x-fromX, y-fromY) > maxDistance {
		return false
	}
	t := ((x-fromX)*dx + (y-fromY)*dy) / lengthSq
	if t > 0.25 {
		return false
	}
	px := fromX + dx*t
	py := fromY + dy*t
	return math.Hypot(x-px, y-py) <= maxLateral
}

func movementWalkDistanceOffset(path []WalkStep, oldPlayer Actor, now time.Time, offset time.Duration, speedMS int, startX, startY float64, hasMoveStart bool) float64 {
	if !oldPlayer.IsMovingAt(now) {
		return 0
	}
	previous := oldPlayer.RenderWalkDistance(now)
	current := pathWalkDistanceWithSpeed(path, offset, speedMS, startX, startY, hasMoveStart)
	return previous - current
}

func clampMovementOffset(offset, duration time.Duration) time.Duration {
	if offset < 0 {
		return 0
	}
	if duration > 0 && offset > duration {
		return duration
	}
	return offset
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
			actor.HeadDir = existing.HeadDir
			actor.Appearance = existing.Appearance
		}
		if !actor.HasObjectType {
			actor.ObjectType = existing.ObjectType
			actor.HasObjectType = existing.HasObjectType
		}
		if actor.Speed <= 0 {
			actor.Speed = existing.Speed
		}
		if !actor.HasState {
			actor.BodyState = existing.BodyState
			actor.HealthState = existing.HealthState
			actor.EffectState = existing.EffectState
			actor.HasState = existing.HasState
		}
		if !actor.HasCartState {
			actor.HasCart = existing.HasCart
			actor.CartNum = existing.CartNum
			actor.HasCartState = existing.HasCartState
		}
		if actor.Moving && actor.FromX == 0 && actor.FromY == 0 {
			actor.FromX = existing.X
			actor.FromY = existing.Y
		}
		if actor.Moving && existing.IsMovingAt(now) {
			actor.WalkDistance = existing.RenderWalkDistance(now)
		}
		if existing.Sitting && !actor.Moving {
			actor.Sitting = true
		}
		if existing.HeadDir != 0 && actor.HeadDir == 0 && !actor.Moving {
			actor.HeadDir = existing.HeadDir
		}
	}
	if actor.Moving {
		actor.Sitting = false
		actor.HeadDir = 0
	}
	if actor.Moving {
		if actor.ToX == 0 && actor.ToY == 0 {
			actor.ToX = actor.X
			actor.ToY = actor.Y
		}
		actor.MoveStarted = now
		actor.MovePath = walkPath(w.GAT, actor.FromX, actor.FromY, actor.ToX, actor.ToY)
		actor.MoveDuration = actorMovementDurationWithSpeed(actor.MovePath, actor.FromX, actor.FromY, actor.ToX, actor.ToY, actorMoveSpeed(actor))
		actor.HasMoveStart = false
	} else {
		actor.FromX = actor.X
		actor.FromY = actor.Y
		actor.ToX = actor.X
		actor.ToY = actor.Y
		actor.MovePath = nil
		actor.HasMoveStart = false
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
		return renderPathPositionWithSpeed(a.MovePath, elapsed, actorMoveSpeed(a), a.MoveStartX, a.MoveStartY, a.HasMoveStart)
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
	fromX, fromY, toX, toY := renderPathSegmentWithSpeed(a.MovePath, elapsed, actorMoveSpeed(a), a.MoveStartX, a.MoveStartY, a.HasMoveStart)
	if math.Abs(fromX-toX) < 0.001 && math.Abs(fromY-toY) < 0.001 {
		return a.Dir
	}
	return DirectionFromFloatDelta(fromX, fromY, toX, toY, a.Dir)
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
		return base + pathWalkDistanceWithSpeed(a.MovePath, elapsed, actorMoveSpeed(a), a.MoveStartX, a.MoveStartY, a.HasMoveStart)
	}
	t := float64(elapsed) / float64(a.MoveDuration)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return base + math.Hypot(float64(a.ToX-a.FromX), float64(a.ToY-a.FromY))*t
}

func actorMovementDurationWithSpeed(path []WalkStep, fromX, fromY, toX, toY int, speedMS int) time.Duration {
	return actorMovementDurationFromWithSpeed(path, fromX, fromY, toX, toY, speedMS, 0, 0, false)
}

func actorMovementDurationFromWithSpeed(path []WalkStep, fromX, fromY, toX, toY int, speedMS int, startX, startY float64, hasMoveStart bool) time.Duration {
	if len(path) >= 2 {
		total := time.Duration(0)
		for i := 1; i < len(path); i++ {
			segmentStartX, segmentStartY := pathSegmentStart(path, i, startX, startY, hasMoveStart)
			total += movementSegmentDurationFloatWithSpeed(float64(path[i].X)-segmentStartX, float64(path[i].Y)-segmentStartY, speedMS)
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

func movementSegmentDurationFloatWithSpeed(dx, dy float64, speedMS int) time.Duration {
	if speedMS <= 0 {
		speedMS = defaultMoveSpeedMS
	}
	dist := math.Hypot(dx, dy)
	if dist <= 0 {
		return 0
	}
	return time.Duration(math.Round(dist*float64(speedMS))) * time.Millisecond
}

func renderPathPositionWithSpeed(path []WalkStep, elapsed time.Duration, speedMS int, startX, startY float64, hasMoveStart bool) (float64, float64) {
	if len(path) == 0 {
		return 0, 0
	}
	if len(path) == 1 {
		return float64(path[0].X), float64(path[0].Y)
	}
	if elapsed < 0 {
		elapsed = 0
	}
	remaining := elapsed
	for i := 1; i < len(path); i++ {
		fromX, fromY := pathSegmentStart(path, i, startX, startY, hasMoveStart)
		toX := float64(path[i].X)
		toY := float64(path[i].Y)
		duration := movementSegmentDurationFloatWithSpeed(toX-fromX, toY-fromY, speedMS)
		if duration <= 0 {
			continue
		}
		if remaining <= duration || i == len(path)-1 {
			t := float64(remaining) / float64(duration)
			if t < 0 {
				t = 0
			} else if t > 1 {
				t = 1
			}
			return fromX + (toX-fromX)*t, fromY + (toY-fromY)*t
		}
		remaining -= duration
	}
	last := path[len(path)-1]
	return float64(last.X), float64(last.Y)
}

func renderPathSegmentWithSpeed(path []WalkStep, elapsed time.Duration, speedMS int, startX, startY float64, hasMoveStart bool) (float64, float64, float64, float64) {
	if len(path) == 0 {
		return 0, 0, 0, 0
	}
	if len(path) == 1 {
		x := float64(path[0].X)
		y := float64(path[0].Y)
		return x, y, x, y
	}
	if elapsed <= 0 {
		fromX, fromY := pathSegmentStart(path, 1, startX, startY, hasMoveStart)
		return fromX, fromY, float64(path[1].X), float64(path[1].Y)
	}
	remaining := elapsed
	for i := 1; i < len(path); i++ {
		fromX, fromY := pathSegmentStart(path, i, startX, startY, hasMoveStart)
		toX := float64(path[i].X)
		toY := float64(path[i].Y)
		duration := movementSegmentDurationFloatWithSpeed(toX-fromX, toY-fromY, speedMS)
		if remaining < duration || i == len(path)-1 {
			return fromX, fromY, toX, toY
		}
		remaining -= duration
	}
	last := path[len(path)-1]
	x := float64(last.X)
	y := float64(last.Y)
	return x, y, x, y
}

func pathElapsedAtPositionWithSpeed(path []WalkStep, x, y float64, speedMS int) (time.Duration, bool) {
	if len(path) < 2 {
		return 0, false
	}
	bestDistance := math.Inf(1)
	bestElapsed := time.Duration(0)
	elapsedBefore := time.Duration(0)
	for i := 1; i < len(path); i++ {
		from := path[i-1]
		to := path[i]
		dx := float64(to.X - from.X)
		dy := float64(to.Y - from.Y)
		segmentDuration := movementSegmentDurationWithSpeed(to.X-from.X, to.Y-from.Y, speedMS)
		segmentLengthSq := dx*dx + dy*dy
		if segmentLengthSq == 0 {
			elapsedBefore += segmentDuration
			continue
		}
		t := ((x-float64(from.X))*dx + (y-float64(from.Y))*dy) / segmentLengthSq
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		px := float64(from.X) + dx*t
		py := float64(from.Y) + dy*t
		distance := math.Hypot(x-px, y-py)
		if distance < bestDistance {
			bestDistance = distance
			bestElapsed = elapsedBefore + time.Duration(float64(segmentDuration)*t)
		}
		elapsedBefore += segmentDuration
	}
	if bestDistance > 0.05 {
		return 0, false
	}
	return bestElapsed, true
}

func pathWalkDistanceWithSpeed(path []WalkStep, elapsed time.Duration, speedMS int, startX, startY float64, hasMoveStart bool) float64 {
	if len(path) < 2 || elapsed <= 0 {
		return 0
	}
	remaining := elapsed
	total := 0.0
	for i := 1; i < len(path); i++ {
		fromX, fromY := pathSegmentStart(path, i, startX, startY, hasMoveStart)
		toX := float64(path[i].X)
		toY := float64(path[i].Y)
		segmentDistance := math.Hypot(toX-fromX, toY-fromY)
		duration := movementSegmentDurationFloatWithSpeed(toX-fromX, toY-fromY, speedMS)
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

func pathSegmentStart(path []WalkStep, segmentIndex int, startX, startY float64, hasMoveStart bool) (float64, float64) {
	if segmentIndex == 1 && hasMoveStart {
		return startX, startY
	}
	from := path[segmentIndex-1]
	return float64(from.X), float64(from.Y)
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
	return DirectionFromFloatDelta(float64(fromX), float64(fromY), float64(toX), float64(toY), fallback)
}

func DirectionFromFloatDelta(fromX, fromY, toX, toY float64, fallback int) int {
	dx := toX - fromX
	dy := toY - fromY
	if math.Abs(dx) < 0.001 && math.Abs(dy) < 0.001 {
		if fallback < 0 {
			return 0
		}
		return fallback & 7
	}
	actionDir := int(math.Round(-math.Atan2(dy, dx)/(math.Pi/4)+6)) & 7
	return (4 - actionDir) & 7
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
