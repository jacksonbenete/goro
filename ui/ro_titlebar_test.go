package ui

import (
	"image"
	"image/color"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

func TestTitleBarGradientEndsWithDarkerBottomBorder(t *testing.T) {
	bounds := geometry.NewRect(3, 5, 80, 4)
	canvas := &uitest.MockCanvas{}

	drawTitleBarGradient(canvas, bounds)

	if len(canvas.Rects) != 0 || len(canvas.Images) != 1 {
		t.Fatalf("title bar draws = %d rectangles and %d images, want one continuous image", len(canvas.Rects), len(canvas.Images))
	}
	call := canvas.Images[0]
	if call.At != bounds.Min || call.Image.Bounds().Size() != image.Pt(80, 4) {
		t.Fatalf("title bar image = at %v size %v, want at %v size 80x4", call.At, call.Image.Bounds().Size(), bounds.Min)
	}
	assertImageColor(t, call.Image, 0, 0, rotheme.Default.Colors.WindowTitleTop)
	assertImageColor(t, call.Image, 0, 2, rotheme.Default.Colors.WindowTitle)
	assertImageColor(t, call.Image, 0, 3, rotheme.Default.Colors.WindowBorder)
}

func TestTitleBarGradientRasterizesOpaqueAtFractionalScale(t *testing.T) {
	canvas := &uitest.MockCanvas{}
	drawTitleBarGradient(canvas, geometry.NewRect(0, 0, 80, 4))
	rasterizer, ok := canvas.Images[0].Image.(interface {
		RasterizeForScale(scale float32, width, height int) image.Image
	})
	if !ok {
		t.Fatal("title bar gradient image does not support native scale rasterization")
	}
	scaled := rasterizer.RasterizeForScale(1.5, 120, 6)
	if scaled.Bounds().Size() != image.Pt(120, 6) {
		t.Fatalf("fractional-scale title bar size = %v, want 120x6", scaled.Bounds().Size())
	}
	for y := 0; y < scaled.Bounds().Dy(); y++ {
		for x := 0; x < scaled.Bounds().Dx(); x++ {
			if alpha := color.RGBAModel.Convert(scaled.At(x, y)).(color.RGBA).A; alpha != 255 {
				t.Fatalf("fractional-scale title bar pixel %d,%d alpha = %d, want opaque", x, y, alpha)
			}
		}
	}
	assertImageColor(t, scaled, 0, 3, rotheme.Default.Colors.WindowTitle)
	assertImageColor(t, scaled, 0, 4, rotheme.Default.Colors.WindowBorder)
	assertImageColor(t, scaled, 0, 5, rotheme.Default.Colors.WindowBorder)
}

func assertImageColor(t *testing.T, img image.Image, x, y int, want widget.Color) {
	t.Helper()
	got := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
	r, g, b, a := want.RGBA8()
	if got != (color.RGBA{R: r, G: g, B: b, A: a}) {
		t.Fatalf("image pixel %d,%d = %v, want rgba(%d,%d,%d,%d)", x, y, got, r, g, b, a)
	}
}
