package network

import (
	"encoding/binary"
	"testing"
)

func TestParseMapInfoNotify(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCNotifyMapInfo)
	binary.LittleEndian.PutUint16(data[2:4], 1)

	notify, ok, err := ParseMapInfoNotify(Packet{ID: PacketZCNotifyMapInfo, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || notify.Result != 1 {
		t.Fatalf("map info notification = %+v, ok=%v", notify, ok)
	}
}

func TestParseMapInfoNotifyRejectsShortPacket(t *testing.T) {
	if _, ok, err := ParseMapInfoNotify(Packet{ID: PacketZCNotifyMapInfo, Data: []byte{0x89, 0x01}}); !ok || err == nil {
		t.Fatalf("short packet = ok %v, err %v", ok, err)
	}
}
