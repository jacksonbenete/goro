package network

import (
	"encoding/binary"
	"testing"
)

func TestBuildEmotionPacket(t *testing.T) {
	packet := BuildEmotionPacket(4)
	if got := ID(packet); got != PacketCZReqEmotion {
		t.Fatalf("opcode = 0x%04X, want 0x%04X", got, PacketCZReqEmotion)
	}
	if len(packet) != 3 {
		t.Fatalf("len = %d, want 3", len(packet))
	}
	if packet[2] != 4 {
		t.Fatalf("emotion type = %d, want 4", packet[2])
	}
}

func TestParseEmotionNotify(t *testing.T) {
	data := make([]byte, 7)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCEmotion)
	binary.LittleEndian.PutUint32(data[2:6], 2000000)
	data[6] = 18
	emotion, ok, err := ParseEmotionNotify(Packet{ID: PacketZCEmotion, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("emotion packet not recognized")
	}
	if emotion.GID != 2000000 || emotion.Type != 18 {
		t.Fatalf("emotion = %+v", emotion)
	}
}
