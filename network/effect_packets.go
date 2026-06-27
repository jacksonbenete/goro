package network

import (
	"encoding/binary"
	"fmt"
)

const (
	SpecialEffectBaseLevelUp = 0
	SpecialEffectJobLevelUp  = 1
)

type SpecialEffectNotify struct {
	AID      uint32
	EffectID uint32
}

func ParseSpecialEffectNotify(packet Packet) (SpecialEffectNotify, bool, error) {
	if packet.ID != 0x019B {
		return SpecialEffectNotify{}, false, nil
	}
	if len(packet.Data) < 10 {
		return SpecialEffectNotify{}, false, fmt.Errorf("ZC_NOTIFY_EFFECT too short: %d", len(packet.Data))
	}
	return SpecialEffectNotify{
		AID:      binary.LittleEndian.Uint32(packet.Data[2:6]),
		EffectID: binary.LittleEndian.Uint32(packet.Data[6:10]),
	}, true, nil
}
