package network

import (
	"encoding/binary"
	"testing"
)

func TestParseTaekwonMission(t *testing.T) {
	data := make([]byte, 32)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCStarSkill)
	copy(data[2:26], "Spore")
	binary.LittleEndian.PutUint32(data[26:30], 1014)
	data[30] = 37
	data[31] = 1

	mission, ok, err := ParseTaekwonMission(Packet{ID: PacketZCStarSkill, Data: data})
	if err != nil || !ok {
		t.Fatalf("parse mission ok=%t err=%v", ok, err)
	}
	if mission.MonsterName != "Spore" || mission.MonsterID != 1014 || mission.Progress != 37 || mission.Result != 1 {
		t.Fatalf("mission = %+v", mission)
	}
}

func TestParseTaekwonPoint(t *testing.T) {
	data := make([]byte, 10)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCTaekwonPoint)
	binary.LittleEndian.PutUint32(data[2:6], uint32(25))
	binary.LittleEndian.PutUint32(data[6:10], uint32(120))

	point, ok, err := ParseTaekwonPoint(Packet{ID: PacketZCTaekwonPoint, Data: data})
	if err != nil || !ok || point.Point != 25 || point.TotalPoint != 120 {
		t.Fatalf("point = %+v ok=%t err=%v", point, ok, err)
	}
}

func TestParseTaekwonRanking(t *testing.T) {
	data := make([]byte, 282)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCTaekwonRank)
	for i := 0; i < taekwonRankEntryCount; i++ {
		name := []byte{'R', 'a', 'n', 'k', byte('0' + i)}
		copy(data[2+i*24:2+(i+1)*24], name)
		pointOffset := 2 + taekwonRankEntryCount*24 + i*4
		binary.LittleEndian.PutUint32(data[pointOffset:pointOffset+4], uint32(1000-i*10))
	}

	ranking, ok, err := ParseTaekwonRanking(Packet{ID: PacketZCTaekwonRank, Data: data})
	if err != nil || !ok {
		t.Fatalf("parse ranking ok=%t err=%v", ok, err)
	}
	if len(ranking.Entries) != 10 || ranking.Entries[0].Name != "Rank0" || ranking.Entries[0].Point != 1000 || ranking.Entries[9].Name != "Rank9" || ranking.Entries[9].Point != 910 {
		t.Fatalf("ranking = %+v", ranking)
	}
}

func TestParseStarPlace(t *testing.T) {
	place, ok, err := ParseStarPlace(Packet{ID: PacketZCStarPlace, Data: []byte{0x53, 0x02, 3}})
	if err != nil || !ok || place.Place != 3 {
		t.Fatalf("place = %+v ok=%t err=%v", place, ok, err)
	}
}

func TestBuildTaekwonRankRequestPacket(t *testing.T) {
	packet := BuildTaekwonRankRequestPacket()
	if len(packet) != 2 || binary.LittleEndian.Uint16(packet) != PacketCZTaekwonRank {
		t.Fatalf("packet = % X", packet)
	}
}

func TestTaekwonPacketLengthsAreFramedFor2008(t *testing.T) {
	lengths := PacketLengths2008()
	want := map[uint16]int{
		PacketZCStarSkill: 32, PacketZCTaekwonPoint: 10,
		PacketZCTaekwonRank: 282, PacketZCStarPlace: 3,
	}
	for packetID, wantLen := range want {
		if got := lengths[packetID]; got != wantLen {
			t.Fatalf("packet 0x%04X length = %d, want %d", packetID, got, wantLen)
		}
	}
}

func TestTaekwonPacketParsersRejectShortPackets(t *testing.T) {
	if _, ok, err := ParseTaekwonMission(Packet{ID: PacketZCStarSkill, Data: make([]byte, 31)}); !ok || err == nil {
		t.Fatalf("short mission ok=%t err=%v", ok, err)
	}
	if _, ok, err := ParseTaekwonPoint(Packet{ID: PacketZCTaekwonPoint, Data: make([]byte, 9)}); !ok || err == nil {
		t.Fatalf("short point ok=%t err=%v", ok, err)
	}
	if _, ok, err := ParseTaekwonRanking(Packet{ID: PacketZCTaekwonRank, Data: make([]byte, 281)}); !ok || err == nil {
		t.Fatalf("short ranking ok=%t err=%v", ok, err)
	}
}
