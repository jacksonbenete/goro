package gamemode

import (
	"strings"
	"testing"
	"time"

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
	if got, ok := firstEmptyCharacterSlot(characters, 9); !ok || got != 0 {
		t.Fatalf("first empty slot = %d, %t, want 0, true", got, ok)
	}
}

func TestCharacterCreateDefaultStatsAreServerValid(t *testing.T) {
	state := defaultCharCreateState(3)
	if state.slot != 3 {
		t.Fatalf("slot = %d, want 3", state.slot)
	}
	assertCreateStatsValid(t, state.stats)
}

func TestCharacterCreateBumpKeepsClassicPairsValid(t *testing.T) {
	stats := defaultCharCreateState(0).stats
	for i := 0; i < 4; i++ {
		if !bumpCreateStat(&stats, createStatStr) {
			t.Fatalf("bump STR %d failed", i)
		}
	}
	if stats[createStatStr] != 9 || stats[createStatInt] != 1 {
		t.Fatalf("STR/INT = %d/%d, want 9/1", stats[createStatStr], stats[createStatInt])
	}
	if bumpCreateStat(&stats, createStatStr) {
		t.Fatal("bump above STR limit succeeded")
	}
	assertCreateStatsValid(t, stats)
}

func TestAppendCharacterNameInputLimitsBytesAndSkipsControls(t *testing.T) {
	got := appendCharacterNameInput("Kiv", "\nuta漢字", 8)
	if got != "Kivuta" {
		t.Fatalf("name input = %q, want Kivuta", got)
	}
}

func assertCreateStatsValid(t *testing.T, stats [createStatCount]uint8) {
	t.Helper()
	sum := 0
	for _, value := range stats {
		if value < 1 || value > 9 {
			t.Fatalf("stat value %d outside 1..9 in %#v", value, stats)
		}
		sum += int(value)
	}
	if sum != 30 {
		t.Fatalf("stat sum = %d, want 30 in %#v", sum, stats)
	}
	if stats[createStatStr]+stats[createStatInt] != 10 {
		t.Fatalf("STR+INT = %d, want 10", stats[createStatStr]+stats[createStatInt])
	}
	if stats[createStatAgi]+stats[createStatLuk] != 10 {
		t.Fatalf("AGI+LUK = %d, want 10", stats[createStatAgi]+stats[createStatLuk])
	}
	if stats[createStatVit]+stats[createStatDex] != 10 {
		t.Fatalf("VIT+DEX = %d, want 10", stats[createStatVit]+stats[createStatDex])
	}
}

func TestCharacterSelectPreviewFacesViewer(t *testing.T) {
	if got := spriteDirectionFromWorldDir(charSelectPreviewDirection); got != 0 {
		t.Fatalf("char select preview sprite direction = %d, want front-facing direction 0", got)
	}
	if charSelectPreviewScale <= 0.82 {
		t.Fatalf("char select preview scale = %.2f, want larger than old preview", charSelectPreviewScale)
	}
	if charSelectPreviewFeetLift <= 0 {
		t.Fatalf("char select preview feet lift = %d, want positive", charSelectPreviewFeetLift)
	}
}

func TestLoginFadeTransitionsThroughBlack(t *testing.T) {
	start := time.Unix(10, 0)
	mode := NewLoginMode()
	mode.startPhaseFade(loginPhaseCharacter, start)
	if got := mode.fadeAlpha(start); got != 0 {
		t.Fatalf("fade alpha at start = %d, want 0", got)
	}
	if mode.updateFade(start.Add(loginTransitionDuration - time.Millisecond)) {
		t.Fatal("fade unexpectedly entered world")
	}
	if mode.phase != loginPhaseAccount {
		t.Fatalf("phase before black = %d, want account", mode.phase)
	}
	if got := mode.fadeAlpha(start.Add(loginTransitionDuration)); got != 255 {
		t.Fatalf("fade alpha at black = %d, want 255", got)
	}
	if mode.updateFade(start.Add(loginTransitionDuration)) {
		t.Fatal("phase fade unexpectedly entered world")
	}
	if mode.phase != loginPhaseCharacter {
		t.Fatalf("phase after black = %d, want character", mode.phase)
	}
	fadeInStart := start.Add(loginTransitionDuration)
	if got := mode.fadeAlpha(fadeInStart); got != 255 {
		t.Fatalf("fade-in alpha at start = %d, want 255", got)
	}
	mode.updateFade(fadeInStart.Add(loginTransitionDuration))
	if got := mode.fadeAlpha(fadeInStart.Add(loginTransitionDuration)); got != 0 {
		t.Fatalf("fade alpha after fade-in = %d, want 0", got)
	}
	if mode.fade.phase != loginFadeNone {
		t.Fatalf("fade phase after fade-in = %d, want none", mode.fade.phase)
	}
}

func TestLoginWorldFadeWaitsForBlack(t *testing.T) {
	start := time.Unix(20, 0)
	mode := NewLoginMode()
	mode.startWorldFade(start)
	if mode.updateFade(start.Add(loginTransitionDuration - time.Millisecond)) {
		t.Fatal("world handoff completed before black")
	}
	if got := mode.fadeAlpha(start.Add(loginTransitionDuration)); got != 255 {
		t.Fatalf("world fade alpha at handoff = %d, want 255", got)
	}
	if !mode.updateFade(start.Add(loginTransitionDuration)) {
		t.Fatal("world handoff did not complete at black")
	}
}

func TestLoginWindowSitsNearTwoThirdsHeight(t *testing.T) {
	ctx := Context{ScreenW: 1280, ScreenH: 720}
	_, y, _, h := loginWindowRect(ctx)
	centerY := y + h/2
	want := (ctx.ScreenH * 2) / 3
	if centerY != want {
		t.Fatalf("login window centerY = %d, want %d", centerY, want)
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
