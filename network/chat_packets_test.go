package network

import (
	"encoding/binary"
	"testing"
)

func TestBuildGlobalChatPacketFor2008ClientDate(t *testing.T) {
	packet := BuildGlobalChatPacketForClientDate("Kivutar", "hello", 20080910)
	if got := ID(packet); got != PacketCZRequestChat {
		t.Fatalf("opcode = 0x%04X, want 0x%04X", got, PacketCZRequestChat)
	}
	if got := binary.LittleEndian.Uint16(packet[2:4]); int(got) != len(packet) {
		t.Fatalf("length = %d, want %d", got, len(packet))
	}
	if got := string(packet[4 : len(packet)-1]); got != "Kivutar : hello" {
		t.Fatalf("payload = %q", got)
	}
	if packet[len(packet)-1] != 0 {
		t.Fatalf("packet is not nul terminated")
	}
}

func TestBuildWhisperPacket(t *testing.T) {
	packet := BuildWhisperPacket("Rekka", "hello")
	if got := ID(packet); got != PacketCZWhisper {
		t.Fatalf("opcode = 0x%04X, want 0x%04X", got, PacketCZWhisper)
	}
	if got := binary.LittleEndian.Uint16(packet[2:4]); int(got) != len(packet) {
		t.Fatalf("length = %d, want %d", got, len(packet))
	}
	if got := string(packet[4:9]); got != "Rekka" {
		t.Fatalf("receiver prefix = %q", got)
	}
	if packet[9] != 0 {
		t.Fatalf("receiver is not nul terminated")
	}
	if got := string(packet[28 : len(packet)-1]); got != "hello" {
		t.Fatalf("message = %q", got)
	}
	if packet[len(packet)-1] != 0 {
		t.Fatalf("packet is not nul terminated")
	}
}

func TestParseNotifyChat(t *testing.T) {
	packet := Packet{ID: PacketZCNotifyChat, Data: []byte{0x8d, 0x00, 0x11, 0x00, 0x44, 0x33, 0x22, 0x11, 'h', 'i', 0}}
	chat, ok, err := ParseChatMessage(packet)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("chat packet not recognized")
	}
	if chat.GID != 0x11223344 || chat.Text != "hi" {
		t.Fatalf("unexpected chat: %+v", chat)
	}
}

func TestParseWhisperMessage(t *testing.T) {
	data := make([]byte, 4+24+6)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCWhisper)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	copy(data[4:28], []byte("Rekka"))
	copy(data[28:], []byte("hello\x00"))
	whisper, ok, err := ParseWhisperMessage(Packet{ID: PacketZCWhisper, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("whisper packet not recognized")
	}
	if whisper.Sender != "Rekka" || whisper.Message != "hello" {
		t.Fatalf("whisper = %+v", whisper)
	}
}

func TestParseWhisperAck(t *testing.T) {
	ack, ok, err := ParseWhisperAck(Packet{ID: PacketZCAckWhisper, Data: []byte{0x98, 0x00, 0x01}})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("whisper ack not recognized")
	}
	if ack.Result != 1 {
		t.Fatalf("result = %d, want 1", ack.Result)
	}
}

func TestParseBroadcastChat(t *testing.T) {
	packet := Packet{ID: PacketZCBroadcast, Data: []byte{0x9a, 0x00, 0x0c, 0x00, 's', 'e', 'r', 'v', 'e', 'r', 0}}
	chat, ok, err := ParseChatMessage(packet)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("broadcast packet not recognized")
	}
	if chat.Text != "server" {
		t.Fatalf("text = %q", chat.Text)
	}
}

func TestParseMsgStringID(t *testing.T) {
	packet := Packet{ID: PacketZCMsg, Data: []byte{0x91, 0x02, 0x2a, 0x00}}
	chat, ok, err := ParseChatMessage(packet)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("message packet not recognized")
	}
	if chat.MessageID != 42 || chat.Text != "" {
		t.Fatalf("chat = %+v", chat)
	}
}
