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

func TestBlendSourceOverPreservesStraightAlphaOnTransparentTarget(t *testing.T) {
	dst := NewImage(1, 1)
	dst.blendPixel(0, 0, color.RGBA{R: 255, G: 240, B: 200, A: 64}, BlendSourceOver)

	got := dst.RGBA().RGBAAt(0, 0)
	want := color.RGBA{R: 255, G: 240, B: 200, A: 64}
	if got != want {
		t.Fatalf("source-over transparent target = %+v, want %+v", got, want)
	}
}

func TestBlendSourceOverMatchesRegularBlendOnOpaqueTarget(t *testing.T) {
	dst := NewImage(1, 1)
	dst.Fill(color.RGBA{R: 10, G: 20, B: 30, A: 255})
	dst.blendPixel(0, 0, color.RGBA{R: 100, G: 80, B: 60, A: 128}, BlendSourceOver)

	got := dst.RGBA().RGBAAt(0, 0)
	want := color.RGBA{R: 55, G: 50, B: 45, A: 255}
	if got != want {
		t.Fatalf("source-over opaque target = %+v, want %+v", got, want)
	}
}
