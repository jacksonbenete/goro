package res

import (
	"encoding/binary"
	"fmt"
	"image"
)

const (
	SPRFramePalette = 0
	SPRFrameRGBA    = 1
)

type SPR struct {
	VersionMajor int
	VersionMinor int
	IndexedCount int
	RGBACount    int
	RGBAIndex    int
	Palette      Palette
	Frames       []SPRFrame
}

type SPRFrame struct {
	Type   int
	Width  int
	Height int
	Data   []byte
}

func ParseSPR(data []byte) (*SPR, error) {
	reader := sprReader{data: data}
	if string(reader.bytes(2)) != "SP" {
		return nil, fmt.Errorf("spr invalid header")
	}
	spr := &SPR{
		VersionMinor: int(reader.u8()),
		VersionMajor: int(reader.u8()),
	}
	spr.IndexedCount = int(reader.u16())
	if spr.versionAtLeast(1, 2) {
		spr.RGBACount = int(reader.u16())
	}
	spr.RGBAIndex = spr.IndexedCount
	spr.Frames = make([]SPRFrame, spr.IndexedCount+spr.RGBACount)

	for i := 0; i < spr.IndexedCount; i++ {
		frame, err := reader.readIndexedFrame(spr.versionAtLeast(2, 1))
		if err != nil {
			return nil, err
		}
		spr.Frames[i] = frame
	}
	for i := 0; i < spr.RGBACount; i++ {
		width := int(reader.i16())
		height := int(reader.i16())
		if width < 0 || height < 0 {
			return nil, fmt.Errorf("spr rgba frame %d negative size %dx%d", i, width, height)
		}
		pixels := reader.bytes(width * height * 4)
		if reader.err != nil {
			return nil, reader.err
		}
		spr.Frames[spr.RGBAIndex+i] = SPRFrame{
			Type:   SPRFrameRGBA,
			Width:  width,
			Height: height,
			Data:   append([]byte(nil), pixels...),
		}
	}

	if spr.versionAtLeast(1, 1) && len(data) >= 1024 {
		palette := data[len(data)-1024:]
		for i := 0; i < 256; i++ {
			copy(spr.Palette[i][:], palette[i*4:i*4+4])
		}
	}
	if reader.err != nil {
		return nil, reader.err
	}
	return spr, nil
}

func (s *SPR) versionAtLeast(major, minor int) bool {
	return s.VersionMajor > major || (s.VersionMajor == major && s.VersionMinor >= minor)
}

func (s *SPR) FrameImage(index int, sprType int) (*image.NRGBA, bool) {
	return s.FrameImageWithPalette(index, sprType, nil)
}

func (s *SPR) FrameImageWithPalette(index int, sprType int, palette *Palette) (*image.NRGBA, bool) {
	if sprType == SPRFrameRGBA {
		index += s.RGBAIndex
	}
	if index < 0 || index >= len(s.Frames) {
		return nil, false
	}
	frame := s.Frames[index]
	if frame.Width <= 0 || frame.Height <= 0 {
		return nil, false
	}
	out := image.NewNRGBA(image.Rect(0, 0, frame.Width, frame.Height))
	framePalette := &s.Palette
	if palette != nil {
		framePalette = palette
	}
	switch frame.Type {
	case SPRFramePalette:
		for y := 0; y < frame.Height; y++ {
			srcRow := y * frame.Width
			dstRow := y * out.Stride
			for x := 0; x < frame.Width; x++ {
				paletteIndex := frame.Data[srcRow+x]
				pal := framePalette[paletteIndex]
				dst := dstRow + x*4
				out.Pix[dst+0] = pal[0]
				out.Pix[dst+1] = pal[1]
				out.Pix[dst+2] = pal[2]
				if paletteIndex == 0 || isROMagenta(pal[0], pal[1], pal[2]) {
					out.Pix[dst+3] = 0
				} else {
					out.Pix[dst+3] = 255
				}
			}
		}
		clearTransparentRGB(out)
	case SPRFrameRGBA:
		for y := 0; y < frame.Height; y++ {
			srcRow := y * frame.Width * 4
			dstRow := (frame.Height - y - 1) * out.Stride
			for x := 0; x < frame.Width; x++ {
				src := srcRow + x*4
				dst := dstRow + x*4
				out.Pix[dst+0] = frame.Data[src+3]
				out.Pix[dst+1] = frame.Data[src+2]
				out.Pix[dst+2] = frame.Data[src+1]
				out.Pix[dst+3] = frame.Data[src+0]
				if isROMagenta(out.Pix[dst+0], out.Pix[dst+1], out.Pix[dst+2]) {
					out.Pix[dst+3] = 0
				}
			}
		}
		clearTransparentRGB(out)
	default:
		return nil, false
	}
	return out, true
}

type sprReader struct {
	data   []byte
	offset int
	err    error
}

func (r *sprReader) readIndexedFrame(rle bool) (SPRFrame, error) {
	width := int(r.u16())
	height := int(r.u16())
	size := width * height
	if width < 0 || height < 0 || size < 0 {
		return SPRFrame{}, fmt.Errorf("spr indexed frame invalid size %dx%d", width, height)
	}
	frame := SPRFrame{Type: SPRFramePalette, Width: width, Height: height}
	if !rle {
		pixels := r.bytes(size)
		if r.err != nil {
			return SPRFrame{}, r.err
		}
		frame.Data = append([]byte(nil), pixels...)
		return frame, nil
	}

	end := int(r.u16()) + r.offset
	if end < r.offset || end > len(r.data) {
		return SPRFrame{}, fmt.Errorf("spr rle frame truncated")
	}
	frame.Data = make([]byte, size)
	write := 0
	for r.offset < end {
		value := r.u8()
		if write >= len(frame.Data) {
			return SPRFrame{}, fmt.Errorf("spr rle frame overflow")
		}
		frame.Data[write] = value
		write++
		if value == 0 && r.offset < end {
			count := int(r.u8())
			if count == 0 {
				if write >= len(frame.Data) {
					return SPRFrame{}, fmt.Errorf("spr rle frame overflow")
				}
				frame.Data[write] = 0
				write++
				continue
			}
			for i := 1; i < count; i++ {
				if write >= len(frame.Data) {
					return SPRFrame{}, fmt.Errorf("spr rle frame overflow")
				}
				frame.Data[write] = 0
				write++
			}
		}
	}
	if write != len(frame.Data) {
		return SPRFrame{}, fmt.Errorf("spr rle frame decoded %d bytes, expected %d", write, len(frame.Data))
	}
	return frame, nil
}

func (r *sprReader) bytes(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || r.offset+n > len(r.data) {
		r.err = fmt.Errorf("spr truncated at offset %d reading %d bytes", r.offset, n)
		return nil
	}
	out := r.data[r.offset : r.offset+n]
	r.offset += n
	return out
}

func (r *sprReader) u8() byte {
	data := r.bytes(1)
	if r.err != nil {
		return 0
	}
	return data[0]
}

func (r *sprReader) u16() uint16 {
	data := r.bytes(2)
	if r.err != nil {
		return 0
	}
	return binary.LittleEndian.Uint16(data)
}

func (r *sprReader) i16() int16 {
	return int16(r.u16())
}
