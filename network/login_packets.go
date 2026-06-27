package network

import (
	"encoding/binary"
	"fmt"
	"time"
)

const (
	PacketCAPlainLogin     uint16 = 0x0064
	PacketCAEnter          uint16 = 0x0065
	PacketCZSelectChar     uint16 = 0x0066
	PacketCZLoadEndAck     uint16 = 0x007D
	PacketCZTickSend       uint16 = 0x0089
	PacketCZTickSendRE     uint16 = 0x0360
	PacketCZReqName        uint16 = 0x0094
	PacketCZReqNameRE      uint16 = 0x0368
	PacketCZWalkToXY       uint16 = 0x00A7
	PacketCZWalkToXYRE     uint16 = 0x035F
	PacketCZRestart        uint16 = 0x00B2
	PacketCZRequestAct     uint16 = 0x0190
	PacketCZRequestAct2    uint16 = 0x0437
	PacketCZRequestAct2012 uint16 = 0x0369
	PacketCZEnter2         uint16 = 0x0436
)

type nameRequestPacketLayout struct {
	date   int
	opcode uint16
	length int
	offset int
}

var nameRequestPacketLayouts = []nameRequestPacketLayout{
	// Keep this table aligned with rAthena's active packetdb branch. Our
	// default 20080910 pre-renewal server uses PACKETVER_MAIN_NUM, not the
	// 20080827 RagexeRE block from roBrowser's cumulative table.
	{date: 20101124, opcode: PacketCZReqNameRE, length: 6, offset: 2},
	{date: 20070212, opcode: 0x008C, length: 11, offset: 7},
}

const (
	ActionAttack uint8 = 7
)

const (
	RestartTypeRespawn    uint8 = 0
	RestartTypeCharSelect uint8 = 1
)

type AccountLogin struct {
	Version    uint32
	Username   string
	Password   string
	ClientType uint8
}

func BuildAccountLoginPacket(login AccountLogin) []byte {
	var w Writer
	w.Uint16(PacketCAPlainLogin)
	w.Uint32(login.Version)
	w.CString(login.Username, 24)
	w.CString(login.Password, 24)
	w.Uint8(login.ClientType)
	return w.Bytes()
}

type CharServerEnter struct {
	AccountID uint32
	AuthCode  uint32
	UserLevel uint32
	Sex       uint8
}

func BuildCharServerEnterPacket(enter CharServerEnter) []byte {
	var w Writer
	w.Uint16(PacketCAEnter)
	w.Uint32(enter.AccountID)
	w.Uint32(enter.AuthCode)
	w.Uint32(enter.UserLevel)
	w.Uint16(0)
	w.Uint8(enter.Sex)
	return w.Bytes()
}

func BuildSelectCharacterPacket(slot uint8) []byte {
	var w Writer
	w.Uint16(PacketCZSelectChar)
	w.Uint8(slot)
	return w.Bytes()
}

func BuildLoadEndAckPacket() []byte {
	var w Writer
	w.Uint16(PacketCZLoadEndAck)
	return w.Bytes()
}

func BuildRestartPacket(restartType uint8) []byte {
	var w Writer
	w.Uint16(PacketCZRestart)
	w.Uint8(restartType)
	return w.Bytes()
}

func BuildTickSendPacket(clientTick uint32) []byte {
	return BuildTickSendPacketForClientDate(clientTick, 20080910)
}

func BuildTickSendPacketForClientDate(clientTick uint32, clientDate int) []byte {
	var w Writer
	if clientDate > 20180307 {
		w.Uint16(PacketCZTickSendRE)
		w.Uint32(clientTick)
	} else {
		w.Uint16(PacketCZTickSend)
		w.Uint8(0)
		w.Uint8(0)
		w.Uint32(clientTick)
	}
	return w.Bytes()
}

func BuildNameRequestPacket(gid uint32) []byte {
	var w Writer
	w.Uint16(PacketCZReqName)
	w.Uint32(gid)
	return w.Bytes()
}

