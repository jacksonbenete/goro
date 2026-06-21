package res

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"math"
)

type GND struct {
	VersionMajor byte
	VersionMinor byte
	Width        int
	Height       int
	Zoom         float32
	Textures     []string
	Surfaces     []GNDSurface
	Cells        []GNDCell
}

type GNDSurface struct {
	U          [4]float32
	V          [4]float32
	TextureID  int
	LightmapID int
	Color      color.RGBA
}

type GNDCell struct {
	Heights [4]float32
	Top     int
	Front   int
	Right   int
}

func ParseGND(data []byte) (*GND, error) {
	reader := gndReader{data: data}
	if string(reader.bytes(4)) != "GRGN" {
		return nil, fmt.Errorf("invalid gnd signature")
	}

	major := reader.u8()
	minor := reader.u8()
	width := int(reader.i32())
	height := int(reader.i32())
	zoom := reader.f32()
	if reader.err != nil {
		return nil, reader.err
	}
	if major != 1 || minor <= 6 {
		return nil, fmt.Errorf("unsupported gnd version %d.%d", major, minor)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid gnd size %dx%d", width, height)
	}

	textureCount := int(reader.i32())
	textureNameLength := int(reader.i32())
	if textureCount < 0 || textureNameLength <= 0 {
		return nil, fmt.Errorf("invalid gnd texture table count=%d length=%d", textureCount, textureNameLength)
	}
	textures := make([]string, textureCount)
	for i := range textures {
		textures[i] = fixedBinaryString(reader.bytes(textureNameLength))
	}

	lightmapCount := int(reader.i32())
	_ = reader.i32()
	_ = reader.i32()
	_ = reader.i32()
	if lightmapCount < 0 {
		return nil, fmt.Errorf("invalid gnd lightmap count=%d", lightmapCount)
	}
	if minor == 7 {
		reader.skip(lightmapCount * 256)
	} else {
		reader.skip(lightmapCount * 16)
		colorChannelCount := int(reader.i32())
		if colorChannelCount < 0 {
			return nil, fmt.Errorf("invalid gnd color channel count=%d", colorChannelCount)
		}
		reader.skip(colorChannelCount * 40)
	}

	surfaceCount := int(reader.i32())
	if surfaceCount < 0 {
		return nil, fmt.Errorf("invalid gnd surface count=%d", surfaceCount)
	}
	surfaces := make([]GNDSurface, surfaceCount)
	for i := range surfaces {
		var surface GNDSurface
		for j := range surface.U {
			surface.U[j] = reader.f32()
		}
		for j := range surface.V {
			surface.V[j] = reader.f32()
		}
		surface.TextureID = int(reader.u16())
		surface.LightmapID = int(reader.u16())
		b := reader.u8()
		g := reader.u8()
		r := reader.u8()
		a := reader.u8()
		surface.Color = color.RGBA{R: r, G: g, B: b, A: a}
		surfaces[i] = surface
	}

	cellCount := width * height
	cells := make([]GNDCell, cellCount)
	for i := range cells {
		var cell GNDCell
		for j := range cell.Heights {
			cell.Heights[j] = reader.f32() * 0.2
		}
		cell.Top = int(reader.i32())
		cell.Front = int(reader.i32())
		cell.Right = int(reader.i32())
		cells[i] = cell
	}
	if reader.err != nil {
		return nil, reader.err
	}

	return &GND{
		VersionMajor: major,
		VersionMinor: minor,
		Width:        width,
		Height:       height,
		Zoom:         zoom,
		Textures:     textures,
		Surfaces:     surfaces,
		Cells:        cells,
	}, nil
}

func (g *GND) InBounds(x, y int) bool {
	return g != nil && x >= 0 && y >= 0 && x < g.Width && y < g.Height
}

func (g *GND) Cell(x, y int) (GNDCell, bool) {
	if !g.InBounds(x, y) {
		return GNDCell{}, false
	}
	return g.Cells[y*g.Width+x], true
}

func (g *GND) Surface(index int) (GNDSurface, bool) {
	if g == nil || index < 0 || index >= len(g.Surfaces) {
		return GNDSurface{}, false
	}
	return g.Surfaces[index], true
}

func fixedBinaryString(data []byte) string {
	if index := bytes.IndexByte(data, 0); index >= 0 {
		data = data[:index]
	}
	for len(data) > 0 && isASCIIWhitespace(data[0]) {
		data = data[1:]
	}
	for len(data) > 0 && isASCIIWhitespace(data[len(data)-1]) {
		data = data[:len(data)-1]
	}
	return string(data)
}

func isASCIIWhitespace(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

type gndReader struct {
	data   []byte
	offset int
	err    error
}

func (r *gndReader) bytes(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || r.offset+n > len(r.data) {
		r.err = fmt.Errorf("gnd truncated at offset %d reading %d bytes", r.offset, n)
		return nil
	}
	out := r.data[r.offset : r.offset+n]
	r.offset += n
	return out
}

func (r *gndReader) skip(n int) {
	_ = r.bytes(n)
}

func (r *gndReader) u8() byte {
	data := r.bytes(1)
	if len(data) == 0 {
		return 0
	}
	return data[0]
}

func (r *gndReader) u16() uint16 {
	data := r.bytes(2)
	if len(data) < 2 {
		return 0
	}
	return binary.LittleEndian.Uint16(data)
}

func (r *gndReader) i32() int32 {
	data := r.bytes(4)
	if len(data) < 4 {
		return 0
	}
	return int32(binary.LittleEndian.Uint32(data))
}

func (r *gndReader) f32() float32 {
	data := r.bytes(4)
	if len(data) < 4 {
		return 0
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(data))
}
