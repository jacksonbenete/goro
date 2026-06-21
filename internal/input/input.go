package input

import "github.com/hajimehoshi/ebiten/v2"

type State struct {
	keys      map[ebiten.Key]bool
	prev      map[ebiten.Key]bool
	buttons   map[ebiten.MouseButton]bool
	prevMouse map[ebiten.MouseButton]bool

	MouseX int
	MouseY int
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

	s.MouseX, s.MouseY = ebiten.CursorPosition()
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
}

var trackedButtons = []ebiten.MouseButton{
	ebiten.MouseButtonLeft,
	ebiten.MouseButtonRight,
}
