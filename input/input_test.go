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

func TestConsumeKeyCodePressPreservesHeldStateAndClaimsLegacyShortcut(t *testing.T) {
	state := NewState()
	code, ok := KeyCodeFromName("KeyW")
	if !ok {
		t.Fatal("KeyW was not recognized")
	}
	state.SetKeyCode(code, true)

	if !state.ConsumeKeyCodePress(code) {
		t.Fatal("ConsumeKeyCodePress = false, want true")
	}
	if state.KeyCodeJustPressed(code) {
		t.Fatal("physical press remained visible after consumption")
	}
	if state.JustPressed(KeyW) {
		t.Fatal("legacy shortcut press remained visible after consumption")
	}
	if !state.KeyCodeDown(code) {
		t.Fatal("held state was cleared after consuming key edge")
	}
	if state.ConsumeKeyCodePress(code) {
		t.Fatal("second ConsumeKeyCodePress = true, want false")
	}

	state.EndFrame()
	state.SetKeyCode(code, false)
	if !state.KeyCodeJustReleased(code) {
		t.Fatal("physical key release was not reported")
	}
}

func TestKeyCodeNamesCoverPhysicalAndNonTextKeys(t *testing.T) {
	for _, name := range []string{"KeyA", "KeyZ", "Digit0", "F12", "ArrowUp", "ControlRight", "NumpadEnter"} {
		if _, ok := KeyCodeFromName(name); !ok {
			t.Fatalf("KeyCodeFromName(%q) = false", name)
		}
	}
	if _, ok := KeyCodeFromName("a"); ok {
		t.Fatal("layout-dependent glyph unexpectedly resolved as a key code")
	}
}
