package res

import (
	"testing"
)

func TestEffectTextureRealArchiveWhenConfigured(t *testing.T) {
	manager := realDataManager(t)
	img, source, err := LoadImage(manager, EffectTextureCandidates("ring_blue"))
	if err != nil {
		t.Fatal(err)
	}
	img = ApplyEffectTransparency(img)
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("invalid %s bounds %v", source, img.Bounds())
	}
	t.Logf("decoded %s bounds=%v", source, img.Bounds())
}
