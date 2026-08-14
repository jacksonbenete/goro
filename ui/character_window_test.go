package ui

import (
	"math"
	"strings"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/session"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestCharacterEXPPanelWrapsBothRowsInRoundedGreyContainer(t *testing.T) {
	const width float32 = characterWindowWidth - 24
	panel := characterEXPPanel(session.Progress{
		BaseLevel:   42,
		JobLevel:    27,
		BaseExp:     25,
		NextBaseExp: 100,
		JobExp:      50,
		NextJobExp:  100,
	}, width)
	panel.Layout(widget.NewContext(), geometry.Constraints{MaxWidth: width, MaxHeight: 100})

	box, ok := panel.(*primitives.BoxWidget)
	if !ok {
		t.Fatalf("EXP panel = %T, want box", panel)
	}
	style := box.Style()
	if style.Radius != characterEXPPanelRadius {
		t.Fatalf("EXP panel radius = %.1f, want %.1f", style.Radius, characterEXPPanelRadius)
	}
	uitest.AssertColorEqual(t, style.Background, rotheme.Default.Colors.WindowFooter)
	if style.Border.Width != 0 {
		t.Fatalf("EXP panel border width = %.1f, want none", style.Border.Width)
	}

	rows := panel.Children()
	if len(rows) != 2 {
		t.Fatalf("EXP panel rows = %d, want Base and Job EXP", len(rows))
	}
	first := rows[0].(interface{ Bounds() geometry.Rect }).Bounds()
	second := rows[1].(interface{ Bounds() geometry.Rect }).Bounds()
	wantRowWidth := width - 2*characterEXPPanelPaddingX
	if first.Min.X != characterEXPPanelPaddingX || first.Min.Y != characterEXPPanelPaddingY || first.Width() != wantRowWidth {
		t.Fatalf("Base EXP row bounds = %v, want inset width %.1f", first, wantRowWidth)
	}
	if second.Min.X != characterEXPPanelPaddingX || second.Min.Y != first.Max.Y+characterEXPPanelGap || second.Width() != wantRowWidth {
		t.Fatalf("Job EXP row bounds = %v, want below Base EXP with %.1f gap", second, characterEXPPanelGap)
	}
	assertCharacterEXPRow := func(row widget.Widget, wantLabel string) {
		t.Helper()
		children := row.Children()
		if len(children) != 2 {
			t.Fatalf("%s row children = %d, want label and bar", wantLabel, len(children))
		}
		labelChildren := children[0].Children()
		if len(labelChildren) != 1 {
			t.Fatalf("%s label children = %d, want text", wantLabel, len(labelChildren))
		}
		text, ok := labelChildren[0].(interface{ Content() string })
		if !ok {
			t.Fatalf("EXP row label = %T, want text", labelChildren[0])
		}
		if text.Content() != wantLabel {
			t.Fatalf("EXP row label = %q, want %q", text.Content(), wantLabel)
		}
		labelBounds := children[0].(interface{ Bounds() geometry.Rect }).Bounds()
		barSlotBounds := children[1].(interface{ Bounds() geometry.Rect }).Bounds()
		if barSlotBounds.Min.X != labelBounds.Max.X+characterEXPLabelBarGap {
			t.Fatalf("%s bar x = %.1f, want %.1f after label", wantLabel, barSlotBounds.Min.X, labelBounds.Max.X+characterEXPLabelBarGap)
		}
		barChildren := children[1].Children()
		if len(barChildren) != 1 {
			t.Fatalf("%s bar slot children = %d, want bar", wantLabel, len(barChildren))
		}
		barBounds := barChildren[0].(interface{ Bounds() geometry.Rect }).Bounds()
		labelCenterY := labelBounds.Center().Y
		barCenterY := barSlotBounds.Min.Y + barBounds.Center().Y
		if math.Abs(float64(labelCenterY-barCenterY)) > 0.001 {
			t.Fatalf("%s vertical centers differ: label=%.3f bar=%.3f", wantLabel, labelCenterY, barCenterY)
		}
	}
	assertCharacterEXPRow(rows[0], "Base Lv. 42")
	assertCharacterEXPRow(rows[1], "Job Lv. 27")
	panelBounds := panel.(interface{ Bounds() geometry.Rect }).Bounds()
	if second.Max.Y+characterEXPPanelPaddingY != panelBounds.Max.Y {
		t.Fatalf("EXP panel bottom = %.1f, want %.1f after bottom inset", panelBounds.Max.Y, second.Max.Y+characterEXPPanelPaddingY)
	}
}

func TestCharacterWindowDoesNotDuplicateLevelOrEXPLabels(t *testing.T) {
	root := (&CharacterWindow{}).widgetTree(Context{Session: &session.Session{
		Progress: session.Progress{BaseLevel: 42, JobLevel: 27},
	}})
	var labels []string
	var walk func(widget.Widget)
	walk = func(current widget.Widget) {
		if text, ok := current.(interface{ Content() string }); ok {
			labels = append(labels, text.Content())
		}
		for _, child := range current.Children() {
			walk(child)
		}
	}
	walk(root)
	joined := strings.Join(labels, "\n")
	for _, label := range []string{"Base Lv. 42", "Job Lv. 27"} {
		if strings.Count(joined, label) != 1 {
			t.Fatalf("character window label %q appears %d times, want once", label, strings.Count(joined, label))
		}
	}
	if strings.Contains(joined, "Base EXP") || strings.Contains(joined, "Job EXP") {
		t.Fatalf("character window still contains EXP percentage labels: %q", joined)
	}
}
