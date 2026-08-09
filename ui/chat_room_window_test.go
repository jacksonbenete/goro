package ui

import (
	"fmt"
	"testing"

	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
)

func TestChatRoomMessagesScrollToMeasuredBottom(t *testing.T) {
	var window ChatRoomWindow
	ctx := client.Context{ScreenW: 800, ScreenH: 600}
	window.Open(ctx, "Room", 20, true, []string{"Kivutar"})
	for i := 0; i < 20; i++ {
		window.AddMessage(ctx, fmt.Sprintf("line %d", i))
	}

	root := window.Window.Widget()
	root.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(
		chatRoomWindowW,
		ROWindowTitleHeight+chatRoomWindowContentH+ROWindowFooterHeight,
	)))
	view := findChatRoomScrollView(root)
	if view == nil {
		t.Fatal("chat room scroll view was not found")
	}
	_, scrollY := view.ScrollOffset()
	want := view.ContentSize().Height - view.ViewportSize().Height
	if scrollY != want {
		t.Fatalf("scroll Y = %.1f, want measured bottom %.1f", scrollY, want)
	}
}

func findChatRoomScrollView(root widget.Widget) *scrollview.Widget {
	if view, ok := root.(*scrollview.Widget); ok {
		return view
	}
	parent, ok := root.(interface{ Children() []widget.Widget })
	if !ok {
		return nil
	}
	for _, child := range parent.Children() {
		if view := findChatRoomScrollView(child); view != nil {
			return view
		}
	}
	return nil
}
