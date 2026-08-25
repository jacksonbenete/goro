package network

import (
	"encoding/binary"
	"testing"
)

func TestBuildAdoptionPackets(t *testing.T) {
	request := BuildAdoptionRequestPacket(2000002)
	if len(request) != 6 || ID(request) != PacketCZReqBaby || binary.LittleEndian.Uint32(request[2:6]) != 2000002 {
		t.Fatalf("bad adoption request packet: % x", request)
	}

	reply := BuildAdoptionReplyPacket(2000001, 2000000, true)
	if len(reply) != 14 || ID(reply) != PacketCZJoinBaby {
		t.Fatalf("bad adoption reply packet: % x", reply)
	}
	if got := binary.LittleEndian.Uint32(reply[2:6]); got != 2000001 {
		t.Fatalf("father account id = %d, want 2000001", got)
	}
	if got := binary.LittleEndian.Uint32(reply[6:10]); got != 2000000 {
		t.Fatalf("mother account id = %d, want 2000000", got)
	}
	if got := binary.LittleEndian.Uint32(reply[10:14]); got != 1 {
		t.Fatalf("accepted = %d, want 1", got)
	}
	if got := binary.LittleEndian.Uint32(BuildAdoptionReplyPacket(1, 2, false)[10:14]); got != 0 {
		t.Fatalf("rejected = %d, want 0", got)
	}
}

func TestParseAdoptionPackets(t *testing.T) {
	request := make([]byte, 34)
	binary.LittleEndian.PutUint16(request[0:2], PacketZCReqBaby)
	binary.LittleEndian.PutUint32(request[2:6], 2000001)
	binary.LittleEndian.PutUint32(request[6:10], 2000000)
	copy(request[10:34], "Zambla")
	parsed, ok, err := ParseAdoptionRequest(Packet{ID: PacketZCReqBaby, Data: request})
	if !ok || err != nil {
		t.Fatalf("ParseAdoptionRequest ok=%t err=%v", ok, err)
	}
	if parsed.FatherAccountID != 2000001 || parsed.MotherAccountID != 2000000 || parsed.FatherName != "Zambla" {
		t.Fatalf("ParseAdoptionRequest = %+v", parsed)
	}

	message := make([]byte, 6)
	binary.LittleEndian.PutUint16(message[0:2], PacketZCBabyMsg)
	binary.LittleEndian.PutUint32(message[2:6], 1)
	parsedMessage, ok, err := ParseAdoptionMessage(Packet{ID: PacketZCBabyMsg, Data: message})
	if !ok || err != nil || parsedMessage.Code != 1 {
		t.Fatalf("ParseAdoptionMessage = %+v ok=%t err=%v", parsedMessage, ok, err)
	}

	if !ParseAdoptionStarted(Packet{ID: PacketZCStartBaby}) {
		t.Fatal("ZC_START_BABY was not recognized")
	}
}

func TestParseAdoptionPacketsRejectShortPayloads(t *testing.T) {
	if _, ok, err := ParseAdoptionRequest(Packet{ID: PacketZCReqBaby, Data: make([]byte, 33)}); !ok || err == nil {
		t.Fatalf("short adoption request ok=%t err=%v", ok, err)
	}
	if _, ok, err := ParseAdoptionMessage(Packet{ID: PacketZCBabyMsg, Data: make([]byte, 5)}); !ok || err == nil {
		t.Fatalf("short adoption message ok=%t err=%v", ok, err)
	}
}
