package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type MapChange struct {
	MapName    string
	X          int
	Y          int
	Address    string
	Port       uint16
	ServerMove bool
}

type MapCellUpdate struct {
	MapName string
	X       int
	Y       int
	RawType uint32
}

func ParseMapChange(packet Packet) (MapChange, bool, error) {
	switch packet.ID {
	case 0x0091, 0x0092, 0x0AC7:
	default:
		return MapChange{}, false, nil
	}
	if len(packet.Data) < 22 {
		return MapChange{}, false, fmt.Errorf("ZC_CHANGEMAP 0x%04X too short: %d", packet.ID, len(packet.Data))
	}
	change := MapChange{
		MapName: normalizeMapName(fixedString(packet.Data[2:18])),
		X:       int(binary.LittleEndian.Uint16(packet.Data[18:20])),
		Y:       int(binary.LittleEndian.Uint16(packet.Data[20:22])),
	}
	if change.MapName == "" {
		return MapChange{}, false, fmt.Errorf("ZC_CHANGEMAP 0x%04X missing map name", packet.ID)
	}
	if packet.ID == 0x0092 || packet.ID == 0x0AC7 {
		if len(packet.Data) < 28 {
			return MapChange{}, false, fmt.Errorf("ZC_CHANGEMAPSERVER too short: %d", len(packet.Data))
		}
		change.Address = net.IPv4(packet.Data[22], packet.Data[23], packet.Data[24], packet.Data[25]).String()
		change.Port = binary.LittleEndian.Uint16(packet.Data[26:28])
		change.ServerMove = change.Address != "0.0.0.0" && change.Port != 0
	}
	return change, true, nil
}

func ParseMapCellUpdate(packet Packet) (MapCellUpdate, bool, error) {
	if packet.ID != 0x0192 {
		return MapCellUpdate{}, false, nil
	}
	if len(packet.Data) < 24 {
		return MapCellUpdate{}, false, fmt.Errorf("ZC_UPDATE_MAPINFO too short: %d", len(packet.Data))
	}
	return MapCellUpdate{
		X:       int(binary.LittleEndian.Uint16(packet.Data[2:4])),
		Y:       int(binary.LittleEndian.Uint16(packet.Data[4:6])),
		RawType: uint32(binary.LittleEndian.Uint16(packet.Data[6:8])),
		MapName: normalizeMapName(fixedString(packet.Data[8:24])),
	}, true, nil
}

func normalizeMapName(mapName string) string {
	mapName = strings.TrimSpace(mapName)
	lower := strings.ToLower(mapName)
	for _, ext := range []string{".gat", ".rsw"} {
		if strings.HasSuffix(lower, ext) {
			return mapName[:len(mapName)-len(ext)]
		}
	}
	return mapName
}
