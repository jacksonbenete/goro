package network

import (
	"encoding/binary"
	"fmt"
)

const PacketZCCompass uint16 = 0x0144

type MinimapCompass struct {
	NPCID uint32
	Type  int
	X     int
	Y     int
	ID    uint8
	Color uint32
}

func ParseMinimapCompass(packet Packet) (MinimapCompass, bool, error) {
	if packet.ID != PacketZCCompass {
		return MinimapCompass{}, false, nil
	}
	if len(packet.Data) < 23 {
		return MinimapCompass{}, true, fmt.Errorf("ZC_COMPASS too short: %d", len(packet.Data))
	}
	return MinimapCompass{
		NPCID: binary.LittleEndian.Uint32(packet.Data[2:6]),
		Type:  int(int32(binary.LittleEndian.Uint32(packet.Data[6:10]))),
		X:     int(int32(binary.LittleEndian.Uint32(packet.Data[10:14]))),
		Y:     int(int32(binary.LittleEndian.Uint32(packet.Data[14:18]))),
		ID:    packet.Data[18],
		Color: binary.LittleEndian.Uint32(packet.Data[19:23]),
	}, true, nil
}
