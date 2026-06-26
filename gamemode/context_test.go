package gamemode

import (
	"testing"

	"github.com/kivutar/goro/core"
)

func TestContextScreenSizeUsesLayoutSizeWhenAvailable(t *testing.T) {
	ctx := Context{
		Config:  core.Config{Window: core.WindowConfig{Width: 1024, Height: 768}},
		ScreenW: 1280,
		ScreenH: 720,
	}

	width, height := ctx.ScreenSize()
	if width != 1280 || height != 720 {
		t.Fatalf("screen size = %dx%d, want 1280x720", width, height)
	}
}

func TestContextScreenSizeFallsBackToConfig(t *testing.T) {
	ctx := Context{
		Config: core.Config{Window: core.WindowConfig{Width: 1024, Height: 768}},
	}

	width, height := ctx.ScreenSize()
	if width != 1024 || height != 768 {
		t.Fatalf("screen size = %dx%d, want 1024x768", width, height)
	}
}
