package network

import (
	"encoding/binary"
	"testing"
)

func TestBuildGuildPackets(t *testing.T) {
	create := BuildCreateGuildPacket(0x01020304, "Knights")
	if len(create) != 30 || ID(create) != PacketCZReqMakeGuild {
		t.Fatalf("BuildCreateGuildPacket len=%d id=0x%04x", len(create), ID(create))
	}
	if got := binary.LittleEndian.Uint32(create[2:6]); got != 0x01020304 {
		t.Fatalf("create char id = 0x%08x", got)
	}
	if got := string(create[6:13]); got != "Knights" {
		t.Fatalf("create name = %q", got)
	}

	invite := BuildRequestGuildInvitePacket(0x11111111, 0x22222222, 0x33333333)
	if len(invite) != 14 || ID(invite) != PacketCZReqJoinGuild {
		t.Fatalf("BuildRequestGuildInvitePacket len=%d id=0x%04x", len(invite), ID(invite))
	}
	if got := binary.LittleEndian.Uint32(invite[2:6]); got != 0x11111111 {
		t.Fatalf("invite target = 0x%08x", got)
	}

	reply := BuildGuildInviteReplyPacket(0x01020304, true)
	if len(reply) != 10 || ID(reply) != PacketCZJoinGuild {
		t.Fatalf("BuildGuildInviteReplyPacket len=%d id=0x%04x", len(reply), ID(reply))
	}
	if got := binary.LittleEndian.Uint32(reply[6:10]); got != 1 {
		t.Fatalf("reply accept = %d", got)
	}
}

func TestParseGuildPackets(t *testing.T) {
	info := make([]byte, 110)
	binary.LittleEndian.PutUint16(info[0:2], PacketZCGuildInfo)
	binary.LittleEndian.PutUint32(info[2:6], 0x01020304)
	copy(info[42:66], []byte("Mandala"))
	parsedInfo, ok, err := ParseGuildInfo(Packet{ID: PacketZCGuildInfo, Data: info})
	if !ok || err != nil || parsedInfo.GuildID != 0x01020304 || parsedInfo.GuildName != "Mandala" {
		t.Fatalf("ParseGuildInfo ok=%t err=%v info=%+v", ok, err, parsedInfo)
	}

	belonging := make([]byte, 43)
	binary.LittleEndian.PutUint16(belonging[0:2], PacketZCUpdateGuildID)
	binary.LittleEndian.PutUint32(belonging[2:6], 0x01020304)
	binary.LittleEndian.PutUint32(belonging[6:10], 7)
	binary.LittleEndian.PutUint32(belonging[10:14], 0x11)
	belonging[14] = 1
	copy(belonging[19:43], []byte("Mandala"))
	parsedBelonging, ok, err := ParseGuildBelonging(Packet{ID: PacketZCUpdateGuildID, Data: belonging})
	if !ok || err != nil || parsedBelonging.GuildID != 0x01020304 || parsedBelonging.EmblemVersion != 7 || parsedBelonging.Mode != 0x11 || !parsedBelonging.IsMaster || parsedBelonging.GuildName != "Mandala" {
		t.Fatalf("ParseGuildBelonging ok=%t err=%v belonging=%+v", ok, err, parsedBelonging)
	}

	createResult := []byte{0x67, 0x01, 2}
	parsedCreate, ok, err := ParseGuildCreationResult(Packet{ID: PacketZCResultMakeGuild, Data: createResult})
	if !ok || err != nil || parsedCreate.Result != 2 {
		t.Fatalf("ParseGuildCreationResult ok=%t err=%v result=%+v", ok, err, parsedCreate)
	}

	inviteAck := []byte{0x69, 0x01, 1}
	parsedAck, ok, err := ParseGuildInviteAck(Packet{ID: PacketZCAckReqJoinGuild, Data: inviteAck})
	if !ok || err != nil || parsedAck.Result != 1 {
		t.Fatalf("ParseGuildInviteAck ok=%t err=%v ack=%+v", ok, err, parsedAck)
	}

	request := make([]byte, 30)
	binary.LittleEndian.PutUint16(request[0:2], PacketZCReqJoinGuild)
	binary.LittleEndian.PutUint32(request[2:6], 0x01020304)
	copy(request[6:30], []byte("Knights"))
	parsedRequest, ok, err := ParseGuildInviteRequest(Packet{ID: PacketZCReqJoinGuild, Data: request})
	if !ok || err != nil || parsedRequest.GuildID != 0x01020304 || parsedRequest.GuildName != "Knights" {
		t.Fatalf("ParseGuildInviteRequest ok=%t err=%v request=%+v", ok, err, parsedRequest)
	}
}

func TestGuildPacketDirections(t *testing.T) {
	lengths := PacketLengths2008()
	for _, id := range []uint16{PacketCZReqMakeGuild, PacketCZReqJoinGuild, PacketCZJoinGuild} {
		if _, ok := lengths[id]; ok {
			t.Fatalf("0x%04X is client-to-server and must not be in the receive framer", id)
		}
	}
	for id, want := range map[uint16]int{
		PacketZCGuildInfo:       110,
		PacketZCGuildInfo2:      114,
		PacketZCResultMakeGuild: 3,
		PacketZCAckReqJoinGuild: 3,
		PacketZCReqJoinGuild:    30,
		PacketZCUpdateGuildID:   43,
	} {
		if got := lengths[id]; got != want {
			t.Fatalf("0x%04X receive length = %d, want %d", id, got, want)
		}
	}
}
