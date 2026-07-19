package render

import (
	"image"
	"image/color"
	"testing"
)

func TestNewImageFromImagePreservesStraightAlphaPixels(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	src.SetNRGBA(0, 0, color.NRGBA{R: 200, G: 80, B: 40, A: 128})

	img := NewImageFromImage(src)
	got := img.RGBA().Pix[:4]
	want := []uint8{200, 80, 40, 128}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pixel bytes = %v, want %v", got, want)
		}
	}
}

func TestFillWritesStraightAlphaPixels(t *testing.T) {
	img := NewImage(1, 1)
	img.Fill(color.RGBA{R: 200, G: 80, B: 40, A: 128})

	got := img.RGBA().Pix[:4]
	want := []uint8{200, 80, 40, 128}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pixel bytes = %v, want %v", got, want)
		}
	}
}
