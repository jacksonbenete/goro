package res

import "testing"

func TestParseItemPairTable(t *testing.T) {
	got := parseItemPairTable([]byte("// comment\n909#Jellopy#\r\n# comment\n0#ignored#\n1002#Poring_Card#extra\n"))
	if got[909] != "Jellopy" {
		t.Fatalf("item 909 = %q", got[909])
	}
	if got[1002] != "Poring_Card" {
		t.Fatalf("item 1002 = %q", got[1002])
	}
	if _, ok := got[0]; ok {
		t.Fatal("zero id should be ignored")
	}
}

func TestNormalizeItemDisplayToken(t *testing.T) {
	if got := normalizeItemDisplayToken("Poring_Card"); got != "Poring Card" {
		t.Fatalf("display = %q", got)
	}
}

func TestItemSpriteResourceCandidates(t *testing.T) {
	got := ItemSpriteResourceCandidates("apple", "spr")
	want := "data\\sprite\\\xBE\xC6\xC0\xCC\xC5\xDB\\apple.spr"
	if len(got) == 0 || got[0] != want {
		t.Fatalf("leading candidates = %#v", got[:minIntForTest(len(got), 2)])
	}
	found := false
	for _, candidate := range got {
		if candidate == "apple.spr" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing bare fallback in %#v", got)
	}
}

func TestItemMetadataLookupFallbacks(t *testing.T) {
	manager := &Manager{
		itemMetadataLoaded: true,
		itemMetadata: map[int]ItemMetadata{
			909: {
				UnidentifiedDisplayName: "Unknown Item",
				IdentifiedDisplayName:   "Jellopy",
				IdentifiedResource:      "jellopy",
			},
		},
	}
	if got, ok := manager.ItemDisplayName(909, true); !ok || got != "Jellopy" {
		t.Fatalf("identified display = %q ok=%v", got, ok)
	}
	if got, ok := manager.ItemDisplayName(909, false); !ok || got != "Unknown Item" {
		t.Fatalf("unidentified display = %q ok=%v", got, ok)
	}
	if got, ok := manager.ItemResourceName(909, false); !ok || got != "jellopy" {
		t.Fatalf("resource fallback = %q ok=%v", got, ok)
	}
}

func TestFormatGroundItemLabel(t *testing.T) {
	if got := FormatGroundItemLabel("Jellopy", 3); got != "Jellopy: 3 ea" {
		t.Fatalf("label = %q", got)
	}
	if got := FormatGroundItemLabel("", 0); got != "Item: 1 ea" {
		t.Fatalf("fallback label = %q", got)
	}
}

func minIntForTest(a, b int) int {
	if a < b {
		return a
	}
	return b
}
