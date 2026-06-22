package network

import (
	"encoding/binary"
	"fmt"
)

type ActorEntry struct {
	ID         uint32
	Job        int16
	Head       int16
	Weapon     int16
	Shield     int16
	HeadTop    int16
	HeadMid    int16
	HeadLow    int16
	Sex        uint8
	Appearance bool
	X          int
	Y          int
	Dir        int
	Moving     bool
	FromX      int
	FromY      int
	ToX        int
	ToY        int
	ObjectType uint8
}

type ActorVanish struct {
	ID     uint32
	Reason uint8
}

type SelfMoveAck struct {
	ServerTick uint32
	FromX      int
	FromY      int
	ToX        int
	ToY        int
}

type ActorSetPosition struct {
	ID uint32
	X  int
	Y  int
}

type ActorLookChange struct {
	ID    uint32
	Type  uint8
	Value uint32
}

func ParseActorEntry(packet Packet) (ActorEntry, bool, error) {
	switch packet.ID {
	case 0x0078:
		if len(packet.Data) < 55 {
			return ActorEntry{}, false, fmt.Errorf("ZC_NOTIFY_STANDENTRY too short: %d", len(packet.Data))
		}
		x, y, dir := unpackPos(packet.Data[47:50])
		return ActorEntry{
			ObjectType: packet.Data[2],
			ID:         binary.LittleEndian.Uint32(packet.Data[3:7]),
			Job:        int16(binary.LittleEndian.Uint16(packet.Data[15:17])),
			Head:       int16(binary.LittleEndian.Uint16(packet.Data[17:19])),
			Weapon:     int16(binary.LittleEndian.Uint16(packet.Data[19:21])),
			HeadLow:    int16(binary.LittleEndian.Uint16(packet.Data[27:29])),
			Shield:     int16(binary.LittleEndian.Uint16(packet.Data[29:31])),
			HeadTop:    int16(binary.LittleEndian.Uint16(packet.Data[31:33])),
			HeadMid:    int16(binary.LittleEndian.Uint16(packet.Data[33:35])),
			Sex:        packet.Data[46],
			Appearance: true,
			X:          x,
			Y:          y,
			Dir:        dir,
		}, true, nil
	case 0x0079, 0x007A:
		if len(packet.Data) < 53 {
			return ActorEntry{}, false, fmt.Errorf("ZC_NOTIFY_ENTRY too short: %d", len(packet.Data))
		}
		x, y, dir := unpackPos(packet.Data[46:49])
		return ActorEntry{
			ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
			Job:        int16(binary.LittleEndian.Uint16(packet.Data[14:16])),
			Head:       int16(binary.LittleEndian.Uint16(packet.Data[16:18])),
			Weapon:     int16(binary.LittleEndian.Uint16(packet.Data[18:20])),
			HeadLow:    int16(binary.LittleEndian.Uint16(packet.Data[26:28])),
			Shield:     int16(binary.LittleEndian.Uint16(packet.Data[28:30])),
			HeadTop:    int16(binary.LittleEndian.Uint16(packet.Data[30:32])),
			HeadMid:    int16(binary.LittleEndian.Uint16(packet.Data[32:34])),
			Sex:        packet.Data[45],
			Appearance: true,
			X:          x,
			Y:          y,
			Dir:        dir,
		}, true, nil
	case 0x007B:
		if len(packet.Data) < 60 {
			return ActorEntry{}, false, fmt.Errorf("ZC_NOTIFY_MOVEENTRY too short: %d", len(packet.Data))
		}
		fromX, fromY, toX, toY := unpackMovePos(packet.Data[50:56])
		return ActorEntry{
			ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
			Job:        int16(binary.LittleEndian.Uint16(packet.Data[14:16])),
			Head:       int16(binary.LittleEndian.Uint16(packet.Data[16:18])),
			Weapon:     int16(binary.LittleEndian.Uint16(packet.Data[18:20])),
			HeadLow:    int16(binary.LittleEndian.Uint16(packet.Data[26:28])),
			Shield:     int16(binary.LittleEndian.Uint16(packet.Data[28:30])),
			HeadTop:    int16(binary.LittleEndian.Uint16(packet.Data[30:32])),
			HeadMid:    int16(binary.LittleEndian.Uint16(packet.Data[32:34])),
			Sex:        packet.Data[49],
			Appearance: true,
			X:          toX,
			Y:          toY,
			FromX:      fromX,
			FromY:      fromY,
			ToX:        toX,
			ToY:        toY,
			Moving:     true,
		}, true, nil
	case 0x007C:
		if len(packet.Data) < 42 {
			return ActorEntry{}, false, fmt.Errorf("ZC_NOTIFY_STANDENTRY_NPC too short: %d", len(packet.Data))
		}
		x, y, dir := unpackPos(packet.Data[37:40])
		return ActorEntry{
			ObjectType: packet.Data[2],
			ID:         binary.LittleEndian.Uint32(packet.Data[3:7]),
			Job:        int16(binary.LittleEndian.Uint16(packet.Data[21:23])),
			Head:       int16(binary.LittleEndian.Uint16(packet.Data[15:17])),
			Sex:        packet.Data[36],
			Appearance: true,
			X:          x,
			Y:          y,
			Dir:        dir,
		}, true, nil
	case 0x0086:
		if len(packet.Data) < 16 {
			return ActorEntry{}, false, fmt.Errorf("ZC_NOTIFY_MOVE too short: %d", len(packet.Data))
		}
		fromX, fromY, toX, toY := unpackMovePos(packet.Data[6:12])
		return ActorEntry{
			ID:     binary.LittleEndian.Uint32(packet.Data[2:6]),
			X:      toX,
			Y:      toY,
			FromX:  fromX,
			FromY:  fromY,
			ToX:    toX,
			ToY:    toY,
			Moving: true,
		}, true, nil
	case 0x01D8:
		if len(packet.Data) < 54 {
			return ActorEntry{}, false, fmt.Errorf("ZC_NOTIFY_STANDENTRY2 too short: %d", len(packet.Data))
		}
		x, y, dir := unpackPos(packet.Data[46:49])
		return ActorEntry{
			ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
			Job:        int16(binary.LittleEndian.Uint16(packet.Data[14:16])),
			Head:       int16(binary.LittleEndian.Uint16(packet.Data[16:18])),
			Weapon:     int16(binary.LittleEndian.Uint16(packet.Data[18:20])),
			HeadLow:    int16(binary.LittleEndian.Uint16(packet.Data[26:28])),
			Shield:     int16(binary.LittleEndian.Uint16(packet.Data[28:30])),
			HeadTop:    int16(binary.LittleEndian.Uint16(packet.Data[30:32])),
			HeadMid:    int16(binary.LittleEndian.Uint16(packet.Data[32:34])),
			Sex:        packet.Data[45],
			Appearance: true,
			X:          x,
			Y:          y,
			Dir:        dir,
		}, true, nil
	case 0x01D9:
		if len(packet.Data) < 53 {
			return ActorEntry{}, false, fmt.Errorf("ZC_NOTIFY_NEWENTRY2 too short: %d", len(packet.Data))
		}
		x, y, dir := unpackPos(packet.Data[46:49])
		return ActorEntry{
			ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
			Job:        int16(binary.LittleEndian.Uint16(packet.Data[14:16])),
			Head:       int16(binary.LittleEndian.Uint16(packet.Data[16:18])),
			Weapon:     int16(binary.LittleEndian.Uint16(packet.Data[18:20])),
			HeadLow:    int16(binary.LittleEndian.Uint16(packet.Data[26:28])),
			Shield:     int16(binary.LittleEndian.Uint16(packet.Data[28:30])),
			HeadTop:    int16(binary.LittleEndian.Uint16(packet.Data[30:32])),
			HeadMid:    int16(binary.LittleEndian.Uint16(packet.Data[32:34])),
			Sex:        packet.Data[45],
			Appearance: true,
			X:          x,
			Y:          y,
			Dir:        dir,
		}, true, nil
	case 0x01DA:
		if len(packet.Data) < 60 {
			return ActorEntry{}, false, fmt.Errorf("ZC_NOTIFY_MOVEENTRY2 too short: %d", len(packet.Data))
		}
		fromX, fromY, toX, toY := unpackMovePos(packet.Data[50:56])
		return ActorEntry{
			ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
			Job:        int16(binary.LittleEndian.Uint16(packet.Data[14:16])),
			Head:       int16(binary.LittleEndian.Uint16(packet.Data[16:18])),
			Weapon:     int16(binary.LittleEndian.Uint16(packet.Data[18:20])),
			HeadLow:    int16(binary.LittleEndian.Uint16(packet.Data[26:28])),
			Shield:     int16(binary.LittleEndian.Uint16(packet.Data[28:30])),
			HeadTop:    int16(binary.LittleEndian.Uint16(packet.Data[30:32])),
			HeadMid:    int16(binary.LittleEndian.Uint16(packet.Data[32:34])),
			Sex:        packet.Data[49],
			Appearance: true,
			X:          toX,
			Y:          toY,
			FromX:      fromX,
			FromY:      fromY,
			ToX:        toX,
			ToY:        toY,
			Moving:     true,
		}, true, nil
	default:
		return ActorEntry{}, false, nil
	}
}

