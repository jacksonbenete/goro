package network

import "fmt"

const PacketZCRestartAck uint16 = 0x00B3

type RestartAck struct {
	Allowed bool
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
