package network

import (
	"encoding/binary"
	"testing"
)

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

func TestFramerWaitsBeforeResyncingUnknownShortPrefix(t *testing.T) {
	framer := NewFramer(LengthTable{0x006B: -1})
	packets, err := framer.Push([]byte{0x81, 0x84, 0x1e, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 0 {
		t.Fatalf("packets before char list = %d", len(packets))
	}

	packets, err = framer.Push([]byte{0x6b, 0x00, 0x04, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 || packets[0].ID != 0x006B || len(packets[0].Data) != 4 {
		t.Fatalf("packets after char list = %+v", packets)
	}
}

func TestPacketLengths2008DoesNotContainAccountIDPreamble(t *testing.T) {
	if _, ok := PacketLengths2008()[0x8480]; ok {
		t.Fatal("char-server account id preamble must not be modeled as packet 0x8480")
	}
}

func TestPacketLengths2008DoesNotContainClientOnlyPartyPackets(t *testing.T) {
	lengths := PacketLengths2008()
	for _, id := range []uint16{
		PacketCZMakeGroup,
		PacketCZReqJoinGroup,
		PacketCZJoinGroup,
		PacketCZReqLeaveGroup,
		PacketCZChangeGroupExp,
		PacketCZReqExpelGroupMember,
	} {
		if _, ok := lengths[id]; ok {
			t.Fatalf("client-only party packet 0x%04X must not be in the receive length table", id)
		}
	}
}

func TestPacketLengths2008ResyncsThroughPartyPayloadToSkillList(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := []byte{
		0x00, 0x01,
		0x00, 0x01,
		0x0f, 0x01, 0x04, 0x00,
	}
	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x010F || len(packets[0].Data) != 4 {
		t.Fatalf("packet = %s", packets[0])
	}
}

func TestPacketLengths2008FramesPetPropertyBeforeSkillList(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := make([]byte, 35+4)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCPropertyPet)
	copy(data[2:26], []byte("Sakurai"))
	data[26] = 1
	binary.LittleEndian.PutUint16(data[27:29], 3)
	binary.LittleEndian.PutUint16(data[29:31], 28)
	binary.LittleEndian.PutUint16(data[31:33], 170)
	binary.LittleEndian.PutUint16(data[33:35], 10007)
	binary.LittleEndian.PutUint16(data[35:37], 0x010F)
	binary.LittleEndian.PutUint16(data[37:39], 4)

	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != PacketZCPropertyPet || len(packets[0].Data) != 35 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x010F || len(packets[1].Data) != 4 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesGuildPositionBeforePartyList(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	packets, err := framer.Push([]byte{
		0xeb, 0x01, 0x80, 0x84, 0x1e, 0x00, 0xf6, 0x00, 0x68, 0x00,
		0xfb, 0x00, 0x06, 0x00, 0xaa, 0xbb,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x01EB || len(packets[0].Data) != 10 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00FB || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesPartyMemberInfoBeforeSkillList(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := make([]byte, 6+81+4+4)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCGroupList)
	binary.LittleEndian.PutUint16(data[2:4], 6)
	binary.LittleEndian.PutUint16(data[6:8], PacketZCAddMemberToGroup2)
	binary.LittleEndian.PutUint32(data[8:12], 2000000)
	binary.LittleEndian.PutUint32(data[12:16], 1)
	binary.LittleEndian.PutUint16(data[16:18], 254)
	binary.LittleEndian.PutUint16(data[18:20], 138)
	data[20] = 1
	copy(data[21:28], []byte("Mandala"))
	copy(data[45:52], []byte("Kivutar"))
	copy(data[69:79], []byte("amatsu.gat"))
	binary.LittleEndian.PutUint16(data[87:89], 0x0199)
	binary.LittleEndian.PutUint16(data[91:93], 0x010F)
	binary.LittleEndian.PutUint16(data[93:95], 4)

	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 4 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != PacketZCGroupList || len(packets[0].Data) != 6 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != PacketZCAddMemberToGroup2 || len(packets[1].Data) != 81 {
		t.Fatalf("second packet = %s", packets[1])
	}
	member, ok, err := ParsePartyMemberJoin(packets[1])
	if !ok || err != nil {
		t.Fatalf("ParsePartyMemberJoin ok=%t err=%v member=%+v", ok, err, member)
	}
	if member.Name != "Kivutar" || member.GroupName != "Mandala" || member.MapName != "amatsu.gat" || member.X != 254 || member.Y != 138 {
		t.Fatalf("member = %+v", member)
	}
	if packets[2].ID != 0x0199 || len(packets[2].Data) != 4 {
		t.Fatalf("third packet = %s", packets[2])
	}
	if packets[3].ID != 0x010F || len(packets[3].Data) != 4 {
		t.Fatalf("fourth packet = %s", packets[3])
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

func TestPacketLengths2008FramesServerMessagePackets(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := []byte{
		0x91, 0x02, 0x2a, 0x00,
		0xe2, 0x07, 0x2b, 0x00, 0x05, 0x00, 0x00, 0x00,
		0xe6, 0x07, 0x1c, 0x00, 0x2c, 0x00, 0x00, 0x00,
		0xf6, 0x07, 0x44, 0x33, 0x22, 0x11, 0xd2, 0x04, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		0xc1, 0x02, 0x0f, 0x00, 0x44, 0x33, 0x22, 0x11, 0xb5, 0xff, 0xb5, 0x00, 'x', 'p', 0x00,
		0xcd, 0x09, 0x2d, 0x00, 0xff, 0xee, 0xdd, 0x00,
	}
	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 6 {
		t.Fatalf("packets = %d", len(packets))
	}
	for i, want := range []struct {
		id  uint16
		len int
	}{
		{0x0291, 4},
		{0x07E2, 8},
		{0x07E6, 8},
		{0x07F6, 14},
		{0x02C1, 15},
		{0x09CD, 8},
	} {
		if packets[i].ID != want.id || len(packets[i].Data) != want.len {
			t.Fatalf("packet %d = %s, want 0x%04X len=%d", i, packets[i], want.id, want.len)
		}
	}
}

func TestPacketLengths2008FramesCartNormalItemList(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	packets, err := framer.Push([]byte{
		0xe9, 0x02, 0x1a, 0x00,
		0x03, 0x00, 0x00, 0x02, 0x00, 0x01, 0x07, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
		0xb6, 0x00, 0x44, 0x33, 0x22, 0x11,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x02E9 || len(packets[0].Data) != 26 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesMakingArrowList(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	packets, err := framer.Push([]byte{
		0xad, 0x01, 0x08, 0x00,
		0x8d, 0x03,
		0x92, 0x03,
		0xb6, 0x00, 0x44, 0x33, 0x22, 0x11,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x01AD || len(packets[0].Data) != 8 {
		t.Fatalf("first packet = %s", packets[0])
	}
	if packets[1].ID != 0x00B6 || len(packets[1].Data) != 6 {
		t.Fatalf("second packet = %s", packets[1])
	}
}

func TestPacketLengths2008FramesCartDeltaPackets(t *testing.T) {
	framer := NewFramer(PacketLengths2008())
	data := make([]byte, 0, 39)
	data = append(data,
		0xc5, 0x01,
		0x03, 0x00,
		0x07, 0x00, 0x00, 0x00,
		0x00, 0x02,
		0x00,
		0x01,
		0x00,
		0x00,
	)
	data = append(data, make([]byte, 8)...)
	data = append(data,
		0x25, 0x01,
		0x03, 0x00,
		0x02, 0x00, 0x00, 0x00,
		0x2c, 0x01, 0x01,
		0xb6, 0x00, 0x44, 0x33, 0x22, 0x11,
	)
	packets, err := framer.Push(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 4 {
		t.Fatalf("packets = %d", len(packets))
	}
	for i, want := range []struct {
		id  uint16
		len int
	}{
		{0x01C5, 22},
		{0x0125, 8},
		{0x012C, 3},
		{0x00B6, 6},
	} {
		if packets[i].ID != want.id || len(packets[i].Data) != want.len {
			t.Fatalf("packet %d = %s, want 0x%04X len=%d", i, packets[i], want.id, want.len)
		}
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
