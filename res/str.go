package res

import (
	"encoding/binary"
	"fmt"
	"math"
)

type STR struct {
	FPS    int
	MaxKey int
	Layers []STRLayer
}

type STRLayer struct {
	Textures   []string
	Animations []STRAnimation
}

type STRAnimation struct {
	Frame     int
	Type      uint32
	Pos       [2]float32
	UV        [8]float32
	XY        [8]float32
	AniFrame  float32
	AniType   uint32
	Delay     float32
	Angle     float32
	Color     [4]float32
	SrcAlpha  uint32
	DestAlpha uint32
	MTPreset  uint32
}

func ParseSTR(data []byte, texturePath string) (*STR, error) {
	reader := strReader{data: data}
	if header := string(reader.bytes(4)); header != "STRM" {
		return nil, fmt.Errorf("str invalid header %q", header)
	}
	version := reader.u32()
	if version != 0x94 {
		return nil, fmt.Errorf("str unsupported version 0x%X", version)
	}
	out := &STR{
		FPS:    int(reader.u32()),
		MaxKey: int(reader.u32()),
	}
	layerCount := int(reader.u32())
	if layerCount < 0 || layerCount > 256 {
		return nil, fmt.Errorf("str invalid layer count %d", layerCount)
	}
	reader.skip(16)
	out.Layers = make([]STRLayer, layerCount)
	for i := range out.Layers {
		layer, err := reader.layer(texturePath)
		if err != nil {
			return nil, err
		}
		out.Layers[i] = layer
	}
	if reader.err != nil {
		return nil, reader.err
	}
	return out, nil
}

func (r *strReader) layer(texturePath string) (STRLayer, error) {
	texCount := int(r.i32())
	if texCount < 0 || texCount > 1024 {
		return STRLayer{}, fmt.Errorf("str invalid texture count %d", texCount)
	}
	layer := STRLayer{Textures: make([]string, texCount)}
	for i := range layer.Textures {
		name := fixedBinaryString(r.bytes(128))
		if name != "" {
			layer.Textures[i] = "data\\texture\\effect\\" + texturePath + name
		}
	}
	keyCount := int(r.i32())
	if keyCount < 0 || keyCount > 4096 {
		return STRLayer{}, fmt.Errorf("str invalid animation count %d", keyCount)
	}
	layer.Animations = make([]STRAnimation, keyCount)
	for i := range layer.Animations {
		layer.Animations[i] = r.animation()
	}
	return layer, r.err
}

func (r *strReader) animation() STRAnimation {
	anim := STRAnimation{
		Frame: int(r.i32()),
		Type:  r.u32(),
		Pos:   [2]float32{r.f32(), r.f32()},
	}
	for i := range anim.UV {
		anim.UV[i] = r.f32()
	}
	for i := range anim.XY {
		anim.XY[i] = r.f32()
	}
	anim.AniFrame = r.f32()
	anim.AniType = r.u32()
	anim.Delay = r.f32()
	anim.Angle = r.f32() / (1024.0 / 360.0)
	for i := range anim.Color {
		anim.Color[i] = r.f32() / 255.0
	}
	anim.SrcAlpha = r.u32()
	anim.DestAlpha = r.u32()
	anim.MTPreset = r.u32()
	return anim
}

type strReader struct {
	data   []byte
	offset int
	err    error
}

func (r *strReader) bytes(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || r.offset+n > len(r.data) {
		r.err = fmt.Errorf("str truncated at offset %d reading %d bytes", r.offset, n)
		return nil
	}
	out := r.data[r.offset : r.offset+n]
	r.offset += n
	return out
}

func (r *strReader) skip(n int) {
	_ = r.bytes(n)
}

func (r *strReader) u32() uint32 {
	data := r.bytes(4)
	if len(data) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(data)
}

func (r *strReader) i32() int32 {
	return int32(r.u32())
}

func (r *strReader) f32() float32 {
	return math.Float32frombits(r.u32())
}
