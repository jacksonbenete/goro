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

func TestBuildWhisperIgnorePackets(t *testing.T) {
	packet := BuildWhisperIgnorePacket("Rekka", false)
	if got := ID(packet); got != PacketCZWhisperIgnore {
		t.Fatalf("opcode = 0x%04X, want 0x%04X", got, PacketCZWhisperIgnore)
	}
	if len(packet) != 27 {
		t.Fatalf("len = %d, want 27", len(packet))
	}
	if got := string(packet[2:7]); got != "Rekka" {
		t.Fatalf("name prefix = %q", got)
	}
	if packet[7] != 0 {
		t.Fatalf("name is not nul terminated")
	}
	if packet[26] != 0 {
		t.Fatalf("type = %d, want deny 0", packet[26])
	}

	packet = BuildWhisperIgnorePacket("123456789012345678901234", false)
	if got := string(packet[2:25]); got != "12345678901234567890123" || packet[25] != 0 {
		t.Fatalf("long name field = %x", packet[2:26])
	}

	packet = BuildWhisperIgnorePacket("Rekka", true)
	if packet[26] != 1 {
		t.Fatalf("type = %d, want allow 1", packet[26])
	}

	packet = BuildWhisperIgnoreAllPacket(false)
	if got := ID(packet); got != PacketCZWhisperIgnoreAll {
		t.Fatalf("opcode = 0x%04X, want 0x%04X", got, PacketCZWhisperIgnoreAll)
	}
	if len(packet) != 3 || packet[2] != 0 {
		t.Fatalf("ignore all packet = %x, want deny", packet)
	}

	packet = BuildWhisperIgnoreAllPacket(true)
	if len(packet) != 3 || packet[2] != 1 {
		t.Fatalf("ignore all packet = %x, want allow", packet)
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

func TestParseWhisperIgnoreAck(t *testing.T) {
	ack, ok, err := ParseWhisperIgnoreAck(Packet{ID: PacketZCWhisperIgnoreAck, Data: []byte{0xd1, 0x00, 0x00, 0x02}})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("whisper ignore ack not recognized")
	}
	if ack.TargetAll || ack.Allow || ack.Result != 2 {
		t.Fatalf("ack = %+v", ack)
	}

	ack, ok, err = ParseWhisperIgnoreAck(Packet{ID: PacketZCWhisperAllAck, Data: []byte{0xd2, 0x00, 0x01, 0x00}})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("whisper all ack not recognized")
	}
	if !ack.TargetAll || !ack.Allow || ack.Result != 0 {
		t.Fatalf("ack = %+v", ack)
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
