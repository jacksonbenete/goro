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

var cursorState = struct {
	sync.Mutex
	app  *gogpu.App
	mode int
}{
	mode: CursorModeNormal,
}

func setCursorApp(app *gogpu.App) {
	cursorState.Lock()
	cursorState.app = app
	mode := cursorState.mode
	cursorState.Unlock()
	applyCursorMode(app, mode)
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

func applyCursorMode(app *gogpu.App, mode int) {
	if app == nil {
		return
	}
	if mode == CursorModeHidden {
		app.SetCursor(gpucontext.CursorNone)
		return
	}
	app.SetCursor(gpucontext.CursorDefault)
}
