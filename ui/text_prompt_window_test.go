package ui

import (
	"testing"

	"github.com/kivutar/goro/client"
)

func TestTextPromptWindowSubmitPublishesAction(t *testing.T) {
	var window TextPromptWindow
	ctx := client.Context{ScreenW: 800, ScreenH: 600}
	window.Open(ctx, "Talkie Box Message", "Message", "Message", 79)
	window.value = " hello "
	window.submit(ctx)

	action := window.PopAction()
	if !action.Submitted || action.Text != "hello" {
		t.Fatalf("action = %+v", action)
	}
	if window.IsOpen() {
		t.Fatal("window remained open after submit")
	}
}
