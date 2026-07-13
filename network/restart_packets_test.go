package network

import "testing"

func TestParseRestartAck(t *testing.T) {
	ack, ok, err := ParseRestartAck(Packet{ID: PacketZCRestartAck, Data: []byte{0xB3, 0x00, 0x01}})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected restart ack")
	}
	if !ack.Allowed {
		t.Fatal("expected allowed ack")
	}

	ack, ok, err = ParseRestartAck(Packet{ID: PacketZCRestartAck, Data: []byte{0xB3, 0x00, 0x00}})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || ack.Allowed {
		t.Fatalf("ack = %+v ok=%t, want denied", ack, ok)
	}
}

func TestParseRestartAckIgnoresOtherPackets(t *testing.T) {
	_, ok, err := ParseRestartAck(Packet{ID: 0x00B0, Data: []byte{0xB0, 0x00}})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("unexpected restart ack")
	}
}

func TestParseQuitGameAck(t *testing.T) {
	ack, ok, err := ParseQuitGameAck(Packet{ID: PacketZCQuitGameAck, Data: []byte{0x8B, 0x01, 0x00, 0x00}})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !ack.Allowed || ack.Result != 0 {
		t.Fatalf("accepted quit ack = %+v ok=%t", ack, ok)
	}

	ack, ok, err = ParseQuitGameAck(Packet{ID: PacketZCQuitGameAck, Data: []byte{0x8B, 0x01, 0x01, 0x00}})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || ack.Allowed || ack.Result != 1 {
		t.Fatalf("refused quit ack = %+v ok=%t", ack, ok)
	}
}
