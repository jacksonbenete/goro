package network

import (
	"encoding/binary"
	"fmt"
	"log"
)

const (
	PacketZCStartCapture      uint16 = 0x019E
	PacketCZTryCaptureMonster uint16 = 0x019F
	PacketZCTryCaptureMonster uint16 = 0x01A0
	PacketZCPetEggList        uint16 = 0x01A6
	PacketCZSelectPetEgg      uint16 = 0x01A7
)

type PetCaptureResult struct {
	Success bool
}

type PetCaptureStart struct{}

type PetEggList struct {
	Indexes []uint16
}

func ParsePetCaptureStart(packet Packet) (PetCaptureStart, bool, error) {
	if packet.ID != PacketZCStartCapture {
		return PetCaptureStart{}, false, nil
	}
	if len(packet.Data) < 2 {
		return PetCaptureStart{}, false, fmt.Errorf("ZC_START_CAPTURE too short: %d", len(packet.Data))
	}
	return PetCaptureStart{}, true, nil
}

func ParsePetCaptureResult(packet Packet) (PetCaptureResult, bool, error) {
	if packet.ID != PacketZCTryCaptureMonster {
		return PetCaptureResult{}, false, nil
	}
	if len(packet.Data) < 3 {
		return PetCaptureResult{}, false, fmt.Errorf("ZC_TRYCAPTURE_MONSTER too short: %d", len(packet.Data))
	}
	return PetCaptureResult{Success: packet.Data[2] != 0}, true, nil
}

func ParsePetEggList(packet Packet) (PetEggList, bool, error) {
	if packet.ID != PacketZCPetEggList {
		return PetEggList{}, false, nil
	}
	if len(packet.Data) < 4 {
		return PetEggList{}, false, fmt.Errorf("ZC_PETEGG_LIST too short: %d", len(packet.Data))
	}
	packetLen := int(binary.LittleEndian.Uint16(packet.Data[2:4]))
	if packetLen <= 0 || packetLen > len(packet.Data) {
		packetLen = len(packet.Data)
	}
	body := packet.Data[4:packetLen]
	if len(body)%2 != 0 {
		return PetEggList{}, false, fmt.Errorf("ZC_PETEGG_LIST bad body len: %d", len(body))
	}
	indexes := make([]uint16, 0, len(body)/2)
	for offset := 0; offset < len(body); offset += 2 {
		indexes = append(indexes, binary.LittleEndian.Uint16(body[offset:offset+2]))
	}
	return PetEggList{Indexes: indexes}, true, nil
}

func BuildTryCaptureMonsterPacket(targetID uint32) []byte {
	packet := make([]byte, 6)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZTryCaptureMonster)
	binary.LittleEndian.PutUint32(packet[2:6], targetID)
	return packet
}

func BuildSelectPetEggPacket(index uint16) []byte {
	packet := make([]byte, 4)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZSelectPetEgg)
	binary.LittleEndian.PutUint16(packet[2:4], index)
	return packet
}

func (c *Client) SendTryCaptureMonster(targetID uint32) error {
	packet := BuildTryCaptureMonsterPacket(targetID)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_TRYCAPTURE_MONSTER opcode=0x%04X target=%d client_date=%d", ID(packet), targetID, c.clientDate)
	} else {
		log.Printf("send CZ_TRYCAPTURE_MONSTER failed opcode=0x%04X len=%d target=%d client_date=%d: %v", ID(packet), len(packet), targetID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendSelectPetEgg(index uint16) error {
	packet := BuildSelectPetEggPacket(index)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_SELECT_PETEGG opcode=0x%04X index=%d client_date=%d", ID(packet), index, c.clientDate)
	} else {
		log.Printf("send CZ_SELECT_PETEGG failed opcode=0x%04X len=%d index=%d client_date=%d: %v", ID(packet), len(packet), index, c.clientDate, err)
	}
	return err
}
