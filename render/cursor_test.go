package render

import (
	"testing"

	"github.com/gogpu/gpucontext"
)

type recordingCursorApp struct {
	cursors []gpucontext.CursorShape
}

func (a *recordingCursorApp) SetCursor(cursor gpucontext.CursorShape) {
	a.cursors = append(a.cursors, cursor)
}

type recordingPlatformProvider struct {
	gpucontext.NullPlatformProvider
	cursors []gpucontext.CursorShape
}

func (p *recordingPlatformProvider) SetCursor(cursor gpucontext.CursorShape) {
	p.cursors = append(p.cursors, cursor)
}

func resetCursorStateForTest(t *testing.T) {
	t.Helper()
	cursorState.Lock()
	oldApp := cursorState.app
	oldMode := cursorState.mode
	oldAppliedApp := cursorState.appliedApp
	oldAppliedMode := cursorState.appliedMode
	oldApplied := cursorState.applied
	cursorState.app = nil
	cursorState.mode = CursorModeNormal
	cursorState.appliedApp = nil
	cursorState.appliedMode = 0
	cursorState.applied = false
	cursorState.Unlock()
	t.Cleanup(func() {
		cursorState.Lock()
		cursorState.app = oldApp
		cursorState.mode = oldMode
		cursorState.appliedApp = oldAppliedApp
		cursorState.appliedMode = oldAppliedMode
		cursorState.applied = oldApplied
		cursorState.Unlock()
	})
}

func TestSetCursorModeSkipsUnchangedMode(t *testing.T) {
	resetCursorStateForTest(t)
	app := &recordingCursorApp{}
	cursorState.Lock()
	cursorState.app = app
	cursorState.Unlock()

	SetCursorMode(CursorModeHidden)
	SetCursorMode(CursorModeHidden)
	reapplyCursorMode()

	if got, want := len(app.cursors), 1; got != want {
		t.Fatalf("cursor applies = %d, want %d", got, want)
	}
	if got, want := app.cursors[0], gpucontext.CursorNone; got != want {
		t.Fatalf("cursor = %v, want %v", got, want)
	}

	SetCursorMode(CursorModeNormal)
	reapplyCursorMode()

	if got, want := len(app.cursors), 2; got != want {
		t.Fatalf("cursor applies after mode change = %d, want %d", got, want)
	}
	if got, want := app.cursors[1], gpucontext.CursorDefault; got != want {
		t.Fatalf("cursor after mode change = %v, want %v", got, want)
	}
}

func TestROCursorPlatformProviderSuppressesUICursorWhenHidden(t *testing.T) {
	resetCursorStateForTest(t)
	base := &recordingPlatformProvider{}
	provider := roCursorPlatformProvider{PlatformProvider: base}

	SetCursorMode(CursorModeNormal)
	provider.SetCursor(gpucontext.CursorPointer)

	SetCursorMode(CursorModeHidden)
	provider.SetCursor(gpucontext.CursorDefault)

	SetCursorMode(CursorModeNormal)
	provider.SetCursor(gpucontext.CursorText)

	if got, want := len(base.cursors), 2; got != want {
		t.Fatalf("delegated UI cursor applies = %d, want %d", got, want)
	}
	if got, want := base.cursors[0], gpucontext.CursorPointer; got != want {
		t.Fatalf("first delegated cursor = %v, want %v", got, want)
	}
	if got, want := base.cursors[1], gpucontext.CursorText; got != want {
		t.Fatalf("second delegated cursor = %v, want %v", got, want)
	}
}
