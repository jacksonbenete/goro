package input

import "testing"

func TestTouchDistance(t *testing.T) {
	got := touchDistance(TouchPoint{X: 10, Y: 20}, TouchPoint{X: 13, Y: 24})
	if got != 5 {
		t.Fatalf("touch distance = %.1f, want 5.0", got)
	}
}

func TestMouseJustReleased(t *testing.T) {
	state := NewState()
	state.SetMouseButton(MouseButtonLeft, true)
	state.EndFrame()

	state.SetMouseButton(MouseButtonLeft, false)
	if !state.MouseJustReleased(MouseButtonLeft) {
		t.Fatal("MouseJustReleased = false, want true")
	}

	state.EndFrame()
	if state.MouseJustReleased(MouseButtonLeft) {
		t.Fatal("MouseJustReleased persisted after EndFrame")
	}
}
