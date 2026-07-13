package network

import (
	"encoding/binary"
	"fmt"
)

const PacketZCRestartAck uint16 = 0x00B3
const PacketZCQuitGameAck uint16 = 0x018B

type RestartAck struct {
	Allowed bool
}

type QuitGameAck struct {
	Allowed bool
	Result  uint16
}

func ParseRestartAck(packet Packet) (RestartAck, bool, error) {
	if packet.ID != PacketZCRestartAck {
		return RestartAck{}, false, nil
	}
	if len(packet.Data) < 3 {
		return RestartAck{}, false, fmt.Errorf("ZC_RESTART_ACK too short: %d", len(packet.Data))
	}
	return RestartAck{Allowed: packet.Data[2] != 0}, true, nil
}

func ParseQuitGameAck(packet Packet) (QuitGameAck, bool, error) {
	if packet.ID != PacketZCQuitGameAck {
		return QuitGameAck{}, false, nil
	}
	if len(packet.Data) < 4 {
		return QuitGameAck{}, false, fmt.Errorf("ZC_ACK_REQ_DISCONNECT too short: %d", len(packet.Data))
	}
	result := binary.LittleEndian.Uint16(packet.Data[2:4])
	return QuitGameAck{Allowed: result == 0, Result: result}, true, nil
}
