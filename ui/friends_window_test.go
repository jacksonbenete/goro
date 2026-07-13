package ui

import (
	"math"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
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

func TestEmptyFriendsListExpandsAndCentersText(t *testing.T) {
	assertEmptyListExpandsAndCenters(t, friendsList(nil), "No friends")
}

func TestEmptyPartyListExpandsAndCentersText(t *testing.T) {
	assertEmptyListExpandsAndCenters(t, partyList(session.Party{}), "No party")
}

func TestFriendsWindowListStartsBelowTabs(t *testing.T) {
	var window FriendsWindow
	root := window.widgetTree(Context{Session: &session.Session{}})

	root.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(friendsWindowWidth, friendsWindowHeight)))
	rootChildren := root.Children()
	if len(rootChildren) < 2 {
		t.Fatalf("window children = %d, want title and content", len(rootChildren))
	}
	contentChildren := rootChildren[1].Children()
	if len(contentChildren) != 1 {
		t.Fatalf("content wrapper children = %d, want content box", len(contentChildren))
	}
	stackChildren := contentChildren[0].Children()
	if len(stackChildren) != 2 {
		t.Fatalf("content stack children = %d, want tabs and list", len(stackChildren))
	}
	tabsBounds := stackChildren[0].(interface{ Bounds() geometry.Rect }).Bounds()
	listBounds := stackChildren[1].(interface{ Bounds() geometry.Rect }).Bounds()
	if listBounds.Min.Y != tabsBounds.Height() {
		t.Fatalf("list top = %.1f, want tabs height %.1f", listBounds.Min.Y, tabsBounds.Height())
	}
}

func assertEmptyListExpandsAndCenters(t *testing.T, list widget.Widget, label string) {
	t.Helper()

	list.Layout(widget.NewContext(), geometry.Tight(geometry.Sz(220, 120)))
	listBounds := list.(interface{ Bounds() geometry.Rect }).Bounds()
	if listBounds.Width() != 220 || listBounds.Height() != 120 {
		t.Fatalf("%s list bounds = %.1fx%.1f, want 220.0x120.0", label, listBounds.Width(), listBounds.Height())
	}

	children := list.Children()
	if len(children) != 1 {
		t.Fatalf("%s list children = %d, want empty state only", label, len(children))
	}
	emptyState, ok := children[0].(*primitives.ExpandedWidget)
	if !ok {
		t.Fatalf("%s empty state = %T, want ExpandedWidget", label, children[0])
	}
	emptyBounds := emptyState.Bounds()
	if emptyBounds.Width() != 220 || emptyBounds.Height() != 120 {
		t.Fatalf("%s empty bounds = %.1fx%.1f, want 220.0x120.0", label, emptyBounds.Width(), emptyBounds.Height())
	}

	inner := emptyState.Children()[0]
	innerChildren := inner.Children()
	if len(innerChildren) != 3 {
		t.Fatalf("%s empty children = %d, want top spacer, text, bottom spacer", label, len(innerChildren))
	}
	textBounds := innerChildren[1].(interface{ Bounds() geometry.Rect }).Bounds()
	if textBounds.Width() != 220 {
		t.Fatalf("%s text width = %.1f, want 220.0", label, textBounds.Width())
	}
	textCenterY := textBounds.Min.Y + textBounds.Height()/2
	if math.Abs(float64(textCenterY-60)) > 0.5 {
		t.Fatalf("%s text center y = %.1f, want 60.0", label, textCenterY)
	}
}
