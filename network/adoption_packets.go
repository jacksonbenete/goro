package network

import (
	"encoding/binary"
	"fmt"

	"github.com/kivutar/goro/glog"
)

const (
	PacketZCReqBaby   uint16 = 0x01F6
	PacketCZJoinBaby  uint16 = 0x01F7
	PacketZCStartBaby uint16 = 0x01F8
	PacketCZReqBaby   uint16 = 0x01F9
	PacketZCBabyMsg   uint16 = 0x0216
)

type AdoptionRequest struct {
	FatherAccountID uint32
	MotherAccountID uint32
	FatherName      string
}

type AdoptionMessage struct {
	Code uint32
}

func ParseAdoptionRequest(packet Packet) (AdoptionRequest, bool, error) {
	if packet.ID != PacketZCReqBaby {
		return AdoptionRequest{}, false, nil
	}
	if len(packet.Data) < 34 {
		return AdoptionRequest{}, true, fmt.Errorf("ZC_REQ_BABY too short: %d", len(packet.Data))
	}
	return AdoptionRequest{
		FatherAccountID: binary.LittleEndian.Uint32(packet.Data[2:6]),
		MotherAccountID: binary.LittleEndian.Uint32(packet.Data[6:10]),
		FatherName:      fixedPacketString(packet.Data[10:34]),
	}, true, nil
}

func ParseAdoptionStarted(packet Packet) bool {
	return packet.ID == PacketZCStartBaby
}

func ParseAdoptionMessage(packet Packet) (AdoptionMessage, bool, error) {
	if packet.ID != PacketZCBabyMsg {
		return AdoptionMessage{}, false, nil
	}
	if len(packet.Data) < 6 {
		return AdoptionMessage{}, true, fmt.Errorf("ZC_BABYMSG too short: %d", len(packet.Data))
	}
	return AdoptionMessage{Code: binary.LittleEndian.Uint32(packet.Data[2:6])}, true, nil
}

func BuildAdoptionRequestPacket(targetAccountID uint32) []byte {
	var w Writer
	w.Uint16(PacketCZReqBaby)
	w.Uint32(targetAccountID)
	return w.Bytes()
}

func BuildAdoptionReplyPacket(fatherAccountID, motherAccountID uint32, accepted bool) []byte {
	var w Writer
	w.Uint16(PacketCZJoinBaby)
	w.Uint32(fatherAccountID)
	w.Uint32(motherAccountID)
	if accepted {
		w.Uint32(1)
	} else {
		w.Uint32(0)
	}
	return w.Bytes()
}

func (c *Client) SendAdoptionRequest(targetAccountID uint32) error {
	packet := BuildAdoptionRequestPacket(targetAccountID)
	err := c.Send(packet)
	if err == nil {
		glog.Debugf("sent CZ_REQ_JOIN_BABY opcode=0x%04X target=%d client_date=%d", ID(packet), targetAccountID, c.clientDate)
	} else {
		glog.Warnf("send CZ_REQ_JOIN_BABY failed opcode=0x%04X target=%d client_date=%d: %v", ID(packet), targetAccountID, c.clientDate, err)
	}
	return err
}

func (c *Client) SendAdoptionReply(fatherAccountID, motherAccountID uint32, accepted bool) error {
	packet := BuildAdoptionReplyPacket(fatherAccountID, motherAccountID, accepted)
	err := c.Send(packet)
	if err == nil {
		glog.Debugf("sent CZ_JOIN_BABY opcode=0x%04X father=%d mother=%d accepted=%t client_date=%d", ID(packet), fatherAccountID, motherAccountID, accepted, c.clientDate)
	} else {
		glog.Warnf("send CZ_JOIN_BABY failed opcode=0x%04X father=%d mother=%d accepted=%t client_date=%d: %v", ID(packet), fatherAccountID, motherAccountID, accepted, c.clientDate, err)
	}
	return err
}
