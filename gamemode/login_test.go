package gamemode

import (
	"strings"
	"testing"
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
