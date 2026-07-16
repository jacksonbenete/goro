package network

import (
	"encoding/binary"
	"testing"
)

func TestPetCapturePackets(t *testing.T) {
	_, ok, err := ParsePetCaptureStart(Packet{ID: PacketZCStartCapture, Data: []byte{0x9E, 0x01}})
	if !ok || err != nil {
		t.Fatalf("ParsePetCaptureStart ok=%t err=%v", ok, err)
	}

	result, ok, err := ParsePetCaptureResult(Packet{ID: PacketZCTryCaptureMonster, Data: []byte{0xA0, 0x01, 1}})
	if !ok || err != nil || !result.Success {
		t.Fatalf("ParsePetCaptureResult = %+v ok=%t err=%v", result, ok, err)
	}

	packet := BuildTryCaptureMonsterPacket(0x11223344)
	if got := binary.LittleEndian.Uint16(packet[0:2]); got != PacketCZTryCaptureMonster {
		t.Fatalf("opcode = 0x%04X", got)
	}
	if got := binary.LittleEndian.Uint32(packet[2:6]); got != 0x11223344 {
		t.Fatalf("target = 0x%08X", got)
	}
}

func TestPetEggPackets(t *testing.T) {
	listData := []byte{0xA6, 0x01, 0x08, 0x00, 7, 0, 9, 0}
	list, ok, err := ParsePetEggList(Packet{ID: PacketZCPetEggList, Data: listData})
	if !ok || err != nil || len(list.Indexes) != 2 || list.Indexes[0] != 7 || list.Indexes[1] != 9 {
		t.Fatalf("ParsePetEggList = %+v ok=%t err=%v", list, ok, err)
	}

	packet := BuildSelectPetEggPacket(9)
	if got := binary.LittleEndian.Uint16(packet[0:2]); got != PacketCZSelectPetEgg {
		t.Fatalf("opcode = 0x%04X", got)
	}
	if got := binary.LittleEndian.Uint16(packet[2:4]); got != 9 {
		t.Fatalf("index = %d", got)
	}
}

func TestPetCommandPackets(t *testing.T) {
	packet := BuildCommandPetPacket(PetCommandFeed)
	if got := binary.LittleEndian.Uint16(packet[0:2]); got != PacketCZCommandPet {
		t.Fatalf("opcode = 0x%04X", got)
	}
	if packet[2] != PetCommandFeed {
		t.Fatalf("command = %d", packet[2])
	}

	packet = BuildPetActPacket(112)
	if got := binary.LittleEndian.Uint16(packet[0:2]); got != PacketCZPetAct {
		t.Fatalf("opcode = 0x%04X", got)
	}
	if got := binary.LittleEndian.Uint32(packet[2:6]); got != 112 {
		t.Fatalf("data = %d", got)
	}
}

func TestPetStatusPackets(t *testing.T) {
	propertyData := make([]byte, 37)
	binary.LittleEndian.PutUint16(propertyData[0:2], PacketZCPropertyPet)
	copy(propertyData[2:26], []byte("Luna"))
	propertyData[26] = 1
	binary.LittleEndian.PutUint16(propertyData[27:29], 12)
	binary.LittleEndian.PutUint16(propertyData[29:31], 4)
	binary.LittleEndian.PutUint16(propertyData[31:33], 900)
	binary.LittleEndian.PutUint16(propertyData[33:35], 10007)
	binary.LittleEndian.PutUint16(propertyData[35:37], 1063)
	property, ok, err := ParsePetProperty(Packet{ID: PacketZCPropertyPet, Data: propertyData})
	if !ok || err != nil {
		t.Fatalf("ParsePetProperty ok=%t err=%v", ok, err)
	}
	if property.Name != "Luna" || property.Level != 12 || property.Fullness != 4 || property.Relationship != 900 || property.AccessoryID != 10007 || property.Job != 1063 {
		t.Fatalf("property = %+v", property)
	}

	feed, ok, err := ParsePetFeedResult(Packet{ID: PacketZCFeedPet, Data: []byte{0xA3, 0x01, 1, 0x11, 0x22}})
	if !ok || err != nil || !feed.Success || feed.ItemID != 0x2211 {
		t.Fatalf("feed = %+v ok=%t err=%v", feed, ok, err)
	}

	stateData := []byte{0xA4, 0x01, 4, 0, 0, 0, 0, 2, 0, 0, 0}
	binary.LittleEndian.PutUint32(stateData[3:7], 1234)
	state, ok, err := ParsePetStateChange(Packet{ID: PacketZCChangeStatePet, Data: stateData})
	if !ok || err != nil || state.Type != 4 || state.ID != 1234 || state.Data != 2 {
		t.Fatalf("state = %+v ok=%t err=%v", state, ok, err)
	}

	actionData := []byte{0xAA, 0x01, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(actionData[2:6], 1234)
	binary.LittleEndian.PutUint32(actionData[6:10], 112)
	action, ok, err := ParsePetAction(Packet{ID: PacketZCPetAct, Data: actionData})
	if !ok || err != nil || action.ID != 1234 || action.Data != 112 {
		t.Fatalf("action = %+v ok=%t err=%v", action, ok, err)
	}
}
