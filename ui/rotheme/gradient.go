package rotheme

import (
	"image"
	"math"
	"sync"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

type verticalGradientImageKey struct {
	width, height             int
	top, bottom, bottomBorder widget.Color
	bottomBorderHeight        float32
}

type scaledVerticalGradientImageKey struct {
	scale         uint32
	width, height int
}

type scaledVerticalGradientImage struct {
	*image.RGBA
	top, bottom, bottomBorder widget.Color
	bottomBorderHeight        float32

	sync.Mutex
	images map[scaledVerticalGradientImageKey]image.Image
}

var verticalGradientImages sync.Map

// DrawVerticalGradient draws a continuous top-to-bottom gradient as one image.
// The image rasterizes itself at the display's physical scale, avoiding seams
// between logical rows at fractional scaling factors.
func DrawVerticalGradient(canvas widget.Canvas, bounds geometry.Rect, top, bottom widget.Color) {
	drawVerticalGradient(canvas, bounds, top, bottom, widget.Color{}, 0)
}

// DrawVerticalGradientWithBottomBorder draws the gradient and its bottom border
// in one image so fractional scaling cannot introduce a seam between them.
func DrawVerticalGradientWithBottomBorder(canvas widget.Canvas, bounds geometry.Rect, top, bottom, border widget.Color, borderHeight float32) {
	drawVerticalGradient(canvas, bounds, top, bottom, border, max(float32(0), borderHeight))
}

func drawVerticalGradient(canvas widget.Canvas, bounds geometry.Rect, top, bottom, bottomBorder widget.Color, bottomBorderHeight float32) {
	if bounds.IsEmpty() {
		return
	}
	width := int(math.Round(float64(bounds.Width())))
	height := int(math.Round(float64(bounds.Height())))
	if width <= 0 || height <= 0 {
		return
	}
	key := verticalGradientImageKey{
		width:              width,
		height:             height,
		top:                top,
		bottom:             bottom,
		bottomBorder:       bottomBorder,
		bottomBorderHeight: bottomBorderHeight,
	}
	value, ok := verticalGradientImages.Load(key)
	if !ok {
		gradient := &scaledVerticalGradientImage{
			RGBA:               rasterizeVerticalGradient(width, height, physicalBorderHeight(bottomBorderHeight, 1, height), top, bottom, bottomBorder),
			top:                top,
			bottom:             bottom,
			bottomBorder:       bottomBorder,
			bottomBorderHeight: bottomBorderHeight,
			images:             make(map[scaledVerticalGradientImageKey]image.Image),
		}
		value, _ = verticalGradientImages.LoadOrStore(key, gradient)
	}
	canvas.DrawImage(value.(image.Image), bounds.Min)
}

func (img *scaledVerticalGradientImage) RasterizeForScale(scale float32, width, height int) image.Image {
	if img == nil || width <= 0 || height <= 0 {
		return nil
	}
	if width == img.Bounds().Dx() && height == img.Bounds().Dy() {
		return img.RGBA
	}
	key := scaledVerticalGradientImageKey{
		scale:  math.Float32bits(scale),
		width:  width,
		height: height,
	}
	img.Lock()
	defer img.Unlock()
	if cached := img.images[key]; cached != nil {
		return cached
	}
	borderHeight := physicalBorderHeight(img.bottomBorderHeight, scale, height)
	rasterized := rasterizeVerticalGradient(width, height, borderHeight, img.top, img.bottom, img.bottomBorder)
	img.images[key] = rasterized
	return rasterized
}

func physicalBorderHeight(logicalHeight, scale float32, totalHeight int) int {
	if logicalHeight <= 0 || totalHeight <= 0 {
		return 0
	}
	if scale <= 0 {
		scale = 1
	}
	return min(totalHeight, max(1, int(math.Round(float64(logicalHeight*scale)))))
}

func rasterizeVerticalGradient(width, height, borderHeight int, top, bottom, border widget.Color) *image.RGBA {
	target := image.NewRGBA(image.Rect(0, 0, max(0, width), max(0, height)))
	if width <= 0 || height <= 0 {
		return target
	}
	borderHeight = min(height, max(0, borderHeight))
	gradientHeight := height - borderHeight
	for y := 0; y < gradientHeight; y++ {
		t := float32(1)
		if gradientHeight > 1 {
			t = float32(y) / float32(gradientHeight-1)
		}
		fillGradientRow(target, y, top.Lerp(bottom, t))
	}
	for y := gradientHeight; y < height; y++ {
		fillGradientRow(target, y, border)
	}
	return target
}

func fillGradientRow(target *image.RGBA, y int, fill widget.Color) {
	r, g, b, a := fill.RGBA8()
	for x := 0; x < target.Bounds().Dx(); x++ {
		offset := target.PixOffset(x, y)
		target.Pix[offset+0] = r
		target.Pix[offset+1] = g
		target.Pix[offset+2] = b
		target.Pix[offset+3] = a
	}
}
