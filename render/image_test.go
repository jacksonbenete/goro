package render

import (
	"image"
	"image/color"
	"testing"
)

func TestNewImageFromStraightAlphaPreservesTransparentRGB(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 32})

	img := NewImageFromStraightAlpha(src)
	got := img.RGBA().RGBAAt(0, 0)
	if got != (color.RGBA{R: 255, G: 255, B: 255, A: 32}) {
		t.Fatalf("straight alpha pixel = %+v, want white rgb with low alpha", got)
	}
}
