package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
)

func TestItemPickupNotificationTextUsesMsgStringAndItemName(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	msgTable := strings.Repeat("ignored#\n", 696) + "- %d obtained.#\n"
	if err := os.WriteFile(filepath.Join(dataDir, "msgstringtable.txt"), []byte(msgTable), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "idnum2itemdisplaynametable.txt"), []byte("938#Apple#\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manager, err := res.NewManager(root)
	if err != nil {
		t.Fatal(err)
	}

	got := itemPickupNotificationTextFor(manager, session.InventoryItem{ItemID: 938, Identified: true}, 2)
	if got != "Apple - 2 obtained." {
		t.Fatalf("notification text = %q", got)
	}
}

func TestItemPickupNotificationExpires(t *testing.T) {
	now := time.Unix(10, 0)
	notification := ItemPickupNotification{}
	notification.Show(Context{}, session.InventoryItem{ItemID: 938}, 1, now)

	if !notification.visible(now.Add(itemPickupNotificationLife)) {
		t.Fatal("notification expired before its life ended")
	}
	if notification.visible(now.Add(itemPickupNotificationLife + time.Millisecond)) {
		t.Fatal("notification stayed visible after its life ended")
	}
}

func TestItemPickupNotificationBoundsCenterAndClamp(t *testing.T) {
	text := "Apple - 2 obtained."
	x, y, w, h := itemPickupNotificationBounds(800, text)
	if y != itemPickupNotificationTopY {
		t.Fatalf("y = %d, want %d", y, itemPickupNotificationTopY)
	}
	centerTwice := x*2 + w
	if centerTwice < 799 || centerTwice > 801 {
		t.Fatalf("bounds x=%d w=%d, want centered on 400", x, w)
	}
	if h != itemPickupNotificationH {
		t.Fatalf("height = %d", h)
	}

	fit := fitItemPickupNotificationText(strings.Repeat("VeryLongName", 12), 160)
	if fit == "" || !strings.HasSuffix(fit, "...") {
		t.Fatalf("fit text = %q, want ellipsis", fit)
	}
	_, _, clampedW, _ := itemPickupNotificationBounds(160, fit)
	if clampedW > 160-itemPickupNotificationMargin*2 {
		t.Fatalf("width = %d, want within screen margin", clampedW)
	}
}
