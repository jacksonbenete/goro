package ui

import (
	"testing"

	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

type countingOverlay struct {
	widget.WidgetBase
	events int
	draws  int
}

func newCountingOverlay() *countingOverlay {
	w := &countingOverlay{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *countingOverlay) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.BiggestFinite(1, 1)
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

func (w *countingOverlay) Draw(widget.Context, widget.Canvas) { w.draws++ }

func (w *countingOverlay) Event(widget.Context, event.Event) bool {
	w.events++
	return false
}

func (w *countingOverlay) Children() []widget.Widget { return nil }

func TestManagerPointerBlockedUsesOverlayBounds(t *testing.T) {
	app := uiapp.New()
	manager := NewManager()
	manager.SetUIApp(basicMenuTestApp{app: app})
	manager.AddOverlay(positionedWidget(newInertOverlay(), 100, 120, 80, 40))
	app.Frame()
	app.Window().DrawTo(&uitest.MockCanvas{})

	if !manager.PointerBlocked(120, 130) {
		t.Fatal("pointer over overlay was not blocked")
	}
	if manager.PointerBlocked(90, 130) {
		t.Fatal("pointer outside overlay was blocked")
	}
}

func TestTopOverlayBlocksLowerOverlayEvents(t *testing.T) {
	app := uiapp.New()
	manager := NewManager()
	manager.SetUIApp(basicMenuTestApp{app: app})
	lower := newCountingOverlay()
	top := newCountingOverlay()
	manager.AddOverlay(positionedWidget(lower, 100, 100, 80, 80))
	manager.AddOverlay(positionedWidget(top, 120, 120, 80, 80))
	app.Frame()
	app.Window().DrawTo(&uitest.MockCanvas{})

	point := geometry.Pt(130, 130)
	app.Window().HandleEvent(event.NewMouseEvent(event.MousePress, event.ButtonLeft, 0, point, point, event.ModNone))
	if top.events != 1 {
		t.Fatalf("top overlay events = %d, want 1", top.events)
	}
	if lower.events != 0 {
		t.Fatalf("lower overlay events = %d, want 0", lower.events)
	}
}

func TestOverlayRootDrawSkipsChildrenOutsideClip(t *testing.T) {
	left := newCountingOverlay()
	right := newCountingOverlay()
	root := newOverlayRoot([]widget.Widget{
		positionedWidget(left, 10, 20, 40, 30),
		positionedWidget(right, 200, 20, 40, 30),
	})
	root.Layout(nil, geometry.Tight(geometry.Sz(300, 100)))

	canvas := clippedOverlayCanvas{
		MockCanvas: uitest.MockCanvas{},
		clip:       geometry.NewRect(190, 0, 80, 80),
	}
	root.Draw(nil, &canvas)

	if left.draws != 0 {
		t.Fatalf("left overlay draws = %d, want 0", left.draws)
	}
	if right.draws != 1 {
		t.Fatalf("right overlay draws = %d, want 1", right.draws)
	}
}

type clippedOverlayCanvas struct {
	uitest.MockCanvas
	clip geometry.Rect
}

func (c *clippedOverlayCanvas) ClipBounds() geometry.Rect {
	return c.clip
}
