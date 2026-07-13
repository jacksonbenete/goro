package network

import (
	"encoding/binary"
	"testing"
)

func TestParseHotkeyList2008(t *testing.T) {
	data := make([]byte, 191)
	binary.LittleEndian.PutUint16(data[0:2], PacketZCShortcutKeyList)
	data[2] = HotkeyTypeSkill
	binary.LittleEndian.PutUint32(data[3:7], 28)
	binary.LittleEndian.PutUint16(data[7:9], 3)
	offset := 2 + 8*7
	data[offset] = HotkeyTypeItem
	binary.LittleEndian.PutUint32(data[offset+1:offset+5], 501)
	binary.LittleEndian.PutUint16(data[offset+5:offset+7], 1)

	list, ok, err := ParseHotkeyList(Packet{ID: PacketZCShortcutKeyList, Data: data})
	if err != nil || !ok {
		t.Fatalf("ParseHotkeyList ok=%t err=%v", ok, err)
	}
	if len(list.Slots) != HotkeyListSlots2008 {
		t.Fatalf("slots = %d, want %d", len(list.Slots), HotkeyListSlots2008)
	}
	if got := list.Slots[0]; got.Type != HotkeyTypeSkill || got.ID != 28 || got.Level != 3 {
		t.Fatalf("slot 0 = %+v", got)
	}
	if got := list.Slots[8]; got.Type != HotkeyTypeItem || got.ID != 501 || got.Level != 1 {
		t.Fatalf("slot 8 = %+v", got)
	}
}

func TestBuildHotkeyPacket(t *testing.T) {
	packet := BuildHotkeyPacket(3, HotkeySlot{Type: HotkeyTypeSkill, ID: 28, Level: 4})
	if len(packet) != 11 || ID(packet) != PacketCZShortcutKeyChange {
		t.Fatalf("packet = % X", packet)
	}
	if got := binary.LittleEndian.Uint16(packet[2:4]); got != 3 {
		t.Fatalf("index = %d, want 3", got)
	}
	if packet[4] != HotkeyTypeSkill {
		t.Fatalf("type = %d, want skill", packet[4])
	}
	if got := binary.LittleEndian.Uint32(packet[5:9]); got != 28 {
		t.Fatalf("id = %d, want 28", got)
	}
	if got := binary.LittleEndian.Uint16(packet[9:11]); got != 4 {
		t.Fatalf("level = %d, want 4", got)
	}
}
