package res

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

func decodeTGA(data []byte) (image.Image, error) {
	if len(data) < 18 {
		return nil, fmt.Errorf("invalid tga: short header")
	}
	idLength := int(data[0])
	colorMapType := data[1]
	imageType := data[2]
	if colorMapType != 0 {
		return nil, fmt.Errorf("unsupported tga color map")
	}
	width := int(binary.LittleEndian.Uint16(data[12:14]))
	height := int(binary.LittleEndian.Uint16(data[14:16]))
	bpp := int(data[16])
	descriptor := data[17]
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid tga size %dx%d", width, height)
	}
	if bpp != 24 && bpp != 32 {
		return nil, fmt.Errorf("unsupported tga bpp %d", bpp)
	}
	offset := 18 + idLength
	if offset > len(data) {
		return nil, fmt.Errorf("invalid tga id length")
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	bytesPerPixel := bpp / 8
	topOrigin := descriptor&0x20 != 0
	switch imageType {
	case 2:
		for i := 0; i < width*height; i++ {
			if offset+bytesPerPixel > len(data) {
				return nil, fmt.Errorf("truncated tga pixels")
			}
			setTGAPixel(img, i, topOrigin, data[offset:offset+bytesPerPixel])
			offset += bytesPerPixel
		}
	case 10:
		for i := 0; i < width*height; {
			if offset >= len(data) {
				return nil, fmt.Errorf("truncated tga rle")
			}
			header := data[offset]
			offset++
			count := int(header&0x7f) + 1
			if header&0x80 != 0 {
				if offset+bytesPerPixel > len(data) {
					return nil, fmt.Errorf("truncated tga rle pixel")
				}
				pixel := data[offset : offset+bytesPerPixel]
				offset += bytesPerPixel
				for j := 0; j < count && i < width*height; j++ {
					setTGAPixel(img, i, topOrigin, pixel)
					i++
				}
				continue
			}
			for j := 0; j < count && i < width*height; j++ {
				if offset+bytesPerPixel > len(data) {
					return nil, fmt.Errorf("truncated tga raw packet")
				}
				setTGAPixel(img, i, topOrigin, data[offset:offset+bytesPerPixel])
				offset += bytesPerPixel
				i++
			}
		}
	default:
		return nil, fmt.Errorf("unsupported tga type %d", imageType)
	}
	return img, nil
}

func setTGAPixel(img *image.NRGBA, index int, topOrigin bool, px []byte) {
	width := img.Bounds().Dx()
	x := index % width
	y := index / width
	if !topOrigin {
		y = img.Bounds().Dy() - 1 - y
	}
	a := uint8(255)
	if len(px) >= 4 {
		a = px[3]
	}
	img.SetNRGBA(x, y, color.NRGBA{R: px[2], G: px[1], B: px[0], A: a})
}
