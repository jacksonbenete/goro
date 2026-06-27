package gamemode

import (
	"image/color"
	"testing"

	"github.com/kivutar/goro/input"
)

func TestNPCDialogTextRunsParseColorCodes(t *testing.T) {
	base := color.RGBA{R: 246, G: 242, B: 232, A: 255}
	runs := npcDialogTextRuns("hello ^FF3300red^000000 base", base)

	if len(runs) != 3 {
		t.Fatalf("run count = %d, want 3: %#v", len(runs), runs)
	}
	if runs[0].text != "hello " || runs[0].color != base {
		t.Fatalf("first run = %#v", runs[0])
	}
	if runs[1].text != "red" || runs[1].color != (color.RGBA{R: 255, G: 51, B: 0, A: 255}) {
		t.Fatalf("colored run = %#v", runs[1])
	}
	if runs[2].text != " base" || runs[2].color != base {
		t.Fatalf("reset run = %#v", runs[2])
	}
}

func TestNPCDialogWrapIgnoresColorCodeWidth(t *testing.T) {
	base := color.RGBA{R: 246, G: 242, B: 232, A: 255}
	lines := wrapNPCDialogLines([]string{"one ^00AAFFtwo^000000 three"}, 10)

	if len(lines) != 2 {
		t.Fatalf("line count = %d, want 2: %#v", len(lines), lines)
	}
	if got := npcDialogPlainText(lines[0]); got != "one two" {
		t.Fatalf("first line = %q", got)
	}
	if got := npcDialogPlainText(lines[1]); got != "three" {
		t.Fatalf("second line = %q", got)
	}
	if lines[0][1].text != "two" || lines[0][1].color == base {
		t.Fatalf("wrapped colored run not preserved: %#v", lines[0])
	}
}

func TestNPCDialogMenuScrollChangesVisibleChoiceRange(t *testing.T) {
	dialog := npcDialogState{
		action:  npcDialogActionMenu,
		options: []string{"one", "two", "three", "four", "five", "six"},
	}
	height := npcMenuTitleH + npcMenuTopPad + 4*npcMenuRowH + npcMenuBottomH

	start, end := dialog.visibleMenuRange(height)
	if start != 0 || end != 4 {
		t.Fatalf("initial range = %d,%d, want 0,4", start, end)
	}

	dialog.scrollMenuBy(-2, height)
	start, end = dialog.visibleMenuRange(height)
	if start != 2 || end != 6 {
		t.Fatalf("scrolled range = %d,%d, want 2,6", start, end)
	}

	dialog.scrollMenuBy(-10, height)
	start, end = dialog.visibleMenuRange(height)
	if start != 2 || end != 6 {
		t.Fatalf("clamped range = %d,%d, want 2,6", start, end)
	}

	dialog.scrollMenuBy(1, height)
	start, end = dialog.visibleMenuRange(height)
	if start != 1 || end != 5 {
		t.Fatalf("scroll up range = %d,%d, want 1,5", start, end)
	}
}

func TestNPCDialogMenuScrollConsumesWheelInsideMenu(t *testing.T) {
	state := input.NewState()
	state.SetMousePosition(20, 20)
	state.AddWheel(0, -1)
	dialog := npcDialogState{
		action:  npcDialogActionMenu,
		options: []string{"one", "two", "three", "four", "five", "six"},
	}
	height := npcMenuTitleH + npcMenuTopPad + 4*npcMenuRowH + npcMenuBottomH

	if !dialog.updateMenuScrollAt(Context{Input: state}, 10, 10, npcMenuWidth, height) {
		t.Fatal("wheel inside menu was not consumed")
	}
	if dialog.menuScroll != 1 {
		t.Fatalf("menu scroll = %d, want 1", dialog.menuScroll)
	}
	if state.WheelY != 0 {
		t.Fatalf("wheel y = %.1f, want consumed", state.WheelY)
	}
}

func npcDialogPlainText(runs []npcDialogTextRun) string {
	text := ""
	for _, run := range runs {
		text += run.text
	}
	return text
}
