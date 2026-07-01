package game

import (
	"path/filepath"
	"testing"

	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/input"
)

type testRuntimeSettings struct {
	fullscreen bool
	vsync      bool
	fps        bool
}

func (s *testRuntimeSettings) Fullscreen() bool {
	return s.fullscreen
}

func (s *testRuntimeSettings) SetFullscreen(value bool) {
	s.fullscreen = value
}

func (s *testRuntimeSettings) VSync() bool {
	return s.vsync
}

func (s *testRuntimeSettings) SetVSync(value bool) {
	s.vsync = value
}

func (s *testRuntimeSettings) FPS() bool {
	return s.fps
}

func (s *testRuntimeSettings) SetFPS(value bool) {
	s.fps = value
}

func TestSettingsWindowEscapeCloses(t *testing.T) {
	var window settingsWindowState
	window.open = true
	inputState := input.NewState()
	ctx := Context{Input: inputState, ScreenW: 800, ScreenH: 600}

	inputState.SetKey(input.KeyEscape, true)
	if !window.update(ctx) {
		t.Fatal("settings window did not consume escape")
	}
	if window.open {
		t.Fatal("settings window stayed open after escape")
	}
}

func TestSettingsWindowCursorOnControls(t *testing.T) {
	var window settingsWindowState
	window.openWindow(Context{ScreenW: 800, ScreenH: 600})
	inputState := input.NewState()
	ctx := Context{Input: inputState, ScreenW: 800, ScreenH: 600}
	x, y, w, h := window.bgmVolumePlusBounds()
	inputState.SetMousePosition(x+w/2, y+h/2)

	action, ok := window.cursorAction(ctx)
	if !ok || action != cursorActionClick {
		t.Fatalf("cursorAction = %d, %t; want click, true", action, ok)
	}
}

func TestSettingsWindowRuntimeToggles(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var window settingsWindowState
	runtime := &testRuntimeSettings{vsync: true}
	inputState := input.NewState()
	ctx := Context{Input: inputState, Runtime: runtime, ScreenW: 800, ScreenH: 600}
	window.openWindow(ctx)

	click := func(bounds func() (int, int, int, int)) {
		x, y, w, h := bounds()
		inputState.SetMousePosition(x+w/2, y+h/2)
		inputState.SetMouseButton(input.MouseButtonLeft, true)
		if !window.update(ctx) {
			t.Fatal("settings window did not consume toggle click")
		}
		inputState.EndFrame()
		inputState.SetMouseButton(input.MouseButtonLeft, false)
		inputState.EndFrame()
	}

	click(window.fullscreenToggleBounds)
	if !runtime.fullscreen {
		t.Fatal("fullscreen toggle did not update runtime state")
	}
	click(window.vsyncToggleBounds)
	if runtime.vsync {
		t.Fatal("vsync toggle did not update runtime state")
	}
	click(window.fpsToggleBounds)
	if !runtime.fps {
		t.Fatal("fps toggle did not update runtime state")
	}
	path, err := config.UserConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "goro.ini" {
		t.Fatalf("settings path = %q, want goro.ini", path)
	}
}

func TestSettingsBoolText(t *testing.T) {
	if settingsBoolText(true) != "on" || settingsBoolText(false) != "off" {
		t.Fatalf("unexpected settings bool text")
	}
}
