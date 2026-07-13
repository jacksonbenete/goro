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
	PacketCZWhisper           uint16 = 0x0096
	PacketCZWhisperIgnore     uint16 = 0x00CF
	PacketCZWhisperIgnoreAll  uint16 = 0x00D0
	PacketZCNotifyChat        uint16 = 0x008D
	PacketZCNotifyPlayerChat  uint16 = 0x008E
	PacketZCWhisper           uint16 = 0x0097
	PacketZCAckWhisper        uint16 = 0x0098
	PacketZCWhisperIgnoreAck  uint16 = 0x00D1
	PacketZCWhisperAllAck     uint16 = 0x00D2
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

type WhisperMessage struct {
	Sender  string
	Message string
}

type WhisperAck struct {
	Result uint8
}

type WhisperIgnoreAck struct {
	TargetAll bool
	Allow     bool
	Result    uint8
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

func BuildWhisperPacket(receiver, message string) []byte {
	receiver = strings.TrimSpace(receiver)
	message = strings.TrimSpace(message)
	if receiver == "" || message == "" {
		return nil
	}
	size := 2 + 2 + 24 + len([]byte(message)) + 1
	if size > 0xffff {
		return nil
	}
	packet := make([]byte, size)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZWhisper)
	binary.LittleEndian.PutUint16(packet[2:4], uint16(size))
	writeFixedCString(packet[4:28], receiver)
	copy(packet[28:], []byte(message))
	return packet
}

func BuildWhisperIgnorePacket(name string, allow bool) []byte {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	packet := make([]byte, 27)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZWhisperIgnore)
	writeFixedName(packet[2:26], name)
	packet[26] = whisperAllowByte(allow)
	return packet
}

func BuildWhisperIgnoreAllPacket(allow bool) []byte {
	return []byte{byte(PacketCZWhisperIgnoreAll), byte(PacketCZWhisperIgnoreAll >> 8), whisperAllowByte(allow)}
}

func whisperAllowByte(allow bool) byte {
	if allow {
		return 1
	}
	return 0
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

func ParseWhisperMessage(packet Packet) (WhisperMessage, bool, error) {
	if packet.ID != PacketZCWhisper {
		return WhisperMessage{}, false, nil
	}
	if len(packet.Data) < 28 {
		return WhisperMessage{}, false, fmt.Errorf("ZC_WHISPER too short: %d", len(packet.Data))
	}
	return WhisperMessage{
		Sender:  packetCString(packet.Data[4:28]),
		Message: packetCString(packet.Data[28:]),
	}, true, nil
}

func ParseWhisperAck(packet Packet) (WhisperAck, bool, error) {
	if packet.ID != PacketZCAckWhisper {
		return WhisperAck{}, false, nil
	}
	if len(packet.Data) < 3 {
		return WhisperAck{}, false, fmt.Errorf("ZC_ACK_WHISPER too short: %d", len(packet.Data))
	}
	return WhisperAck{Result: packet.Data[2]}, true, nil
}

func ParseWhisperIgnoreAck(packet Packet) (WhisperIgnoreAck, bool, error) {
	switch packet.ID {
	case PacketZCWhisperIgnoreAck:
		if len(packet.Data) < 4 {
			return WhisperIgnoreAck{}, false, fmt.Errorf("ZC_SETTING_WHISPER_PC too short: %d", len(packet.Data))
		}
		return WhisperIgnoreAck{
			Allow:  packet.Data[2] != 0,
			Result: packet.Data[3],
		}, true, nil
	case PacketZCWhisperAllAck:
		if len(packet.Data) < 4 {
			return WhisperIgnoreAck{}, false, fmt.Errorf("ZC_SETTING_WHISPER_STATE too short: %d", len(packet.Data))
		}
		return WhisperIgnoreAck{
			TargetAll: true,
			Allow:     packet.Data[2] != 0,
			Result:    packet.Data[3],
		}, true, nil
	default:
		return WhisperIgnoreAck{}, false, nil
	}
}

func packetCString(data []byte) string {
	if i := bytes.IndexByte(data, 0); i >= 0 {
		data = data[:i]
	}
	return strings.TrimSpace(string(data))
}

func writeFixedCString(dst []byte, value string) {
	src := []byte(value)
	copy(dst, src)
	if len(src) < len(dst) {
		dst[len(src)] = 0
	}
}

func writeFixedName(dst []byte, value string) {
	if len(dst) == 0 {
		return
	}
	src := []byte(value)
	if len(src) >= len(dst) {
		src = src[:len(dst)-1]
	}
	copy(dst, src)
	dst[len(src)] = 0
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

func (c *Client) SendWhisper(receiver, message string) error {
	packet := BuildWhisperPacket(receiver, message)
	if len(packet) == 0 {
		return fmt.Errorf("empty whisper")
	}
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_WHISPER opcode=0x%04X len=%d receiver=%q client_date=%d", ID(packet), len(packet), receiver, c.clientDate)
	} else {
		log.Printf("send CZ_WHISPER failed opcode=0x%04X len=%d receiver=%q client_date=%d: %v", ID(packet), len(packet), receiver, c.clientDate, err)
	}
	return err
}

func (c *Client) SendWhisperIgnore(name string, allow bool) error {
	packet := BuildWhisperIgnorePacket(name, allow)
	if len(packet) == 0 {
		return fmt.Errorf("empty whisper ignore target")
	}
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_SETTING_WHISPER_PC opcode=0x%04X name=%q allow=%t client_date=%d", ID(packet), name, allow, c.clientDate)
	} else {
		log.Printf("send CZ_SETTING_WHISPER_PC failed opcode=0x%04X name=%q allow=%t client_date=%d: %v", ID(packet), name, allow, c.clientDate, err)
	}
	return err
}

func (c *Client) SendWhisperIgnoreAll(allow bool) error {
	packet := BuildWhisperIgnoreAllPacket(allow)
	err := c.Send(packet)
	if err == nil {
		log.Printf("sent CZ_SETTING_WHISPER_STATE opcode=0x%04X allow=%t client_date=%d", ID(packet), allow, c.clientDate)
	} else {
		log.Printf("send CZ_SETTING_WHISPER_STATE failed opcode=0x%04X allow=%t client_date=%d: %v", ID(packet), allow, c.clientDate, err)
	}
	return err
}
