package ui

import (
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/db"
)

func TestEmotePageCountMatchesRobrowserWindow(t *testing.T) {
	if got := emotePageCount(); got != 3 {
		t.Fatalf("emotePageCount() = %d, want 3", got)
	}
}

func TestEmoteCellSelectsOnClickAndPlaysOnQuickSecondClick(t *testing.T) {
	selects := 0
	plays := 0
	cell := newEmoteCellWidget(
		db.Emotion{ID: 14, Frame: 4, Command: "lv2"},
		nil,
		func() { selects++ },
		func() { plays++ },
	)
	ctx := widget.NewContext()
	pos := geometry.Pt(8, 8)
	start := time.Unix(1000, 0)

	cell.Event(ctx, event.NewMouseEventWithTime(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, pos, pos, event.ModNone, start))
	if selects != 1 || plays != 0 {
		t.Fatalf("after first click selects=%d plays=%d, want 1/0", selects, plays)
	}
	cell.Event(ctx, event.NewMouseEventWithTime(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, pos, pos, event.ModNone, start.Add(100*time.Millisecond)))
	if selects != 1 || plays != 1 {
		t.Fatalf("after second click selects=%d plays=%d, want 1/1", selects, plays)
	}
}

func TestEmoteCellPlaysOnDoubleClickEvent(t *testing.T) {
	plays := 0
	cell := newEmoteCellWidget(
		db.Emotion{ID: 4, Frame: 5, Command: "swt"},
		nil,
		nil,
		func() { plays++ },
	)
	pos := geometry.Pt(8, 8)
	cell.Event(widget.NewContext(), event.NewMouseEvent(event.MouseDoubleClick, event.ButtonLeft, event.ButtonStateLeft, pos, pos, event.ModNone))
	if plays != 1 {
		t.Fatalf("double click plays = %d, want 1", plays)
	}
}

func TestEmoteWindowSelectEmotionWritesConsoleCommand(t *testing.T) {
	var window EmoteWindow
	var console ChatConsole

	window.selectEmotion(Context{}, &console, db.Emotion{ID: 4, Frame: 5, Command: "swt"})
	if console.input != "/swt" || !console.active {
		t.Fatalf("console input=%q active=%t, want /swt active", console.input, console.active)
	}
}
