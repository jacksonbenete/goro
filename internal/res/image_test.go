package res

import (
	"image"
	"image/color"
	"testing"
)

func TestApplyROTransparency(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 255, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 120, G: 40, B: 200, A: 255})

	out := applyROTransparency(img)
	if got := out.NRGBAAt(0, 0).A; got != 0 {
		t.Fatalf("magenta alpha = %d, want 0", got)
	}
	if got := out.NRGBAAt(1, 0).A; got != 255 {
		t.Fatalf("non-key alpha = %d, want 255", got)
	}
}
