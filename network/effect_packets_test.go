package network

import (
	"encoding/binary"
	"testing"
)

func TestParseSpecialEffectNotify(t *testing.T) {
	data := make([]byte, 10)
	binary.LittleEndian.PutUint16(data[0:2], 0x019B)
	binary.LittleEndian.PutUint32(data[2:6], 0x11223344)
	binary.LittleEndian.PutUint32(data[6:10], SpecialEffectBaseLevelUp)

	notify, ok, err := ParseSpecialEffectNotify(Packet{ID: 0x019B, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("special effect notify not parsed")
	}
	if notify.AID != 0x11223344 || notify.EffectID != SpecialEffectBaseLevelUp {
		t.Fatalf("notify = %+v", notify)
	}
}

func TestParseSpecialEffectNotifyIgnoresOtherPackets(t *testing.T) {
	_, ok, err := ParseSpecialEffectNotify(Packet{ID: 0x019A, Data: make([]byte, 10)})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("unexpected special effect notify")
	}
}
