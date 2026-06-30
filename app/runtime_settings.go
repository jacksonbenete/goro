package app

import "sync/atomic"

type runtimeSettings struct {
	fullscreen atomic.Bool
	vsync      atomic.Bool
	fps        atomic.Bool
}

func newRuntimeSettings(fullscreen, vsync, fps bool) *runtimeSettings {
	settings := &runtimeSettings{}
	settings.fullscreen.Store(fullscreen)
	settings.vsync.Store(vsync)
	settings.fps.Store(fps)
	return settings
}

func (s *runtimeSettings) Fullscreen() bool {
	if s == nil {
		return false
	}
	return s.fullscreen.Load()
}

func (s *runtimeSettings) SetFullscreen(value bool) {
	if s != nil {
		s.fullscreen.Store(value)
	}
}

func (s *runtimeSettings) VSync() bool {
	if s == nil {
		return false
	}
	return s.vsync.Load()
}

func (s *runtimeSettings) SetVSync(value bool) {
	if s != nil {
		s.vsync.Store(value)
	}
}

func (s *runtimeSettings) FPS() bool {
	if s == nil {
		return false
	}
	return s.fps.Load()
}

func (s *runtimeSettings) SetFPS(value bool) {
	if s != nil {
		s.fps.Store(value)
	}
}
