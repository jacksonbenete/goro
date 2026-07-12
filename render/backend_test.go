package render

import (
	"os"
	"testing"

	"github.com/kivutar/goro/config"
)

func TestConfigureGogpuVSyncDisablesWaylandFrameGateWhenVSyncOff(t *testing.T) {
	t.Setenv("GOGPU_WAYLAND_FRAME_CALLBACK", "")
	if err := os.Unsetenv("GOGPU_WAYLAND_FRAME_CALLBACK"); err != nil {
		t.Fatal(err)
	}

	configureGogpuVSync(config.RenderConfig{VSync: false})

	if got := os.Getenv("GOGPU_WAYLAND_FRAME_CALLBACK"); got != "0" {
		t.Fatalf("GOGPU_WAYLAND_FRAME_CALLBACK = %q, want 0", got)
	}
}

func TestConfigureGogpuVSyncLeavesWaylandFrameGateWhenVSyncOn(t *testing.T) {
	t.Setenv("GOGPU_WAYLAND_FRAME_CALLBACK", "")
	if err := os.Unsetenv("GOGPU_WAYLAND_FRAME_CALLBACK"); err != nil {
		t.Fatal(err)
	}

	configureGogpuVSync(config.RenderConfig{VSync: true})

	if got, ok := os.LookupEnv("GOGPU_WAYLAND_FRAME_CALLBACK"); ok {
		t.Fatalf("GOGPU_WAYLAND_FRAME_CALLBACK = %q, want unset", got)
	}
}

func TestConfigureGogpuVSyncDisablesWaylandFrameGateForBenchmarks(t *testing.T) {
	t.Setenv("GOGPU_WAYLAND_FRAME_CALLBACK", "")
	if err := os.Unsetenv("GOGPU_WAYLAND_FRAME_CALLBACK"); err != nil {
		t.Fatal(err)
	}

	configureGogpuVSync(config.RenderConfig{VSync: true, BenchSeconds: 10})

	if got := os.Getenv("GOGPU_WAYLAND_FRAME_CALLBACK"); got != "0" {
		t.Fatalf("GOGPU_WAYLAND_FRAME_CALLBACK = %q, want 0", got)
	}
}

func TestConfigureGogpuVSyncPreservesExplicitWaylandFrameGate(t *testing.T) {
	t.Setenv("GOGPU_WAYLAND_FRAME_CALLBACK", "1")

	configureGogpuVSync(config.RenderConfig{VSync: false})

	if got := os.Getenv("GOGPU_WAYLAND_FRAME_CALLBACK"); got != "1" {
		t.Fatalf("GOGPU_WAYLAND_FRAME_CALLBACK = %q, want explicit value", got)
	}
}
