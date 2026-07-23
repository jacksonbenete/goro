package render

import (
	"os"
	"testing"

	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
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

func TestRequestUIRedrawMarksCleanTreeDirty(t *testing.T) {
	app := uiapp.New(uiapp.WithRenderMode(uiapp.RenderModeFrameworkManaged))
	root := primitives.Box().Width(24).Height(12)
	app.SetRoot(root)
	app.Frame()
	if !app.Window().DrawTo(&uitest.MockCanvas{}) {
		t.Fatal("initial UI draw did not render")
	}
	if app.Window().NeedsRedraw() {
		t.Fatal("window stayed dirty after initial draw")
	}
	if widget.NeedsRedrawInTree(root) {
		t.Fatal("root stayed dirty after initial draw")
	}

	(&runner{ui: app}).requestUIRedraw()

	if !app.Window().NeedsRedraw() {
		t.Fatal("UI redraw request did not dirty the window")
	}
	if !widget.NeedsRedrawInTree(root) {
		t.Fatal("UI redraw request did not dirty the widget tree")
	}
	if !app.Window().DrawTo(&uitest.MockCanvas{}) {
		t.Fatal("requested UI redraw did not render")
	}
	if app.Window().NeedsRedraw() {
		t.Fatal("window stayed dirty after requested redraw")
	}
	if widget.NeedsRedrawInTree(root) {
		t.Fatal("root stayed dirty after requested redraw")
	}
}
