package res

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestParseSPRIndexedFrame(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("SP")
	buf.WriteByte(1)
	buf.WriteByte(1)
	writeTestU16(&buf, 1)
	writeTestU16(&buf, 2)
	writeTestU16(&buf, 1)
	buf.Write([]byte{1, 0})
	palette := make([]byte, 1024)
	palette[4] = 10
	palette[5] = 20
	palette[6] = 30
	palette[7] = 255
	buf.Write(palette)

	spr, err := ParseSPR(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if spr.VersionMajor != 1 || spr.VersionMinor != 1 || len(spr.Frames) != 1 {
		t.Fatalf("unexpected spr header: v%d.%d frames=%d", spr.VersionMajor, spr.VersionMinor, len(spr.Frames))
	}
	img, ok := spr.FrameImage(0, SPRFramePalette)
	if !ok {
		t.Fatal("frame image missing")
	}
	if got := img.Pix[:8]; got[0] != 10 || got[1] != 20 || got[2] != 30 || got[3] != 255 || got[7] != 0 {
		t.Fatalf("unexpected pixels: %v", got)
	}
}

func TestSPRFrameImageWithPaletteOverride(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("SP")
	buf.WriteByte(1)
	buf.WriteByte(1)
	writeTestU16(&buf, 1)
	writeTestU16(&buf, 1)
	writeTestU16(&buf, 1)
	buf.WriteByte(1)
	buf.Write(make([]byte, 1024))

	spr, err := ParseSPR(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	var palette Palette
	palette[1] = [4]byte{40, 50, 60, 255}
	img, ok := spr.FrameImageWithPalette(0, SPRFramePalette, &palette)
	if !ok {
		t.Fatal("frame image missing")
	}
	if got := img.Pix[:4]; got[0] != 40 || got[1] != 50 || got[2] != 60 || got[3] != 255 {
		t.Fatalf("unexpected override pixel: %v", got)
	}
}

func TestSPRFrameImageClearsTransparentRGB(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("SP")
	buf.WriteByte(2)
	buf.WriteByte(1)
	writeTestU16(&buf, 0)
	writeTestU16(&buf, 1)
	writeTestU16(&buf, 1)
	writeTestU16(&buf, 1)
	buf.Write([]byte{0, 128, 255, 0})
	buf.Write(make([]byte, 1024))

	spr, err := ParseSPR(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	img, ok := spr.FrameImage(0, SPRFrameRGBA)
	if !ok {
		t.Fatal("frame image missing")
	}
	if got := img.Pix[:4]; got[0] != 0 || got[1] != 0 || got[2] != 0 || got[3] != 0 {
		t.Fatalf("transparent rgba pixel = %v, want fully zero", got)
	}
}

func TestSPRRGBAFrameImageUsesABGRByteOrder(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("SP")
	buf.WriteByte(2)
	buf.WriteByte(1)
	writeTestU16(&buf, 0)
	writeTestU16(&buf, 1)
	writeTestU16(&buf, 1)
	writeTestU16(&buf, 1)
	buf.Write([]byte{128, 30, 20, 10})
	buf.Write(make([]byte, 1024))

	spr, err := ParseSPR(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	img, ok := spr.FrameImage(0, SPRFrameRGBA)
	if !ok {
		t.Fatal("frame image missing")
	}
	if got := img.Pix[:4]; got[0] != 10 || got[1] != 20 || got[2] != 30 || got[3] != 128 {
		t.Fatalf("rgba pixel = %v, want [10 20 30 128]", got)
	}
}

func TestParseSPRRLEFrame(t *testing.T) {
	var rle bytes.Buffer
	rle.WriteByte(5)
	rle.WriteByte(0)
	rle.WriteByte(3)
	rle.WriteByte(7)

	var buf bytes.Buffer
	buf.WriteString("SP")
	buf.WriteByte(1)
	buf.WriteByte(2)
	writeTestU16(&buf, 1)
	writeTestU16(&buf, 0)
	writeTestU16(&buf, 5)
	writeTestU16(&buf, 1)
	writeTestU16(&buf, uint16(rle.Len()))
	buf.Write(rle.Bytes())
	buf.Write(make([]byte, 1024))

	spr, err := ParseSPR(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if got := spr.Frames[0].Data; !bytes.Equal(got, []byte{5, 0, 0, 0, 7}) {
		t.Fatalf("rle decoded %v", got)
	}
}

func TestParseACTLayer(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("AC")
	buf.WriteByte(4)
	buf.WriteByte(2)
	writeTestU16(&buf, 1)
	buf.Write(make([]byte, 10))
	writeTestU32(&buf, 1)
	buf.Write(make([]byte, 32))
	writeTestU32(&buf, 1)
	writeTestI32(&buf, 12)
	writeTestI32(&buf, -34)
	writeTestI32(&buf, 3)
	writeTestI32(&buf, 1)
	buf.Write([]byte{255, 128, 64, 32})
	writeTestF32(&buf, 1.5)
	writeTestF32(&buf, 0.75)
	writeTestI32(&buf, 45)
	writeTestI32(&buf, 1)
	writeTestI32(&buf, -1)
	writeTestI32(&buf, 0)
	writeTestI32(&buf, 0)
	writeTestF32(&buf, 6)

	act, err := ParseACT(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	action, ok := act.ActionFor(0, 0)
	if !ok || len(action.Animations) != 1 {
		t.Fatalf("missing action: ok=%v animations=%d", ok, len(action.Animations))
	}
	layer := action.Animations[0].Layers[0]
	if layer.X != 12 || layer.Y != -34 || layer.Index != 3 || !layer.Mirror || layer.SPRType != 1 {
		t.Fatalf("bad layer: %+v", layer)
	}
	if math.Abs(float64(layer.ScaleX-1.5)) > 0.001 || math.Abs(float64(layer.ScaleY-0.75)) > 0.001 {
		t.Fatalf("bad scale: %f %f", layer.ScaleX, layer.ScaleY)
	}
	if math.Abs(float64(action.DelayMS-150)) > 0.001 {
		t.Fatalf("delay = %f", action.DelayMS)
	}
}

func TestPlayerPaletteResourceCandidates(t *testing.T) {
	body := PlayerBodyPaletteResourceCandidates(0, 1, 3, "pal")
	if len(body) != 1 || body[0] != "data\\palette\\몸\\초보자_남_3.pal" {
		t.Fatalf("body palette candidates = %#v", body)
	}
	head := PlayerHeadPaletteResourceCandidates(0, 2, 0, 4, "pal")
	if len(head) != 1 || head[0] != "data\\palette\\머리\\머리2_여_4.pal" {
		t.Fatalf("head palette candidates = %#v", head)
	}
}

func writeTestU16(buf *bytes.Buffer, value uint16) {
	var tmp [2]byte
	binary.LittleEndian.PutUint16(tmp[:], value)
	buf.Write(tmp[:])
}

func writeTestU32(buf *bytes.Buffer, value uint32) {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], value)
	buf.Write(tmp[:])
}

func writeTestI32(buf *bytes.Buffer, value int32) {
	writeTestU32(buf, uint32(value))
}

func writeTestF32(buf *bytes.Buffer, value float32) {
	writeTestU32(buf, math.Float32bits(value))
}
