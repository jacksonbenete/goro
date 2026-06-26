package network

import (
	"encoding/binary"
	"testing"
)

func TestParseNPCSayDialog(t *testing.T) {
	data := make([]byte, 8+len("hello")+1)
	binary.LittleEndian.PutUint16(data[0:2], 0x00B4)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint32(data[4:8], 1234)
	copy(data[8:], "hello")

	dialog, ok, err := ParseNPCDialog(Packet{ID: 0x00B4, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("dialog packet not recognized")
	}
	if dialog.Kind != NPCDialogSay || dialog.NPCID != 1234 || dialog.Message != "hello" {
		t.Fatalf("dialog = %+v", dialog)
	}
}

func TestParseNPCMenuDialog(t *testing.T) {
	data := make([]byte, 8+len("A:B:C")+1)
	binary.LittleEndian.PutUint16(data[0:2], 0x00B7)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	binary.LittleEndian.PutUint32(data[4:8], 77)
	copy(data[8:], "A:B:C")

	dialog, ok, err := ParseNPCDialog(Packet{ID: 0x00B7, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("menu packet not recognized")
	}
	if dialog.Kind != NPCDialogMenu || dialog.NPCID != 77 || len(dialog.Options) != 3 || dialog.Options[1] != "B" {
		t.Fatalf("dialog = %+v", dialog)
	}
}

func TestBuildNPCDialogPackets(t *testing.T) {
	if got := BuildNPCContactPacket(0x11223344, 0); ID(got) != 0x0090 || len(got) != 7 || binary.LittleEndian.Uint32(got[2:6]) != 0x11223344 || got[6] != 0 {
		t.Fatalf("contact packet = %x", got)
	}
	if got := BuildNPCNextPacket(0x11223344); ID(got) != 0x00B9 || len(got) != 6 || binary.LittleEndian.Uint32(got[2:6]) != 0x11223344 {
		t.Fatalf("next packet = %x", got)
	}
	if got := BuildNPCMenuChoicePacket(0x11223344, 2); ID(got) != 0x00B8 || len(got) != 7 || binary.LittleEndian.Uint32(got[2:6]) != 0x11223344 || got[6] != 2 {
		t.Fatalf("choice packet = %x", got)
	}
	if got := BuildNPCClosePacket(0x11223344); ID(got) != 0x0146 || len(got) != 6 || binary.LittleEndian.Uint32(got[2:6]) != 0x11223344 {
		t.Fatalf("close packet = %x", got)
	}
}
