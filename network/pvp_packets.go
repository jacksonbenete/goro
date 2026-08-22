package network

import (
	"encoding/binary"
	"fmt"
)

const (
	PacketZCNotifyMapProperty uint16 = 0x0199
	PacketZCNotifyRanking     uint16 = 0x019A
	PacketCZRequestPvPInfo    uint16 = 0x020F
	PacketZCAckPvPInfo        uint16 = 0x0210
)

type MapPropertyNotify struct {
	Property uint16
}

type PvPRanking struct {
	ActorID uint32
	Rank    int32
	Total   int32
}

type PvPInfo struct {
	CharID    uint32
	AccountID uint32
	Wins      int32
	Losses    int32
	Points    int32
}

func ParseMapPropertyNotify(packet Packet) (MapPropertyNotify, bool, error) {
	if packet.ID != PacketZCNotifyMapProperty {
		return MapPropertyNotify{}, false, nil
	}
	if len(packet.Data) < 4 {
		return MapPropertyNotify{}, true, fmt.Errorf("ZC_NOTIFY_MAPPROPERTY too short: %d", len(packet.Data))
	}
	return MapPropertyNotify{
		Property: binary.LittleEndian.Uint16(packet.Data[2:4]),
	}, true, nil
}

func ParsePvPRanking(packet Packet) (PvPRanking, bool, error) {
	if packet.ID != PacketZCNotifyRanking {
		return PvPRanking{}, false, nil
	}
	if len(packet.Data) < 14 {
		return PvPRanking{}, true, fmt.Errorf("ZC_NOTIFY_RANKING too short: %d", len(packet.Data))
	}
	return PvPRanking{
		ActorID: binary.LittleEndian.Uint32(packet.Data[2:6]),
		Rank:    int32(binary.LittleEndian.Uint32(packet.Data[6:10])),
		Total:   int32(binary.LittleEndian.Uint32(packet.Data[10:14])),
	}, true, nil
}

func BuildPvPInfoRequestPacket(charID, accountID uint32) []byte {
	var w Writer
	w.Uint16(PacketCZRequestPvPInfo)
	w.Uint32(charID)
	w.Uint32(accountID)
	return w.Bytes()
}

func ParsePvPInfo(packet Packet) (PvPInfo, bool, error) {
	if packet.ID != PacketZCAckPvPInfo {
		return PvPInfo{}, false, nil
	}
	if len(packet.Data) < 22 {
		return PvPInfo{}, true, fmt.Errorf("ZC_ACK_PVPPOINT too short: %d", len(packet.Data))
	}
	return PvPInfo{
		CharID:    binary.LittleEndian.Uint32(packet.Data[2:6]),
		AccountID: binary.LittleEndian.Uint32(packet.Data[6:10]),
		Wins:      int32(binary.LittleEndian.Uint32(packet.Data[10:14])),
		Losses:    int32(binary.LittleEndian.Uint32(packet.Data[14:18])),
		Points:    int32(binary.LittleEndian.Uint32(packet.Data[18:22])),
	}, true, nil
}

func (c *Client) SendPvPInfoRequest(charID, accountID uint32) error {
	return c.Send(BuildPvPInfoRequestPacket(charID, accountID))
}
