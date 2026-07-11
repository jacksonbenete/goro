package network

import (
	"encoding/binary"
	"testing"
)

func TestBuildTradePackets(t *testing.T) {
	request := BuildTradeRequestPacket(1234)
	if len(request) != 6 || ID(request) != PacketCZReqExchangeItem || binary.LittleEndian.Uint32(request[2:6]) != 1234 {
		t.Fatalf("bad trade request packet: % x", request)
	}
	ack := BuildTradeAckPacket(true)
	if len(ack) != 3 || ID(ack) != PacketCZAckExchangeItem || ack[2] != TradeAckAccept {
		t.Fatalf("bad trade ack packet: % x", ack)
	}
	reject := BuildTradeAckPacket(false)
	if reject[2] != TradeAckCancel {
		t.Fatalf("bad trade reject packet: % x", reject)
	}
	add := BuildTradeAddItemPacket(7, 42)
	if len(add) != 8 || ID(add) != PacketCZAddExchangeItem || binary.LittleEndian.Uint16(add[2:4]) != 7 || binary.LittleEndian.Uint32(add[4:8]) != 42 {
		t.Fatalf("bad trade add packet: % x", add)
	}
	if len(BuildTradeConcludePacket()) != 2 || ID(BuildTradeConcludePacket()) != PacketCZConcludeExchangeItem {
		t.Fatalf("bad trade conclude packet")
	}
	if len(BuildTradeCancelPacket()) != 2 || ID(BuildTradeCancelPacket()) != PacketCZCancelExchangeItem {
		t.Fatalf("bad trade cancel packet")
	}
	if len(BuildTradeCommitPacket()) != 2 || ID(BuildTradeCommitPacket()) != PacketCZExecExchangeItem {
		t.Fatalf("bad trade commit packet")
	}
}

func TestParseTradePackets(t *testing.T) {
	request := make([]byte, 32)
	binary.LittleEndian.PutUint16(request[0:2], PacketZCReqExchangeItem2)
	copy(request[2:26], []byte("Mira"))
	binary.LittleEndian.PutUint32(request[26:30], 2000001)
	binary.LittleEndian.PutUint16(request[30:32], 42)
	parsedRequest, ok, err := ParseTradeRequest(Packet{ID: PacketZCReqExchangeItem2, Data: request})
	if !ok || err != nil || parsedRequest.Name != "Mira" || parsedRequest.TargetID != 2000001 || parsedRequest.Level != 42 {
		t.Fatalf("ParseTradeRequest = %+v ok=%t err=%v", parsedRequest, ok, err)
	}

	response := []byte{0xF5, 0x01, 3, 0, 0, 0, 0, 7, 0}
	binary.LittleEndian.PutUint32(response[3:7], 2000002)
	parsedResponse, ok, err := ParseTradeResponse(Packet{ID: PacketZCAckExchangeItem2, Data: response})
	if !ok || err != nil || parsedResponse.Result != 3 || parsedResponse.TargetID != 2000002 || parsedResponse.Level != 7 {
		t.Fatalf("ParseTradeResponse = %+v ok=%t err=%v", parsedResponse, ok, err)
	}

	item := make([]byte, 19)
	binary.LittleEndian.PutUint16(item[0:2], PacketZCAddExchangeItem)
	binary.LittleEndian.PutUint32(item[2:6], 5)
	binary.LittleEndian.PutUint16(item[6:8], 512)
	item[8] = 1
	item[10] = 2
	binary.LittleEndian.PutUint16(item[11:13], 4001)
	parsedItem, ok, err := ParseTradeItem(Packet{ID: PacketZCAddExchangeItem, Data: item})
	if !ok || err != nil || parsedItem.Amount != 5 || parsedItem.ItemID != 512 || !parsedItem.Identified || parsedItem.Refine != 2 || parsedItem.Cards[0] != 4001 {
		t.Fatalf("ParseTradeItem = %+v ok=%t err=%v", parsedItem, ok, err)
	}

	addAck := []byte{0xEA, 0x00, 9, 0, 0}
	parsedAck, ok, err := ParseTradeAddItemAck(Packet{ID: PacketZCAckAddExchangeItem, Data: addAck})
	if !ok || err != nil || parsedAck.Index != 9 || parsedAck.Result != 0 {
		t.Fatalf("ParseTradeAddItemAck = %+v ok=%t err=%v", parsedAck, ok, err)
	}

	conclude, ok, err := ParseTradeConclude(Packet{ID: PacketZCConcludeExchangeItem, Data: []byte{0xEC, 0x00, 1}})
	if !ok || err != nil || !conclude.Other {
		t.Fatalf("ParseTradeConclude = %+v ok=%t err=%v", conclude, ok, err)
	}

	exec, ok, err := ParseTradeExec(Packet{ID: PacketZCExecExchangeItem, Data: []byte{0xF0, 0x00, 0}})
	if !ok || err != nil || exec.Result != 0 {
		t.Fatalf("ParseTradeExec = %+v ok=%t err=%v", exec, ok, err)
	}
}
