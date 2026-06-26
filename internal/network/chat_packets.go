package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strings"
)

const (
	PacketCZRequestChatLegacy uint16 = 0x008C
	PacketCZRequestChat       uint16 = 0x00F3
	PacketZCNotifyChat        uint16 = 0x008D
	PacketZCNotifyPlayerChat  uint16 = 0x008E
	PacketZCBroadcast         uint16 = 0x009A
	PacketZCMsg               uint16 = 0x0291
	PacketZCMsgValue          uint16 = 0x07E2
	PacketZCMsgSkill          uint16 = 0x07E6
	PacketZCMsgColor          uint16 = 0x09CD
)

type ChatMessage struct {
	GID       uint32
	Text      string
	MessageID int
	Value     int32
	SkillID   uint16
	Color     uint32
}

func BuildGlobalChatPacket(name, message string) []byte {
	return BuildGlobalChatPacketForClientDate(name, message, 20080910)
}

func BuildGlobalChatPacketForClientDate(name, message string, clientDate int) []byte {
	payload := strings.TrimSpace(name) + " : " + strings.TrimSpace(message)
	if strings.TrimSpace(name) == "" || strings.TrimSpace(message) == "" {
		return nil
	}
	size := 4 + len([]byte(payload)) + 1
	if size > 0xffff {
		return nil
	}
	opcode := PacketCZRequestChat
	if clientDate < 20040726 {
		opcode = PacketCZRequestChatLegacy
	}
	packet := make([]byte, size)
	binary.LittleEndian.PutUint16(packet[0:2], opcode)
	binary.LittleEndian.PutUint16(packet[2:4], uint16(size))
	copy(packet[4:], []byte(payload))
	return packet
}

func ParseChatMessage(packet Packet) (ChatMessage, bool, error) {
	switch packet.ID {
	case PacketZCNotifyChat:
		if len(packet.Data) < 8 {
			return ChatMessage{}, false, fmt.Errorf("ZC_NOTIFY_CHAT too short: %d", len(packet.Data))
		}
		return ChatMessage{
			GID:  binary.LittleEndian.Uint32(packet.Data[4:8]),
			Text: packetCString(packet.Data[8:]),
		}, true, nil
	case PacketZCNotifyPlayerChat:
		if len(packet.Data) < 4 {
			return ChatMessage{}, false, fmt.Errorf("ZC_NOTIFY_PLAYERCHAT too short: %d", len(packet.Data))
		}
		return ChatMessage{
			Text: packetCString(packet.Data[4:]),
		}, true, nil
	case PacketZCBroadcast:
		if len(packet.Data) < 4 {
			return ChatMessage{}, false, fmt.Errorf("ZC_BROADCAST too short: %d", len(packet.Data))
		}
		return ChatMessage{
			Text: packetCString(packet.Data[4:]),
		}, true, nil
	case PacketZCMsg:
		if len(packet.Data) < 4 {
			return ChatMessage{}, false, fmt.Errorf("ZC_MSG too short: %d", len(packet.Data))
		}
		return ChatMessage{MessageID: int(binary.LittleEndian.Uint16(packet.Data[2:4]))}, true, nil
	case PacketZCMsgValue:
		if len(packet.Data) < 8 {
			return ChatMessage{}, false, fmt.Errorf("ZC_MSG_VALUE too short: %d", len(packet.Data))
		}
		messageID := binary.LittleEndian.Uint16(packet.Data[2:4])
		value := int32(binary.LittleEndian.Uint32(packet.Data[4:8]))
		return ChatMessage{MessageID: int(messageID), Value: value}, true, nil
	case PacketZCMsgSkill:
		if len(packet.Data) < 8 {
			return ChatMessage{}, false, fmt.Errorf("ZC_MSG_SKILL too short: %d", len(packet.Data))
		}
		skillID := binary.LittleEndian.Uint16(packet.Data[2:4])
		messageID := int32(binary.LittleEndian.Uint32(packet.Data[4:8]))
		return ChatMessage{MessageID: int(messageID), SkillID: skillID}, true, nil
	case PacketZCMsgColor:
		if len(packet.Data) < 8 {
			return ChatMessage{}, false, fmt.Errorf("ZC_MSG_COLOR too short: %d", len(packet.Data))
		}
		messageID := binary.LittleEndian.Uint16(packet.Data[2:4])
		color := binary.LittleEndian.Uint32(packet.Data[4:8])
		return ChatMessage{MessageID: int(messageID), Color: color}, true, nil
	default:
		return ChatMessage{}, false, nil
	}
}

func packetCString(data []byte) string {
	if i := bytes.IndexByte(data, 0); i >= 0 {
		data = data[:i]
	}
	return strings.TrimSpace(string(data))
}

func (c *Client) SendGlobalChat(name, message string) error {
	packet := BuildGlobalChatPacketForClientDate(name, message, c.clientDate)
	if len(packet) == 0 {
		return fmt.Errorf("empty chat message")
	}
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_REQUEST_CHAT opcode=0x%04X len=%d name=%q client_date=%d", ID(packet), len(packet), name, c.clientDate)
	} else {
		log.Printf("send CZ_REQUEST_CHAT failed opcode=0x%04X len=%d name=%q client_date=%d: %v", ID(packet), len(packet), name, c.clientDate, err)
	}
	return err
}
