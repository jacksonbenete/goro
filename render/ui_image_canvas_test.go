package render

import (
	"image"
	"testing"
)

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
