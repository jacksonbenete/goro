package res

import (
	"errors"
	"os"
	"testing"
)

func TestEffectTextureRealArchiveWhenConfigured(t *testing.T) {
	path := os.Getenv("GORO_TEST_GRF")
	if path == "" {
		t.Skip("set GORO_TEST_GRF to run against a real archive")
	}
	grf, err := OpenGRF(path)
	if err != nil {
		if errors.Is(err, ErrGRFUnsupportedVersion) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer grf.Close()

	manager := &Manager{Archives: []*GRF{grf}}
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
