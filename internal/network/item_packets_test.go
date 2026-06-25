package network

import (
	"encoding/binary"
	"testing"
)

func TestParseFloorItemEntryExistingItem(t *testing.T) {
	data := make([]byte, 17)
	binary.LittleEndian.PutUint16(data[0:2], 0x009D)
	binary.LittleEndian.PutUint32(data[2:6], 1001)
	binary.LittleEndian.PutUint16(data[6:8], 909)
	data[8] = 1
	binary.LittleEndian.PutUint16(data[9:11], 150)
	binary.LittleEndian.PutUint16(data[11:13], 160)
	binary.LittleEndian.PutUint16(data[13:15], 3)
	data[15] = 4
	data[16] = 8

	item, ok, err := ParseFloorItemEntry(Packet{ID: 0x009D, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if item.ID != 1001 || item.ItemID != 909 || !item.Identified || item.X != 150 || item.Y != 160 || item.Amount != 3 || item.SubX != 4 || item.SubY != 8 || item.Falling {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestParseFloorItemEntryFallingItem(t *testing.T) {
	data := make([]byte, 17)
	binary.LittleEndian.PutUint16(data[0:2], 0x009E)
	binary.LittleEndian.PutUint32(data[2:6], 2002)
	binary.LittleEndian.PutUint16(data[6:8], 512)
	data[8] = 0
	binary.LittleEndian.PutUint16(data[9:11], 30)
	binary.LittleEndian.PutUint16(data[11:13], 40)
	data[13] = 5
	data[14] = 7
	binary.LittleEndian.PutUint16(data[15:17], 9)

	item, ok, err := ParseFloorItemEntry(Packet{ID: 0x009E, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if item.ID != 2002 || item.ItemID != 512 || item.Identified || item.X != 30 || item.Y != 40 || item.Amount != 9 || item.SubX != 5 || item.SubY != 7 || !item.Falling {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestParseFloorItemDisappear(t *testing.T) {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], 0x00A1)
	binary.LittleEndian.PutUint32(data[2:6], 3003)

	disappear, ok, err := ParseFloorItemDisappear(Packet{ID: 0x00A1, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || disappear.ID != 3003 {
		t.Fatalf("unexpected disappear: ok=%v value=%+v", ok, disappear)
	}
}

func TestBuildItemPickupPacket(t *testing.T) {
	packet := BuildItemPickupPacket(0x11223344)
	if len(packet) != 6 {
		t.Fatalf("len = %d, want 6", len(packet))
	}
	if got := ID(packet); got != 0x009F {
		t.Fatalf("opcode = 0x%04X, want 0x009F", got)
	}
	if got := binary.LittleEndian.Uint32(packet[2:6]); got != 0x11223344 {
		t.Fatalf("gid = 0x%08X", got)
	}
}

func TestBuildItemPickupPacketForClientDate20080910(t *testing.T) {
	packet := BuildItemPickupPacketForClientDate(0x11223344, 20080910)
	if len(packet) != 8 {
		t.Fatalf("len = %d, want 8", len(packet))
	}
	if got := ID(packet); got != 0x00F5 {
		t.Fatalf("opcode = 0x%04X, want 0x00F5", got)
	}
	if got := binary.LittleEndian.Uint32(packet[4:8]); got != 0x11223344 {
		t.Fatalf("gid = 0x%08X", got)
	}
	if padding := binary.LittleEndian.Uint16(packet[2:4]); padding != 0 {
		t.Fatalf("padding = 0x%04X, want 0", padding)
	}
}

func TestBuildItemPickupPacketForClientDateShuffledBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		clientDate int
		opcode     uint16
		length     int
		offset     int
	}{
		{name: "legacy", clientDate: 20040712, opcode: 0x009F, length: 6, offset: 2},
		{name: "20040713", clientDate: 20040713, opcode: 0x009F, length: 10, offset: 6},
		{name: "20040726", clientDate: 20040726, opcode: 0x0094, length: 10, offset: 6},
		{name: "20070212", clientDate: 20070212, opcode: 0x00F5, length: 8, offset: 4},
		{name: "20101124", clientDate: 20101124, opcode: 0x0362, length: 6, offset: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packet := BuildItemPickupPacketForClientDate(0x11223344, tc.clientDate)
			if len(packet) != tc.length {
				t.Fatalf("len = %d, want %d", len(packet), tc.length)
			}
			if got := ID(packet); got != tc.opcode {
				t.Fatalf("opcode = 0x%04X, want 0x%04X", got, tc.opcode)
			}
			if got := binary.LittleEndian.Uint32(packet[tc.offset : tc.offset+4]); got != 0x11223344 {
				t.Fatalf("gid = 0x%08X", got)
			}
		})
	}
}
