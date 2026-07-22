package network

import (
	"encoding/binary"
	"testing"
)

func TestParseMinimapCompass(t *testing.T) {
	data := make([]byte, 23)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCCompass)
	binary.LittleEndian.PutUint32(data[2:6], 0x11223344)
	binary.LittleEndian.PutUint32(data[6:10], uint32(1))
	binary.LittleEndian.PutUint32(data[10:14], uint32(120))
	binary.LittleEndian.PutUint32(data[14:18], uint32(80))
	data[18] = 7
	binary.LittleEndian.PutUint32(data[19:23], 0x00AABBCC)

	got, ok, err := ParseMinimapCompass(Packet{ID: PacketZCCompass, Data: data})
	if err != nil {
		t.Fatalf("ParseMinimapCompass error: %v", err)
	}
	if !ok {
		t.Fatal("ParseMinimapCompass did not recognize packet")
	}
	if got.NPCID != 0x11223344 || got.Type != 1 || got.X != 120 || got.Y != 80 || got.ID != 7 || got.Color != 0x00AABBCC {
		t.Fatalf("compass = %+v, want decoded packet fields", got)
	}
}

func TestParseMinimapCompassSignedFields(t *testing.T) {
	data := make([]byte, 23)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCCompass)
	binary.LittleEndian.PutUint32(data[6:10], 0xffffffff)
	binary.LittleEndian.PutUint32(data[10:14], 0xffffffec)
	binary.LittleEndian.PutUint32(data[14:18], 0xffffffe2)

	got, ok, err := ParseMinimapCompass(Packet{ID: PacketZCCompass, Data: data})
	if err != nil {
		t.Fatalf("ParseMinimapCompass error: %v", err)
	}
	if !ok {
		t.Fatal("ParseMinimapCompass did not recognize packet")
	}
	if got.Type != -1 || got.X != -20 || got.Y != -30 {
		t.Fatalf("signed fields = type:%d x:%d y:%d, want -1,-20,-30", got.Type, got.X, got.Y)
	}
}

func TestParseMinimapCompassShortPacket(t *testing.T) {
	_, ok, err := ParseMinimapCompass(Packet{ID: PacketZCCompass, Data: make([]byte, 22)})
	if !ok {
		t.Fatal("short ZC_COMPASS should be recognized")
	}
	if err == nil {
		t.Fatal("short ZC_COMPASS should fail")
	}
}

func TestParseMinimapCompassIgnoresOtherPackets(t *testing.T) {
	_, ok, err := ParseMinimapCompass(Packet{ID: 0x0080, Data: make([]byte, 7)})
	if err != nil {
		t.Fatalf("other packet error: %v", err)
	}
	if ok {
		t.Fatal("other packet should not be recognized")
	}
}