func BuildNameRequestPacketForClientDate(gid uint32, clientDate int) ([]byte, bool) {
	for _, layout := range nameRequestPacketLayouts {
		if clientDate >= layout.date {
			packet := make([]byte, layout.length)
			binary.LittleEndian.PutUint16(packet[0:2], layout.opcode)
			binary.LittleEndian.PutUint32(packet[layout.offset:layout.offset+4], gid)
			return packet, true
		}
	}
	return nil, false
}

type MapServerEnter struct {
	AccountID  uint32
	CharID     uint32
	AuthCode   uint32
	ClientTick uint32
	Sex        uint8
}

func BuildMapServerEnterPacket(enter MapServerEnter) []byte {
	return BuildMapServerEnterPacketForClientDate(enter, 20080910)
}

func BuildMapServerEnterPacketForClientDate(enter MapServerEnter, clientDate int) []byte {
	var w Writer
	w.Uint16(PacketCZEnter2)
	w.Uint32(enter.AccountID)
	w.Uint32(enter.CharID)
	w.Uint32(enter.AuthCode)
	w.Uint32(enter.ClientTick)
	if clientDate >= 20211103 {
		w.Uint32(0)
	}
	w.Uint8(enter.Sex)
	return w.Bytes()
}

func BuildWalkToXYPacket(x, y int) ([]byte, bool) {
	return BuildWalkToXYPacketForClientDate(x, y, 20080910)
}

func BuildWalkToXYPacketForClientDate(x, y, clientDate int) ([]byte, bool) {
	dest, ok := EncodeMoveDestination(x, y)
	if !ok {
		return nil, false
	}

	var w Writer
	if clientDate >= 20211103 {
		w.Uint16(PacketCZWalkToXYRE)
	} else {
		w.Uint16(PacketCZWalkToXY)
		w.Uint8(0)
		w.Uint8(0)
		w.Uint8(0)
	}
	_, _ = w.Write(dest[:])
	return w.Bytes(), true
}

func BuildActionRequestPacket(targetGID uint32, action uint8) []byte {
	return BuildActionRequestPacketForClientDate(targetGID, action, 20080910)
}

func BuildActionRequestPacketForClientDate(targetGID uint32, action uint8, clientDate int) []byte {
	switch {
	case clientDate >= 20211103:
		return buildCompactActionRequestPacket(PacketCZRequestAct2, targetGID, action)
	case clientDate >= 20120410:
		return buildCompactActionRequestPacket(PacketCZRequestAct2012, targetGID, action)
	default:
		return buildLegacyActionRequestPacket(targetGID, action)
	}
}

func buildCompactActionRequestPacket(opcode uint16, targetGID uint32, action uint8) []byte {
	var w Writer
	w.Uint16(opcode)
	w.Uint32(targetGID)
	w.Uint8(action)
	return w.Bytes()
}

func buildLegacyActionRequestPacket(targetGID uint32, action uint8) []byte {
	packet := make([]byte, 19)
	binary.LittleEndian.PutUint16(packet[0:2], PacketCZRequestAct)
	fillLegacyPacketPadding(packet[2:5])
	binary.LittleEndian.PutUint32(packet[5:9], targetGID)
	fillLegacyPacketPadding(packet[9:18])
	packet[18] = action
	return packet
}

func fillLegacyPacketPadding(pad []byte) {
	if len(pad) == 0 {
		return
	}
	now := uint32(time.Now().UnixMilli())
	text := fmt.Sprintf("%x", now^(now>>uint(len(pad))))
	if text == "" {
		text = "0"
	}
	src := len(text) - 1
	fallback := text[len(text)-1]
	for i := range pad {
		if src >= 0 {
			pad[i] = text[src]
			src--
		} else {
			pad[i] = fallback
		}
	}
	pad[len(pad)-1] = 0
}

func EncodeMoveDestination(x, y int) ([3]byte, bool) {
	if x < 0 || x > 1023 || y < 0 || y > 1023 {
		return [3]byte{}, false
	}
	return [3]byte{
		byte((x >> 2) & 0xff),
		byte(((x & 0x03) << 6) | ((y >> 4) & 0x3f)),
		byte((y & 0x0f) << 4),
	}, true
}
