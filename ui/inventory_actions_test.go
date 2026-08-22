package ui

import (
	"strings"
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/session"
)

func TestUseInventoryItemRejectsDeadPlayer(t *testing.T) {
	err := UseInventoryItem(client.Context{
		Session: &session.Session{Dead: true},
	}, session.InventoryItem{
		Index:  7,
		ItemID: 501,
		Type:   db.ItemTypeHealing,
	})

	if err == nil || !strings.Contains(err.Error(), "dead") {
		t.Fatalf("error = %v, want dead-player rejection", err)
	}
}
