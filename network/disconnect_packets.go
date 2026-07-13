package network

import "fmt"

const (
	PacketSCNotifyBan      uint16 = 0x0081
	PacketZCISVRDisconnect uint16 = 0x02D5
)

type NotifyBan struct {
	ErrorCode uint8
}

func ParseNotifyBan(packet Packet) (NotifyBan, bool, error) {
	if packet.ID != PacketSCNotifyBan {
		return NotifyBan{}, false, nil
	}
	if len(packet.Data) < 3 {
		return NotifyBan{}, true, fmt.Errorf("SC_NOTIFY_BAN too short: %d", len(packet.Data))
	}
	return NotifyBan{ErrorCode: packet.Data[2]}, true, nil
}

func IsInterServerDisconnect(packet Packet) bool {
	return packet.ID == PacketZCISVRDisconnect
}
