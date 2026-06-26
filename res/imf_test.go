package res

import (
	"bytes"
	"testing"
)

func TestParseIMF(t *testing.T) {
	var buf bytes.Buffer
	writeI32(&buf, 1)
	writeI32(&buf, 99)
	writeI32(&buf, 2)

	writeI32(&buf, 1)
	writeI32(&buf, 2)
	writeI32(&buf, 4)
	writeI32(&buf, 10)
	writeI32(&buf, 20)
	writeI32(&buf, 5)
	writeI32(&buf, 11)
	writeI32(&buf, 21)

	writeI32(&buf, 1)
	writeI32(&buf, 2)
	writeI32(&buf, 0)
	writeI32(&buf, 30)
	writeI32(&buf, 40)
	writeI32(&buf, 1)
	writeI32(&buf, 31)
	writeI32(&buf, 41)

	imf, err := ParseIMF(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if imf.Version != 1 || imf.Checksum != 99 || len(imf.Layers) != 2 {
		t.Fatalf("unexpected imf header: %+v", imf)
	}
	if layer := imf.LayerForPriority(5, 0, 1); layer != 0 {
		t.Fatalf("layer for priority = %d, want 0", layer)
	}
	if layer := imf.LayerForPriority(1, 0, 1); layer != 1 {
		t.Fatalf("layer for priority = %d, want 1", layer)
	}
	x, y := imf.Point(1, 0, 1)
	if x != 31 || y != 41 {
		t.Fatalf("point = (%d,%d), want (31,41)", x, y)
	}
}

func TestParseIMFRejectsInvalidLayerCount(t *testing.T) {
	var buf bytes.Buffer
	writeI32(&buf, 1)
	writeI32(&buf, 0)
	writeI32(&buf, 16)
	if _, err := ParseIMF(buf.Bytes()); err == nil {
		t.Fatal("expected error")
	}
}
