package world

import (
	"time"

	"github.com/kivutar/goro/internal/res"
)

type World struct {
	MapName string
	Player  Actor
	Actors  map[uint32]Actor
	Camera  Camera
	Dir     int
	GAT     *res.GAT
	GND     *res.GND
	RSW     *res.RSW
	RSM     map[string]*res.RSM
	RSMFail int
}

type Actor struct {
	ID           uint32
	Name         string
	X            int
	Y            int
	Dir          int
	Job          int16
	Head         int16
	Sex          byte
	Appearance   bool
	Moving       bool
	FromX        int
	FromY        int
	ToX          int
	ToY          int
	MoveStarted  time.Time
	MoveDuration time.Duration
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
	w.Player.MoveDuration = actorMovementDuration(fromX, fromY, toX, toY)
	w.Dir = dir
}

func (w *World) UpsertActor(actor Actor) {
	if actor.ID == 0 {
		return
	}
	if actor.Name == "" {
		actor.Name = "actor"
	}
	if existing, ok := w.Actors[actor.ID]; ok {
		if !actor.Appearance {
			actor.Job = existing.Job
			actor.Head = existing.Head
			actor.Sex = existing.Sex
			actor.Appearance = existing.Appearance
		}
		if actor.Moving && actor.FromX == 0 && actor.FromY == 0 {
			actor.FromX = existing.X
			actor.FromY = existing.Y
		}
	}
	if actor.Moving {
		if actor.ToX == 0 && actor.ToY == 0 {
			actor.ToX = actor.X
			actor.ToY = actor.Y
		}
		actor.MoveStarted = time.Now()
		actor.MoveDuration = actorMovementDuration(actor.FromX, actor.FromY, actor.ToX, actor.ToY)
	} else {
		actor.FromX = actor.X
		actor.FromY = actor.Y
		actor.ToX = actor.X
		actor.ToY = actor.Y
	}
	w.Actors[actor.ID] = actor
}

func (w *World) RemoveActor(id uint32) {
	delete(w.Actors, id)
}

func (a Actor) IsMovingAt(now time.Time) bool {
	return a.Moving && !a.MoveStarted.IsZero() && now.Before(a.MoveStarted.Add(a.MoveDuration))
}

func (a Actor) RenderPosition(now time.Time) (float64, float64) {
	if !a.IsMovingAt(now) || a.MoveDuration <= 0 {
		return float64(a.X), float64(a.Y)
	}
	elapsed := now.Sub(a.MoveStarted)
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

func actorMovementDuration(fromX, fromY, toX, toY int) time.Duration {
	dx := absInt(toX - fromX)
	dy := absInt(toY - fromY)
	steps := dx
	if dy > steps {
		steps = dy
	}
	if steps < 1 {
		steps = 1
	}
	return time.Duration(steps) * 150 * time.Millisecond
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
