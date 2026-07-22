package ui

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/gogpu/ui/widget"
)

func TestWrapChatTextKeepsWordsWithinWidth(t *testing.T) {
	lines := wrapChatText("alpha beta gamma delta", chatTextWidth("alpha beta"))

	want := []string{"alpha beta", "gamma", "delta"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("wrapped lines = %#v, want %#v", lines, want)
	}
}

func TestWrapChatTextSplitsLongWords(t *testing.T) {
	maxWidth := chatTextWidth("abc")
	lines := wrapChatText("abcdefghijkl", maxWidth)

	if len(lines) <= 1 {
		t.Fatalf("wrapped lines = %#v, want multiple lines", lines)
	}
	for _, line := range lines {
		if got := chatTextWidth(line); got > maxWidth {
			t.Fatalf("line %q width = %d, want <= %d", line, got, maxWidth)
		}
	}
}

func TestWrapConsoleMessagesPreservesColor(t *testing.T) {
	lineColor := color.RGBA{R: 12, G: 34, B: 56, A: 255}
	lines := wrapConsoleMessages([]ConsoleMessage{
		{Text: "alpha beta gamma", Color: lineColor},
	}, chatTextWidth("alpha beta"))

	if len(lines) != 2 {
		t.Fatalf("wrapped lines = %#v, want 2 lines", lines)
	}
	for _, line := range lines {
		if line.color != lineColor {
			t.Fatalf("line color = %#v, want %#v", line.color, lineColor)
		}
	}
}

func TestWrapChatRoomLinesPreservesColor(t *testing.T) {
	lineColor := widget.RGBA8(12, 34, 56, 255)
	lines := wrapChatRoomLines([]chatRoomLine{
		{text: "alpha beta gamma", color: lineColor},
	}, chatTextWidth("alpha beta"))

	if len(lines) != 2 {
		t.Fatalf("wrapped lines = %#v, want 2 lines", lines)
	}
	for _, line := range lines {
		if line.color != lineColor {
			t.Fatalf("line color = %#v, want %#v", line.color, lineColor)
		}
	}
}
