package res

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestParseGAT(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("GRAT")
	buf.WriteByte(1)
	buf.WriteByte(3)
	writeU32(&buf, 2)
	writeU32(&buf, 1)
	writeGATCell(&buf, [4]float32{10, 20, 30, 40}, 0)
	writeGATCell(&buf, [4]float32{5, 5, 5, 5}, 3)

	gat, err := ParseGAT(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if gat.VersionMajor != 1 || gat.VersionMinor != 3 || gat.Width != 2 || gat.Height != 1 {
		t.Fatalf("unexpected header: %+v", gat)
	}
	first, ok := gat.Cell(0, 0)
	if !ok {
		t.Fatal("missing first cell")
	}
	if first.Type&GATTypeWalkable == 0 || first.Heights[0] != -2 {
		t.Fatalf("unexpected first cell: %+v", first)
	}
	second, ok := gat.Cell(1, 0)
	if !ok {
		t.Fatal("missing second cell")
	}
	if second.Type&GATTypeWater == 0 {
		t.Fatalf("expected water cell: %+v", second)
	}
}

func TestSetCellRawTypeUpdatesWalkability(t *testing.T) {
	gat := &GAT{
		Width:  1,
		Height: 1,
		Cells:  []GATCell{{RawType: 0, Type: GATCellType(0)}},
	}
	if !gat.Walkable(0, 0) {
		t.Fatal("initial cell should be walkable")
	}
	if !gat.SetCellRawType(0, 0, 5) {
		t.Fatal("cell update failed")
	}
	if gat.Walkable(0, 0) {
		t.Fatalf("raw type 5 should not be walkable: %+v", gat.Cells[0])
	}
	if got := gat.Cells[0].Type; got != GATTypeSnipable {
		t.Fatalf("cell type = %d, want snipable only", got)
	}
}

func writeGATCell(buf *bytes.Buffer, heights [4]float32, rawType uint32) {
	for _, h := range heights {
		var tmp [4]byte
		binary.LittleEndian.PutUint32(tmp[:], math.Float32bits(h))
		buf.Write(tmp[:])
	}
	writeU32(buf, rawType)
}