func ParseActorLookChange(packet Packet) (ActorLookChange, bool, error) {
	switch packet.ID {
	case 0x00C3:
		if len(packet.Data) < 8 {
			return ActorLookChange{}, false, fmt.Errorf("ZC_CHANGE_LOOK too short: %d", len(packet.Data))
		}
		return ActorLookChange{
			ID:    binary.LittleEndian.Uint32(packet.Data[2:6]),
			Type:  packet.Data[6],
			Value: uint32(packet.Data[7]),
		}, true, nil
	case 0x01D7, 0x0229:
		if len(packet.Data) < 11 {
			return ActorLookChange{}, false, fmt.Errorf("ZC_CHANGE_LOOK2 too short: %d", len(packet.Data))
		}
		return ActorLookChange{
			ID:    binary.LittleEndian.Uint32(packet.Data[2:6]),
			Type:  packet.Data[6],
			Value: binary.LittleEndian.Uint32(packet.Data[7:11]),
		}, true, nil
	default:
		return ActorLookChange{}, false, nil
	}
}

func ParseActorVanish(packet Packet) (ActorVanish, bool, error) {
	if packet.ID != 0x0080 {
		return ActorVanish{}, false, nil
	}
	if len(packet.Data) < 7 {
		return ActorVanish{}, false, fmt.Errorf("ZC_NOTIFY_VANISH too short: %d", len(packet.Data))
	}
	return ActorVanish{
		ID:     binary.LittleEndian.Uint32(packet.Data[2:6]),
		Reason: packet.Data[6],
	}, true, nil
}

