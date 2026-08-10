package appicon

import "testing"

func TestImage(t *testing.T) {
	bounds := Image().Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Fatalf("icon dimensions = %dx%d, want 32x32", bounds.Dx(), bounds.Dy())
	}
}
