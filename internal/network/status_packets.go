package network

import (
	"encoding/binary"
	"fmt"
)

const (
	StatusBaseExp     uint16 = 1
	StatusJobExp      uint16 = 2
	StatusHP          uint16 = 5
	StatusMaxHP       uint16 = 6
	StatusSP          uint16 = 7
	StatusMaxSP       uint16 = 8
	StatusBaseLevel   uint16 = 11
	StatusNextBaseExp uint16 = 22
	StatusNextJobExp  uint16 = 23
	StatusJobLevel    uint16 = 55
)

type ParameterChange struct {
	VarID uint16
	Value int64
}

func ParseParameterChange(packet Packet) (ParameterChange, bool, error) {
	if packet.ID != 0x00B0 && packet.ID != 0x00B1 {
		return ParameterChange{}, false, nil
	}
	if len(packet.Data) < 8 {
		return ParameterChange{}, false, fmt.Errorf("ZC_PAR_CHANGE too short: %d", len(packet.Data))
	}
	return ParameterChange{
		VarID: binary.LittleEndian.Uint16(packet.Data[2:4]),
		Value: int64(binary.LittleEndian.Uint32(packet.Data[4:8])),
	}, true, nil
}
