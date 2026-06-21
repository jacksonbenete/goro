package res

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestParseGND(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("GRGN")
	buf.WriteByte(1)
	buf.WriteByte(8)
	writeI32(&buf, 1)
	writeI32(&buf, 1)
	writeF32(&buf, 10)

	writeI32(&buf, 1)
	writeI32(&buf, 16)
	buf.WriteString("stone.bmp")
	buf.Write(make([]byte, 16-len("stone.bmp")))

	writeI32(&buf, 0)
	writeI32(&buf, 8)
	writeI32(&buf, 8)
	writeI32(&buf, 1)
	writeI32(&buf, 0)

	writeI32(&buf, 1)
	for i := 0; i < 8; i++ {
		writeF32(&buf, float32(i)/10)
	}
	writeU16(&buf, 0)
	writeU16(&buf, 0)
	buf.Write([]byte{10, 20, 30, 255})

	for _, h := range []float32{10, 20, 30, 40} {
		writeF32(&buf, h)
	}
	writeI32(&buf, 0)
	writeI32(&buf, -1)
	writeI32(&buf, -1)

	gnd, err := ParseGND(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if gnd.VersionMajor != 1 || gnd.VersionMinor != 8 || gnd.Width != 1 || gnd.Height != 1 || gnd.Zoom != 10 {
		t.Fatalf("unexpected header: %+v", gnd)
	}
	if len(gnd.Textures) != 1 || gnd.Textures[0] != "stone.bmp" {
		t.Fatalf("unexpected textures: %#v", gnd.Textures)
	}
	surface, ok := gnd.Surface(0)
	if !ok {
		t.Fatal("missing surface")
	}
	if surface.TextureID != 0 || surface.Color.R != 30 || surface.Color.G != 20 || surface.Color.B != 10 || surface.Color.A != 255 {
		t.Fatalf("unexpected surface: %+v", surface)
	}
	cell, ok := gnd.Cell(0, 0)
	if !ok {
		t.Fatal("missing cell")
	}
	if cell.Top != 0 || cell.Front != -1 || cell.Right != -1 || cell.Heights[0] != 2 {
		t.Fatalf("unexpected cell: %+v", cell)
	}
}

func TestFixedBinaryStringPreservesNonUTF8Bytes(t *testing.T) {
	raw := []byte{' ', 0xc7, 0xca, '\\', 'P', 'R', 'T', 0xf8, '.', 'b', 'm', 'p', ' ', 0, 'x'}
	got := fixedBinaryString(raw)
	want := string([]byte{0xc7, 0xca, '\\', 'P', 'R', 'T', 0xf8, '.', 'b', 'm', 'p'})
	if got != want {
		t.Fatalf("fixedBinaryString bytes = % x, want % x", []byte(got), []byte(want))
	}
}

func TestUnpackGNDLightmap5BitChannel(t *testing.T) {
	var channel [40]byte
	for i := range channel {
		channel[i] = 0xff
	}
	values := unpackGNDLightmap5BitChannel(channel)
	for i, value := range values {
		if value != 248 {
			t.Fatalf("value[%d] = %d, want 248", i, value)
		}
	}
}

func TestDecodeGNDLightmapRaw(t *testing.T) {
	raw := make([]byte, 256)
	for i, sample := range gndLightmapVertexSamples {
		pixel := sample.y*8 + sample.x
		raw[pixel] = byte(32 + i)
		base := 64 + pixel*3
		raw[base] = byte(64 + i)
		raw[base+1] = byte(96 + i)
		raw[base+2] = byte(128 + i)
	}
	lightmap := decodeGNDLightmapRaw(raw)
	for i := range lightmap.Alpha {
		if lightmap.Alpha[i] != byte(32+i) {
			t.Fatalf("alpha[%d] = %d", i, lightmap.Alpha[i])
		}
		if lightmap.Color[i].R != byte(64+i) || lightmap.Color[i].G != byte(96+i) || lightmap.Color[i].B != byte(128+i) {
			t.Fatalf("color[%d] = %+v", i, lightmap.Color[i])
		}
	}
}

func writeI32(buf *bytes.Buffer, value int32) {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], uint32(value))
	buf.Write(tmp[:])
}

func writeU16(buf *bytes.Buffer, value uint16) {
	var tmp [2]byte
	binary.LittleEndian.PutUint16(tmp[:], value)
	buf.Write(tmp[:])
}

func writeF32(buf *bytes.Buffer, value float32) {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], math.Float32bits(value))
	buf.Write(tmp[:])
}
