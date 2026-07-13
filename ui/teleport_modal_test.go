package ui

import (
	"testing"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/session"
)

func TestTeleportModalSelectsFirstEnabledDestination(t *testing.T) {
	modal := TeleportModal{}

	modal.OpenWarpPointList(network.WarpPointList{
		SkillID:  teleportSkillID,
		MapNames: []string{teleportRandomMap},
	}, session.Skill{ID: teleportSkillID, Level: 2})

	if modal.row != 0 {
		t.Fatalf("selected row = %d, want Random", modal.row)
	}
	destinations := modal.destinations()
	if len(destinations) != 2 {
		t.Fatalf("destination count = %d, want 2", len(destinations))
	}
	if !destinations[0].enabled || destinations[1].enabled {
		t.Fatalf("destination enabled states = %#v, want Random enabled and Save Point disabled", destinations)
	}
}

func TestTeleportModalWarpPortalDestinationsSkipEmptySlots(t *testing.T) {
	modal := TeleportModal{}

	modal.OpenWarpPointList(network.WarpPointList{
		SkillID:  warpPortalSkillID,
		MapNames: []string{"prontera", "", "geffen"},
	}, session.Skill{ID: warpPortalSkillID, Level: 1})

	destinations := modal.destinations()
	if len(destinations) != 2 {
		t.Fatalf("destination count = %d, want 2", len(destinations))
	}
	if destinations[0].label != "Save Point: prontera" || destinations[0].mapName != "prontera" {
		t.Fatalf("first destination = %#v", destinations[0])
	}
	if destinations[1].label != "geffen" || destinations[1].mapName != "geffen" {
		t.Fatalf("second destination = %#v", destinations[1])
	}
	if modal.row != 0 {
		t.Fatalf("selected row = %d, want first warp destination", modal.row)
	}
}
