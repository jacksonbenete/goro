package render

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var debugTextCache = make(map[string]*Image)

func DrawRect(dst *Image, x, y, w, h float64, c color.Color) {
	if dst == nil || dst.pix == nil || w <= 0 || h <= 0 {
		return
	}
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	if dst.screen {
		drawSolidQuad(dst, x, y, w, h, rgba)
		return
	}
	x0 := clampInt(int(math.Floor(x)), 0, dst.pix.Bounds().Dx())
	y0 := clampInt(int(math.Floor(y)), 0, dst.pix.Bounds().Dy())
	x1 := clampInt(int(math.Ceil(x+w)), 0, dst.pix.Bounds().Dx())
	y1 := clampInt(int(math.Ceil(y+h)), 0, dst.pix.Bounds().Dy())
	for yy := y0; yy < y1; yy++ {
		for xx := x0; xx < x1; xx++ {
			dst.blendPixel(xx, yy, rgba, BlendSourceOver)
		}
	}
}

func DrawLine(dst *Image, x0, y0, x1, y1 float64, c color.Color) {
	if dst == nil || dst.pix == nil {
		return
	}
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	if dst.screen {
		steps := int(math.Max(math.Abs(x1-x0), math.Abs(y1-y0)))
		if steps <= 0 {
			drawSolidQuad(dst, math.Round(x0), math.Round(y0), 1, 1, rgba)
			return
		}
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(steps)
			drawSolidQuad(dst, math.Round(x0+(x1-x0)*t), math.Round(y0+(y1-y0)*t), 1, 1, rgba)
		}
		return
	}
	dx := x1 - x0
	dy := y1 - y0
	steps := int(math.Max(math.Abs(dx), math.Abs(dy)))
	if steps <= 0 {
		dst.blendPixel(clampInt(int(math.Round(x0)), 0, dst.pix.Bounds().Dx()-1), clampInt(int(math.Round(y0)), 0, dst.pix.Bounds().Dy()-1), rgba, BlendSourceOver)
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(x0 + dx*t))
		y := int(math.Round(y0 + dy*t))
		if imageContains(dst, x, y) {
			dst.blendPixel(x, y, rgba, BlendSourceOver)
		}
	}
}

func DebugPrintAt(dst *Image, text string, x, y int) {
	if dst == nil || dst.pix == nil || text == "" {
		return
	}
	if dst.screen {
		img := cachedDebugText(text)
		var opts DrawImageOptions
		opts.GeoM.Translate(float64(x), float64(y))
		opts.Filter = FilterNearest
		dst.DrawImage(img, &opts)
		return
	}
	d := &font.Drawer{
		Dst:  dst.pix,
		Src:  image.NewUniform(color.RGBA{R: 255, G: 255, B: 255, A: 255}),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y+13),
	}
	d.DrawString(text)
}

func cachedDebugText(text string) *Image {
	if img := debugTextCache[text]; img != nil {
		return img
	}
	w := len(text) * 7
	if w < 1 {
		w = 1
	}
	img := NewImage(w, 13)
	DebugPrintAt(img, text, 0, -1)
	debugTextCache[text] = img
	if len(debugTextCache) > 512 {
		for key := range debugTextCache {
			delete(debugTextCache, key)
			if len(debugTextCache) <= 384 {
				break
			}
		}
	}
	return img
}

func imageContains(dst *Image, x, y int) bool {
	b := dst.pix.Bounds()
	return x >= b.Min.X && y >= b.Min.Y && x < b.Max.X && y < b.Max.Y
}

func drawSolidQuad(dst *Image, x, y, w, h float64, c color.RGBA) {
	white := WhiteImage()
	r := float32(c.R) / 255
	g := float32(c.G) / 255
	b := float32(c.B) / 255
	a := float32(c.A) / 255
	vertices := []Vertex{
		{DstX: float32(x), DstY: float32(y), SrcX: 0, SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: float32(x + w), DstY: float32(y), SrcX: 1, SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: float32(x), DstY: float32(y + h), SrcX: 0, SrcY: 1, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: float32(x + w), DstY: float32(y + h), SrcX: 1, SrcY: 1, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
	}
	dst.DrawTriangles(vertices, []uint16{0, 1, 2, 2, 1, 3}, white, &DrawTrianglesOptions{Filter: FilterNearest, Address: AddressUnsafe})
}
