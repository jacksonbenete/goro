package game

import (
	"reflect"
	"testing"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
)

func TestHandleChatMessageRoutesToOpenChatRoomOnly(t *testing.T) {
	mode := &WorldMode{}
	ctx := client.Context{
		UIManager: &worldModeTestUIManager{},
	}
	mode.ui.chatRoom.Open(ctx, "Room", 20, true, []string{"Kivutar"})

	mode.handleChatMessage(ctx, network.ChatMessage{
		Text: "Kivutar : hello",
	}, time.Now())

	lines := chatRoomLineTexts(t, mode)
	if len(lines) != 1 {
		t.Fatalf("chat room lines = %d, want 1", len(lines))
	}
	if lines[0] != "Kivutar : hello" {
		t.Fatalf("chat room line = %q", lines[0])
	}
	if len(mode.speechBubbles) != 0 {
		t.Fatalf("speech bubbles = %d, want 0", len(mode.speechBubbles))
	}
}

func chatRoomLineTexts(t *testing.T, mode *WorldMode) []string {
	t.Helper()
	lines := reflect.ValueOf(&mode.ui.chatRoom).Elem().FieldByName("lines")
	out := make([]string, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		out = append(out, lines.Index(i).FieldByName("text").String())
	}
	return out
}
