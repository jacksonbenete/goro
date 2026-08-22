package network

import (
	"encoding/binary"
	"testing"
)

func TestParseMapPropertyNotify(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCNotifyMapProperty)
	binary.LittleEndian.PutUint16(data[2:4], 5)

	notify, ok, err := ParseMapPropertyNotify(Packet{ID: PacketZCNotifyMapProperty, Data: data})
	if err != nil || !ok {
		t.Fatalf("parse map property ok=%t err=%v", ok, err)
	}
	if notify.Property != 5 {
		t.Fatalf("property = %d, want 5", notify.Property)
	}
}

func TestParsePvPRankingPreservesSignedRank(t *testing.T) {
	data := make([]byte, 14)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCNotifyRanking)
	binary.LittleEndian.PutUint32(data[2:6], 2000000)
	binary.LittleEndian.PutUint32(data[6:10], uint32(0xFFFFFFFF))
	binary.LittleEndian.PutUint32(data[10:14], 12)

	ranking, ok, err := ParsePvPRanking(Packet{ID: PacketZCNotifyRanking, Data: data})
	if err != nil || !ok {
		t.Fatalf("parse pvp ranking ok=%t err=%v", ok, err)
	}
	if ranking.ActorID != 2000000 || ranking.Rank != -1 || ranking.Total != 12 {
		t.Fatalf("ranking = %+v, want actor=2000000 rank=-1 total=12", ranking)
	}
}

func TestBuildAndParsePvPInfoPackets(t *testing.T) {
	request := BuildPvPInfoRequestPacket(150000, 2000000)
	if len(request) != 10 || binary.LittleEndian.Uint16(request[0:2]) != PacketCZRequestPvPInfo {
		t.Fatalf("request = % X", request)
	}
	if charID := binary.LittleEndian.Uint32(request[2:6]); charID != 150000 {
		t.Fatalf("request char id = %d, want 150000", charID)
	}
	if accountID := binary.LittleEndian.Uint32(request[6:10]); accountID != 2000000 {
		t.Fatalf("request account id = %d, want 2000000", accountID)
	}

	data := make([]byte, 22)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCAckPvPInfo)
	binary.LittleEndian.PutUint32(data[2:6], 150000)
	binary.LittleEndian.PutUint32(data[6:10], 2000000)
	binary.LittleEndian.PutUint32(data[10:14], 7)
	binary.LittleEndian.PutUint32(data[14:18], 3)
	binary.LittleEndian.PutUint32(data[18:22], 42)

	info, ok, err := ParsePvPInfo(Packet{ID: PacketZCAckPvPInfo, Data: data})
	if err != nil || !ok {
		t.Fatalf("parse pvp info ok=%t err=%v", ok, err)
	}
	if info.CharID != 150000 || info.AccountID != 2000000 || info.Wins != 7 || info.Losses != 3 || info.Points != 42 {
		t.Fatalf("pvp info = %+v", info)
	}
}

func TestPvPInfoPacketLengthsAreFramedFor2008(t *testing.T) {
	lengths := PacketLengths2008()
	if lengths[PacketCZRequestPvPInfo] != 10 || lengths[PacketZCAckPvPInfo] != 22 {
		t.Fatalf("pvp packet lengths = request:%d ack:%d", lengths[PacketCZRequestPvPInfo], lengths[PacketZCAckPvPInfo])
	}
}

func TestParsePvPPacketsRejectShortPayloads(t *testing.T) {
	if _, ok, err := ParseMapPropertyNotify(Packet{ID: PacketZCNotifyMapProperty, Data: []byte{0x99, 0x01}}); !ok || err == nil {
		t.Fatalf("short map property ok=%t err=%v", ok, err)
	}
	if _, ok, err := ParsePvPRanking(Packet{ID: PacketZCNotifyRanking, Data: make([]byte, 13)}); !ok || err == nil {
		t.Fatalf("short ranking ok=%t err=%v", ok, err)
	}
	if _, ok, err := ParsePvPInfo(Packet{ID: PacketZCAckPvPInfo, Data: make([]byte, 21)}); !ok || err == nil {
		t.Fatalf("short pvp info ok=%t err=%v", ok, err)
	}
}
