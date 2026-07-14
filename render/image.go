package render

import (
	"image"
	"image/color"
	"image/draw"
)

type Image struct {
	pix             *image.RGBA
	screen          bool
	screenScaleX    float32
	screenScaleY    float32
	version         uint64
	commands        []DrawCommand
	worldCommands   []WorldCommand
	worldMeshes     []WorldMeshCommand
	worldBillboards []WorldBillboardCommand
	uiRects         []UIRectCommand
	uiTextBoxes     []UITextBoxCommand
	uiTextLabels    []UITextLabelCommand
	clear           color.RGBA
	camera          Camera3D
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
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
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
	if i.screen {
		i.clear = color.RGBAModel.Convert(c).(color.RGBA)
		return
	}
	draw.Draw(i.pix, i.pix.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	i.version++
}
