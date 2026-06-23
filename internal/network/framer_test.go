package network

import "testing"

func TestFramerFixedPacket(t *testing.T) {
	framer := NewFramer(LengthTable{0x0073: 3})
	packets, err := framer.Push([]byte{0x73, 0x00, 0x01})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x0073 || len(packets[0].Data) != 3 {
		t.Fatalf("unexpected packet: %+v", packets[0])
	}
}

func TestFramerVariablePacket(t *testing.T) {
	framer := NewFramer(LengthTable{0x0069: -1})
	packets, err := framer.Push([]byte{0x69, 0x00, 0x06, 0x00, 0xaa, 0xbb})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x0069 || len(packets[0].Data) != 6 {
		t.Fatalf("unexpected packet: %+v", packets[0])
	}
}

func TestFramerWaitsForFullPacket(t *testing.T) {
	framer := NewFramer(LengthTable{0x0073: 3})
	packets, err := framer.Push([]byte{0x73, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 0 {
		t.Fatalf("packets = %d", len(packets))
	}
	packets, err = framer.Push([]byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 {
		t.Fatalf("packets = %d", len(packets))
	}
}

func TestPacketLengths2008FramesNotifyEffectDirect(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := []byte{
		0xf3, 0x01, 0x01, 0x13, 0xa1, 0x8e, 0x06, 0x65, 0x00, 0x00,
		0xb6, 0x00, 0x44, 0x33, 0x22, 0x11,
	}
	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x01F3 || len(packets[0].Data) != 10 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesVariable01F1(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	packets, err := framer.Push([]byte{
		0xf1, 0x01, 0x05, 0x00, 0xaa,
		0xb6, 0x00, 0x44, 0x33, 0x22, 0x11,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x01F1 || len(packets[0].Data) != 5 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesActionNotify2(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := make([]byte, 0, 39)
	data = append(data, 0xe1, 0x02)
	data = append(data, make([]byte, 31)...)
	data = append(data, 0xb6, 0x00, 0x44, 0x33, 0x22, 0x11)

	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x02E1 || len(packets[0].Data) != 33 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesActionPosition(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := make([]byte, 0, 29)
	data = append(data, 0x8b, 0x00)
	data = append(data, make([]byte, 21)...)
	data = append(data, 0xb6, 0x00, 0x44, 0x33, 0x22, 0x11)

	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x008B || len(packets[0].Data) != 23 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesItemFallEntry(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := make([]byte, 0, 23)
	data = append(data, 0x9e, 0x00)
	data = append(data, make([]byte, 15)...)
	data = append(data, 0xb6, 0x00, 0x44, 0x33, 0x22, 0x11)

	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x009E || len(packets[0].Data) != 17 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesPartyHPUpdate(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := []byte{
		0x06, 0x01,
		0x01, 0x80, 0x00, 0x00,
		0x58, 0x00,
		0xb0, 0x00,
		0xb6, 0x00, 0x44, 0x33, 0x22, 0x11,
	}
	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x0106 || len(packets[0].Data) != 10 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestFramerResyncsAfterUnknownBytes(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	packets, err := framer.Push([]byte{
		0x8e, 0x06, 0x06, 0x80,
		0xb6, 0x00, 0x44, 0x33, 0x22, 0x11,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x00B6 || len(packets[0].Data) != 6 {
		t.Fatalf("packet = %s", packets[0])
	}
}

func TestPacketLengths2008FramesPartyHPUpdateR2(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := []byte{
		0x0e, 0x08,
		0x01, 0x80, 0x00, 0x00,
		0x58, 0x00, 0x00, 0x00,
		0xb0, 0x00, 0x00, 0x00,
		0xb6, 0x00, 0x44, 0x33, 0x22, 0x11,
	}
	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x080E || len(packets[0].Data) != 14 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesVariableChatRoomUpdate(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := []byte{
		0xdf, 0x00, 0x0c, 0x00,
		0x44, 0x33, 0x22, 0x11,
		0xb4, 0x8e, 0x06, 0x8e,
		0xb6, 0x00, 0x44, 0x33, 0x22, 0x11,
	}
	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x00DF || len(packets[0].Data) != 12 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}
