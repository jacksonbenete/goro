package render

import (
	"sync"

	"github.com/gogpu/gogpu"
	"github.com/gogpu/gpucontext"
)

const (
	CursorModeNormal = 0
	CursorModeHidden = 1
)

type cursorApp interface {
	SetCursor(gpucontext.CursorShape)
}

var cursorState = struct {
	sync.Mutex
	app         cursorApp
	mode        int
	appliedApp  cursorApp
	appliedMode int
	applied     bool
}{
	mode: CursorModeNormal,
}

func setCursorApp(app *gogpu.App) {
	var cursorApp cursorApp
	if app != nil {
		cursorApp = app
	}
	cursorState.Lock()
	cursorState.app = cursorApp
	cursorState.appliedApp = nil
	cursorState.applied = false
	mode := cursorState.mode
	cursorState.Unlock()
	applyCursorMode(cursorApp, mode)
}

func SetCursorMode(mode int) {
	cursorState.Lock()
	cursorState.mode = mode
	app := cursorState.app
	cursorState.Unlock()
	applyCursorMode(app, mode)
}

func reapplyCursorMode() {
	cursorState.Lock()
	mode := cursorState.mode
	app := cursorState.app
	cursorState.Unlock()
	applyCursorMode(app, mode)
}

func applyCursorMode(app cursorApp, mode int) {
	if app == nil {
		return
	}
	shape := cursorShapeForMode(mode)
	cursorState.Lock()
	if cursorState.applied && cursorState.appliedApp == app && cursorState.appliedMode == mode {
		cursorState.Unlock()
		return
	}
	cursorState.Unlock()
	app.SetCursor(shape)
	cursorState.Lock()
	if cursorState.app == app && cursorState.mode == mode {
		cursorState.appliedApp = app
		cursorState.appliedMode = mode
		cursorState.applied = true
	}
	cursorState.Unlock()
}

func cursorShapeForMode(mode int) gpucontext.CursorShape {
	if mode == CursorModeHidden {
		return gpucontext.CursorNone
	}
	return gpucontext.CursorDefault
}

func cursorModeHidden() bool {
	cursorState.Lock()
	hidden := cursorState.mode == CursorModeHidden
	cursorState.Unlock()
	return hidden
}

type roCursorPlatformProvider struct {
	gpucontext.PlatformProvider
}

func (p roCursorPlatformProvider) SetCursor(cursor gpucontext.CursorShape) {
	if cursorModeHidden() {
		return
	}
	p.PlatformProvider.SetCursor(cursor)
}
