package render

import (
	"math"
	"os"
	"testing"

	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/geometry"
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

func TestConfigureGogpuVSyncLeavesWaylandFrameGateForVSyncBenchmarks(t *testing.T) {
	t.Setenv("GOGPU_WAYLAND_FRAME_CALLBACK", "")
	if err := os.Unsetenv("GOGPU_WAYLAND_FRAME_CALLBACK"); err != nil {
		t.Fatal(err)
	}

	configureGogpuVSync(config.RenderConfig{VSync: true, BenchSeconds: 10})

	if got, ok := os.LookupEnv("GOGPU_WAYLAND_FRAME_CALLBACK"); ok {
		t.Fatalf("GOGPU_WAYLAND_FRAME_CALLBACK = %q, want unset", got)
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

func TestSetUIImageReleasesPreviousGPUTexture(t *testing.T) {
	oldImage := NewImage(1, 1)
	nextImage := NewImage(1, 1)
	gpu := &gpuRenderer{
		textures: map[*Image]*gpuImageTexture{
			oldImage: {},
		},
	}
	r := &runner{gpu: gpu, uiImage: oldImage}

	r.setUIImage(nextImage)

	if r.uiImage != nextImage {
		t.Fatal("runner did not publish the replacement UI image")
	}
	if _, ok := gpu.textures[oldImage]; ok {
		t.Fatal("old UI image texture was not released from the GPU cache")
	}
}

func TestSetUIImageKeepsCurrentGPUTextureWhenImageUnchanged(t *testing.T) {
	image := NewImage(1, 1)
	gpu := &gpuRenderer{
		textures: map[*Image]*gpuImageTexture{
			image: {},
		},
	}
	r := &runner{gpu: gpu, uiImage: image}

	r.setUIImage(image)

	if _, ok := gpu.textures[image]; !ok {
		t.Fatal("current UI image texture was released even though the image did not change")
	}
}

func TestSetUIDragLayerReleasesPreviousGPUTexture(t *testing.T) {
	oldImage := NewImage(1, 1)
	nextImage := NewImage(1, 1)
	gpu := &gpuRenderer{
		textures: map[*Image]*gpuImageTexture{
			oldImage: {},
		},
	}
	r := &runner{
		gpu: gpu,
		uiDrag: uiDragLayer{
			image:  oldImage,
			active: true,
		},
	}

	r.setUIDragLayer(uiDragLayer{image: nextImage, active: true})

	if r.uiDrag.image != nextImage {
		t.Fatal("runner did not publish the replacement drag image")
	}
	if _, ok := gpu.textures[oldImage]; ok {
		t.Fatal("old drag image texture was not released from the GPU cache")
	}
}

func TestShouldRecordAsyncUIBackpressuresWhenPendingListExists(t *testing.T) {
	r := &runner{
		uiAsyncBusy:    true,
		uiPendingLists: []uiDrawList{{}},
	}

	if r.shouldRecordAsyncUI(true) {
		t.Fatal("async UI recording was not backpressured while a raster and pending list were already queued")
	}
	if !r.lastUIWork {
		t.Fatal("backpressured UI work was not reported")
	}
}

func TestShouldRecordAsyncUIAllowsFirstPendingListWhileBusy(t *testing.T) {
	r := &runner{uiAsyncBusy: true}

	if !r.shouldRecordAsyncUI(true) {
		t.Fatal("async UI did not allow recording the first pending draw list")
	}
}

func TestDrawActorLabelOverlayUsesSharedSnappedOrigin(t *testing.T) {
	screen := NewFrame(320, 240)
	screen.SetScreenScale(1.25, 1.5)
	cached := cachedOverlayImage{
		image:  NewImage(120, 68),
		width:  60,
		height: 34,
	}
	label := UIActorLabelCommand{
		Emblem:  NewImage(24, 24),
		CenterX: 100.82,
		Y:       50.42,
	}

	drawActorLabelOverlay(screen, cached, label)

	if len(screen.commands) != 2 {
		t.Fatalf("draw commands = %d, want text and emblem", len(screen.commands))
	}
	textOrigin := screen.commands[0].Vertices[0]
	emblemOrigin := screen.commands[1].Vertices[0]
	blockLeft, blockTop := snapScreenPoint(screen, label.CenterX-float64(cached.width+actorLabelEmblemSize+actorLabelEmblemGap)/2, label.Y)
	assertFloatClose(t, "text x", float64(textOrigin.DstX), blockLeft+actorLabelEmblemSize+actorLabelEmblemGap)
	assertFloatClose(t, "text y", float64(textOrigin.DstY), blockTop)
	assertFloatClose(t, "emblem x", float64(emblemOrigin.DstX), blockLeft+2)
	assertFloatClose(t, "emblem y", float64(emblemOrigin.DstY), blockTop+5)
}

func TestUIDragLayerDrawRectPreservesPhysicalCropSize(t *testing.T) {
	r := &runner{
		uiImage: NewImage(1000, 750),
		width:   800,
		height:  600,
	}
	frame := geometry.NewRect(11, 21, 100, 80)
	capture := r.captureUIImageRect(frame)
	if capture.image == nil {
		t.Fatal("capture returned nil image")
	}
	if got := capture.image.Bounds().Dx(); got != 126 {
		t.Fatalf("capture width = %d, want 126 physical px", got)
	}
	if got := capture.image.Bounds().Dy(); got != 101 {
		t.Fatalf("capture height = %d, want 101 physical px", got)
	}

	if !r.beginUIDragLayer("window", frame) {
		t.Fatal("drag layer did not start")
	}
	r.moveUIDragLayer("window", geometry.NewRect(40, 55, 100, 80))
	drawRect := r.uiDrag.drawRect()
	assertFloatClose(t, "draw x", float64(drawRect.Min.X), 39.4)
	assertFloatClose(t, "draw y", float64(drawRect.Min.Y), 54.8)
	assertFloatClose(t, "draw width", float64(drawRect.Width()), 100.8)
	assertFloatClose(t, "draw height", float64(drawRect.Height()), 80.8)
}

func assertFloatClose(t *testing.T, label string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.001 {
		t.Fatalf("%s = %.4f, want %.4f", label, got, want)
	}
}
