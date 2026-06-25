package input

import (
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

type State struct {
	keys      map[ebiten.Key]bool
	prev      map[ebiten.Key]bool
	buttons   map[ebiten.MouseButton]bool
	prevMouse map[ebiten.MouseButton]bool

	MouseX   int
	MouseY   int
	MouseDX  int
	MouseDY  int
	hasMouse bool

	WheelX        float64
	WheelY        float64
	TouchPoints   []TouchPoint
	PinchDelta    float64
	touchIDs      []ebiten.TouchID
	hasPinch      bool
	pinchDistance float64
}

type TouchPoint struct {
	ID ebiten.TouchID
	X  int
	Y  int
}

func NewState() *State {
	return &State{
		keys:      make(map[ebiten.Key]bool),
		prev:      make(map[ebiten.Key]bool),
		buttons:   make(map[ebiten.MouseButton]bool),
		prevMouse: make(map[ebiten.MouseButton]bool),
	}
}

func (s *State) Update() {
	for key, down := range s.keys {
		s.prev[key] = down
	}
	for button, down := range s.buttons {
		s.prevMouse[button] = down
	}

	for _, key := range trackedKeys {
		s.keys[key] = ebiten.IsKeyPressed(key)
	}
	for _, button := range trackedButtons {
		s.buttons[button] = ebiten.IsMouseButtonPressed(button)
	}

	mouseX, mouseY := ebiten.CursorPosition()
	if s.hasMouse {
		s.MouseDX = mouseX - s.MouseX
		s.MouseDY = mouseY - s.MouseY
	} else {
		s.MouseDX = 0
		s.MouseDY = 0
		s.hasMouse = true
	}
	s.MouseX, s.MouseY = mouseX, mouseY

	s.WheelX, s.WheelY = ebiten.Wheel()
	s.updateTouches()
}

func (s *State) Pressed(key ebiten.Key) bool {
	return s.keys[key]
}

func (s *State) JustPressed(key ebiten.Key) bool {
	return s.keys[key] && !s.prev[key]
}

func (s *State) MousePressed(button ebiten.MouseButton) bool {
	return s.buttons[button]
}

func (s *State) MouseJustPressed(button ebiten.MouseButton) bool {
	return s.buttons[button] && !s.prevMouse[button]
}

var trackedKeys = []ebiten.Key{
	ebiten.KeyEnter,
	ebiten.KeyEscape,
	ebiten.KeyTab,
	ebiten.KeyArrowUp,
	ebiten.KeyArrowDown,
	ebiten.KeyArrowLeft,
	ebiten.KeyArrowRight,
	ebiten.KeyQ,
	ebiten.KeyE,
	ebiten.KeyR,
}

var trackedButtons = []ebiten.MouseButton{
	ebiten.MouseButtonLeft,
	ebiten.MouseButtonRight,
}

func (s *State) updateTouches() {
	s.touchIDs = ebiten.AppendTouchIDs(s.touchIDs[:0])
	sort.Slice(s.touchIDs, func(i, j int) bool {
		return s.touchIDs[i] < s.touchIDs[j]
	})
	s.TouchPoints = s.TouchPoints[:0]
	for _, id := range s.touchIDs {
		x, y := ebiten.TouchPosition(id)
		s.TouchPoints = append(s.TouchPoints, TouchPoint{ID: id, X: x, Y: y})
	}
	s.PinchDelta = 0
	if len(s.TouchPoints) < 2 {
		s.hasPinch = false
		s.pinchDistance = 0
		return
	}
	distance := touchDistance(s.TouchPoints[0], s.TouchPoints[1])
	if s.hasPinch {
		s.PinchDelta = distance - s.pinchDistance
	} else {
		s.hasPinch = true
	}
	s.pinchDistance = distance
}

func touchDistance(a, b TouchPoint) float64 {
	return math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y))
}
