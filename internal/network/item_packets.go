package network

import (
	"encoding/binary"
	"fmt"
	"log"
)

const PacketCZItemPickup uint16 = 0x009F

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
	if packet.ID != 0x00A0 {
		return ItemPickupAck{}, false, nil
	}
	if len(packet.Data) < 23 {
		return ItemPickupAck{}, false, fmt.Errorf("ZC_ITEM_PICKUP_ACK too short: %d", len(packet.Data))
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
	_ = clientDate
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
