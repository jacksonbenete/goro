package network

import (
	"encoding/binary"
	"testing"
)

func TestParseParameterChange(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint16(data[0:2], 0x00B0)
	binary.LittleEndian.PutUint16(data[2:4], StatusHP)
	binary.LittleEndian.PutUint32(data[4:8], 1234)

	change, ok, err := ParseParameterChange(Packet{ID: 0x00B0, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("parameter change not parsed")
	}
	if change.VarID != StatusHP || change.Value != 1234 {
		t.Fatalf("change = %+v", change)
	}
}

func TestParseLongParameterChange(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint16(data[0:2], 0x00B1)
	binary.LittleEndian.PutUint16(data[2:4], StatusMaxHP)
	binary.LittleEndian.PutUint32(data[4:8], 123456)

	change, ok, err := ParseParameterChange(Packet{ID: 0x00B1, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("long parameter change not parsed")
	}
	if change.VarID != StatusMaxHP || change.Value != 123456 {
		t.Fatalf("change = %+v", change)
	}
}
