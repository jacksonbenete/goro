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

	menu := BuildGuildMenuRequestPacket(3)
	if len(menu) != 6 || ID(menu) != PacketCZReqGuildMenu {
		t.Fatalf("BuildGuildMenuRequestPacket len=%d id=0x%04x", len(menu), ID(menu))
	}
	if got := binary.LittleEndian.Uint32(menu[2:6]); got != 3 {
		t.Fatalf("menu tab = %d", got)
	}
}

func TestParseGuildPackets(t *testing.T) {
	info := make([]byte, 110)
	binary.LittleEndian.PutUint16(info[0:2], PacketZCGuildInfo)
	binary.LittleEndian.PutUint32(info[2:6], 0x01020304)
	binary.LittleEndian.PutUint32(info[6:10], 5)
	binary.LittleEndian.PutUint32(info[10:14], 7)
	binary.LittleEndian.PutUint32(info[14:18], 16)
	binary.LittleEndian.PutUint32(info[18:22], 42)
	binary.LittleEndian.PutUint32(info[22:26], 1234)
	binary.LittleEndian.PutUint32(info[26:30], 9000)
	binary.LittleEndian.PutUint32(info[30:34], 11)
	binary.LittleEndian.PutUint32(info[34:38], 12)
	binary.LittleEndian.PutUint32(info[38:42], 13)
	binary.LittleEndian.PutUint32(info[42:46], 7)
	copy(info[46:70], []byte("Mandala"))
	copy(info[70:94], []byte("Arcer"))
	copy(info[94:110], []byte("Prontera"))
	parsedInfo, ok, err := ParseGuildInfo(Packet{ID: PacketZCGuildInfo, Data: info})
	if !ok || err != nil || parsedInfo.GuildID != 0x01020304 || parsedInfo.Level != 5 || parsedInfo.UserNum != 7 || parsedInfo.MaxUserNum != 16 || parsedInfo.UserAverageLevel != 42 || parsedInfo.Exp != 1234 || parsedInfo.MaxExp != 9000 || parsedInfo.Point != 11 || parsedInfo.Honor != 12 || parsedInfo.Virtue != 13 || parsedInfo.EmblemVersion != 7 || parsedInfo.GuildName != "Mandala" || parsedInfo.MasterName != "Arcer" || parsedInfo.ManageLand != "Prontera" {
		t.Fatalf("ParseGuildInfo ok=%t err=%v info=%+v", ok, err, parsedInfo)
	}

	info2 := append([]byte(nil), info...)
	info2 = append(info2, 0, 0, 0, 0)
	binary.LittleEndian.PutUint16(info2[0:2], PacketZCGuildInfo2)
	binary.LittleEndian.PutUint32(info2[110:114], 765)
	parsedInfo2, ok, err := ParseGuildInfo(Packet{ID: PacketZCGuildInfo2, Data: info2})
	if !ok || err != nil || parsedInfo2.Zeny != 765 || parsedInfo2.GuildName != "Mandala" {
		t.Fatalf("ParseGuildInfo2 ok=%t err=%v info=%+v", ok, err, parsedInfo2)
	}

	members := make([]byte, 4+104)
	binary.LittleEndian.PutUint16(members[0:2], PacketZCGuildMembers)
	binary.LittleEndian.PutUint16(members[2:4], uint16(len(members)))
	member := members[4:]
	binary.LittleEndian.PutUint32(member[0:4], 0x01020304)
	binary.LittleEndian.PutUint32(member[4:8], 0x05060708)
	binary.LittleEndian.PutUint16(member[8:10], 3)
	binary.LittleEndian.PutUint16(member[10:12], 4)
	binary.LittleEndian.PutUint16(member[12:14], 1)
	binary.LittleEndian.PutUint16(member[14:16], 4)
	binary.LittleEndian.PutUint16(member[16:18], 55)
	binary.LittleEndian.PutUint32(member[18:22], 12345)
	binary.LittleEndian.PutUint32(member[22:26], 1)
	binary.LittleEndian.PutUint32(member[26:30], 2)
	copyFixedName(member[30:80], "Memo")
	copyFixedName(member[80:104], "Arcer")
	parsedMembers, ok, err := ParseGuildMembers(Packet{ID: PacketZCGuildMembers, Data: members})
	if !ok || err != nil || len(parsedMembers) != 1 || parsedMembers[0].CharName != "Arcer" || parsedMembers[0].Memo != "Memo" || parsedMembers[0].Level != 55 || parsedMembers[0].PositionID != 2 {
		t.Fatalf("ParseGuildMembers ok=%t err=%v members=%+v", ok, err, parsedMembers)
	}

	positions := make([]byte, 4+16)
	binary.LittleEndian.PutUint16(positions[0:2], PacketZCGuildPositions)
	binary.LittleEndian.PutUint16(positions[2:4], uint16(len(positions)))
	position := positions[4:]
	binary.LittleEndian.PutUint32(position[0:4], 2)
	binary.LittleEndian.PutUint32(position[4:8], 0x11)
	binary.LittleEndian.PutUint32(position[8:12], 3)
	binary.LittleEndian.PutUint32(position[12:16], 50)
	parsedPositions, ok, err := ParseGuildPositions(Packet{ID: PacketZCGuildPositions, Data: positions})
	if !ok || err != nil || len(parsedPositions) != 1 || parsedPositions[0].PositionID != 2 || parsedPositions[0].Right != 0x11 || parsedPositions[0].Ranking != 3 || parsedPositions[0].PayRate != 50 {
		t.Fatalf("ParseGuildPositions ok=%t err=%v positions=%+v", ok, err, parsedPositions)
	}

	positionNames := make([]byte, 4+28)
	binary.LittleEndian.PutUint16(positionNames[0:2], PacketZCGuildPosNames)
	binary.LittleEndian.PutUint16(positionNames[2:4], uint16(len(positionNames)))
	binary.LittleEndian.PutUint32(positionNames[4:8], 2)
	copyFixedName(positionNames[8:32], "Leader")
	parsedPositionNames, ok, err := ParseGuildPositionNames(Packet{ID: PacketZCGuildPosNames, Data: positionNames})
	if !ok || err != nil || len(parsedPositionNames) != 1 || parsedPositionNames[0].PositionID != 2 || parsedPositionNames[0].PosName != "Leader" {
		t.Fatalf("ParseGuildPositionNames ok=%t err=%v positions=%+v", ok, err, parsedPositionNames)
	}

	guildSkills := make([]byte, 6+37)
	binary.LittleEndian.PutUint16(guildSkills[0:2], PacketZCGuildSkillInfo)
	binary.LittleEndian.PutUint16(guildSkills[2:4], uint16(len(guildSkills)))
	binary.LittleEndian.PutUint16(guildSkills[4:6], 3)
	writeSkillInfoEntry(guildSkills[6:], 1, 10000, 1, 0, 1, "Official Guild Approval", true)
	parsedGuildSkills, ok, err := ParseGuildSkillInfo(Packet{ID: PacketZCGuildSkillInfo, Data: guildSkills})
	if !ok || err != nil || parsedGuildSkills.SkillPoints != 3 || len(parsedGuildSkills.Skills) != 1 || parsedGuildSkills.Skills[0].ID != 10000 || parsedGuildSkills.Skills[0].Name != "Official Guild Approval" {
		t.Fatalf("ParseGuildSkillInfo ok=%t err=%v info=%+v", ok, err, parsedGuildSkills)
	}

	expelHistory := make([]byte, 4+88)
	binary.LittleEndian.PutUint16(expelHistory[0:2], PacketZCGuildBanList)
	binary.LittleEndian.PutUint16(expelHistory[2:4], uint16(len(expelHistory)))
	copyFixedName(expelHistory[4:28], "Chjara")
	copyFixedName(expelHistory[28:52], "Kivutar")
	copyFixedName(expelHistory[52:92], "Testing")
	parsedExpelHistory, ok, err := ParseGuildExpelHistory(Packet{ID: PacketZCGuildBanList, Data: expelHistory})
	if !ok || err != nil || len(parsedExpelHistory) != 1 || parsedExpelHistory[0].CharName != "Chjara" || parsedExpelHistory[0].Account != "Kivutar" || parsedExpelHistory[0].Reason != "Testing" {
		t.Fatalf("ParseGuildExpelHistory ok=%t err=%v history=%+v", ok, err, parsedExpelHistory)
	}

	notice := make([]byte, 182)
	binary.LittleEndian.PutUint16(notice[0:2], PacketZCGuildNotice)
	copyFixedName(notice[2:62], "Maintenance")
	copyFixedName(notice[62:182], "Gather in Prontera.")
	parsedNotice, ok, err := ParseGuildNotice(Packet{ID: PacketZCGuildNotice, Data: notice})
	if !ok || err != nil || parsedNotice.Subject != "Maintenance" || parsedNotice.Notice != "Gather in Prontera." {
		t.Fatalf("ParseGuildNotice ok=%t err=%v notice=%+v", ok, err, parsedNotice)
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

	emblem := make([]byte, 16)
	binary.LittleEndian.PutUint16(emblem[0:2], PacketZCGuildEmblem)
	binary.LittleEndian.PutUint16(emblem[2:4], uint16(len(emblem)))
	binary.LittleEndian.PutUint32(emblem[4:8], 0x01020304)
	binary.LittleEndian.PutUint32(emblem[8:12], 9)
	copy(emblem[12:], []byte{1, 2, 3, 4})
	parsedEmblem, ok, err := ParseGuildEmblemImage(Packet{ID: PacketZCGuildEmblem, Data: emblem})
	if !ok || err != nil || parsedEmblem.GuildID != 0x01020304 || parsedEmblem.EmblemVersion != 9 || string(parsedEmblem.Data) != string([]byte{1, 2, 3, 4}) {
		t.Fatalf("ParseGuildEmblemImage ok=%t err=%v emblem=%+v", ok, err, parsedEmblem)
	}

	change := make([]byte, 12)
	binary.LittleEndian.PutUint16(change[0:2], PacketZCChangeGuild)
	binary.LittleEndian.PutUint32(change[2:6], 0x11111111)
	binary.LittleEndian.PutUint32(change[6:10], 0x01020304)
	binary.LittleEndian.PutUint16(change[10:12], 11)
	parsedChange, ok, err := ParseGuildEmblemChange(Packet{ID: PacketZCChangeGuild, Data: change})
	if !ok || err != nil || parsedChange.ActorID != 0x11111111 || parsedChange.GuildID != 0x01020304 || parsedChange.EmblemVersion != 11 {
		t.Fatalf("ParseGuildEmblemChange ok=%t err=%v change=%+v", ok, err, parsedChange)
	}
}

func TestGuildPacketDirections(t *testing.T) {
	lengths := PacketLengths2008()
	for _, id := range []uint16{PacketCZReqMakeGuild, PacketCZReqJoinGuild, PacketCZJoinGuild, PacketCZReqGuildEmblem, PacketCZRegGuildEmblem} {
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
		PacketZCGuildMembers:    -1,
		PacketZCGuildPositions:  -1,
		PacketZCGuildSkillInfo:  -1,
		PacketZCGuildBanList:    -1,
		PacketZCGuildPosNames:   -1,
		PacketZCGuildNotice:     182,
		PacketZCUpdateGuildID:   43,
		PacketZCGuildEmblem:     -1,
		PacketZCChangeGuild:     12,
	} {
		if got := lengths[id]; got != want {
			t.Fatalf("0x%04X receive length = %d, want %d", id, got, want)
		}
	}
}
