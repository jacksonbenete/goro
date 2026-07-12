package ui

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
)

func TestPartyFooterAlignsButtonsRight(t *testing.T) {
	var window FriendsWindow
	footer := window.partyFooter(session.Party{Name: "Party"})

	footer.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(180, 24)))

	children := footer.Children()
	if len(children) != 3 {
		t.Fatalf("party footer children = %d, want spacer and two buttons", len(children))
	}
	leave := children[2].(interface{ Bounds() geometry.Rect }).Bounds()
	if leave.Max.X != 180 {
		t.Fatalf("leave button right edge = %.1f, want 180.0", leave.Max.X)
	}
}
