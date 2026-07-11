package network

import (
	"encoding/binary"
	"testing"
)

func TestBuildPartyPackets(t *testing.T) {
	makeParty := BuildMakePartyPacket("Goro")
	if len(makeParty) != 26 || ID(makeParty) != PacketCZMakeGroup || string(makeParty[2:6]) != "Goro" {
		t.Fatalf("BuildMakePartyPacket = len %d id 0x%04x data %q", len(makeParty), ID(makeParty), makeParty[2:6])
	}

	invite := BuildPartyInvitePacket(0x11223344, "Alice")
	if len(invite) != 6 || ID(invite) != PacketCZReqJoinGroup || binary.LittleEndian.Uint32(invite[2:6]) != 0x11223344 {
		t.Fatalf("BuildPartyInvitePacket = len %d id 0x%04x data %x", len(invite), ID(invite), invite)
	}

	ack := BuildPartyInviteAckPacket(0x01020304, true)
	if len(ack) != 10 || ID(ack) != PacketCZJoinGroup || binary.LittleEndian.Uint32(ack[2:6]) != 0x01020304 || binary.LittleEndian.Uint32(ack[6:10]) != 1 {
		t.Fatalf("BuildPartyInviteAckPacket = len %d id 0x%04x data %x", len(ack), ID(ack), ack)
	}

	leave := BuildLeavePartyPacket()
	if len(leave) != 2 || ID(leave) != PacketCZReqLeaveGroup {
		t.Fatalf("BuildLeavePartyPacket = len %d id 0x%04x", len(leave), ID(leave))
	}

	opt := BuildPartyOptionPacket(1)
	if len(opt) != 6 || ID(opt) != PacketCZChangeGroupExp || binary.LittleEndian.Uint32(opt[2:6]) != 1 {
		t.Fatalf("BuildPartyOptionPacket = len %d id 0x%04x data %x", len(opt), ID(opt), opt)
	}
}

func TestParsePartyPackets(t *testing.T) {
	create := []byte{0xfa, 0x00, 0x00}
	parsedCreate, ok, err := ParsePartyCreateResult(Packet{ID: PacketZCAckMakeGroup, Data: create})
	if !ok || err != nil || parsedCreate.Result != 0 {
		t.Fatalf("ParsePartyCreateResult ok=%t err=%v result=%+v", ok, err, parsedCreate)
	}

	list := make([]byte, 28+46)
	binary.LittleEndian.PutUint16(list[0:2], PacketZCGroupList)
	binary.LittleEndian.PutUint16(list[2:4], uint16(len(list)))
	copy(list[4:28], []byte("Beta Party"))
	binary.LittleEndian.PutUint32(list[28:32], 0x11223344)
	copy(list[32:56], []byte("Alice"))
	copy(list[56:72], []byte("prontera"))
	list[72] = 0
	list[73] = 0
	parsedList, ok, err := ParsePartyList(Packet{ID: PacketZCGroupList, Data: list})
	if !ok || err != nil || parsedList.Name != "Beta Party" || len(parsedList.Members) != 1 || parsedList.Members[0].Name != "Alice" || parsedList.Members[0].AccountID != 0x11223344 {
		t.Fatalf("ParsePartyList ok=%t err=%v list=%+v", ok, err, parsedList)
	}

	join := make([]byte, 79)
	binary.LittleEndian.PutUint16(join[0:2], PacketZCAddMemberToGroup)
	binary.LittleEndian.PutUint32(join[2:6], 0x22334455)
	binary.LittleEndian.PutUint32(join[6:10], 1)
	binary.LittleEndian.PutUint16(join[10:12], 12)
	binary.LittleEndian.PutUint16(join[12:14], 34)
	join[14] = 0
	copy(join[15:39], []byte("Beta Party"))
	copy(join[39:63], []byte("Bob"))
	copy(join[63:79], []byte("payon"))
	parsedJoin, ok, err := ParsePartyMemberJoin(Packet{ID: PacketZCAddMemberToGroup, Data: join})
	if !ok || err != nil || parsedJoin.Name != "Bob" || parsedJoin.MapName != "payon" || parsedJoin.X != 12 || parsedJoin.Y != 34 {
		t.Fatalf("ParsePartyMemberJoin ok=%t err=%v member=%+v", ok, err, parsedJoin)
	}

	hp := make([]byte, 10)
	binary.LittleEndian.PutUint16(hp[0:2], PacketZCNotifyHPToGroup)
	binary.LittleEndian.PutUint32(hp[2:6], 0x22334455)
	binary.LittleEndian.PutUint16(hp[6:8], 123)
	binary.LittleEndian.PutUint16(hp[8:10], 456)
	parsedHP, ok, err := ParsePartyMemberHP(Packet{ID: PacketZCNotifyHPToGroup, Data: hp})
	if !ok || err != nil || parsedHP.HP != 123 || parsedHP.MaxHP != 456 {
		t.Fatalf("ParsePartyMemberHP ok=%t err=%v hp=%+v", ok, err, parsedHP)
	}
}
