package network

import (
	"encoding/binary"
	"testing"
)

func TestPetCapturePackets(t *testing.T) {
	_, ok, err := ParsePetCaptureStart(Packet{ID: PacketZCStartCapture, Data: []byte{0x9E, 0x01}})
	if !ok || err != nil {
		t.Fatalf("ParsePetCaptureStart ok=%t err=%v", ok, err)
	}

	result, ok, err := ParsePetCaptureResult(Packet{ID: PacketZCTryCaptureMonster, Data: []byte{0xA0, 0x01, 1}})
	if !ok || err != nil || !result.Success {
		t.Fatalf("ParsePetCaptureResult = %+v ok=%t err=%v", result, ok, err)
	}

	packet := BuildTryCaptureMonsterPacket(0x11223344)
	if got := binary.LittleEndian.Uint16(packet[0:2]); got != PacketCZTryCaptureMonster {
		t.Fatalf("opcode = 0x%04X", got)
	}
	if got := binary.LittleEndian.Uint32(packet[2:6]); got != 0x11223344 {
		t.Fatalf("target = 0x%08X", got)
	}
}

func TestPetEggPackets(t *testing.T) {
	listData := []byte{0xA6, 0x01, 0x08, 0x00, 7, 0, 9, 0}
	list, ok, err := ParsePetEggList(Packet{ID: PacketZCPetEggList, Data: listData})
	if !ok || err != nil || len(list.Indexes) != 2 || list.Indexes[0] != 7 || list.Indexes[1] != 9 {
		t.Fatalf("ParsePetEggList = %+v ok=%t err=%v", list, ok, err)
	}

	packet := BuildSelectPetEggPacket(9)
	if got := binary.LittleEndian.Uint16(packet[0:2]); got != PacketCZSelectPetEgg {
		t.Fatalf("opcode = 0x%04X", got)
	}
	if got := binary.LittleEndian.Uint16(packet[2:4]); got != 9 {
		t.Fatalf("index = %d", got)
	}
}
