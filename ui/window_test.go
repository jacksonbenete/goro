package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
)

func TestFixedHeightFooterStretchesContent(t *testing.T) {
	button := primitives.Box().Width(30).Height(10)
	row := primitives.HBox(
		primitives.Expanded(primitives.Box()),
		button,
	)
	window := Win(
		TitleBar(false),
		Size(200, 80),
		FooterHeight(24),
		FooterPadding(10),
		Footer(row),
	)

	window.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(200, 80)))

	if got := row.Bounds().Width(); got != 180 {
		t.Fatalf("footer row width = %.1f, want 180.0", got)
	}
	if got := button.Bounds().Min.X; got != 150 {
		t.Fatalf("right footer button x = %.1f, want 150.0", got)
	}
}

func TestFooterHeightCreatesEmptyFooterBand(t *testing.T) {
	window := Win(
		TitleBar(false),
		Size(200, 80),
		FooterHeight(24),
		FooterPadding(10),
	)

	window.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(200, 80)))

	children := window.Children()
	if len(children) != 1 {
		t.Fatalf("window children = %d, want 1 footer", len(children))
	}
	footer := children[0]
	if got := footer.(interface{ Bounds() geometry.Rect }).Bounds().Width(); got != 200 {
		t.Fatalf("footer width = %.1f, want 200.0", got)
	}
	footerChildren := footer.Children()
	if len(footerChildren) != 2 {
		t.Fatalf("footer children = %d, want divider and body", len(footerChildren))
	}
	body := footerChildren[1]
	if got := body.(interface{ Bounds() geometry.Rect }).Bounds().Width(); got != 200 {
		t.Fatalf("footer body width = %.1f, want 200.0", got)
	}
}

func TestWindowFullRedrawPublishIsIdempotent(t *testing.T) {
	manager := &escapeMenuTestUIManager{}
	ctx := client.Context{UIManager: manager, ScreenW: 800, ScreenH: 600}
	window := NewWindow(100, 80)
	window.SetFullRedraw(true)
	window.OpenAt(10, 20, primitives.Box())
	window.Publish(ctx)
	if len(manager.overlays) != 1 {
		t.Fatalf("published overlays = %d, want 1", len(manager.overlays))
	}

	root := manager.overlays[0]
	widget.ClearRedrawInTree(root)
	redraw, ok := root.(interface{ NeedsRedraw() bool })
	if !ok {
		t.Fatal("published root does not expose redraw state")
	}
	if redraw.NeedsRedraw() {
		t.Fatal("published root stayed dirty after clear")
	}

	window.Publish(ctx)
	if redraw.NeedsRedraw() {
		t.Fatal("unchanged full redraw publish dirtied the published root")
	}
	if manager.overlays[0] != root {
		t.Fatal("unchanged full redraw publish replaced the published root")
	}

	window.SetContent(primitives.Box())
	window.Publish(ctx)
	if manager.overlays[0] == root {
		t.Fatal("content change did not replace the published root")
	}
	replaced, ok := manager.overlays[0].(interface{ NeedsRedraw() bool })
	if !ok {
		t.Fatal("replaced root does not expose redraw state")
	}
	if !replaced.NeedsRedraw() {
		t.Fatal("full redraw content replacement did not dirty the new root")
	}
}
