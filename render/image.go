package render

import (
	"image"
	"image/color"
)

type Image struct {
	// pix stores straight-alpha RGBA bytes. The CPU rasterizer and GPU upload
	// sample Pix directly and blend with straight alpha, matching robr/WebGL.
	pix     *image.RGBA
	version uint64
}

var whiteImage *Image

func WhiteImage() *Image {
	if whiteImage == nil {
		whiteImage = NewImage(1, 1)
		whiteImage.Fill(color.White)
	}
	return whiteImage
}

func NewImage(width, height int) *Image {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	return &Image{pix: image.NewRGBA(image.Rect(0, 0, width, height))}
}

func NewImageFromImage(src image.Image) *Image {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	copyStraightAlphaImage(dst, src, b)
	return &Image{pix: dst}
}

func (i *Image) Bounds() image.Rectangle {
	if i == nil || i.pix == nil {
		return image.Rectangle{}
	}
	return i.pix.Bounds()
}

func (i *Image) RGBA() *image.RGBA {
	if i == nil {
		return nil
	}
	return i.pix
}

func (i *Image) Fill(c color.Color) {
	if i == nil || i.pix == nil {
		return
	}
	fill := straightAlphaColor(c)
	b := i.pix.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			off := i.pix.PixOffset(x, y)
			i.pix.Pix[off+0] = fill.R
			i.pix.Pix[off+1] = fill.G
			i.pix.Pix[off+2] = fill.B
			i.pix.Pix[off+3] = fill.A
		}
	}
	i.version++
}

func copyStraightAlphaImage(dst *image.RGBA, src image.Image, srcBounds image.Rectangle) {
	width, height := srcBounds.Dx(), srcBounds.Dy()
	if width <= 0 || height <= 0 {
		return
	}
	switch src := src.(type) {
	case *image.NRGBA:
		for y := 0; y < height; y++ {
			srcOff := src.PixOffset(srcBounds.Min.X, srcBounds.Min.Y+y)
			dstOff := dst.PixOffset(0, y)
			copy(dst.Pix[dstOff:dstOff+width*4], src.Pix[srcOff:srcOff+width*4])
		}
		return
	case *image.RGBA:
		for y := 0; y < height; y++ {
			srcOff := src.PixOffset(srcBounds.Min.X, srcBounds.Min.Y+y)
			dstOff := dst.PixOffset(0, y)
			copy(dst.Pix[dstOff:dstOff+width*4], src.Pix[srcOff:srcOff+width*4])
		}
		return
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.NRGBAModel.Convert(src.At(srcBounds.Min.X+x, srcBounds.Min.Y+y)).(color.NRGBA)
			off := dst.PixOffset(x, y)
			dst.Pix[off+0] = c.R
			dst.Pix[off+1] = c.G
			dst.Pix[off+2] = c.B
			dst.Pix[off+3] = c.A
		}
	}
}

func straightAlphaColor(c color.Color) color.RGBA {
	switch c := c.(type) {
	case color.RGBA:
		return c
	case color.NRGBA:
		return color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
	default:
		nrgba := color.NRGBAModel.Convert(c).(color.NRGBA)
		return color.RGBA{R: nrgba.R, G: nrgba.G, B: nrgba.B, A: nrgba.A}
	}
}
