package render

import (
	"image"
	"image/color"
	"testing"
)

type testScaledImage struct {
	*image.RGBA
	calls int
}

func (img *testScaledImage) RasterizeForScale(_ float32, width, height int) image.Image {
	img.calls++
	rasterized := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			rasterized.SetRGBA(x, y, color.RGBA{R: 0x7f, A: 0xff})
		}
	}
	return rasterized
}

func TestScaledCanvasPhysicalSizeRoundsLogicalSize(t *testing.T) {
	tests := []struct {
		logical int
		scale   float32
		want    int
	}{
		{logical: 24, scale: 1.5, want: 36},
		{logical: 25, scale: 1.25, want: 31},
		{logical: 32, scale: 2, want: 64},
	}
	for _, tt := range tests {
		if got := scaledCanvasPhysicalSize(tt.logical, tt.scale); got != tt.want {
			t.Fatalf("scaledCanvasPhysicalSize(%d, %.2f) = %d, want %d", tt.logical, tt.scale, got, tt.want)
		}
	}
}

func TestScaledCanvasSceneImageUsesPhysicalTargetSize(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 24, 24))
	img := scaledCanvasSceneImage(src, 24, 24, 1.5)
	if img == nil {
		t.Fatal("scaledCanvasSceneImage returned nil")
	}
	if img.Width != 36 || img.Height != 36 {
		t.Fatalf("scene image size = %dx%d, want 36x36", img.Width, img.Height)
	}
}

func TestScaledCanvasSceneImageUsesNativeScaleRasterizer(t *testing.T) {
	src := &testScaledImage{RGBA: image.NewRGBA(image.Rect(0, 0, 12, 8))}
	img := scaledCanvasSceneImage(src, 12, 8, 2)
	if img == nil {
		t.Fatal("scaledCanvasSceneImage returned nil")
	}
	if src.calls != 1 {
		t.Fatalf("native-scale rasterizer calls = %d, want 1", src.calls)
	}
	if img.Width != 24 || img.Height != 16 {
		t.Fatalf("scene image size = %dx%d, want 24x16", img.Width, img.Height)
	}
	if got := img.Data[:4]; got[0] != 0x7f || got[1] != 0 || got[2] != 0 || got[3] != 0xff {
		t.Fatalf("first scene pixel = %v, want native rasterizer pixel", got)
	}
}
