package network

import "testing"

func TestParseNotifyBan(t *testing.T) {
	notify, ok, err := ParseNotifyBan(Packet{ID: PacketSCNotifyBan, Data: []byte{0x81, 0x00, 15}})
	if err != nil || !ok {
		t.Fatalf("notify ban ok=%t err=%v", ok, err)
	}
	if notify.ErrorCode != 15 {
		t.Fatalf("notify ban code = %d, want 15", notify.ErrorCode)
	}
}

func TestParseNotifyBanIgnoresOtherPackets(t *testing.T) {
	if _, ok, err := ParseNotifyBan(Packet{ID: 0x0069, Data: []byte{0x69, 0x00}}); ok || err != nil {
		t.Fatalf("notify ban parsed other packet ok=%t err=%v", ok, err)
	}
}

func TestParseNotifyBanRejectsShortPacket(t *testing.T) {
	if _, ok, err := ParseNotifyBan(Packet{ID: PacketSCNotifyBan, Data: []byte{0x81, 0x00}}); !ok || err == nil {
		t.Fatalf("short notify ban ok=%t err=%v, want ok true err", ok, err)
	}
}
