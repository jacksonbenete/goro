package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

const skillInfoEntryLen = 37

type SkillInfo struct {
	ID         uint16
	Type       uint32
	Level      int
	SPCost     int
	Range      int
	Name       string
	Upgradable bool
}

type SkillInfoList struct {
	Skills []SkillInfo
}

type SkillInfoUpdate struct {
	Skill SkillInfo
}

func ParseSkillInfoList(packet Packet) (SkillInfoList, bool, error) {
	if packet.ID != 0x010F {
		return SkillInfoList{}, false, nil
	}
	if len(packet.Data) < 4 {
		return SkillInfoList{}, false, fmt.Errorf("ZC_SKILLINFO_LIST too short: %d", len(packet.Data))
	}
	packetLen := int(binary.LittleEndian.Uint16(packet.Data[2:4]))
	if packetLen <= 0 || packetLen > len(packet.Data) {
		packetLen = len(packet.Data)
	}
	body := packet.Data[4:packetLen]
	if len(body)%skillInfoEntryLen != 0 {
		return SkillInfoList{}, false, fmt.Errorf("ZC_SKILLINFO_LIST bad body len: %d", len(body))
	}
	skills := make([]SkillInfo, 0, len(body)/skillInfoEntryLen)
	for offset := 0; offset < len(body); offset += skillInfoEntryLen {
		skills = append(skills, parseSkillInfoEntry(body[offset:offset+skillInfoEntryLen], 0))
	}
	return SkillInfoList{Skills: skills}, true, nil
}

func ParseSkillInfoUpdate(packet Packet) (SkillInfoUpdate, bool, error) {
	switch packet.ID {
	case 0x010E:
		if len(packet.Data) < 11 {
			return SkillInfoUpdate{}, false, fmt.Errorf("ZC_SKILLINFO_UPDATE too short: %d", len(packet.Data))
		}
		return SkillInfoUpdate{Skill: SkillInfo{
			ID:         binary.LittleEndian.Uint16(packet.Data[2:4]),
			Level:      int(binary.LittleEndian.Uint16(packet.Data[4:6])),
			SPCost:     int(binary.LittleEndian.Uint16(packet.Data[6:8])),
			Range:      int(binary.LittleEndian.Uint16(packet.Data[8:10])),
			Upgradable: packet.Data[10] != 0,
		}}, true, nil
	case 0x0111:
		if len(packet.Data) < 39 {
			return SkillInfoUpdate{}, false, fmt.Errorf("ZC_ADD_SKILL too short: %d", len(packet.Data))
		}
		return SkillInfoUpdate{Skill: parseSkillInfoEntry(packet.Data[2:39], 0)}, true, nil
	default:
		return SkillInfoUpdate{}, false, nil
	}
}

func BuildSkillLevelUpPacket(skillID uint16) []byte {
	packet := make([]byte, 4)
	binary.LittleEndian.PutUint16(packet[0:2], 0x0112)
	binary.LittleEndian.PutUint16(packet[2:4], skillID)
	return packet
}

func parseSkillInfoEntry(data []byte, offset int) SkillInfo {
	return SkillInfo{
		ID:         binary.LittleEndian.Uint16(data[offset : offset+2]),
		Type:       binary.LittleEndian.Uint32(data[offset+2 : offset+6]),
		Level:      int(binary.LittleEndian.Uint16(data[offset+6 : offset+8])),
		SPCost:     int(binary.LittleEndian.Uint16(data[offset+8 : offset+10])),
		Range:      int(binary.LittleEndian.Uint16(data[offset+10 : offset+12])),
		Name:       decodeROFixedString(data[offset+12 : offset+36]),
		Upgradable: data[offset+36] != 0,
	}
}

func decodeROFixedString(data []byte) string {
	if end := bytes.IndexByte(data, 0); end >= 0 {
		data = data[:end]
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return ""
	}
	decoded, _, err := transform.Bytes(korean.EUCKR.NewDecoder(), data)
	if err != nil {
		return string(data)
	}
	return strings.TrimSpace(string(decoded))
}
