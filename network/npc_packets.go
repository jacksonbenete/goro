package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type NPCDialogKind int

const (
	NPCDialogSay NPCDialogKind = iota
	NPCDialogNext
	NPCDialogClose
	NPCDialogMenu
	NPCDialogClear
)

type NPCDialog struct {
	Kind    NPCDialogKind
	NPCID   uint32
	Message string
	Options []string
}

func ParseNPCDialog(packet Packet) (NPCDialog, bool, error) {
	switch packet.ID {
	case 0x00B4:
		if len(packet.Data) < 8 {
			return NPCDialog{}, true, fmt.Errorf("ZC_SAY_DIALOG too short: %d", len(packet.Data))
		}
		return NPCDialog{
			Kind:    NPCDialogSay,
			NPCID:   binary.LittleEndian.Uint32(packet.Data[4:8]),
			Message: trimPacketCString(packet.Data[8:]),
		}, true, nil
	case 0x00B5:
		if len(packet.Data) < 6 {
			return NPCDialog{}, true, fmt.Errorf("ZC_WAIT_DIALOG too short: %d", len(packet.Data))
		}
		return NPCDialog{Kind: NPCDialogNext, NPCID: binary.LittleEndian.Uint32(packet.Data[2:6])}, true, nil
	case 0x00B6:
		if len(packet.Data) < 6 {
			return NPCDialog{}, true, fmt.Errorf("ZC_CLOSE_DIALOG too short: %d", len(packet.Data))
		}
		return NPCDialog{Kind: NPCDialogClose, NPCID: binary.LittleEndian.Uint32(packet.Data[2:6])}, true, nil
	case 0x00B7:
		if len(packet.Data) < 8 {
			return NPCDialog{}, true, fmt.Errorf("ZC_MENU_LIST too short: %d", len(packet.Data))
		}
		raw := trimPacketCString(packet.Data[8:])
		return NPCDialog{
			Kind:    NPCDialogMenu,
			NPCID:   binary.LittleEndian.Uint32(packet.Data[4:8]),
			Message: raw,
			Options: splitNPCMenuOptions(raw),
		}, true, nil
	case 0x08D6:
		if len(packet.Data) < 6 {
			return NPCDialog{}, true, fmt.Errorf("ZC_CLEAR_DIALOG too short: %d", len(packet.Data))
		}
		return NPCDialog{Kind: NPCDialogClear, NPCID: binary.LittleEndian.Uint32(packet.Data[2:6])}, true, nil
	default:
		return NPCDialog{}, false, nil
	}
}

func BuildNPCContactPacket(npcID uint32, contactType uint8) []byte {
	var w Writer
	w.Uint16(0x0090)
	w.Uint32(npcID)
	w.Uint8(contactType)
	return w.Bytes()
}

func BuildNPCMenuChoicePacket(npcID uint32, choice uint8) []byte {
	var w Writer
	w.Uint16(0x00B8)
	w.Uint32(npcID)
	w.Uint8(choice)
	return w.Bytes()
}

func BuildNPCNextPacket(npcID uint32) []byte {
	var w Writer
	w.Uint16(0x00B9)
	w.Uint32(npcID)
	return w.Bytes()
}

func BuildNPCClosePacket(npcID uint32) []byte {
	var w Writer
	w.Uint16(0x0146)
	w.Uint32(npcID)
	return w.Bytes()
}

func trimPacketCString(data []byte) string {
	if i := bytes.IndexByte(data, 0); i >= 0 {
		data = data[:i]
	}
	return strings.TrimSpace(string(data))
}

func splitNPCMenuOptions(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ":")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
