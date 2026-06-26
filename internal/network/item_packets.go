package network

import (
	"encoding/binary"
	"fmt"
	"log"
)

const PacketCZItemPickup uint16 = 0x009F

type itemPickupPacketLayout struct {
	date   int
	opcode uint16
	length int
	offset int
}

var itemPickupPacketLayouts = []itemPickupPacketLayout{
	// Keep this table aligned with rAthena's clif_packetdb.hpp and roBrowser's
	// PacketVersions.js. For 20080910 the last active main-client remap is the
	// 20070212 shuffled 0x00F5 packet.
	{date: 20101124, opcode: 0x0362, length: 6, offset: 2},
	{date: 20070212, opcode: 0x00F5, length: 8, offset: 4},
	{date: 20070108, opcode: 0x00F5, length: 11, offset: 7},
	{date: 20050719, opcode: 0x00F5, length: 13, offset: 9},
	{date: 20050718, opcode: 0x00F5, length: 7, offset: 3},
	{date: 20050628, opcode: 0x00F5, length: 13, offset: 9},
	{date: 20050509, opcode: 0x00F5, length: 8, offset: 4},
	{date: 20050110, opcode: 0x00F5, length: 9, offset: 5},
	{date: 20041129, opcode: 0x00A2, length: 7, offset: 3},
	{date: 20041025, opcode: 0x0113, length: 9, offset: 5},
	{date: 20041005, opcode: 0x0113, length: 10, offset: 6},
	{date: 20040920, opcode: 0x0113, length: 14, offset: 10},
	{date: 20040906, opcode: 0x0113, length: 11, offset: 7},
	{date: 20040809, opcode: 0x0094, length: 13, offset: 9},
	{date: 20040726, opcode: 0x0094, length: 10, offset: 6},
	{date: 20040713, opcode: 0x009F, length: 10, offset: 6},
}

type FloorItemEntry struct {
	ID         uint32
	ItemID     uint16
	Identified bool
	X          int
	Y          int
	SubX       uint8
	SubY       uint8
	Amount     uint16
	Falling    bool
}

type FloorItemDisappear struct {
	ID uint32
}

type ItemPickupAck struct {
	Index      uint16
	Amount     uint16
	ItemID     uint16
	Identified bool
	Result     uint8
}

func ParseFloorItemEntry(packet Packet) (FloorItemEntry, bool, error) {
	switch packet.ID {
	case 0x009D:
		if len(packet.Data) < 17 {
			return FloorItemEntry{}, false, fmt.Errorf("ZC_ITEM_ENTRY too short: %d", len(packet.Data))
		}
		return FloorItemEntry{
			ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
			ItemID:     binary.LittleEndian.Uint16(packet.Data[6:8]),
			Identified: packet.Data[8] != 0,
			X:          int(binary.LittleEndian.Uint16(packet.Data[9:11])),
			Y:          int(binary.LittleEndian.Uint16(packet.Data[11:13])),
			Amount:     binary.LittleEndian.Uint16(packet.Data[13:15]),
			SubX:       packet.Data[15],
			SubY:       packet.Data[16],
		}, true, nil
	case 0x009E:
		if len(packet.Data) < 17 {
			return FloorItemEntry{}, false, fmt.Errorf("ZC_ITEM_FALL_ENTRY too short: %d", len(packet.Data))
		}
		return FloorItemEntry{
			ID:         binary.LittleEndian.Uint32(packet.Data[2:6]),
			ItemID:     binary.LittleEndian.Uint16(packet.Data[6:8]),
			Identified: packet.Data[8] != 0,
			X:          int(binary.LittleEndian.Uint16(packet.Data[9:11])),
			Y:          int(binary.LittleEndian.Uint16(packet.Data[11:13])),
			SubX:       packet.Data[13],
			SubY:       packet.Data[14],
			Amount:     binary.LittleEndian.Uint16(packet.Data[15:17]),
			Falling:    true,
		}, true, nil
	default:
		return FloorItemEntry{}, false, nil
	}
}

func ParseFloorItemDisappear(packet Packet) (FloorItemDisappear, bool, error) {
	if packet.ID != 0x00A1 {
		return FloorItemDisappear{}, false, nil
	}
	if len(packet.Data) < 6 {
		return FloorItemDisappear{}, false, fmt.Errorf("ZC_ITEM_DISAPPEAR too short: %d", len(packet.Data))
	}
	return FloorItemDisappear{ID: binary.LittleEndian.Uint32(packet.Data[2:6])}, true, nil
}

func ParseItemPickupAck(packet Packet) (ItemPickupAck, bool, error) {
	switch packet.ID {
	case 0x00A0, 0x029A, 0x02D4:
	default:
		return ItemPickupAck{}, false, nil
	}
	if len(packet.Data) < 23 {
		return ItemPickupAck{}, false, fmt.Errorf("ZC_ITEM_PICKUP_ACK 0x%04X too short: %d", packet.ID, len(packet.Data))
	}
	return ItemPickupAck{
		Index:      binary.LittleEndian.Uint16(packet.Data[2:4]),
		Amount:     binary.LittleEndian.Uint16(packet.Data[4:6]),
		ItemID:     binary.LittleEndian.Uint16(packet.Data[6:8]),
		Identified: packet.Data[8] != 0,
		Result:     packet.Data[22],
	}, true, nil
}

func BuildItemPickupPacket(gid uint32) []byte {
	var w Writer
	w.Uint16(PacketCZItemPickup)
	w.Uint32(gid)
	return w.Bytes()
}

func BuildItemPickupPacketForClientDate(gid uint32, clientDate int) []byte {
	for _, layout := range itemPickupPacketLayouts {
		if clientDate >= layout.date {
			packet := make([]byte, layout.length)
			binary.LittleEndian.PutUint16(packet[0:2], layout.opcode)
			binary.LittleEndian.PutUint32(packet[layout.offset:layout.offset+4], gid)
			return packet
		}
	}
	return BuildItemPickupPacket(gid)
}

func (c *Client) SendItemPickup(gid uint32) error {
	packet := BuildItemPickupPacketForClientDate(gid, c.clientDate)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_ITEM_PICKUP opcode=0x%04X target=%d client_date=%d", ID(packet), gid, c.clientDate)
	} else {
		log.Printf("send CZ_ITEM_PICKUP failed opcode=0x%04X len=%d target=%d client_date=%d: %v", ID(packet), len(packet), gid, c.clientDate, err)
	}
	return err
}
