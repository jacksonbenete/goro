package gamemode

import (
	"strings"
	"testing"

	"github.com/kivutar/goro/session"
)

func TestLoginBackgroundSetsPrefer2008SingleImage(t *testing.T) {
	sets := loginBackgroundSets(20080910)
	if len(sets) == 0 || len(sets[0]) != 1 || sets[0][0] != "bgi_temp.bmp" {
		t.Fatalf("first 2008 login background set = %#v", sets)
	}
}

func TestLoginBackgroundSetsIncludeModernTiles(t *testing.T) {
	sets := loginBackgroundSets(20181114)
	if len(sets) == 0 || len(sets[0]) != 12 {
		t.Fatalf("first 2018 login background set = %#v", sets)
	}
}

func TestLoginInterfaceCandidatesUseROInterfacePath(t *testing.T) {
	candidates := loginInterfaceCandidates("bgi_temp.bmp")
	if len(candidates) == 0 {
		t.Fatal("no candidates")
	}
	if !strings.HasPrefix(candidates[0], "data\\texture\\") || !strings.HasSuffix(candidates[0], "\\bgi_temp.bmp") {
		t.Fatalf("first candidate = %q", candidates[0])
	}
}

func TestTrimLastRune(t *testing.T) {
	if got := trimLastRune("abé"); got != "ab" {
		t.Fatalf("trimLastRune = %q, want ab", got)
	}
	if got := trimLastRune(""); got != "" {
		t.Fatalf("trimLastRune empty = %q", got)
	}
}

func TestLoginBackgroundRealDataWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	img, source, ok := loadLoginBackgroundImage(manager, "bgi_temp.bmp")
	if !ok {
		t.Skip("bgi_temp.bmp not present in configured client data")
	}
	if img == nil || img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("invalid login background from %s: %#v", source, img)
	}
}

func TestCharacterSelectSlotHelpers(t *testing.T) {
	characters := []session.Character{
		{ID: 10, Slot: 5},
		{ID: 11, Slot: 2},
	}
	if got := firstOccupiedCharacterSlot(characters); got != 2 {
		t.Fatalf("first occupied slot = %d, want 2", got)
	}
	if got := charSelectMaxSlots(characters); got != 9 {
		t.Fatalf("max slots = %d, want 9", got)
	}
	if got := charSelectPage(5); got != 1 {
		t.Fatalf("page = %d, want 1", got)
	}
	character, ok := characterBySlot(characters, 5)
	if !ok || character.ID != 10 {
		t.Fatalf("characterBySlot = %+v, %t", character, ok)
	}
	if got := clampCharacterSlot(99, 9); got != 8 {
		t.Fatalf("clamp high = %d, want 8", got)
	}
}

func TestCharacterSelectPreviewFacesViewer(t *testing.T) {
	if got := spriteDirectionFromWorldDir(charSelectPreviewDirection); got != 0 {
		t.Fatalf("char select preview sprite direction = %d, want front-facing direction 0", got)
	}
}

func TestCharacterSelectSkinRealDataWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	for _, name := range []string{"login_interface/win_select.bmp", "login_interface/box_select.bmp"} {
		img, source, ok := loadLoginBackgroundImage(manager, name)
		if !ok {
			t.Skipf("%s not present in configured client data", name)
		}
		if img == nil || img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
			t.Fatalf("invalid char select skin from %s: %#v", source, img)
		}
	}
}
