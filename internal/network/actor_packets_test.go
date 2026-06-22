package network

import (
	"encoding/binary"
	"testing"
)

func TestParseActorStandEntry2(t *testing.T) {
	data := make([]byte, 54)
	binary.LittleEndian.PutUint16(data[0:2], 0x01D8)
	binary.LittleEndian.PutUint32(data[2:6], 2000001)
	binary.LittleEndian.PutUint16(data[14:16], 1002)
	binary.LittleEndian.PutUint16(data[16:18], 7)
	data[45] = 1
	data[46], data[47], data[48] = packPosition(102, 134, 3)

	entry, ok, err := ParseActorEntry(Packet{ID: 0x01D8, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if !entry.Appearance {
		t.Fatal("stand entry should include appearance")
	}
	if entry.ID != 2000001 || entry.Job != 1002 || entry.Head != 7 || entry.Sex != 1 || entry.X != 102 || entry.Y != 134 || entry.Dir != 3 {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}

func TestParseActorStandEntryLegacy(t *testing.T) {
	data := make([]byte, 55)
	binary.LittleEndian.PutUint16(data[0:2], 0x0078)
	data[2] = 5
	binary.LittleEndian.PutUint32(data[3:7], 2000003)
	binary.LittleEndian.PutUint16(data[15:17], 1011)
	binary.LittleEndian.PutUint16(data[17:19], 2)
	data[46] = 0
	data[47], data[48], data[49] = packPosition(44, 55, 6)

	entry, ok, err := ParseActorEntry(Packet{ID: 0x0078, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("not parsed")
	}
	if entry.ID != 2000003 || entry.Job != 1011 || entry.ObjectType != 5 || entry.X != 44 || entry.Y != 55 || entry.Dir != 6 {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}

func TestParseActorMoveEntry2(t *testing.T) {
	data := make([]byte, 60)
	binary.LittleEndian.PutUint16(data[0:2], 0x01DA)
	binary.LittleEndian.PutUint32(data[2:6], 2000002)
	binary.LittleEndian.PutUint16(data[14:16], 1002)
	data[49] = 0
	data[50], data[51], data[52], data[53], data[54], data[55] = packMovePosition(10, 20, 30, 40)

	entry, ok, err := ParseActorEntry(Packet{ID: 0x01DA, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !entry.Moving {
		t.Fatalf("move not parsed: ok=%v entry=%+v", ok, entry)
	}
	if !entry.Appearance {
		t.Fatal("move entry should include appearance")
	}
	if entry.FromX != 10 || entry.FromY != 20 || entry.ToX != 30 || entry.ToY != 40 || entry.X != 30 || entry.Y != 40 {
		t.Fatalf("unexpected move entry: %+v", entry)
	}
}

func TestParseActorMoveUpdate(t *testing.T) {
	data := make([]byte, 16)
	binary.LittleEndian.PutUint16(data[0:2], 0x0086)
	binary.LittleEndian.PutUint32(data[2:6], 2000004)
	data[6], data[7], data[8], data[9], data[10], data[11] = packMovePosition(11, 21, 31, 41)
	binary.LittleEndian.PutUint32(data[12:16], 123456)

	entry, ok, err := ParseActorEntry(Packet{ID: 0x0086, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !entry.Moving {
		t.Fatalf("move update not parsed: ok=%v entry=%+v", ok, entry)
	}
	if entry.Appearance {
		t.Fatal("move-only update should not include appearance")
	}
	if entry.ID != 2000004 || entry.FromX != 11 || entry.FromY != 21 || entry.ToX != 31 || entry.ToY != 41 || entry.X != 31 || entry.Y != 41 {
		t.Fatalf("unexpected move update: %+v", entry)
	}
}

func TestParseActorVanish(t *testing.T) {
	data := make([]byte, 7)
	binary.LittleEndian.PutUint16(data[0:2], 0x0080)
	binary.LittleEndian.PutUint32(data[2:6], 2000005)
	data[6] = 1

	vanish, ok, err := ParseActorVanish(Packet{ID: 0x0080, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("vanish not parsed")
	}
	if vanish.ID != 2000005 || vanish.Reason != 1 {
		t.Fatalf("unexpected vanish: %+v", vanish)
	}
}

func TestParseActorLookChangeModern(t *testing.T) {
	data := make([]byte, 11)
	binary.LittleEndian.PutUint16(data[0:2], 0x01D7)
	binary.LittleEndian.PutUint32(data[2:6], 2000006)
	data[6] = 2
	binary.LittleEndian.PutUint32(data[7:11], uint32(2101)<<16|1201)

	look, ok, err := ParseActorLookChange(Packet{ID: 0x01D7, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("look change not parsed")
	}
	if look.ID != 2000006 || look.Type != 2 || look.Value != uint32(2101)<<16|1201 {
		t.Fatalf("unexpected look change: %+v", look)
	}
}

func TestParseActorLookChangeLegacy(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint16(data[0:2], 0x00C3)
	binary.LittleEndian.PutUint32(data[2:6], 2000006)
	data[6] = 4
	data[7] = 7

	look, ok, err := ParseActorLookChange(Packet{ID: 0x00C3, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || look.ID != 2000006 || look.Type != 4 || look.Value != 7 {
		t.Fatalf("unexpected legacy look change: ok=%v look=%+v", ok, look)
	}
}

func TestParseSelfMoveAck(t *testing.T) {
	data := make([]byte, 12)
	binary.LittleEndian.PutUint16(data[0:2], 0x0087)
	binary.LittleEndian.PutUint32(data[2:6], 123456)
	data[6], data[7], data[8], data[9], data[10], data[11] = packMovePosition(120, 121, 122, 123)

	ack, ok, err := ParseSelfMoveAck(Packet{ID: 0x0087, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("self move ack not parsed")
	}
	if ack.ServerTick != 123456 || ack.FromX != 120 || ack.FromY != 121 || ack.ToX != 122 || ack.ToY != 123 {
		t.Fatalf("unexpected self move ack: %+v", ack)
	}
}

func TestParseActorSetPosition(t *testing.T) {
	data := make([]byte, 10)
	binary.LittleEndian.PutUint16(data[0:2], 0x0088)
	binary.LittleEndian.PutUint32(data[2:6], 2000006)
	x := int16(-12)
	binary.LittleEndian.PutUint16(data[8:10], uint16(int16(34)))
	binary.LittleEndian.PutUint16(data[6:8], uint16(x))

	position, ok, err := ParseActorSetPosition(Packet{ID: 0x0088, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("actor set position not parsed")
	}
	if position.ID != 2000006 || position.X != -12 || position.Y != 34 {
		t.Fatalf("unexpected actor position: %+v", position)
	}
}

func packMovePosition(fromX, fromY, toX, toY int) (byte, byte, byte, byte, byte, byte) {
	return byte(fromX >> 2),
		byte(((fromX & 0x03) << 6) | ((fromY >> 4) & 0x3f)),
		byte(((fromY & 0x0f) << 4) | ((toX >> 6) & 0x0f)),
		byte(((toX & 0x3f) << 2) | ((toY >> 8) & 0x03)),
		byte(toY),
		0
}