func ParseSelfMoveAck(packet Packet) (SelfMoveAck, bool, error) {
	if packet.ID != 0x0087 {
		return SelfMoveAck{}, false, nil
	}
	if len(packet.Data) < 12 {
		return SelfMoveAck{}, false, fmt.Errorf("ZC_NOTIFY_PLAYERMOVE too short: %d", len(packet.Data))
	}
	fromX, fromY, toX, toY := unpackMovePos(packet.Data[6:12])
	return SelfMoveAck{
		ServerTick: binary.LittleEndian.Uint32(packet.Data[2:6]),
		FromX:      fromX,
		FromY:      fromY,
		ToX:        toX,
		ToY:        toY,
	}, true, nil
}

func ParseActorSetPosition(packet Packet) (ActorSetPosition, bool, error) {
	if packet.ID != 0x0088 {
		return ActorSetPosition{}, false, nil
	}
	if len(packet.Data) < 10 {
		return ActorSetPosition{}, false, fmt.Errorf("ZC_STOPMOVE too short: %d", len(packet.Data))
	}
	return ActorSetPosition{
		ID: binary.LittleEndian.Uint32(packet.Data[2:6]),
		X:  int(int16(binary.LittleEndian.Uint16(packet.Data[6:8]))),
		Y:  int(int16(binary.LittleEndian.Uint16(packet.Data[8:10]))),
	}, true, nil
}

func unpackPos(data []byte) (x, y, dir int) {
	if len(data) < 3 {
		return 0, 0, 0
	}
	return int(data[0])<<2 | int(data[1]>>6),
		int(data[1]&0x3f)<<4 | int(data[2]>>4),
		int(data[2] & 0x0f)
}

func unpackMovePos(data []byte) (fromX, fromY, toX, toY int) {
	if len(data) < 6 {
		return 0, 0, 0, 0
	}
	a, b, c, d, e := data[0], data[1], data[2], data[3], data[4]
	return int(a)<<2 | int((b&0xc0)>>6),
		int(b&0x3f)<<4 | int((c&0xf0)>>4),
		int((d&0xfc)>>2) | int(c&0x0f)<<6,
		int(d&0x03)<<8 | int(e)
}
