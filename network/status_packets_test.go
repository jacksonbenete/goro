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

func TestParseLongParameterChangeKeepsUnsignedValue(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint16(data[0:2], 0x00B1)
	binary.LittleEndian.PutUint16(data[2:4], StatusBaseExp)
	binary.LittleEndian.PutUint32(data[4:8], 0xFFFFFFFF)

	change, ok, err := ParseParameterChange(Packet{ID: 0x00B1, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("long parameter change not parsed")
	}
	if change.VarID != StatusBaseExp || change.Value != 4294967295 {
		t.Fatalf("change = %+v", change)
	}
}

func TestParseCompactStatusChange(t *testing.T) {
	data := make([]byte, 5)
	binary.LittleEndian.PutUint16(data[0:2], 0x00BE)
	binary.LittleEndian.PutUint16(data[2:4], StatusStr)
	data[4] = 31

	change, ok, err := ParseParameterChange(Packet{ID: 0x00BE, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("compact status change not parsed")
	}
	if change.VarID != StatusStr || change.Value != 31 {
		t.Fatalf("change = %+v", change)
	}
}

func TestParseRecovery(t *testing.T) {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], 0x013D)
	binary.LittleEndian.PutUint16(data[2:4], StatusHP)
	binary.LittleEndian.PutUint16(data[4:6], 42)

	recovery, ok, err := ParseRecovery(Packet{ID: 0x013D, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("recovery not parsed")
	}
	if recovery.StatusID != StatusHP || recovery.Amount != 42 {
		t.Fatalf("recovery = %+v", recovery)
	}
}

func TestParseStatusSnapshot(t *testing.T) {
	data := make([]byte, 44)
	binary.LittleEndian.PutUint16(data[0:2], 0x00BD)
	binary.LittleEndian.PutUint16(data[2:4], 7)
	data[4] = 11
	data[5] = 13
	data[6] = 12
	data[7] = 12
	data[8] = 13
	data[9] = 14
	data[10] = 14
	data[11] = 16
	data[12] = 15
	data[13] = 15
	data[14] = 16
	data[15] = 17
	binary.LittleEndian.PutUint16(data[16:18], 42)
	binary.LittleEndian.PutUint16(data[18:20], 3)
	binary.LittleEndian.PutUint16(data[20:22], 30)
	binary.LittleEndian.PutUint16(data[22:24], 20)
	binary.LittleEndian.PutUint16(data[40:42], 640)

	snapshot, ok, err := ParseStatusSnapshot(Packet{ID: 0x00BD, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("status snapshot not parsed")
	}
	if snapshot.Points != 7 || snapshot.Str != 11 || snapshot.StrCost != 13 || snapshot.Attack != 42 || snapshot.ASPD != 640 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
}

func TestParseStatusChangeAck(t *testing.T) {
	data := make([]byte, 6)
	binary.LittleEndian.PutUint16(data[0:2], 0x00BC)
	binary.LittleEndian.PutUint16(data[2:4], StatusDex)
	data[4] = 1
	data[5] = 22

	ack, ok, err := ParseStatusChangeAck(Packet{ID: 0x00BC, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("status ack not parsed")
	}
	if ack.StatusID != StatusDex || !ack.Success || ack.Value != 22 {
		t.Fatalf("ack = %+v", ack)
	}
}

func TestBuildStatusIncreasePacket(t *testing.T) {
	packet := BuildStatusIncreasePacket(StatusStr)
	if got := ID(packet); got != 0x00BB {
		t.Fatalf("opcode = 0x%04X", got)
	}
	if len(packet) != 5 {
		t.Fatalf("len = %d", len(packet))
	}
	if got := binary.LittleEndian.Uint16(packet[2:4]); got != StatusStr {
		t.Fatalf("status id = %d", got)
	}
	if packet[4] != 1 {
		t.Fatalf("amount = %d", packet[4])
	}
}
